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
//
// Compare and contrast with test helpers in testutil package.
func Compare(docValue, filterValue any) CompareResult {
	if docValue == nil {
		panic("compare: docValue is nil")
	}
	if filterValue == nil {
		panic("compare: filterValue is nil")
	}

	switch docValue := docValue.(type) {
	case *Document:
		if filterDoc, ok := filterValue.(*Document); ok {
			return compareDocuments(docValue, filterDoc)
		}

		return Incomparable
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
				return res
			}
		}

		return Incomparable

	default:
		return compareScalars(docValue, filterValue)
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
// returns Equal when an array equals to filter array;
// returns Less when an index of the document array is less than the index of the filter array;
// returns Greater when an index of the document array is greater than the index of the filter array;
// returns Incomparable when an index comparison detects Composite types.
func compareArrays(filterArr, docArr *Array) CompareResult {
	if filterArr.Len() == 0 && docArr.Len() == 0 {
		return Equal
	}
	if filterArr.Len() > 0 && docArr.Len() == 0 {
		return Less
	}

	if filterArr.Len() == 0 {
		return Greater
	}

	for i := 0; i < docArr.Len(); i++ {
		docValue := must.NotFail(docArr.Get(i))

		filterValue, err := filterArr.Get(i)
		if err != nil {
			if docArr.Len() > filterArr.Len() {
				return Greater
			}

			continue
		}

		orderResult := CompareOrder(docValue, filterValue, Ascending)
		if orderResult != Equal {
			return orderResult
		}

		iterationResult := compareScalars(docValue, filterValue)
		if iterationResult != Equal {
			return iterationResult
		}
	}

	if docArr.Len() < filterArr.Len() {
		return Less
	}

	return Equal
}

// compareDocuments compares documents recursively by
// comparing them in the order of types, field names and field values.
func compareDocuments(a, b *Document) CompareResult {
	if a.Len() == 0 && b.Len() == 0 {
		return Equal
	}

	if a.Len() == 0 && b.Len() > 0 {
		return Less
	}

	if b.Len() == 0 {
		return Greater
	}

	aKeys := a.Keys()
	bKeys := b.Keys()
	bValues := b.Values()
	aValues := a.Values()

	for i, aKey := range aKeys {
		if b.Len() == i {
			return Greater
		}

		// compare type
		if result := compareTypeOrder(aValues[i], bValues[i]); result != Equal {
			return result
		}

		// compare keys
		if result := compareScalars(aKey, bKeys[i]); result != Equal {
			return result
		}

		// compare values
		if result := Compare(aValues[i], bValues[i]); result != Equal {
			return result
		}
	}

	if a.Len() < b.Len() {
		return Less
	}

	return Equal
}
