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
	"math"
	"strconv"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

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
				supportedProjectionTypes := []string{"$elemMatch", "$slice"}
				if !slices.Contains(supportedProjectionTypes, projectionType) {
					err = lazyerrors.Errorf("projection of %s is not supported", projectionType)
					return
				}

				switch projectionType {
				case "$elemMatch":
					inclusion = true
				case "$slice":
					inclusion = false
				default:
					panic(projectionType + " not supported")
				}
			}

		case float64, int32, int64:
			result := types.Compare(v, int32(0))
			if types.ContainsCompareResult(result, types.Equal) {
				if inclusion {
					err = NewCommandErrorMsgWithArgument(ErrProjectionExIn,
						fmt.Sprintf("Cannot do exclusion on field %s in inclusion projection", k),
						"projection",
					)
					return
				}
				exclusion = true
			} else {
				if exclusion {
					err = NewCommandErrorMsgWithArgument(ErrProjectionInEx,
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
					err = NewCommandErrorMsgWithArgument(ErrProjectionInEx,
						fmt.Sprintf("Cannot do inclusion on field %s in exclusion projection", k),
						"projection",
					)
					return
				}
				inclusion = true
			} else {
				if inclusion {
					err = NewCommandErrorMsgWithArgument(ErrProjectionExIn,
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
			result := types.Compare(projectionVal, int32(0))
			if types.ContainsCompareResult(result, types.Equal) {
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
		switch projectionType {
		case "$elemMatch":
			var docValueA any
			docValueA, err = doc.GetByPath(types.NewPath([]string{k1}))
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
			}
		case "$slice":
			var docValue any
			docValue, err = doc.Get(k1)
			if err != nil { // the field can't be obtained, so there is nothing to do
				return
			}
			// $slice works only for arrays, so docValue must be an array
			arr, ok := docValue.(*types.Array)
			if !ok {
				return
			}
			projectionVal := must.NotFail(projectionVal.Get(projectionType))
			res, err := filterFieldArraySlice(arr, projectionVal)
			if err != nil {
				return err
			}

			if res == nil {
				doc.Set(k1, types.Null)
				return nil
			}

			doc.Set(k1, res)
		default:
			return NewCommandError(ErrCommandNotFound,
				lazyerrors.Errorf("applyComplexProjection: unknown projection operator: %q", projectionType),
			)
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
						doc.RemoveByPath(types.NewPath([]string{k1, strconv.Itoa(j)}))
						continue
					}
					result := types.Compare(docVal, elemMatchFieldCondition)
					if types.ContainsCompareResult(result, types.Equal) {
						// elemMatch to return first matching, all others are to be removed
						found = j
						break
					}
					doc.RemoveByPath(types.NewPath([]string{k1, strconv.Itoa(j)}))
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

// filterFieldArraySlice implements $slice projection query.
func filterFieldArraySlice(docValue *types.Array, projectionValue any) (*types.Array, error) {
	switch projectionValue := projectionValue.(type) {
	case *types.Array:
		if projectionValue.Len() < 2 || projectionValue.Len() > 3 {
			return nil, NewCommandErrorMsgWithArgument(
				ErrInvalidArg,
				fmt.Sprintf(
					"Invalid $slice syntax. The given syntax "+
						"did not match the find() syntax because :: Location31272: "+
						"$slice array argument should be of form [skip, limit] :: "+
						"The given syntax did not match the expression "+
						"$slice syntax. :: caused by :: "+
						"Expression $slice takes at least 2 arguments, and at most 3, but %d were passed in.",
					projectionValue.Len(),
				),
				"$slice",
			)
		}

		if projectionValue.Len() == 3 {
			// this is the error MongoDB 5.0 is returning in this case
			return nil, NewCommandErrorMsgWithArgument(
				ErrSliceFirstArg,
				fmt.Sprintf(
					"First argument to $slice must be an array, but is of type: %s",
					AliasFromType(must.NotFail(projectionValue.Get(0))),
				),
				"$slice",
			)
		}

		return projectionSliceMultiArgs(docValue, projectionValue)

	case float64, int32, int64:
		return projectionSliceSingleArg(docValue, projectionValue), nil

	default:
		return nil, NewCommandErrorMsgWithArgument(
			ErrInvalidArg,
			"Invalid $slice syntax. The given syntax "+
				"did not match the find() syntax because :: Location31273: "+
				"$slice only supports numbers and [skip, limit] arrays :: "+
				"The given syntax did not match the expression $slice syntax. :: caused by :: "+
				"Expression $slice takes at least 2 arguments, and at most 3, but 1 were passed in.",
			"$slice",
		)
	}
}

func projectionSliceSingleArg(arr *types.Array, arg any) *types.Array {
	var n int
	switch v := arg.(type) {
	case float64:
		if math.IsNaN(v) {
			break // because n == 0 already
		}
		if math.IsInf(v, -1) || v < math.MinInt {
			n = math.MinInt
			break
		}
		if math.IsInf(v, +1) || v > math.MaxInt {
			n = math.MaxInt
			break
		}
		n = int(v)

	case int32:
		n = int(v)

	case int64:
		if v > math.MaxInt {
			n = math.MaxInt
			break
		}
		if v < math.MinInt {
			n = math.MinInt
			break
		}
		n = int(v)
	}

	// negative n is OK in case of a single argument
	var skip, limit int
	if n < 0 {
		skip, limit = arr.Len()+n, arr.Len()
		n = -n
	} else {
		skip, limit = 0, n
	}
	if n < arr.Len() {
		res := types.MakeArray(limit)
		for i := skip; i < limit; i++ {
			must.NoError(res.Append(must.NotFail(arr.Get(i))))
		}
		return res
	}
	// otherwise return arr as is
	return arr
}

func projectionSliceMultiArgs(arr, args *types.Array) (*types.Array, error) {
	var skip, limit int
	pair := [2]int{}
	for i := range pair {
		switch v := must.NotFail(args.Get(i)).(type) {
		case float64:
			if math.IsNaN(v) {
				break // because pair[i] == 0 already
			}
			if math.IsInf(v, -1) || v < math.MinInt {
				pair[i] = math.MinInt
				break
			}
			if math.IsInf(v, +1) || v > math.MaxInt {
				pair[i] = math.MaxInt
				break
			}
			pair[i] = int(v)

		case types.NullType:
			return nil, nil //nolint:nilnil // nil is a valid value

		case int32:
			pair[i] = int(v)

		case int64:
			if v > math.MaxInt {
				pair[i] = math.MaxInt
				break
			}
			if v < math.MinInt {
				pair[i] = math.MinInt
				break
			}
			pair[i] = int(v)

		default:
			return nil, NewCommandErrorMsgWithArgument(
				ErrSliceFirstArg,
				fmt.Sprintf(
					"First argument to $slice must be an array, but is of type: %s",
					AliasFromType(must.NotFail(args.Get(0))),
				),
				"$slice",
			)
		}

		if i == 1 && pair[i] < 0 { // limit can't be negative in case of 2 arguments
			return nil, NewCommandErrorMsgWithArgument(
				ErrSliceFirstArg,
				fmt.Sprintf(
					"First argument to $slice must be an array, but is of type: %s",
					AliasFromType(must.NotFail(args.Get(0))),
				),
				"$slice",
			)
		}
	}

	skip, limit = pair[0], pair[1]

	if skip < 0 {
		if -skip >= arr.Len() {
			skip = 0
		} else {
			skip = arr.Len() + skip
		}
	} else {
		if skip > arr.Len() {
			return types.MakeArray(0), nil
		}
	}
	limit += skip
	if limit >= arr.Len() {
		limit = arr.Len()
	}
	res := types.MakeArray(limit)
	for i := skip; i < limit; i++ {
		must.NoError(res.Append(must.NotFail(arr.Get(i))))
	}
	return res, nil
}
