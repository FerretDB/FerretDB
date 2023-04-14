// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

var (
	errProjectionEmpty = errors.New("projection is empty")
)

// validateProjection check projection document.
// Document fields could be either included or excluded but not both.
// Exception is for the _id field that could be included or excluded.
func validateProjection(projection *types.Document) (*types.Document, bool, error) {
	validated := types.MakeDocument(0)

	if projection == nil {
		return nil, false, errProjectionEmpty
	}

	var projectionVal *bool

	iter := projection.Iterator()
	defer iter.Close()

	for {
		key, value, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, false, lazyerrors.Error(err)
		}

		if key == "_id" { // _id is a special case and can be included or excluded
			continue
		}

		var result bool

		switch value := value.(type) {
		case *types.Document:
			return nil, false, commonerrors.NewCommandErrorMsg(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("projection expression %s is not supported", types.FormatAnyValue(value)),
			)

		case float64, int32, int64:
			comparison := types.Compare(value, int32(0))
			if comparison != types.Equal {
				result = true
			}
		case bool:
			if value {
				result = true
			}
		default:
			return nil, false, lazyerrors.Errorf("unsupported operation %s %value (%T)", key, value, value)
		}

		// set the value with boolean result to omit type assertion when we will apply projection
		validated.Set(key, result)

		// if projectionVal is nil we are processing the first field
		if projectionVal == nil {
			projectionVal = &result
			continue
		}

		if *projectionVal != result {
			if *projectionVal {
				return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrProjectionExIn,
					fmt.Sprintf("Cannot do exclusion on field %s in inclusion projection", key),
					"projection",
				)
			} else {
				return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrProjectionInEx,
					fmt.Sprintf("Cannot do inclusion on field %s in exclusion projection", key),
					"projection",
				)
			}
		}
	}

	return validated, *projectionVal, nil
}

// projectDocument applies projection to the copy of the document.
func projectDocument(doc, projection *types.Document, inclusion bool) (*types.Document, error) {
	var projected *types.Document

	if inclusion {
		projected = types.MakeDocument(1)

		projected.Set("_id", must.NotFail(doc.Get("_id")))
	} else {
		projected = doc.DeepCopy()
	}

	iter := projection.Iterator()
	defer iter.Close()

	for {
		key, value, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		path, err := types.NewPathFromString(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		switch value := value.(type) { // found in the projection
		case *types.Document: // field: { $elemMatch: { field2: value }}
			return nil, commonerrors.NewCommandErrorMsg(
				commonerrors.ErrCommandNotFound,
				fmt.Sprintf("projection %s is not supported",
					types.FormatAnyValue(value),
				),
			)

		case bool: // field: bool
			// process top level fields
			if path.Len() == 1 {
				if inclusion {
					if doc.Has(key) {
						projected.Set(key, must.NotFail(doc.Get(key)))
					}
				} else {
					projected.Remove(key)
				}

				continue
			}

			// TODO: process dot notation here https://github.com/FerretDB/FerretDB/issues/2430
		default:
			return nil, lazyerrors.Errorf("unsupported operation %s %v (%T)", key, value, value)
		}
	}

	return projected, nil
}
