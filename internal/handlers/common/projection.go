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

// isProjectionInclusion: projection can be only inclusion or exclusion. Validate and return true if inclusion.
// Exception for the _id field.
func isProjectionInclusion(projection *types.Document) (bool, error) {
	inclusion, _, err := validateExpression(projection, 0, false, false)
	return inclusion, err
}

func validateExpression(projection *types.Document, depth int, inclusion, exclusion bool) (bool, bool, error) {
	var err error
	for _, k := range projection.Keys() {
		if k == "_id" { // _id is a special case and can be both
			continue
		}

		v := must.NotFail(projection.Get(k))
		switch v := v.(type) {
		case *types.Document:

			for _, key := range v.Keys() {
				val := must.NotFail(v.Get(key))
				switch val := val.(type) {
				case *types.Document:
					if key == "$elemMatch" && depth >= 1 {
						err = NewErrorMsg(ErrElemMatchNestedField,
							"Cannot use $elemMatch projection on a nested field.",
						)
						return false, false, err
					}
					inclusion, exclusion, err = validateExpression(val, depth+1, inclusion, exclusion)
					return inclusion, exclusion, err

				default:
					switch key {
					case "$eq",
						"$ne",
						"$gt", "$gte",
						"$lt", "$lte":
						inclusion = true

					case "$in":
						switch must.NotFail(v.Get(key)).(type) {
						case *types.Array:
							// ok
						default:
							err = NewErrorMsg(ErrBadValue, "$in needs an array")
							return false, false, err
						}
					case "$nin", "$not":
						exclusion = true

					default: // $mod, etc
						err = NewErrorMsg(ErrNotImplemented, key+" is not supported")
						return inclusion, exclusion, err
					}
				}
			}
		default: // scalars and arrays

			if k == "$elemMatch" {
				err = NewError(ErrElemMatchObjectRequired,
					fmt.Errorf("elemMatch: Invalid argument, object required, but got %T", v),
				)
				return false, false, err
			}

			switch v := v.(type) {
			case float64, int32, int64:
				if types.Compare(v, int32(0)) == types.Equal {
					if inclusion {
						err = NewError(ErrElemMatchExclusionInInclusion,
							fmt.Errorf("Cannot do exclusion on field %s in inclusion projection", k),
						)
						return false, false, err
					}
					exclusion = true
				} else {
					if exclusion {
						err = NewError(ErrElemMatchInclusionInExclusion,
							fmt.Errorf("Cannot do inclusion on field %s in exclusion projection", k),
						)
						return false, false, err
					}
					inclusion = true
				}

			case bool:
				if v {
					if exclusion {
						err = NewError(ErrElemMatchInclusionInExclusion,
							fmt.Errorf("Cannot do inclusion on field %s in exclusion projection", k),
						)
						return false, false, err
					}
					inclusion = true
				} else {
					if inclusion {
						err = NewError(ErrElemMatchExclusionInInclusion,
							fmt.Errorf("Cannot do exclusion on field %s in inclusion projection", k),
						)
						return false, false, err
					}
					exclusion = true
				}
			}
			return inclusion, exclusion, err
		}
	}
	return inclusion, exclusion, err
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

	for k1, k1Val := range doc.Map() {
		k1Projection, ok := projectionMap[k1]
		if !ok {
			if k1 == "_id" { // if _id is not in projection map, do not do anything with it
				continue
			}
			if inclusion { // k1 from doc is absent in projection, remove from doc only if projection type inclusion
				doc.Remove(k1)
			}
			continue
		}

		switch k1Projection := k1Projection.(type) { // found in the projection
		case *types.Document: // in projection doc: k1: { k2: value }}, k1Projection == { k2: value }}
			if err := applyDocProjection(k1, doc, k1Projection); err != nil {
				return err
			}

		case *types.Array: // in projection doc: { k1: [value1, value2... ], k1Projection = [ value1, value2.. ]
			return NewErrorMsg(ErrNotImplemented, k1+" not supported")

		case float64, // in projection doc: { k1: k1Projection } where k1Projection is a number
			int32,
			int64:
			if types.Compare(k1Projection, int32(0)) == types.Equal {
				doc.Remove(k1)
			}

		case bool: // in projection doc: { k1: k1Projection }
			if !k1Projection {
				doc.Remove(k1)
			}

		default:
			return lazyerrors.Errorf("unsupported operation %s %v (%T)", k1, k1Val, k1Val)
		}
	}
	return nil
}

