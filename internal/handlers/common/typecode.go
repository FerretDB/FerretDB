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
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../../bin/stringer -linecomment -type typeCode

// typeCode represents BSON type codes.
// BSON type codes represent corresponding codes in BSON specification.
// They could be used to query fields with particular type values using $type operator.
// Type code `number` is added to support MongoDB surrogate alias `number` which matches double, int and long type values.
type typeCode int32

const (
	typeCodeDouble    = typeCode(1)  // double
	typeCodeString    = typeCode(2)  // string
	typeCodeObject    = typeCode(3)  // object
	typeCodeArray     = typeCode(4)  // array
	typeCodeBinData   = typeCode(5)  // binData
	typeCodeObjectID  = typeCode(7)  // objectId
	typeCodeBool      = typeCode(8)  // bool
	typeCodeDate      = typeCode(9)  // date
	typeCodeNull      = typeCode(10) // null
	typeCodeRegex     = typeCode(11) // regex
	typeCodeInt       = typeCode(16) // int
	typeCodeTimestamp = typeCode(17) // timestamp
	typeCodeLong      = typeCode(18) // long
	// Not implemented.
	typeCodeDecimal = typeCode(19)  // decimal
	typeCodeMinKey  = typeCode(-1)  // minKey
	typeCodeMaxKey  = typeCode(127) // maxKey
	// Not actual type code. `number` matches double, int and long.
	typeCodeNumber = typeCode(-128) // number
)

// newTypeCode returns typeCode and error by given code.
func newTypeCode(code int32) (typeCode, error) {
	c := typeCode(code)
	switch c {
	case typeCodeDouble, typeCodeString, typeCodeObject, typeCodeArray,
		typeCodeBinData, typeCodeObjectID, typeCodeBool, typeCodeDate,
		typeCodeNull, typeCodeRegex, typeCodeInt, typeCodeTimestamp, typeCodeLong, typeCodeNumber:
		return c, nil
	case typeCodeDecimal, typeCodeMinKey, typeCodeMaxKey:
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

// hasSameTypeElements returns true if types.Array elements has the same type.
// MongoDB consider int32, int64 and float64 that could be converted to int as the same type.
func hasSameTypeElements(array *types.Array) bool {
	var prev string
	for i := 0; i < array.Len(); i++ {
		var cur string

		element := must.NotFail(array.Get(i))

		if isWholeNumber(element) {
			cur = typeCodeNumber.String()
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

// aliasToTypeCode matches string type aliases to the corresponding typeCode value.
var aliasToTypeCode = map[string]typeCode{}

func init() {
	for _, i := range []typeCode{
		typeCodeDouble, typeCodeString, typeCodeObject, typeCodeArray,
		typeCodeBinData, typeCodeObjectID, typeCodeBool, typeCodeDate, typeCodeNull,
		typeCodeRegex, typeCodeInt, typeCodeTimestamp, typeCodeLong, typeCodeNumber,
	} {
		aliasToTypeCode[i.String()] = i
	}
}

// AliasFromType returns type alias name for given value.
func AliasFromType(v any) string {
	switch v := v.(type) {
	case *types.Document:
		return typeCodeObject.String()
	case *types.Array:
		return typeCodeArray.String()
	case float64:
		return typeCodeDouble.String()
	case string:
		return typeCodeString.String()
	case types.Binary:
		return typeCodeBinData.String()
	case types.ObjectID:
		return typeCodeObjectID.String()
	case bool:
		return typeCodeBool.String()
	case time.Time:
		return typeCodeDate.String()
	case types.NullType:
		return typeCodeNull.String()
	case types.Regex:
		return typeCodeRegex.String()
	case int32:
		return typeCodeInt.String()
	case types.Timestamp:
		return typeCodeTimestamp.String()
	case int64:
		return typeCodeLong.String()
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
