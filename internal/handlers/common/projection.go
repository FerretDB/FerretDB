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
	"strconv"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// isProjectionInclusion: projection can be only inclusion or exlusion. Validate and return true if inclusion.
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
				supportedProjectionTypes := []string{"$elemMatch"}
				if !slices.Contains(supportedProjectionTypes, projectionType) {
					err = lazyerrors.Errorf("projecion of %s is not supported", projectionType)
					return
				}

				switch projectionType {
				case "$elemMatch":
					inclusion = true
				default:
					panic(projectionType + " not supported")
				}
			}

		case float64, int32, int64:
			if types.Compare(v, int32(0)) == types.Equal {
				if inclusion {
					err = NewError(ErrProjectionExIn,
						fmt.Errorf("Cannot do exclusion on field %s in inclusion projection", k),
					)
					return
				}
				exclusion = true
			} else {
				if exclusion {
					err = NewError(ErrProjectionInEx,
						fmt.Errorf("Cannot do inclusion on field %s in exclusion projection", k),
					)
					return
				}
				inclusion = true
			}

		case bool:
			if v {
				if exclusion {
					err = NewError(ErrProjectionInEx,
						fmt.Errorf("Cannot do inclusion on field %s in exclusion projection", k),
					)
					return
				}
				inclusion = true
			} else {
				if inclusion {
					err = NewError(ErrProjectionExIn,
						fmt.Errorf("Cannot do exclusion on field %s in inclusion projection", k),
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
			if err := applyComplexProjection(k1, doc, projectionVal); err != nil {
				return err
			}

		case float64, int32, int64: // field: number
			if types.Compare(projectionVal, int32(0)) == types.Equal {
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

func applyComplexProjection(k1 string, doc, projectionVal *types.Document) (err error) {
	for _, projectionType := range projectionVal.Keys() {
		supportedProjections := []string{"$elemMatch"}
		if !slices.Contains(supportedProjections, projectionType) {
			return fmt.Errorf("projecion %s is not supported", projectionType)
		}

		// for now it's only $elemMatch further
		// if the corresponding value is not an array, skip

		var docValueA any
		docValueA, err = doc.GetByPath(k1)
		if err != nil {
			continue
		}

		// $elemMatch works only for arrays, it must be an array
		docValueArray, ok := docValueA.(*types.Array)
		if !ok {
			doc.Remove(k1)
			return
		}

		// get the elemMatch conditions
		conditions := must.NotFail(projectionVal.Get(projectionType)).(*types.Document)

		var found int
		found, err = filterFieldArrayElemMatch(k1, doc, conditions, docValueArray)
		if found < 0 {
			doc.Remove(k1)
			return
		}
	}
	return
}

// filterFieldArrayElemMatch is for elemMatch conditions.
func filterFieldArrayElemMatch(k1 string, doc, conditions *types.Document, docValueArray *types.Array) (found int, err error) {
	for k2ConditionField, conditionValue := range conditions.Map() {
		switch elemMatchFieldCondition := conditionValue.(type) {
		case *types.Document: // TODO field2: { $gte: 10 }

		case *types.Array:
			panic("unexpected code")

		default: // field2: value
			found = -1 // >= 0 means found

			for j := 0; j < docValueArray.Len(); j++ {
				var cmpVal any
				cmpVal, err = docValueArray.Get(j)
				if err != nil {
					continue
				}
				switch cmpVal := cmpVal.(type) {
				case *types.Document:
					docVal, err := cmpVal.Get(k2ConditionField)
					if err != nil {
						doc.RemoveByPath(k1, strconv.Itoa(j))
						continue
					}
					if types.Compare(docVal, elemMatchFieldCondition) == types.Equal {
						// elemMatch to return first matching, all others are to be removed
						found = j
						break
					}
					doc.RemoveByPath(k1, strconv.Itoa(j))
					j = j - 1
				}
			}

			if found < 0 {
				return
			}
		}
	}
	return
}
