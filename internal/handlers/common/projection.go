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
)

// isProjectionInclusion: projection can be only inclusion or exclusion. Validate and return true if inclusion.
// Exception for the _id field.
func isProjectionInclusion(projection *types.Document) (bool, error) {
	var exclusion bool
	var inclusion bool

	iter := projection.Iterator()
	for {
		key, value, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return false, lazyerrors.Error(err)
		}

		if key == "_id" { // _id is a special case and can be both
			continue
		}

		switch value := value.(type) {
		case *types.Document:
			return false, commonerrors.NewCommandErrorMsg(
				commonerrors.ErrNotImplemented,
				"projection expressions is not supported",
			)

		case float64, int32, int64:
			result := types.Compare(value, int32(0))
			if result == types.Equal {
				if inclusion {
					return false, commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrProjectionExIn,
						fmt.Sprintf("Cannot do exclusion on field %s in inclusion projection", key),
						"projection",
					)
				}
				exclusion = true
			} else {
				if exclusion {
					return false, commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrProjectionInEx,
						fmt.Sprintf("Cannot do inclusion on field %s in exclusion projection", key),
						"projection",
					)
				}
				inclusion = true
			}

		case bool:
			if value {
				if exclusion {
					return false, commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrProjectionInEx,
						fmt.Sprintf("Cannot do inclusion on field %s in exclusion projection", key),
						"projection",
					)
				}
				inclusion = true
			} else {
				if inclusion {
					return false, commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrProjectionExIn,
						fmt.Sprintf("Cannot do exclusion on field %s in inclusion projection", key),
						"projection",
					)
				}
				exclusion = true
			}

		default:
			return false, lazyerrors.Errorf("unsupported operation %s %value (%T)", key, value, value)
		}
	}

	return inclusion, nil
}

func projectDocument(inclusion bool, doc *types.Document, projection *types.Document) error {
	iter := doc.Iterator()
	defer iter.Close()

	for {
		key, _, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		projectionVal, err := projection.Get(key)
		if err != nil {
			if key == "_id" { // if _id is not in projection map, do not do anything with it
				continue
			}
			if inclusion { // k1 from doc is absent in projection, remove from doc only if projection type inclusion
				doc.Remove(key)
			}
			continue
		}

		switch projectionVal := projectionVal.(type) { // found in the projection
		case *types.Document: // field: { $elemMatch: { field2: value }}
			if err := applyComplexProjection(projectionVal); err != nil {
				return err
			}

		case float64, int32, int64: // field: number
			result := types.Compare(projectionVal, int32(0))
			if result == types.Equal {
				doc.Remove(key)
			}

		case bool: // field: bool
			if !projectionVal {
				doc.Remove(key)
			}

		default:
			return lazyerrors.Errorf("unsupported operation %s %v (%T)", key, projectionVal, projectionVal)
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

	return nil
}
