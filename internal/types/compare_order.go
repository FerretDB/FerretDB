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
	"fmt"
	"math"
	"time"
)

//go:generate ../../bin/stringer -linecomment -type compareTypeOrderResult
//go:generate ../../bin/stringer -linecomment -type numberOrderResult
//go:generate ../../bin/stringer -linecomment -type SortType

// compareTypeOrderResult represents the comparison order of data types.
type compareTypeOrderResult uint8

// TODO: handle sorting for documentDataType and arrayDataType; https://github.com/FerretDB/FerretDB/issues/457
const (
	_ compareTypeOrderResult = iota
	nullDataType
	nanDataType
	numbersDataType
	stringDataType
	documentDataType
	arrayDataType
	binDataType
	objectIDDataType
	booleanDataType
	dateDataType
	timestampDataType
	regexDataType
)

// detectDataType returns a sequence for build-in type.
func detectDataType(value any) compareTypeOrderResult {
	switch value := value.(type) {
	case *Document:
		return documentDataType
	case *Array:
		return arrayDataType
	case float64:
		if math.IsNaN(value) {
			return nanDataType
		}
		return numbersDataType
	case string:
		return stringDataType
	case Binary:
		return binDataType
	case ObjectID:
		return objectIDDataType
	case bool:
		return booleanDataType
	case time.Time:
		return dateDataType
	case NullType:
		return nullDataType
	case Regex:
		return regexDataType
	case int32:
		return numbersDataType
	case Timestamp:
		return timestampDataType
	case int64:
		return numbersDataType
	default:
		panic(fmt.Sprintf("value cannot be defined, value is %[1]v, data type of value is %[1]T", value))
	}
}

// numberOrderResult represents the comparison order of numbers.
type numberOrderResult uint8

const (
	_ numberOrderResult = iota
	doubleNegativeZero
	doubleDT
	int32DT
	int64DT
)

// detectNumberType returns a sequence for float64, int32 and int64 types.
func detectNumberType(value any) numberOrderResult {
	switch value := value.(type) {
	case float64:
		if value == 0 && math.Signbit(value) {
			return doubleNegativeZero
		}
		return doubleDT
	case int32:
		return int32DT
	case int64:
		return int64DT
	default:
		panic(fmt.Sprintf("detectNumberType: value cannot be defined, value is %[1]v, data type of value is %[1]T", value))
	}
}

// SortType represents sort type for $sort aggregation.
type SortType int8

const (
	// Ascending is used for sort in ascending order.
	Ascending SortType = 1

	// Descending is used for sort in descending order.
	Descending SortType = -1
)

// CompareOrder detects the data type for two values and compares them.
// When the types are equal, it compares their values using Compare.
// Use this it for getting sort order.
func CompareOrder(a, b any, order SortType) CompareResult {
	if a == nil {
		panic("CompareOrder: a is nil")
	}
	if b == nil {
		panic("CompareOrder: b is nil")
	}
	if order != Ascending && order != Descending {
		panic(fmt.Sprintf("CompareOrder: order is %v", order))
	}

	result := compareTypeOrder(a, b)
	if result != Equal {
		return result
	}

	return Compare(a, b)
}

// CompareForSort detects the data type for two values and compares them.
// If a or b is array, the minimum element of each array
// is used for Ascending sort, and maximum element for
// Descending sort. Hence, an element of an array is used
// for type comparison and value comparison.
// Empty Array is smaller than Null.
//
// Use it for sorting documents such as sort().
func CompareForSort(a, b any, order SortType) CompareResult {
	if a == nil {
		panic("CompareOrder: a is nil")
	}
	if b == nil {
		panic("CompareOrder: b is nil")
	}
	if order != Ascending && order != Descending {
		panic(fmt.Sprintf("CompareOrder: order is %v", order))
	}

	a, b = unwrapArray(a, b, order)
	arrA, isAArray := a.(*Array)
	arrB, isBArray := b.(*Array)

	switch {
	case isAArray && arrA.Len() == 0 && isBArray && arrB.Len() == 0:
		return Equal
	case isAArray && arrA.Len() == 0:
		return getLess(order)
	case isBArray && arrB.Len() == 0:
		return getGreater(order)
	}

	result := CompareOrder(a, b, order)
	if result == Equal {
		return Equal
	}
	if result == Less {
		return getLess(order)
	}
	return getGreater(order)
}

