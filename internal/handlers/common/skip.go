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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// GetSkipParam validates the given skip value for find and count commands.
//
// If the value is valid, it returns its int64 representation,
// otherwise it returns a command error with the given command being mentioned.
func GetSkipParam(command string, value any) (int64, error) {
	skipValue, err := GetWholeNumberParam(value)

	if err == nil {
		if skipValue < 0 {
			return 0, commonerrors.NewCommandError(
				commonerrors.ErrValueNegative,
				fmt.Errorf("BSON field 'skip' value must be >= 0, actual value '%d'", skipValue),
			)
		}

		return skipValue, nil
	}

	switch {
	case errors.Is(err, errUnexpectedType):
		if _, ok := value.(types.NullType); ok {
			return 0, nil
		}

		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			fmt.Sprintf(
				`BSON field '%s.skip' is the wrong type '%s', expected types '[long, int, decimal, double]'`,
				command, AliasFromType(value),
			),
			"skip",
		)
	case errors.Is(err, errNotWholeNumber):
		if math.Signbit(value.(float64)) {
			return 0, commonerrors.NewCommandError(
				commonerrors.ErrValueNegative,
				fmt.Errorf("BSON field 'skip' value must be >= 0, actual value '%d'", int(math.Ceil(value.(float64)))),
			)
		}

		// for non-integer numbers, skip value is rounded to the greatest integer value less than the given value.
		return int64(math.Floor(value.(float64))), nil
	default:
		return 0, lazyerrors.Error(err)
	}
}

// SkipDocuments returns a subslice of given documents according to the given skip value.
func SkipDocuments(docs []*types.Document, skip int64) ([]*types.Document, error) {
	switch {
	case skip == 0:
		return docs, nil
	case skip > 0:
		if int64(len(docs)) < skip {
			return []*types.Document{}, nil
		}

		return docs[skip:], nil
	default:
		return nil, lazyerrors.Errorf("unexpected skip value: %d", skip)
	}
}
