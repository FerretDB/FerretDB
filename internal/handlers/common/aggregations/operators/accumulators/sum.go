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

package accumulators

import (
	"errors"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/operators"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// sum represents $sum aggregation operator.
type sum struct {
	expression *aggregations.Expression
	operator   operators.Operator
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
			"The $sum accumulator is a unary operator",
			"$sum (accumulator)",
		)
	case *types.Document:
		if !operators.IsOperator(expr) {
			//TODO
			accumulator.number = int32(0)
			break
		}

		op, err := operators.NewOperator(expr)
		if err != nil {
			var opErr operators.OperatorError
			if !errors.As(err, &opErr) {
				return nil, lazyerrors.Error(err)
			}

			return nil, opErr
		}

		accumulator.operator = op

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
func (s *sum) Accumulate(iter types.DocumentsIterator) (any, error) {
	var numbers []any

	for {
		_, doc, err := iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if s.operator != nil {
			v, err := s.operator.Process(doc)
			if err != nil {
				// TODO
				return nil, err
			}

			switch v.(type) {
			case float64, int32, int64:
				numbers = append(numbers, v)
			}
			continue
		}

		if s.expression != nil {
			value, err := s.expression.Evaluate(doc)

			// sum fields that exist
			if err == nil {
				numbers = append(numbers, value)
			}

			continue
		}

		switch number := s.number.(type) {
		case float64, int32, int64:
			// For number types, the result is equivalent of iterator len*number,
			// with conversion handled upon overflow of int32 and int64.
			// For example, { $sum: 1 } is equivalent of { $count: { } }.
			numbers = append(numbers, number)
		default:
			// $sum returns 0 on non-existent and non-numeric field.
			return int32(0), nil
		}
	}

	return aggregations.SumNumbers(numbers...), nil
}

// check interfaces
var (
	_ Accumulator = (*sum)(nil)
)
