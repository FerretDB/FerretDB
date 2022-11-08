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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// GetRequiredParam returns doc's value for key or protocol error for missing or invalid parameter.
func GetRequiredParam[T types.Type](doc *types.Document, key string) (T, error) {
	var zero T

	v, err := doc.Get(key)
	if err != nil {
		msg := fmt.Sprintf("required parameter %q is missing", key)
		return zero, NewCommandErrorMsgWithArgument(ErrBadValue, msg, key)
	}

	res, ok := v.(T)
	if !ok {
		msg := fmt.Sprintf("required parameter %q has type %T (expected %T)", key, v, zero)
		return zero, NewCommandErrorMsgWithArgument(ErrBadValue, msg, key)
	}

	return res, nil
}

// GetOptionalParam returns doc's value for key, default value for missing parameter, or protocol error for invalid parameter.
func GetOptionalParam[T types.Type](doc *types.Document, key string, defaultValue T) (T, error) {
	v, err := doc.Get(key)
	if err != nil {
		return defaultValue, nil
	}

	res, ok := v.(T)
	if !ok {
		msg := fmt.Sprintf(
			`BSON field '%s' is the wrong type '%s', expected type '%s'`,
			key, AliasFromType(v), AliasFromType(defaultValue),
		)

		return defaultValue, NewCommandErrorMsgWithArgument(ErrTypeMismatch, msg, key)
	}

	return res, nil
}

// GetBoolOptionalParam returns doc's bool value for key.
// Non-zero double, long, and int values return true.
// Zero values for those types, as well as nulls and missing fields, return false.
// Other types return a protocol error.
func GetBoolOptionalParam(doc *types.Document, key string) (bool, error) {
	v, err := doc.Get(key)
	if err != nil {
		return false, nil
	}

	switch v := v.(type) {
	case float64:
		return v != 0, nil
	case bool:
		return v, nil
	case types.NullType:
		return false, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	default:
		msg := fmt.Sprintf(
			`BSON field '%s' is the wrong type '%s', expected types '[bool, long, int, decimal, double]'`,
			key,
			AliasFromType(v),
		)

		return false, NewCommandErrorMsgWithArgument(ErrTypeMismatch, msg, key)
	}
}

// AssertType asserts value's type, returning protocol error for unexpected types.
//
// If a custom error is needed, use a normal Go type assertion instead:
//
//	d, ok := value.(*types.Document)
//	if !ok {
//	  return NewCommandErrorMsg(ErrBadValue, "expected document")
//	}
func AssertType[T types.Type](value any) (T, error) {
	res, ok := value.(T)
	if !ok {
		msg := fmt.Sprintf("got type %T, expected %T", value, res)
		return res, NewCommandErrorMsg(ErrBadValue, msg)
	}

	return res, nil
}

var (
	errUnexpectedType        = fmt.Errorf("unexpected type")
	errNotWholeNumber        = fmt.Errorf("not a whole number")
	errNegativeNumber        = fmt.Errorf("negative number")
	errNotBinaryMask         = fmt.Errorf("not a binary mask")
	errUnexpectedLeftOpType  = fmt.Errorf("unexpected left operand type")
	errUnexpectedRightOpType = fmt.Errorf("unexpected right operand type")
	errLongExceeded          = fmt.Errorf("long exceeded")
	errIntExceeded           = fmt.Errorf("int exceeded")
	errNaN                   = fmt.Errorf("not a number")
	errInfinity              = fmt.Errorf("infinity")
)

// GetWholeNumberParam checks if the given value is int32, int64, or float64 containing a whole number,
// such as used in the limit, $size, etc.
func GetWholeNumberParam(value any) (int64, error) {
	switch value := value.(type) {
	// TODO: add string support https://github.com/FerretDB/FerretDB/issues/1089
	case float64:
		// TODO check float negative zero (math.Copysign(0, -1))
		if math.IsNaN(value) {
			return 0, errNaN
		}

		if math.IsInf(value, 1) {
			return 0, errInfinity
		}

		if value != math.Trunc(value) {
			return 0, errNotWholeNumber
		}

		return int64(value), nil
	case int32:
		return int64(value), nil
	case int64:
		return value, nil
	default:
		return 0, errUnexpectedType
	}
}

