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

// Package operators provides aggregation operators.
package operators

import (
	"errors"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// sum represents `$sum` operator.
type sum struct {
	expression *aggregations.Expression
	numbers    []any
}

// newSum returns `$sum` operator.
func newSum(operation *types.Document) (Operator, error) {
	expression := must.NotFail(operation.Get("$sum"))

	operator := new(sum)

	switch expr := expression.(type) {
	case *types.Array:
		operator.numbers = []any{expr}
	case float64:
		operator.numbers = []any{expr}
	case string:
		var err error
		if operator.expression, err = aggregations.NewExpression(expr); err != nil {
			// $sum returns 0 on non-existent field.
			operator.numbers = []any{int32(0)}
		}
	case int32, int64:
		operator.numbers = []any{expr}
	default:
		operator.numbers = []any{int32(0)}
		// $sum returns 0 on non-numeric field
	}

	return operator, nil
}

// Process implements Operator interface.
func (s *sum) Process(doc *types.Document) (any, error) {
	var numbers []any

	if s.expression != nil {
		value, err := s.expression.Evaluate(doc)
		if err != nil {
			// $sum returns 0 on non-existent field.
			return int32(0), nil
		}

		switch v := value.(type) {
		case *types.Array:
			iter := v.Iterator()
			defer iter.Close()

			for {
				_, v, err := iter.Next()
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}

				if err != nil {
					return nil, lazyerrors.Error(err)
				}

				numbers = append(numbers, v)
			}
		default:
			numbers = append(numbers, value)
		}
	}

	for _, number := range s.numbers {
		switch number := number.(type) {
		case float64, int32, int64:
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
	_ Operator = (*sum)(nil)
)