func applyDocProjection(k1 string, doc *types.Document, k1Projection *types.Document) error {
	var err error
	for _, projectionName := range k1Projection.Keys() {
		if projectionName != "$elemMatch" {
			panic(projectionName + " not supported!") // checks must be done in projection check func above
		}
		conditions := must.NotFail(k1Projection.Get(projectionName)).(*types.Document)
		var found bool
		found, err = findDocElemMatch(k1, doc, conditions)
		if err != nil {
			return err
		}
		if !found {
			doc.Remove(k1)
			return nil
		}
	}
	return err
}

func findInArray(k1, k2 string, value any, doc *types.Document, compareRes []types.CompareResult) bool {
	docValueArray := must.NotFail(doc.GetByPath(k1)).(*types.Array)

	found := -1
	for j := 0; j < docValueArray.Len(); j++ {
		e, err := docValueArray.Get(j)
		if err != nil {
			continue
		}

		if found >= 0 {
			doc.RemoveByPath(k1, strconv.Itoa(j))
			j -= 1
			continue
		}
		switch e := e.(type) {
		case *types.Document:
			var d any
			d, err = e.Get(k2)
			if err != nil {
				doc.RemoveByPath(k1, strconv.Itoa(j))
				j -= 1
				continue
			}
			switch value := value.(type) {
			case *types.Document, *types.Array: // TODO
			default:
				cmp := types.Compare(d, value)
				if slices.Contains(compareRes, cmp) {
					found = j
					continue
				}
			}

			doc.RemoveByPath(k1, strconv.Itoa(j))
			j -= 1

		default:
			doc.RemoveByPath(k1, strconv.Itoa(j))
			j -= 1
			continue
		}
	}

	if found < 0 {
		doc.RemoveByPath(k1)
	}
	return found >= 0
}

// findDocElemMatch is for elemMatch conditions.
func findDocElemMatch(k1 string, doc, conditions *types.Document) (bool, error) {
	found := false

	// for sure it's here - see code above
	docValueA := must.NotFail(doc.GetByPath(k1))

	// $elemMatch works only for arrays, it must be an array
	docValueArray, ok := docValueA.(*types.Array)
	if !ok {
		doc.Remove(k1)
		return found, nil
	}

	for k2, condition := range conditions.Map() {
		switch condition := condition.(type) {
		// in condition: { $eq: 42 }
		case *types.Document:
			for operand, value := range condition.Map() {
				var err error
				switch operand {
				case "$eq":
					found = findInArray(k1, k2, value, doc, []types.CompareResult{types.Equal})

				case "$ne":
					found = findInArray(k1, k2, value, doc, []types.CompareResult{types.Less, types.Greater})

				case "$gt":
					found = findInArray(k1, k2, value, doc, []types.CompareResult{types.Greater})

				case "$gte":
					found = findInArray(k1, k2, value, doc, []types.CompareResult{types.Greater, types.Equal})

				case "$lt":
					found = findInArray(k1, k2, value, doc, []types.CompareResult{types.Less})

				case "$lte":
					found = findInArray(k1, k2, value, doc, []types.CompareResult{types.Less, types.Equal})

				case "$nin":
					switch inValue := value.(type) {
					case *types.Array:
						for i := 0; i < inValue.Len(); i++ {
							x := must.NotFail(inValue.Get(i))
							found = findInArray(k1, k2, x, doc,
								[]types.CompareResult{types.Less, types.Greater, types.NotEqual},
							)
							if found {
								return found, err
							}
						}
					default:
						err = NewErrorMsg(ErrBadValue, "$nin needs an array")
						return found, err
					}
					if !found {
						return found, err
					}

				case "$in":
					switch inValue := value.(type) {
					case *types.Array:
						for i := 0; i < inValue.Len(); i++ {
							x := must.NotFail(inValue.Get(i))
							found = findInArray(k1, k2, x, doc, []types.CompareResult{types.Equal})
							if found {
								return found, err
							}
						}
					default:
						err = NewErrorMsg(ErrBadValue, "array values supported for $in only")
						return found, err
					}
					if !found {
						return found, err
					}

					// operand is not an operand possible: <scalar value> OR field: {nested projection}
				default:

					for j := 0; j < docValueArray.Len(); j++ {
						e := must.NotFail(docValueArray.Get(j))

						switch e := e.(type) {
						case *types.Document:
							docVal, err := e.Get(k2)
							if err != nil {
								doc.RemoveByPath(k1, strconv.Itoa(j))
								continue
							}
							if types.Compare(docVal, value) == types.Equal {
								found = true
								break
							}
						default: // field2: value
							if types.Compare(e, value) == types.Equal {
								found = true
								break
							}
						}
						doc.RemoveByPath(k1, strconv.Itoa(j))
						j = j - 1
					}
					err = NewErrorMsg(ErrBadValue, k2+" not supported")
					return found, err
				}
				return found, err
			}
		}
	}
	return found, nil
}
