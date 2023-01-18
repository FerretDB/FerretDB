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

// SortType represents sort type for $sort aggregation.
type SortType int8

const (
	// Ascending is used for sort in ascending order.
	Ascending SortType = 1

	// Descending is used for sort in descending order.
	Descending SortType = -1
)

// numberOrderResult represents the comparison order of numbers.
type numberOrderResult uint8

const (
	_ numberOrderResult = iota
	doubleDT
	int32DT
	int64DT
)

// detectNumberType returns a comparison order of a number type
// for float64, int32 and int64 types.
func detectNumberType(value any) numberOrderResult {
	switch value := value.(type) {
	case float64:
		return doubleDT
	case int32:
		return int32DT
	case int64:
		return int64DT
	default:
		panic(fmt.Sprintf("detectNumberType: value cannot be defined, value is %[1]v, data type of value is %[1]T", value))
	}
}

// CompareOrder detects the data type for two values and compares them.
// When the types are equal, it compares their values using Compare.
// This is used by update operator $max.
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

// IsSameForUpdate detects the data type for two values and compares them.
// When the types are equal, it compares their values using Compare.
// If they are equal and number type, it compares number type.
// This is used by update operator $set.
func IsSameForUpdate(a, b any, order SortType) bool {
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
		return false
	}

	if result = Compare(a, b); result != Equal {
		return false
	}

	if isNumber(a) {
		return isSameNumberType(a, b)
	}

	return true
}

// CompareOrderForSort detects the data type for two values and compares them.
// If a or b is array, the minimum element of each array
// is used for Ascending sort, and maximum element for
// Descending sort. Hence, an array is used
// for type comparison and value comparison.
// Empty Array is smaller than Null.
//
// This is used by sort operation.
func CompareOrderForSort(a, b any, order SortType) CompareResult {
	if a == nil {
		panic("CompareOrderForSort: a is nil")
	}

	if b == nil {
		panic("CompareOrderForSort: b is nil")
	}

	if order != Ascending && order != Descending {
		panic(fmt.Sprintf("CompareOrder: order is %v", order))
	}

	arrA, isAArray := a.(*Array)
	arrB, isBArray := b.(*Array)

	// empty array is the lowest on the sort order.
	switch {
	case isAArray && arrA.Len() == 0 && isBArray && arrB.Len() == 0:
		return Equal
	case isAArray && arrA.Len() == 0:
		if order == Ascending {
			return Less
		}

		return Greater
	case isBArray && arrB.Len() == 0:
		if order == Ascending {
			return Greater
		}

		return Less
	}

	// sort does not compare array itself, it compares
	// minimum element in array for ascending sort and
	// maximum element in array for descending sort.
	if isAArray {
		a = getComparisonElementFromArray(arrA, order)
	}

	if isBArray {
		b = getComparisonElementFromArray(arrB, order)
	}

	if result := compareTypeOrder(a, b); result != Equal {
		if order == Ascending {
			return result
		}

		// invert the result for descending sort.
		return compareInvert(result)
	}

	result := Compare(a, b)
	if order == Ascending {
		return result
	}

	// invert the result for descending sort.
	return compareInvert(result)
}

// CompareOrderForOperator detects the data type for two values and compares them.
// If a is an array, the array is filtered by the same type as
// b type.
// It is used by $gt, $gte, $lt and $lte comparison.
func CompareOrderForOperator(a, b any, order SortType) CompareResult {
	if a == nil {
		panic("CompareOrderForOperator: a is nil")
	}

	if b == nil {
		panic("CompareOrderForOperator: b is nil")
	}

	if order != Ascending && order != Descending {
		panic(fmt.Sprintf("CompareOrder: order is %v", order))
	}

	arrA, isAArray := a.(*Array)
	arrB, isBArray := b.(*Array)

	if isAArray && !isBArray {
		// if only a is array, filter the array by
		// only keeping the same type as b.
		arrA = arrA.FilterArrayByType(b)
	}

	if isAArray && arrA.Len() == 0 && isBArray && arrB.Len() == 0 {
		return Equal
	}

	if !isAArray && isBArray && arrB.Len() == 0 {
		// empty array is the lowest on the sort order.
		if order == Ascending {
			return Greater
		}

		return Less
	}

	if isAArray && !isBArray {
		a = getComparisonElementFromArray(arrA, order)
	}

	if result := compareTypeOrder(a, b); result != Equal {
		if order == Ascending {
			return Greater
		}

		return Less
	}

	return Compare(a, b)
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

// isNumber returns true if a value is numbersDataType.
func isNumber(value any) bool {
	return detectDataType(value) == numbersDataType
}

// isSameNumberType returns true if a and b are the same
// number type.
func isSameNumberType(a, b any) bool {
	aType := detectNumberType(a)
	bType := detectNumberType(b)

	return aType == bType
}

// getComparisonElementFromArray gets an element use for
// comparison according to the sort order.
// For Ascending order minimum element is retrieved, and
// for descending order maximum element is retrieved.
func getComparisonElementFromArray(arr *Array, order SortType) any {
	if arr.Len() == 0 {
		return arr
	}

	if order == Ascending {
		return arr.Min()
	}

	if order == Descending {
		return arr.Max()
	}

	panic("unsupported sort type")
}
