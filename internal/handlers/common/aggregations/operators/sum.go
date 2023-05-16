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

package operators

import (
	"context"
	"math"
	"math/big"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// sum represents $sum aggregation operator.
type sum struct {
	expression aggregations.Expression
	number     any
}

// newSum creates a new $sum aggregation operator.
func newSum(accumulation *types.Document) (Accumulator, error) {
	expression := must.NotFail(accumulation.Get("$sum"))
	accumulator := new(sum)

	switch expr := expression.(type) {
	case *types.Array:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupUnaryOperator,
			"The $sum accumulator is a unary accumulator",
			"$sum (accumulator)",
		)
	case float64:
		accumulator.number = expr
	case string:
		var err error
		if accumulator.expression, err = aggregations.NewExpression(expr); err != nil {
			// $sum returns 0 on non-existent field.
			accumulator.number = int32(0)
		}
	case int32, int64:
		accumulator.number = expr
	default:
		accumulator.number = int32(0)
		// $sum returns 0 on non-numeric field
	}

	return accumulator, nil
}

// Accumulate implements Accumulator interface.
func (s *sum) Accumulate(ctx context.Context, groupID any, grouped []*types.Document) (any, error) {
	if s.expression != nil {
		var values []any

		for _, doc := range grouped {
			v := s.expression.Evaluate(doc)
			values = append(values, v)
		}

		return sumNumbers(values...), nil
	}

	switch number := s.number.(type) {
	case float64, int32, int64:
		// Below is equivalent of len(grouped)*number,
		// with conversion handling upon overflow of int32 and int64.
		// For example, { $sum: 1 } is equivalent of $count.
		numbers := make([]any, len(grouped))
		for i := 0; i < len(grouped); i++ {
			numbers[i] = number
		}

		return sumNumbers(numbers...), nil
	}

	// $sum returns 0 on non-existent and non-numeric field.
	return int32(0), nil
}

// sumNumbers accumulate numbers and returns the result of summation.
// The result has the same type as the input, except when the result
// cannot be presented accurately. Then int32 is converted to int64,
// and int64 is converted to float64. It ignores non-number values.
// This should only be used for aggregation, aggregation does not return
// error on overflow.
func sumNumbers(vs ...any) any {
	// use big.Int to accumulate values larger than math.MaxInt64.
	intSum := big.NewInt(0)

	// TODO: handle accumulation of doubles close to max precision.
	// https://github.com/FerretDB/FerretDB/issues/2300
	var floatSum float64

	var hasFloat64, hasInt64 bool

	for _, v := range vs {
		switch v := v.(type) {
		case float64:
			hasFloat64 = true

			floatSum = floatSum + v
		case int32:
			intSum.Add(intSum, big.NewInt(int64(v)))
		case int64:
			hasInt64 = true

			intSum.Add(intSum, big.NewInt(v))
		default:
			// ignore non-number
		}
	}

	// handle float64 or intSum bigger than the maximum of int64.
	if hasFloat64 || !intSum.IsInt64() {
		// ignore accuracy because there is no rounding from int64.
		intAsFloat, _ := new(big.Float).SetInt(intSum).Float64()

		return intAsFloat + floatSum
	}

	integer := intSum.Int64()

	// handle int32
	if !hasInt64 && integer <= math.MaxInt32 && integer >= math.MinInt32 {
		// convert to int32 if input has no int64 and can be represented in int32.
		return int32(integer)
	}

	// return int64
	return integer
}

// check interfaces
var (
	_ Accumulator = (*sum)(nil)
)
