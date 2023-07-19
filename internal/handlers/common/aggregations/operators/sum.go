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
	expressions   []*aggregations.Expression
	operatorExprs []*types.Document
	numbers       []any
	arrayLen      int
}

// newSum returns `$sum` operator.
func newSum(operation *types.Document) (Operator, error) {
	expression := must.NotFail(operation.Get("$sum"))
	operator := new(sum)

	switch expr := expression.(type) {
	case *types.Document:
		if IsOperator(expr) {
			operator.operatorExprs = []*types.Document{expr}
		}
	case *types.Array:
		operator.arrayLen = expr.Len()

		iter := expr.Iterator()
		defer iter.Close()

		for {
			_, v, err := iter.Next()

			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			switch elemExpr := v.(type) {
			case *types.Document:
				if IsOperator(elemExpr) {
					operator.operatorExprs = append(operator.operatorExprs, elemExpr)
				}
			case float64:
				operator.numbers = append(operator.numbers, elemExpr)
			case string:
				ex, err := aggregations.NewExpression(elemExpr)
				if err != nil {
					return nil, err
				}

				operator.expressions = append(operator.expressions, ex)
			case int32, int64:
				operator.numbers = append(operator.numbers, elemExpr)
			default:
				continue
			}
		}
	case float64:
		operator.numbers = []any{expr}
	case string:
		ex, err := aggregations.NewExpression(expr)
		if err != nil {
			return nil, err
		}

		operator.expressions = []*aggregations.Expression{ex}
	case int32, int64:
		operator.numbers = []any{expr}
	default:
	}

	return operator, nil
}

// Process implements Operator interface.
func (s *sum) Process(doc *types.Document) (any, error) {
	var numbers []any

	for _, expression := range s.expressions {
		value, err := expression.Evaluate(doc)
		if err != nil {
			// expression cannot evaluate doc, nothing to do
			continue
		}

		switch v := value.(type) {
		case *types.Array:
			if s.arrayLen > 1 {
				// This handles strange behaviour of MongoDB.
				// When $sum has more than one argument,
				// expression is ignored if evaluated path contains *array*.
				// Below case, `$sum` has two arguments, so "$v" is ignored.
				// `{$sum: ["$v", 1]}` and doc is `{v: [2, 3]}` => sum is `1`
				// Below case, `$sum` has one argument, "$v" is evaluated.
				// `{$sum: ["$v"]}` and doc is `{v: [2, 3]}` => sum is `5`
				continue
			}

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

	for _, operatorExpr := range s.operatorExprs {
		op, err := NewOperator(operatorExpr)
		if err != nil {
			return nil, err
		}

		v, err := op.Process(doc)
		if err != nil {
			return nil, err
		}

		numbers = append(numbers, v)
	}

	for _, number := range s.numbers {
		switch number := number.(type) {
		case float64, int32, int64:
			numbers = append(numbers, number)
		default:
		}
	}

	return aggregations.SumNumbers(numbers...), nil
}

// check interfaces
var (
	_ Operator = (*sum)(nil)
)
