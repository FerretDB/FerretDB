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
)

var (
	ErrNegativeNumber        = fmt.Errorf("negative number")
	ErrNotWholeNumber        = fmt.Errorf("not a whole number")
	ErrNotBinaryMask         = fmt.Errorf("not a binary mask")
	ErrUnexpectedLeftOpType  = fmt.Errorf("unexpected left operand type")
	ErrUnexpectedRightOpType = fmt.Errorf("unexpected right operand type")
	ErrLongExceededPositive  = fmt.Errorf("long exceeded - positive value")
	ErrLongExceededNegative  = fmt.Errorf("long exceeded - negative value")
	ErrIntExceeded           = fmt.Errorf("int exceeded")
	ErrInfinity              = fmt.Errorf("infinity")
	ErrUnexpectedType        = fmt.Errorf("unexpected type")
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
