package accumulators

import (
	"errors"

	"github.com/FerretDB/FerretDB/internal/handler/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handler/common/aggregations/operators"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// avg represents $avg aggregation operator.
type avg struct {
	expression *aggregations.Expression
	operator   operators.Operator
	number     any
}

// newAvg creates a new $avg aggregation operator.
func newAvg(args ...any) (Accumulator, error) {
	accumulator := new(avg)

	if len(args) != 1 {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrStageGroupUnaryOperator,
			"The $avg accumulator is a unary operator",
			"$avg (accumulator)",
		)
	}

	for _, arg := range args {
		switch arg := arg.(type) {
		case *types.Document:
			if !operators.IsOperator(arg) {
				accumulator.number = int32(0)
				break
			}

			op, err := operators.NewOperator(arg)
			if err != nil {
				var opErr operators.OperatorError
				if !errors.As(err, &opErr) {
					return nil, lazyerrors.Error(err)
				}

				return nil, opErr
			}

			accumulator.operator = op
		case float64:
			accumulator.number = arg
		case string:
			var err error
			if accumulator.expression, err = aggregations.NewExpression(arg, nil); err != nil {
				// $avg returns 0 on non-existent field.
				accumulator.number = int32(0)
			}
		case int32, int64:
			accumulator.number = arg
		default:
			accumulator.number = int32(0)
			// $avg returns 0 on non-numeric field
		}
	}

	return accumulator, nil
}

// Accumulate implements Accumulator interface.
func (s *avg) Accumulate(iter types.DocumentsIterator) (any, error) {
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
				return nil, err
			}

			numbers = append(numbers, v)

			continue

		case s.expression != nil:
			value, err := s.expression.Evaluate(doc)

			// avg fields that exist
			if err == nil {
				numbers = append(numbers, value)
			}

			continue
		}

		switch number := s.number.(type) {
		case float64, int32, int64:
			// For number types, the result is equivalent of iterator len*number,
			// with conversion handled upon overflow of int32 and int64.
			// For example, { $avg: 1 } is equivalent of { $count: { } }.
			numbers = append(numbers, number)
		default:
			// $avg returns 0 on non-existent and non-numeric field.
			return int32(0), nil
		}
	}

	return aggregations.AvgNumbers(numbers...), nil
}

// check interfaces
var (
	_ Accumulator = (*avg)(nil)
)