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
	"errors"
	"fmt"
	"math"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// GetScaleParam validates the given scale value for collStats and dbStats command.
//
// If the value is valid, it returns its int32 representation,
// otherwise it returns a command error with the given command being mentioned.
func GetScaleParam(command string, value any) (int32, error) {
	scaleValue, err := commonparams.GetWholeNumberParam(value)

	if err == nil {
		if scaleValue <= 0 {
			return 0, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrValueNegative,
				fmt.Sprintf("BSON field 'scale' value must be >= 1, actual value '%d'", scaleValue),
				"scale",
			)
		}

		if scaleValue > math.MaxInt32 {
			return math.MaxInt32, nil
		}

		return int32(scaleValue), nil
	}

	switch {
	case errors.Is(err, commonparams.ErrUnexpectedType):
		if _, ok := value.(types.NullType); ok {
			return 1, nil
		}

		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			fmt.Sprintf(
				`BSON field '%s.scale' is the wrong type '%s', expected types '[long, int, decimal, double]'`,
				command, commonparams.AliasFromType(value),
			),
			"scale",
		)
	case errors.Is(err, commonparams.ErrNotWholeNumber):
		if math.Signbit(value.(float64)) {
			return 0, commonerrors.NewCommandError(
				commonerrors.ErrValueNegative,
				fmt.Errorf("BSON field 'scale' value must be >= 1, actual value '%d'", int(math.Ceil(value.(float64)))),
			)
		}

		// for non-integer numbers, scale value is rounded to the greatest integer value less than the given value.
		return int32(math.Floor(value.(float64))), nil

	case errors.Is(err, commonparams.ErrLongExceededPositive):
		return math.MaxInt32, nil

	case errors.Is(err, commonparams.ErrLongExceededNegative):
		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrValueNegative,
			fmt.Sprintf("BSON field 'scale' value must be >= 1, actual value '%d'", math.MinInt32),
			"scale",
		)

	default:
		return 0, lazyerrors.Error(err)
	}
}
