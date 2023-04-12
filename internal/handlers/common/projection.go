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

// validateProjection check projection document.
// Document fields could be either included or excluded but not both.
// Exception is for the _id field that could be included or excluded.
func validateProjection(projection *types.Document) error {
	var projectionVal *bool

	iter := projection.Iterator()
	defer iter.Close()

	for {
		key, value, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		if key == "_id" { // _id is a special case and can be included or excluded
			continue
		}

		var result bool

		switch value := value.(type) {
		case *types.Document:
			return commonerrors.NewCommandErrorMsg(
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
			return lazyerrors.Errorf("unsupported operation %s %value (%T)", key, value, value)
		}

		// set the value with boolean result to omit type assertion in the next iteration
		projection.Set(key, result)

		// if projectionVal is nil, it means that we are processing the first field
		if projectionVal == nil {
			projectionVal = &result
			continue
		}

		if *projectionVal != result {
			if *projectionVal {
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrProjectionExIn,
					fmt.Sprintf("Cannot do exclusion on field %s in inclusion projection", key),
					"projection",
				)
			} else {
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrProjectionInEx,
					fmt.Sprintf("Cannot do inclusion on field %s in exclusion projection", key),
					"projection",
				)
			}
		}
	}

	return nil
}

func projectDocument(doc *types.Document, projection *types.Document) (*types.Document, error) {
	if projection == nil {
		return doc, nil
	}

	projected := types.MakeDocument(1)

	projected.Set("_id", must.NotFail(doc.Get("_id")))

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
			// if projection value is false, we should skip the field
			if !value {
				projected.RemoveByPath(path)
				continue
			}

		default:
			return nil, lazyerrors.Errorf("unsupported operation %s %v (%T)", key, value, value)
		}

		// if doc has field set it to the projected document
		if doc.HasByPath(path) {
			projected.SetByPath(path, must.NotFail(doc.GetByPath(path)))
		}
	}

	return projected, nil
}
