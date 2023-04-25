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

package commonparams

import (
	"fmt"
	"math"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../../bin/stringer -linecomment -type TypeCode

// TypeCode represents BSON type codes.
// BSON type codes represent corresponding codes in BSON specification.
// They could be used to query fields with particular type values using $type operator.
// Type code `number` is added to support MongoDB surrogate alias `number` which matches double, int and long type values.
type TypeCode int32

const (
	// TypeCodeDouble is a double type code.
	TypeCodeDouble = TypeCode(1) // double
	// TypeCodeString is a string type code.
	TypeCodeString = TypeCode(2) // string
	// TypeCodeObject is an object type code.
	TypeCodeObject = TypeCode(3) // object
	// TypeCodeArray is an array type code.
	TypeCodeArray = TypeCode(4) // array
	// TypeCodeBinData is a binary data type code.
	TypeCodeBinData = TypeCode(5) // binData
	// TypeCodeObjectID is an object id type code.
	TypeCodeObjectID = TypeCode(7) // objectId
	// TypeCodeBool is a boolean type code.
	TypeCodeBool = TypeCode(8) // bool
	// TypeCodeDate is a date type code.
	TypeCodeDate = TypeCode(9) // date
	// TypeCodeNull is a null type code.
	TypeCodeNull = TypeCode(10) // null
	// TypeCodeRegex is a regex type code.
	TypeCodeRegex = TypeCode(11) // regex
	// TypeCodeInt is an int type code.
	TypeCodeInt = TypeCode(16) // int
	// TypeCodeTimestamp is a timestamp type code.
	TypeCodeTimestamp = TypeCode(17) // timestamp
	// TypeCodeLong is a long type code.
	TypeCodeLong = TypeCode(18) // long

	// Not implemented.

	// TypeCodeDecimal is a decimal type code.
	TypeCodeDecimal = TypeCode(19) // decimal
	// TypeCodeMinKey is a minKey type code.
	TypeCodeMinKey = TypeCode(-1) // minKey
	// TypeCodeMaxKey is a maxKey type code.
	TypeCodeMaxKey = TypeCode(127) // maxKey

	// Not actual type code. `number` matches double, int and long.

	// TypeCodeNumber is a number type code.
	TypeCodeNumber = TypeCode(-128) // number
)

// NewTypeCode returns TypeCode and error by given code.
func NewTypeCode(code int32) (TypeCode, error) {
	c := TypeCode(code)
	switch c {
	case TypeCodeDouble, TypeCodeString, TypeCodeObject, TypeCodeArray,
		TypeCodeBinData, TypeCodeObjectID, TypeCodeBool, TypeCodeDate,
		TypeCodeNull, TypeCodeRegex, TypeCodeInt, TypeCodeTimestamp, TypeCodeLong, TypeCodeNumber:
		return c, nil
	case TypeCodeDecimal, TypeCodeMinKey, TypeCodeMaxKey:
		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			fmt.Sprintf(`Type code %v not implemented`, code),
			"$type",
		)
	default:
		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf(`Invalid numerical type code: %d`, code),
			"$type",
		)
	}
}

// HasSameTypeElements returns true if types.Array elements has the same type.
// MongoDB consider int32, int64 and float64 that could be converted to int as the same type.
func HasSameTypeElements(array *types.Array) bool {
	var prev string

	for i := 0; i < array.Len(); i++ {
		var cur string

		element := must.NotFail(array.Get(i))

		if isWholeNumber(element) {
			cur = TypeCodeNumber.String()
		} else {
			cur = AliasFromType(element)
		}

		if prev == "" {
			prev = cur
			continue
		}

		if prev != cur {
			return false
		}
	}

	return true
}

// aliasToTypeCode matches string type aliases to the corresponding TypeCode value.
var aliasToTypeCode = map[string]TypeCode{}

func init() {
	for _, i := range []TypeCode{
		TypeCodeDouble, TypeCodeString, TypeCodeObject, TypeCodeArray,
		TypeCodeBinData, TypeCodeObjectID, TypeCodeBool, TypeCodeDate, TypeCodeNull,
		TypeCodeRegex, TypeCodeInt, TypeCodeTimestamp, TypeCodeLong, TypeCodeNumber,
	} {
		aliasToTypeCode[i.String()] = i
	}
}

// AliasFromType returns type alias name for given value.
func AliasFromType(v any) string {
	switch v := v.(type) {
	case *types.Document:
		return TypeCodeObject.String()
	case *types.Array:
		return TypeCodeArray.String()
	case float64:
		return TypeCodeDouble.String()
	case string:
		return TypeCodeString.String()
	case types.Binary:
		return TypeCodeBinData.String()
	case types.ObjectID:
		return TypeCodeObjectID.String()
	case bool:
		return TypeCodeBool.String()
	case time.Time:
		return TypeCodeDate.String()
	case types.NullType:
		return TypeCodeNull.String()
	case types.Regex:
		return TypeCodeRegex.String()
	case int32:
		return TypeCodeInt.String()
	case types.Timestamp:
		return TypeCodeTimestamp.String()
	case int64:
		return TypeCodeLong.String()
	default:
		panic(fmt.Sprintf("not supported type %T", v))
	}
}

// isWholeNumber checks is the given value represents a whole number type.
func isWholeNumber(v any) bool {
	switch v := v.(type) {
	case float64:
		return !(v != math.Trunc(v) || math.IsNaN(v) || math.IsInf(v, 0))
	case int32:
		return true
	case int64:
		return true
	default:
		return false
	}
}

// ParseTypeCode returns TypeCode and error by given type code alias.
func ParseTypeCode(alias string) (TypeCode, error) {
	code, ok := aliasToTypeCode[alias]
	if !ok {
		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf(`Unknown type name alias: %s`, alias),
			"$type",
		)
	}

	return code, nil
}
