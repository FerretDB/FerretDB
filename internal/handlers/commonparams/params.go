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
	// ErrNegativeNumber is returned when a negative number is given.
	ErrNegativeNumber = fmt.Errorf("negative number")
	// ErrNotWholeNumber is returned when a non-whole number is given.
	ErrNotWholeNumber = fmt.Errorf("not a whole number")
	// ErrNotBinaryMask is returned when a non-binary mask is given.
	ErrNotBinaryMask = fmt.Errorf("not a binary mask")
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
func GetWholeParamStrict(command string, param string, value any) (int64, error) {
	whole, err := GetWholeNumberParam(value)

	if err == nil {
		if whole < 0 {
			return 0, commonerrors.NewCommandError(
				commonerrors.ErrValueNegative,
				fmt.Errorf("BSON field '%s' value must be >= 0, actual value '%d'", param, whole),
			)
		}

		return whole, nil
	}

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
			return 0, commonerrors.NewCommandError(
				commonerrors.ErrValueNegative,
				fmt.Errorf("BSON field '%s' value must be >= 0, actual value '%d'", param, int(math.Ceil(value.(float64)))),
			)
		}

		// for non-integer numbers, value is rounded to the greatest integer value less than the given value.
		return int64(math.Floor(value.(float64))), nil

	case errors.Is(err, ErrLongExceededPositive):
		return math.MaxInt64, nil

	case errors.Is(err, ErrLongExceededNegative):
		return 0, commonerrors.NewCommandError(
			commonerrors.ErrValueNegative,
			fmt.Errorf("BSON field '%s' value must be >= 0, actual value '%d'", param, int(math.Ceil(value.(float64)))),
		)

	default:
		return 0, lazyerrors.Error(err)
	}
}