// CompareOrderForOperator detects the data type for two values and compares them.
// If a or b is array, the array is filtered by the same type as
// non-array type.
// Use it for $gt $lt comparison.
func CompareOrderForOperator(a, b any, order SortType) CompareResult {
	if a == nil {
		panic("CompareOrder: a is nil")
	}
	if b == nil {
		panic("CompareOrder: b is nil")
	}
	if order != Ascending && order != Descending {
		panic(fmt.Sprintf("CompareOrder: order is %v", order))
	}

	arrA, isAArray := a.(*Array)
	arrB, isBArray := b.(*Array)

	switch {
	case isAArray && isBArray:
	case isAArray && !isBArray:
		arrA = arrA.FilterArrayByType(b)
		if arrA.Len() == 0 {
			if order == Ascending {
				return Greater
			}
			return Less
		}
		a = getComparisonElement(arrA, order)
	case isBArray && arrB.Len() == 0:
		arrB = arrB.FilterArrayByType(a)
		if arrB.Len() == 0 {
			return getGreater(order)
		}
		b = getComparisonElement(arrB, order)
	}

	if result := compareTypeOrder(a, b); result != Equal {
		// they are not the same type, so no match
		if order == Ascending {
			return Greater
		}
		return Less
	}

	result := compareTypeOrder(a, b)
	if result != Equal {
		return result
	}

	return Compare(a, b)
}

// compareNumberOrder detects the number type for two values and compares them.
func compareNumberOrder(a, b any, order SortType) CompareResult {
	if a == nil {
		panic("compareNumberOrder: a is nil")
	}
	if b == nil {
		panic("compareNumberOrder: b is nil")
	}

	aNumberType := detectNumberType(a)
	bNumberType := detectNumberType(b)
	switch {
	case aNumberType < bNumberType && order == Ascending:
		return Less
	case aNumberType > bNumberType && order == Ascending:
		return Greater
	case aNumberType < bNumberType && order == Descending:
		return Greater
	case aNumberType > bNumberType && order == Descending:
		return Less
	default:
		return Equal
	}
}

// compareTypeOrder detects the data type for two values and compares them.
func compareTypeOrder(a, b any) CompareResult {
	aType := detectDataType(a)
	bType := detectDataType(b)

	switch {
	case aType < bType:
		return Less
	case aType > bType:
		return Greater
	default:
		return Equal
	}
}

// unwrapArray gets an element used for comparison from an array.
// If a or b is array, the minimum element of each array
// is used for Ascending sort, and maximum element for
// Descending sort. Hence, an element of an array is used
// for type comparison and value comparison.
func unwrapArray(a, b any, order SortType) (any, any) {
	switch arrA := a.(type) {
	case *Array:
		switch arrB := b.(type) {
		case *Array:
			if arrB.Len() == 0 {
				return arrA, arrB
			}

			b = getComparisonElement(arrB, order)
		default:
			// a is an array but b is not an array
		}

		if arrA.Len() == 0 {
			return arrA, b
		}

		a = getComparisonElement(arrA, order)
		return a, b

	default:
		switch arrB := b.(type) {
		case *Array:
			// a is not an array but b is an array
			if arrB.Len() == 0 {
				return arrA, arrB
			}

			b = getComparisonElement(arrB, order)
		default:
			// both not array
		}
		return a, b
	}
}

// getComparisonElement gets an element from an array to
// use for comparison.
func getComparisonElement(arr *Array, order SortType) any {
	if order == Ascending {
		return arr.Min()
	}

	if order == Descending {
		return arr.Max()
	}

	panic("unsupported sort type")
}

// getLess computes response Less for sort type.
func getLess(order SortType) CompareResult {
	if order == Ascending {
		return Less
	}
	return Greater
}

// getGreater computes response Greater for sort type.
func getGreater(order SortType) CompareResult {
	if order == Ascending {
		return Greater
	}
	return Less
}
