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
	case *types.Document:
		if !operators.IsOperator(expr) {
			accumulator.number = int32(0)
			break
		}

		op, err := operators.NewOperator(expr)
		if err != nil {
			var opErr operators.OperatorError
			if !errors.As(err, &opErr) {
				return nil, lazyerrors.Error(err)
			}

			return nil, processOperatorError(opErr)
		}

		accumulator.operator = op
	case *types.Array:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupUnaryOperator,
			"The $sum accumulator is a unary operator",
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

		switch {
		case s.operator != nil:
			v, err := s.operator.Process(doc)
			if err != nil {
				// TODO
				return nil, processOperatorError(err)
			}

			numbers = append(numbers, v)

			continue

		case s.expression != nil:
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

// processOperatorError takes internal error related to operator evaluation and
// returns proper CommandError that can be returned by $group aggregation stage.
func processOperatorError(err error) error {
	var opErr operators.OperatorError
	var exErr *aggregations.ExpressionError

	switch {
	case errors.As(err, &opErr):
		switch opErr.Code() {
		case operators.ErrTooManyFields:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrExpressionWrongLenOfFields,
				"An object representing an expression must have exactly one field",
				"$group (stage)",
			)
		case operators.ErrNotImplemented:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"Invalid $group :: caused by :: "+opErr.Error(),
				"$group (stage)",
			)
		case operators.ErrArgsInvalidLen:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrOperatorWrongLenOfArgs,
				opErr.Error(),
				"$group (stage)",
			)
		case operators.ErrInvalidExpression:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidPipelineOperator,
				opErr.Error(),
				"$group (stage)",
			)
		case operators.ErrInvalidNestedExpression:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidPipelineOperator,
				opErr.Error(),
				"$group (stage)",
			)
		}

	case errors.As(err, &exErr):
		switch exErr.Code() {
		case aggregations.ErrNotExpression:
			// handled by upstream and this should not be reachable for existing expression implementation
			fallthrough
		case aggregations.ErrInvalidExpression:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				"'$' starts with an invalid character for a user variable name",
				"$group (stage)",
			)
		case aggregations.ErrEmptyFieldPath:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrGroupInvalidFieldPath,
				"'$' by itself is not a valid FieldPath",
				"$group (stage)",
			)
		case aggregations.ErrUndefinedVariable:
			// TODO https://github.com/FerretDB/FerretDB/issues/2275
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"Aggregation expression variables are not implemented yet",
				"$group (stage)",
			)
		case aggregations.ErrEmptyVariable:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				"empty variable names are not allowed",
				"$group (stage)",
			)
		}
	}

	return lazyerrors.Error(err)
}

// check interfaces
var (
	_ Accumulator = (*sum)(nil)
)
