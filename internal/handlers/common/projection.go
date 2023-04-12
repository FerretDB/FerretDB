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

var errProjectionEmpty = errors.New("projection is empty")

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

	return nil
}

// isProjectionInclusion: projection can be only inclusion or exclusion. Validate and return true if inclusion.
// Exception for the _id field.
func isProjectionInclusion(projection *types.Document) (inclusion bool, err error) {
	var exclusion bool
	for _, k := range projection.Keys() {
		if k == "_id" { // _id is a special case and can be both
			continue
		}
		v := must.NotFail(projection.Get(k))
		switch v := v.(type) {
		case *types.Document:
			for _, projectionType := range v.Keys() {
				err = commonerrors.NewCommandError(
					commonerrors.ErrNotImplemented,
					fmt.Errorf("projection of %s is not supported", projectionType),
				)

				return
			}

		switch value := value.(type) {
		case *types.Document, *types.Array, string:
			return nil, false, commonerrors.NewCommandErrorMsg(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("projection expression %s is not supported", types.FormatAnyValue(value)),
			)
		case float64, int32, int64:
			result := types.Compare(v, int32(0))
			if result == types.Equal {
				if inclusion {
					err = commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrProjectionExIn,
						fmt.Sprintf("Cannot do exclusion on field %s in inclusion projection", k),
						"projection",
					)
					return
				}
				exclusion = true
			} else {
				if exclusion {
					err = commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrProjectionInEx,
						fmt.Sprintf("Cannot do inclusion on field %s in exclusion projection", k),
						"projection",
					)
					return
				}
				inclusion = true
			}

			if comparison != types.Equal {
				result = true
			}
		case bool:
			if v {
				if exclusion {
					err = commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrProjectionInEx,
						fmt.Sprintf("Cannot do inclusion on field %s in exclusion projection", k),
						"projection",
					)
					return
				}
				inclusion = true
			} else {
				if inclusion {
					err = commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrProjectionExIn,
						fmt.Sprintf("Cannot do exclusion on field %s in inclusion projection", k),
						"projection",
					)
					return
				}
				exclusion = true
			}

		default:
			return nil, false, lazyerrors.Errorf("unsupported operation %s %value (%T)", key, value, value)
		}

		// set the value with boolean result to omit type assertion when we will apply projection
		validated.Set(key, result)

		if projection.Len() == 1 && key == "_id" {
			return validated, result, nil
		}

		// if projectionVal is nil we are processing the first field
		if projectionVal == nil {
			if key == "_id" {
				continue
			}

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
			}

			return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrProjectionInEx,
				fmt.Sprintf("Cannot do inclusion on field %s in exclusion projection", key),
				"projection",
			)
		}
	}

	return validated, *projectionVal, nil
}

// projectDocument applies projection to the copy of the document.
func projectDocument(doc, projection *types.Document, inclusion bool) (*types.Document, error) {
	projected, err := types.NewDocument("_id", must.NotFail(doc.Get("_id")))
	if err != nil {
		return nil, err
	}

	if projection.Has("_id") {
		idValue := must.NotFail(projection.Get("_id"))

		var set bool

		switch idValue := idValue.(type) {
		case *types.Document: // field: { $elemMatch: { field2: value }}
			return nil, commonerrors.NewCommandErrorMsg(
				commonerrors.ErrCommandNotFound,
				fmt.Sprintf("projection %s is not supported",
					types.FormatAnyValue(idValue),
				),
			)
		case bool:
			set = idValue
		default:
			return nil, lazyerrors.Errorf("unsupported operation %s %v (%T)", "_id", idValue, idValue)
		}

		if !set {
			projected.Remove("_id")
		}
	}

	projectedWithoutID, err := projectDocumentWithoutID(doc, projection, inclusion)
	if err != nil {
		return nil, err
	}

	for _, key := range projectedWithoutID.Keys() {
		projected.Set(key, must.NotFail(projectedWithoutID.Get(key)))
	}

	return projected, nil
}

// projectDocumentWithoutID applies projection to the copy of the document and returns projected document.
// It ignores _id field in the projection.
func projectDocumentWithoutID(doc *types.Document, projection *types.Document, inclusion bool) (*types.Document, error) {
	projectionWithoutID := projection.DeepCopy()
	projectionWithoutID.Remove("_id")

	docWithoutID := doc.DeepCopy()
	docWithoutID.Remove("_id")

	projected := types.MakeDocument(0)

	if !inclusion {
		projected = docWithoutID.DeepCopy()
	}

	iter := projectionWithoutID.Iterator()
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
					if docWithoutID.Has(key) {
						projected.Set(key, must.NotFail(docWithoutID.Get(key)))
					}

					continue
				}

				projected.Remove(key)
			}

			// TODO: process dot notation here https://github.com/FerretDB/FerretDB/issues/2430
		default:
			return lazyerrors.Errorf("unsupported operation %s %v (%T)", k1, projectionVal, projectionVal)
		}
	}
	return nil
}

func applyComplexProjection(projectionVal *types.Document) error {
	for _, projectionType := range projectionVal.Keys() {
		switch projectionType {
		case "$elemMatch", "$slice":
			return commonerrors.NewCommandError(
				commonerrors.ErrNotImplemented,
				fmt.Errorf("projection of %s is not supported", projectionType),
			)
		default:
			return commonerrors.NewCommandError(
				commonerrors.ErrCommandNotFound,
				fmt.Errorf("projection of %s is not supported", projectionType),
			)
		}
	}

	return projected, nil
}
