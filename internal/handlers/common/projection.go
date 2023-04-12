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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ProjectDocuments modifies given documents in places according to the given projection.
func ProjectDocuments(docs []*types.Document, projection *types.Document) error {
	if projection.Len() == 0 {
		return nil
	}

	inclusion, err := isProjectionInclusion(projection)
	if err != nil {
		return err
	}

	for i := 0; i < len(docs); i++ {
		err = projectDocument(inclusion, docs[i], projection)
		if err != nil {
			return err
		}
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
			err = lazyerrors.Errorf("unsupported operation %s %v (%T)", k, v, v)
			return
		}
	}
	return
}

func projectDocument(inclusion bool, doc *types.Document, projection *types.Document) error {
	projectionMap := projection.Map()

	for k1 := range doc.Map() {
		projectionVal, ok := projectionMap[k1]
		if !ok {
			if k1 == "_id" { // if _id is not in projection map, do not do anything with it
				continue
			}
			if inclusion { // k1 from doc is absent in projection, remove from doc only if projection type inclusion
				doc.Remove(k1)
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
				doc.Remove(k1)
			}

		case bool: // field: bool
			if !projectionVal {
				doc.Remove(k1)
			}

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

	return nil
}
