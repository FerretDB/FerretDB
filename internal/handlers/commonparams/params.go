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
	"errors"
	"fmt"
	"math"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

var (
	// ErrNotWholeNumber is returned when a non-whole number is given.
	ErrNotWholeNumber = fmt.Errorf("not a whole number")
	// ErrUnexpectedLeftOpType is returned when an unexpected left operand type is given.
	ErrUnexpectedLeftOpType = fmt.Errorf("unexpected left operand type")
	// ErrUnexpectedRightOpType is returned when an unexpected right operand type is given.
	ErrUnexpectedRightOpType = fmt.Errorf("unexpected right operand type")
	// ErrLongExceededPositive is returned when a positive long value is given that exceeds the maximum value.
	ErrLongExceededPositive = fmt.Errorf("long exceeded - positive value")
	// ErrLongExceededNegative is returned when a negative long value is given that exceeds the minimum value.
	ErrLongExceededNegative = fmt.Errorf("long exceeded - negative value")
	// ErrIntExceeded is returned when an int value is given that exceeds the maximum value.
	ErrIntExceeded = fmt.Errorf("int exceeded")
	// ErrInfinity is returned when an infinity value is given.
	ErrInfinity = fmt.Errorf("infinity")
	// ErrUnexpectedType is returned when an unexpected type is given.
	ErrUnexpectedType = fmt.Errorf("unexpected type")
)

// GetWholeNumberParam checks if the given value is int32, int64, or float64 containing a whole number,
// such as used in the limit, $size, etc.
func GetWholeNumberParam(value any) (int64, error) {
	switch value := value.(type) {
	// TODO: add string support https://github.com/FerretDB/FerretDB/issues/1089
	case float64:
		switch {
		case math.IsInf(value, 1):
			return 0, ErrInfinity
		case value > float64(math.MaxInt64):
			return 0, ErrLongExceededPositive
		case value < float64(math.MinInt64):
			return 0, ErrLongExceededNegative
		case value != math.Trunc(value):
			return 0, ErrNotWholeNumber
		}

		return int64(value), nil
	case int32:
		return int64(value), nil
	case int64:
		return value, nil
	default:
		return 0, ErrUnexpectedType
	}
}

// GetWholeParamStrict validates the given value for find and count commands.
//
// If the value is valid, it returns its int64 representation,
// otherwise it returns a command error with the given command being mentioned.
func GetWholeParamStrict(command string, param string, value any, boundary int64) (int64, error) {
	whole, err := GetWholeNumberParam(value)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnexpectedType):
			if _, ok := value.(types.NullType); ok {
				return 0, nil
			}

			return 0, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				fmt.Sprintf(
					`BSON field '%s.%s' is the wrong type '%s', expected types '[long, int, decimal, double]'`,
					command, param, AliasFromType(value),
				),
				param,
			)
		case errors.Is(err, ErrNotWholeNumber):
			if math.Signbit(value.(float64)) {
				return 0, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrValueNegative,
					fmt.Sprintf(
						"BSON field '%s' value must be >= %d, actual value '%d'",
						param, boundary, int(math.Ceil(value.(float64))),
					),
					param,
				)
			}

			// for non-integer numbers, value is rounded to the greatest integer value less than the given value.
			return int64(math.Floor(value.(float64))), nil

		case errors.Is(err, ErrLongExceededPositive):
			return math.MaxInt64, nil

		case errors.Is(err, ErrLongExceededNegative):
			return 0, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrValueNegative,
				fmt.Sprintf(
					"BSON field '%s' value must be >= %d, actual value '%d'",
					param, boundary, int(math.Ceil(value.(float64))),
				),
				param,
			)

		default:
			return 0, lazyerrors.Error(err)
		}
	}

	if whole < boundary {
		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrValueNegative,
			fmt.Sprintf("BSON field '%s' value must be >= %d, actual value '%d'", param, boundary, whole),
			param,
		)
	}

	return whole, nil
}

// getOptionalPositiveNumber returns doc's value for key or protocol error for invalid parameter.
func getOptionalPositiveNumber(key string, value any) (int64, error) {
	whole, err := GetWholeNumberParam(value)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnexpectedType):
			return 0, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf("%s must be a number", key),
				key,
			)
		case errors.Is(err, ErrNotWholeNumber):
			if _, ok := value.(float64); ok {
				return 0, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrBadValue,
					fmt.Sprintf("%v has non-integral value", key),
					key,
				)
			}

			return 0, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf("%s must be a whole number", key),
				key,
			)
		default:
			return 0, lazyerrors.Error(err)
		}
	}

	if whole > math.MaxInt32 || whole < math.MinInt32 {
		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("%v value for %s is out of range", whole, key),
			key,
		)
	}

	if whole < 0 {
		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("%v value for %s is out of range", value, key),
			key,
		)
	}

	return whole, nil
}

// GetBoolOptionalParam returns bool value of v.
// Non-zero double, long, and int values return true.
// Zero values for those types, as well as nulls and missing fields, return false.
// Other types return a protocol error.
func GetBoolOptionalParam(key string, v any) (bool, error) {
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

		return false, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrTypeMismatch, msg, key)
	}
}
