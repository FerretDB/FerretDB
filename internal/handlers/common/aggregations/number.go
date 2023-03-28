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

package aggregations

import (
	"math"
	"math/big"
)

// sumNumbers accumulate numbers and returns the result of summation.
// The result has the same type as the input, except when the result
// cannot be presented accurately. Then int32 is converted to int64,
// and int64 is converted to float64.
// It ignores non-number values.
// This should only be used for aggregation, aggregation does not return
// error on overflow.
func sumNumbers(vs ...any) any {
	// use big.Int to accumulate values larger than math.MaxInt64.
	sumInt := big.NewInt(0)

	// TODO: handle accumulation of doubles close to max precision.
	// https://github.com/FerretDB/FerretDB/issues/2300
	var sumFloat float64

	var hasFloat64, hasInt64 bool

	for _, v := range vs {
		switch v := v.(type) {
		case float64:
			hasFloat64 = true

			sumFloat = sumFloat + v
		case int32:
			sumInt.Add(sumInt, big.NewInt(int64(v)))
		case int64:
			hasInt64 = true

			sumInt.Add(sumInt, big.NewInt(v))
		default:
			// ignore non-number
		}
	}

	if hasFloat64 {
		// return float64
		return float64(sumInt.Int64()) + sumFloat
	}

	if !sumInt.IsInt64() {
		// int64 is bigger than maximum of int64, convert to float64.
		bigFloat := new(big.Float).SetInt(sumInt)

		// ignore accuracy because there is no rounding from int64.
		res, _ := bigFloat.Float64()

		return res
	}

	res := sumInt.Int64()

	if !hasInt64 && res <= math.MaxInt32 && res >= math.MinInt32 {
		// convert to int32 if input has no int64 and can be represented in int32.
		return int32(res)
	}

	// return int64
	return res
}
