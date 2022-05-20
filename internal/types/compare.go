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
//go:generate ../../bin/stringer -linecomment -type dataTypeOrderResult
//go:generate ../../bin/stringer -linecomment -type numberDataTypeOrderResult
//go:generate ../../bin/stringer -linecomment -type SortType

// CompareResult represents the result of a comparison.
type CompareResult int8

// Values match results of comparison functions such as bytes.Compare.
// They do not match MongoDB `sort` values where 1 means ascending order and -1 means descending.
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
func Compare(v1, v2 any) CompareResult {
	if v1 == nil || v2 == nil {
		panic(fmt.Sprintf("compare: values should not be nil, v1 is %v, v2 is %v", v1, v2))
	}

	switch v1 := v1.(type) {
	case *Document:
		// TODO: implement document comparing
		return Incomparable

	case *Array:
		for i := 0; i < v1.Len(); i++ {
			v := must.NotFail(v1.Get(i))
			switch v.(type) {
			case *Document, *Array:
				continue
			}

			if res := compareScalars(v, v2); res != Incomparable {
				return res
			}
		}
		return Incomparable

	default:
		return compareScalars(v1, v2)
	}
}

// compareScalars compares BSON scalar values.
func compareScalars(v1, v2 any) CompareResult {
	compareEnsureScalar(v1)
	compareEnsureScalar(v2)

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

// compareEnsureScalar panics if v is not a BSON scalar value.
func compareEnsureScalar(v any) {
	if v == nil {
		panic("v is nil")
	}

	switch v.(type) {
	case float64, string, Binary, ObjectID, bool, time.Time, NullType, Regex, int32, Timestamp, int64:
		return
	}

	panic(fmt.Sprintf("non-scalar type %T", v))
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

// dataTypeOrderResult represents the comparison order of data types.
type dataTypeOrderResult uint8

const (
	_ dataTypeOrderResult = iota
	nullDataType
	nanDataType
	numbersDataType
	stringDataType
	objectDataType
	arrayDataType
	binDataType
	objectIdDataType
	booleanDataType
	dateDataType
	timestampDataType
	regexDataType
)

// defineDataType define which type has value and returns a sequence the type has.
func defineDataType(value any) dataTypeOrderResult {
	switch value := value.(type) {
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
		return objectIdDataType
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

// numberDataTypeOrderResult represents the comparison order of numbers.
type numberDataTypeOrderResult uint8

const (
	_ numberDataTypeOrderResult = iota
	doubleNegativeZero
	doubleDT
	int32DT
	int64DT
)

// defineNumberDataType define which number type has value and returns a sequence the type has.
func defineNumberDataType(value any) numberDataTypeOrderResult {
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
	}

	panic(fmt.Sprintf("defineNumberDataType: value cannot be defined, value is %v, data type of value is %T", value, value))
}

// SortType represents sort type for $sort aggregation.
type SortType int8

const (
	Ascending  SortType = 1  // asc
	Descending SortType = -1 // desc
)

// CompareOrder defines the data type for the two values and compares them.
// When the types are equal, it compares their values using Compare.
func CompareOrder(a, b any, order SortType) CompareResult {
	if a == nil || b == nil {
		panic(fmt.Sprintf("CompareOrder: values should not be nil, a is %v, b is %v", a, b))
	}

	aType := defineDataType(a)
	bType := defineDataType(b)
	if aType == bType {
		res := Compare(a, b)
		if res == Equal && aType == numbersDataType {
			aNumberType := defineNumberDataType(a)
			bNumberType := defineNumberDataType(b)
			switch {
			case aNumberType < bNumberType && order == Ascending:
				return Less
			case aNumberType > bNumberType && order == Ascending:
				return Greater
			case aNumberType < bNumberType && order == Descending:
				return Greater
			case aNumberType > bNumberType && order == Descending:
				return Less
			}
		}

		return res
	}

	switch {
	case aType < bType:
		return Less
	case aType > bType:
		return Greater
	}

	panic("CompareOrder: not reached")
}
