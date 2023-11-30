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

	"github.com/FerretDB/FerretDB/internal/handler/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handler/common/aggregations/operators"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// first represents the `$first` operator.
type first struct {
	expression *aggregations.Expression
	operator   operators.Operator
	value      any
}

// newFirst creates a new $count aggregation operator.
func newFirst(args ...any) (Accumulator, error) {
	accumulator := new(first)

	if len(args) != 1 {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrStageGroupUnaryOperator,
			"The $first accumulator is a unary operator",
			"$first (accumulator)",
		)
	}

	for _, arg := range args {
		switch arg := arg.(type) {
		case *types.Document:
			if !operators.IsOperator(arg) {
				accumulator.value = arg
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
		case string:
			var err error
			if accumulator.expression, err = aggregations.NewExpression(arg, nil); err != nil {
				accumulator.value = arg
			}
		default:
			accumulator.value = arg
		}
	}

	return accumulator, nil
}

// Accumulate implements Accumulator interface.
func (f *first) Accumulate(iter types.DocumentsIterator) (any, error) {
	var res any

	for {
		_, doc, err := iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		switch {
		case f.operator != nil:
			v, err := f.operator.Process(doc)
			if err != nil {
				return nil, err
			}

			res = v

			continue

		case f.expression != nil:
			value, err := f.expression.Evaluate(doc)
			if err != nil {
				return types.Null, nil
			}

			res = value

			continue
		default:
			res = f.value
		}
	}

	return res, nil
}

// check interfaces
var (
	_ Accumulator = (*first)(nil)
)
