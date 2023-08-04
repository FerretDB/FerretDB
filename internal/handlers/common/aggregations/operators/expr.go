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

// expr represents `$expr` operator.
type expr struct {
	value any
}

// NewExpr validates and creates $expr operator.
func NewExpr(doc *types.Document) (Operator, error) {
	v := must.NotFail(doc.Get("$expr"))
	if err := validateExpr(v); err != nil {
		return nil, err
	}

	return &expr{
		value: v,
	}, nil
}

// Process implements Operator interface.
func (e *expr) Process(doc *types.Document) (any, error) {
	v, ok := e.value.(*types.Document)
	if !ok {
		return e.value, nil
	}

	return evaluateExpr(v, doc)
}

// validateExpr recursively validates operators and expressions.
func validateExpr(exprValue any) error {
	switch exprValue := exprValue.(type) {
	case *types.Document:
		_, isAggregationOperator := Operators[exprValue.Command()]

		if IsOperator(exprValue) && isAggregationOperator {
			op, err := NewOperator(exprValue)
			if err != nil {
				return err
			}

			_, err = op.Process(nil)
			if err != nil {
				// TODO https://github.com/FerretDB/FerretDB/issues/3129
				return err
			}

			return nil
		}

		iter := exprValue.Iterator()
		defer iter.Close()

		for {
			_, v, err := iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return lazyerrors.Error(err)
			}

			if err = validateExpr(v); err != nil {
				return err
			}
		}
	case *types.Array:
		iter := exprValue.Iterator()
		defer iter.Close()

		for {
			_, v, err := iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return lazyerrors.Error(err)
			}

			if err = validateExpr(v); err != nil {
				return err
			}
		}
	case string:
		_, err := aggregations.NewExpression(exprValue)
		var exprErr *aggregations.ExpressionError

		if errors.As(err, &exprErr) && exprErr.Code() == aggregations.ErrNotExpression {
			err = nil
		}

		if err != nil {
			return err
		}
	default:
	}

	return nil
}

// evaluateExpr recursively evaluates document's field expressions and operators.
func evaluateExpr(exprValue any, doc *types.Document) (any, error) {
	switch exprValue := exprValue.(type) {
	case *types.Document:
		if IsOperator(exprValue) {
			op, err := NewOperator(exprValue)
			if err != nil {
				return nil, err
			}

			v, err := op.Process(doc)
			if err != nil {
				return nil, err
			}

			return v, nil
		}

		iter := exprValue.Iterator()
		defer iter.Close()

		res := new(types.Document)

		for {
			k, v, err := iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			eval, err := evaluateExpr(v, doc)
			if err != nil {
				return nil, err
			}

			res.Set(k, eval)
		}

		return res, nil
	case *types.Array:
		iter := doc.Iterator()
		defer iter.Close()

		res := types.MakeArray(exprValue.Len())

		for {
			_, v, err := iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			eval, err := evaluateExpr(v, doc)
			if err != nil {
				return nil, err
			}

			res.Append(eval)
		}

		return res, nil
	case string:
		expression, err := aggregations.NewExpression(exprValue)

		var exprErr *aggregations.ExpressionError
		if errors.As(err, &exprErr) && exprErr.Code() == aggregations.ErrNotExpression {
			return exprValue, nil
		}

		if err != nil {
			// expression error was validated in NewExpr.
			return nil, lazyerrors.Error(err)
		}

		return expression.Evaluate(doc)
	default:
		return exprValue, nil
	}
}

// check interfaces
var (
	_ Operator = (*expr)(nil)
)
