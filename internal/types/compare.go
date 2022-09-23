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

package types

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"time"

	"golang.org/x/exp/constraints"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../bin/stringer -linecomment -type CompareResult

// CompareResult represents the result of a comparison.
type CompareResult int8

// Values match results of comparison functions such as bytes.Compare.
// They do not match MongoDB SortType values where 1 means ascending order and -1 means descending.
const (
	Equal        CompareResult = 0   // ==
	Less         CompareResult = -1  // <
	Greater      CompareResult = 1   // >
	Incomparable CompareResult = 127 // ≹
)

// Compare compares any BSON values in the same way as MongoDB does it for filtering.
//
// It converts types as needed; that may result in different types being equal.
// For that reason, it typically should not be used in tests.
// It returns a slice of result because a filter might match the document as
// Less, Greater and Equal at the same time (mostly for composite data type, e.g. embedded Array).
//
// Compare and contrast with test helpers in testutil package.
func Compare(docValue, filterValue any) []CompareResult {
	if docValue == nil {
		panic("compare: docValue is nil")
	}
	if filterValue == nil {
		panic("compare: filterValue is nil")
	}

	switch docValue := docValue.(type) {
	case *Document:
		// TODO: implement document comparing https://github.com/FerretDB/FerretDB/issues/457
		return []CompareResult{Incomparable}

	case *Array:
		if filterArr, ok := filterValue.(*Array); ok {
			return compareArrays(filterArr, docValue)
		}

		for i := 0; i < docValue.Len(); i++ {
			docValue := must.NotFail(docValue.Get(i))
			switch docValue.(type) {
			case *Document, *Array:
				continue
			}

			if res := compareScalars(docValue, filterValue); res != Incomparable {
				return []CompareResult{res}
			}
		}
		return []CompareResult{Incomparable}

	default:
		return []CompareResult{compareScalars(docValue, filterValue)}
	}
}

// compareScalars compares BSON scalar values.
func compareScalars(v1, v2 any) CompareResult {
	if !isScalar(v1) || !isScalar(v2) {
		return Incomparable
	}

	switch v1 := v1.(type) {
	case float64:
		switch v2 := v2.(type) {
		case float64:
			if math.IsNaN(v1) && math.IsNaN(v2) {
				return Equal
			}
			return compareOrdered(v1, v2)
		case int32:
			return compareNumbers(v1, int64(v2))
		case int64:
			return compareNumbers(v1, v2)
		default:
			return Incomparable
		}

	case string:
		v2, ok := v2.(string)
		if ok {
			return compareOrdered(v1, v2)
		}
		return Incomparable

	case Binary:
		v2, ok := v2.(Binary)
		if !ok {
			return Incomparable
		}
		v1l, v2l := len(v1.B), len(v2.B)
		if v1l != v2l {
			return compareOrdered(v1l, v2l)
		}
		if v1.Subtype != v2.Subtype {
			return compareOrdered(v1.Subtype, v2.Subtype)
		}
		return CompareResult(bytes.Compare(v1.B, v2.B))

	case ObjectID:
		v2, ok := v2.(ObjectID)
		if !ok {
			return Incomparable
		}
		return CompareResult(bytes.Compare(v1[:], v2[:]))

	case bool:
		v2, ok := v2.(bool)
		if !ok {
			return Incomparable
		}
		if v1 == v2 {
			return Equal
		}
		if v2 {
			return Less
		}
		return Greater

	case time.Time:
		v2, ok := v2.(time.Time)
		if !ok {
			return Incomparable
		}
		return compareOrdered(v1.UnixMilli(), v2.UnixMilli())

	case NullType:
		_, ok := v2.(NullType)
		if ok {
			return Equal
		}
		return Incomparable

	case Regex:
		v2, ok := v2.(Regex)
		if ok {
			v1 := must.NotFail(v1.Compile())
			v2 := must.NotFail(v2.Compile())
			return compareOrdered(v1.String(), v2.String())
		}
		return Incomparable

	case int32:
		switch v2 := v2.(type) {
		case float64:
			return compareInvert(compareNumbers(v2, int64(v1)))
		case int32:
			return compareOrdered(v1, v2)
		case int64:
			return compareOrdered(int64(v1), v2)
		default:
			return Incomparable
		}

	case Timestamp:
		v2, ok := v2.(Timestamp)
		if ok {
			return compareOrdered(v1, v2)
		}
		return Incomparable

	case int64:
		switch v2 := v2.(type) {
		case float64:
			return compareInvert(compareNumbers(v2, v1))
		case int32:
			return compareOrdered(v1, int64(v2))
		case int64:
			return compareOrdered(v1, v2)
		default:
			return Incomparable
		}
	}

	panic("not reached")
}

// isScalar check if v is a BSON scalar value.
func isScalar(v any) bool {
	if v == nil {
		panic("v is nil")
	}

	switch v.(type) {
	case float64, string, Binary, ObjectID, bool, time.Time, NullType, Regex, int32, Timestamp, int64:
		return true
	}

	return false
}

// compareInvert swaps Less and Greater, keeping Equal and Incomparable.
func compareInvert(res CompareResult) CompareResult {
	switch res {
	case Equal:
		return Equal
	case Less:
		return Greater
	case Greater:
		return Less
	case Incomparable:
		return Incomparable
	}

	panic("not reached")
}

// compareOrdered compares BSON values of the same type using ==, <, > operators.
func compareOrdered[T constraints.Ordered](a, b T) CompareResult {
	switch {
	case a == b:
		return Equal
	case a < b:
		return Less
	case a > b:
		return Greater
	default:
		return Incomparable
	}
}