// getBinaryMaskParam matches value type, returning bit mask and error if match failed.
// Possible values are: position array ([1,3,5] == 010101), whole number value and types.Binary value.
func getBinaryMaskParam(mask any) (uint64, error) {
	var bitmask uint64

	switch mask := mask.(type) {
	case *types.Array:
		// {field: {$bitsAllClear: [position1, position2]}}
		for i := 0; i < mask.Len(); i++ {
			val := must.NotFail(mask.Get(i))
			b, ok := val.(int32)
			if !ok {
				return 0, NewCommandError(
					ErrBadValue,
					fmt.Errorf(`Failed to parse bit position. Expected a number in: %d: %#v`, i, val),
				)
			}

			if b < 0 {
				return 0, NewCommandError(
					ErrBadValue,
					fmt.Errorf("Failed to parse bit position. Expected a non-negative number in: %d: %d", i, b),
				)
			}

			bitmask |= 1 << uint64(math.Min(float64(b), 63))
		}

	case float64:
		// {field: {$bitsAllClear: bitmask}}
		// TODO check float negative zero
		if mask != math.Trunc(mask) || math.IsNaN(mask) || math.IsInf(mask, 0) {
			return 0, errNotWholeNumber
		}

		if mask < 0 {
			return 0, errNegativeNumber
		}

		bitmask = uint64(mask)

	case types.Binary:
		// {field: {$bitsAllClear: BinData()}}
		for b := 0; b < len(mask.B); b++ {
			byteAt := mask.B[b]

			if byteAt == 0 {
				continue
			}

			if b < 8 {
				bitmask |= uint64(byteAt) << uint64(b*8)
			} else {
				bitmask |= 1 << 63
			}
		}

	case int32:
		// {field: {$bitsAllClear: bitmask}}
		if mask < 0 {
			return 0, errNegativeNumber
		}

		bitmask = uint64(mask)

	case int64:
		// {field: {$bitsAllClear: bitmask}}
		if mask < 0 {
			return 0, errNegativeNumber
		}

		bitmask = uint64(mask)

	default:
		return 0, errNotBinaryMask
	}

	return bitmask, nil
}

// parseTypeCode returns typeCode and error by given type code alias.
func parseTypeCode(alias string) (typeCode, error) {
	code, ok := aliasToTypeCode[alias]
	if !ok {
		return 0, NewCommandErrorMsgWithArgument(ErrBadValue, fmt.Sprintf(`Unknown type name alias: %s`, alias), "$type")
	}

	return code, nil
}

// addNumbers returns the result of v1 and v2 addition and error if addition failed.
// The v1 and v2 parameters could be float64, int32, int64.
// The result would be the broader type possible, i.e. int32 + int64 produces int64.
func addNumbers(v1, v2 any) (any, error) {
	switch v1 := v1.(type) {
	case float64:
		switch v2 := v2.(type) {
		case float64:
			return v1 + v2, nil
		case int32:
			return v1 + float64(v2), nil
		case int64:
			return v1 + float64(v2), nil
		default:
			return nil, errUnexpectedRightOpType
		}
	case int32:
		switch v2 := v2.(type) {
		case float64:
			return v2 + float64(v1), nil
		case int32:
			if v2 == math.MaxInt32 && v1 > 0 {
				return int64(v1) + int64(v2), nil
			}

			if v2 == math.MinInt32 && v1 < 0 {
				return int64(v1) + int64(v2), nil
			}

			return v1 + v2, nil
		case int64:
			if v2 > 0 {
				if int64(v1) > math.MaxInt64-v2 {
					return nil, errLongExceeded
				}
			} else {
				if int64(v1) < math.MinInt64-v2 {
					return nil, errLongExceeded
				}
			}

			return v2 + int64(v1), nil
		default:
			return nil, errUnexpectedRightOpType
		}
	case int64:
		switch v2 := v2.(type) {
		case float64:
			return v2 + float64(v1), nil
		case int32:
			if v2 > 0 {
				if v1 > math.MaxInt64-int64(v2) {
					return nil, errIntExceeded
				}
			} else {
				if v1 < math.MinInt64-int64(v2) {
					return nil, errIntExceeded
				}
			}

			return v1 + int64(v2), nil
		case int64:
			if v2 > 0 {
				if v1 > math.MaxInt64-v2 {
					return nil, errLongExceeded
				}
			} else {
				if v1 < math.MinInt64-v2 {
					return nil, errLongExceeded
				}
			}

			return v1 + v2, nil
		default:
			return nil, errUnexpectedRightOpType
		}
	default:
		return nil, errUnexpectedLeftOpType
	}
}

// GetOptionalPositiveNumber returns doc's value for key or protocol error for invalid parameter.
func GetOptionalPositiveNumber(document *types.Document, key string) (int32, error) {
	v, err := document.Get(key)
	if err != nil {
		return 0, nil
	}

	wholeNumberParam, err := GetWholeNumberParam(v)
	if err != nil {
		switch err {
		case errUnexpectedType:
			return 0, NewCommandErrorMsgWithArgument(ErrBadValue, fmt.Sprintf("%s must be a number", key), key)
		case errNotWholeNumber:
			if _, ok := v.(float64); ok {
				return 0, NewCommandErrorMsgWithArgument(
					ErrBadValue,
					fmt.Sprintf("%v has non-integral value", key),
					key,
				)
			}

			return 0, NewCommandErrorMsgWithArgument(ErrBadValue, fmt.Sprintf("%s must be a whole number", key), key)
		default:
			return 0, err
		}
	}

	if wholeNumberParam > math.MaxInt32 || wholeNumberParam < math.MinInt32 {
		return 0, NewCommandErrorMsgWithArgument(
			ErrBadValue,
			fmt.Sprintf("%v value for %s is out of range", int64(wholeNumberParam), key),
			key,
		)
	}

	value := int32(wholeNumberParam)
	if value < 0 {
		return 0, NewCommandErrorMsgWithArgument(
			ErrBadValue,
			fmt.Sprintf("%v value for %s is out of range", value, key),
			key,
		)
	}

	return value, nil
}