// compareNumbers compares BSON numbers.
func compareNumbers(a float64, b int64) CompareResult {
	if math.IsNaN(a) {
		return Incomparable
	}

	// TODO figure out correct precision
	bigA := new(big.Float).SetFloat64(a).SetPrec(100000)
	bigB := new(big.Float).SetInt64(b).SetPrec(100000)

	return CompareResult(bigA.Cmp(bigB))
}

// compareArrays compares indices of a filter array according to indices of a document array;
// returns Equal even when an array contains subarray equals to filter array;
// returns both Less and Greater when subarrays satisfy filter array. Example:
// document : [[44, 50], [43, 49]]
// filter : [44,50]
// result : Equal, Less, Greater.
func compareArrays(filterArr, docArr *Array) []CompareResult {
	if filterArr.Len() == 0 && docArr.Len() == 0 {
		return []CompareResult{Equal}
	}
	if filterArr.Len() > 0 && docArr.Len() == 0 {
		return []CompareResult{Less}
	}

	var entireCompareResult []CompareResult
	var subArrayEquality, gtAndLt, subArray bool

	for i := 0; i < docArr.Len(); i++ {
		docValue := must.NotFail(docArr.Get(i))
		switch docValue := docValue.(type) {
		case *Array:
			filterValue, errEmptyFilterArray := filterArr.Get(i)
			switch filterArrValue := filterValue.(type) {
			case *Array:
				iterationResult := compareArrays(filterArrValue, docValue)
				if i == 0 {
					entireCompareResult = append(entireCompareResult, iterationResult...)
				}
				entireCompareResult = append(entireCompareResult, Less) // always Less

			default:
				// handle empty embedded array
				_, err := docValue.Get(0)
				if err != nil && errEmptyFilterArray != nil {
					return []CompareResult{Greater, Equal}
				}
				if err != nil {
					return []CompareResult{Greater, Less}
				}

				subArray = true

				iterationResult := compareArrays(filterArr, docValue)
				if ContainsCompareResult(iterationResult, Equal) {
					subArrayEquality = true
				}

				if gtAndLt {
					continue // looking for subArrayEquality
				}

				if i == 0 {
					if errEmptyFilterArray != nil {
						return []CompareResult{Greater}
					}

					entireCompareResult = append(entireCompareResult, iterationResult...)
					entireCompareResult = append(entireCompareResult, CompareOrder(docValue, filterValue, Ascending))
				}

				entireCompareResult, gtAndLt = handleInconsistencyInResults(entireCompareResult, iterationResult, subArray)
			}

		default:
			if gtAndLt {
				continue // looking for subArrayEquality
			}

			filterValue, err := filterArr.Get(i)
			if err != nil {
				if filterArr.Len() == 0 { // here check (instead of beginning) to get #handle empty embedded array
					return []CompareResult{Greater}
				}

				if ContainsCompareResult(entireCompareResult, Equal) && docArr.Len() > filterArr.Len() {
					entireCompareResult = []CompareResult{Greater}
				}

				continue
			}

			if i == 0 { // set first non-Incomparable result
				entireCompareResult = []CompareResult{CompareOrder(docValue, filterValue, Ascending)}
				continue
			}

			iterationResult := compareScalars(docValue, filterValue)

			if !ContainsCompareResult(entireCompareResult, iterationResult) { // check inconsistency
				entireCompareResult, gtAndLt = handleInconsistencyInResults(entireCompareResult, iterationResult, subArray)
			}
		}
	}

	if ContainsCompareResult(entireCompareResult, Equal) && docArr.Len() < filterArr.Len() {
		return []CompareResult{Less}
	}

	if subArrayEquality {
		entireCompareResult = append(entireCompareResult, Equal)
	}

	return entireCompareResult
}

// ContainsCompareResult returns true if the result is in an array.
func ContainsCompareResult[T CompareResult](result []T, value T) bool {
	for _, v := range result {
		if v == value {
			return true
		}
	}
	return false
}

// deleteFromCompareResult removes an element from the CompareResult array.
func deleteFromCompareResult[T CompareResult](arr []T, value T) []T {
	for i, v := range arr {
		if v == value {
			arr[i] = arr[len(arr)-1]
			return arr[:len(arr)-1]
		}
	}

	return arr
}

// handleInconsistencyInResults resolves variability in the results of the entire result with the current iteration result.
func handleInconsistencyInResults(globalResult []CompareResult, iterationResult any, subArray bool) ([]CompareResult, bool) {
	var gtAndLt bool

	switch iterationResult := iterationResult.(type) {
	case []CompareResult:
		if ContainsCompareResult(globalResult, Equal) && !ContainsCompareResult(iterationResult, Equal) {
			globalResult = deleteFromCompareResult(globalResult, Equal)
			globalResult = append(globalResult, iterationResult...)
		}

		if (ContainsCompareResult(iterationResult, Less) && ContainsCompareResult(globalResult, Greater)) ||
			(ContainsCompareResult(iterationResult, Greater) && ContainsCompareResult(globalResult, Less)) {
			globalResult = []CompareResult{Greater, Less}
			gtAndLt = true
		}

		return globalResult, gtAndLt

	case CompareResult:
		if ContainsCompareResult(globalResult, Equal) {
			globalResult = deleteFromCompareResult(globalResult, Equal)
			globalResult = append(globalResult, iterationResult)
		}

		if !subArray { // two results can only be if there is a subarray, so skip it.
			return globalResult, false
		}

		if (iterationResult == Less && ContainsCompareResult(globalResult, Greater)) ||
			(iterationResult == Greater && ContainsCompareResult(globalResult, Less)) {
			globalResult = []CompareResult{Greater, Less}
			gtAndLt = true
		}

		return globalResult, gtAndLt

	default:
		panic(fmt.Sprintf("wrong iterationResult type: %[1]T, value: %[1]v", iterationResult))
	}
}
