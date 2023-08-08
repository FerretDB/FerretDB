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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// expr represents `$expr` operator.
type expr struct {
	exprValue   any
	errArgument string
}

// NewExpr validates and creates $expr operator.
//
// It returns CommandError for invalid value of $expr operator.
func NewExpr(exprValue *types.Document, errArgument string) (Operator, error) {
	v := must.NotFail(exprValue.Get("$expr"))
	e := &expr{
		exprValue:   v,
		errArgument: errArgument,
	}

	if err := e.validateExpr(v); err != nil {
		return nil, err
	}

	return e, nil
}

// Process implements Operator interface.
func (e *expr) Process(doc *types.Document) (any, error) {
	return e.processExpr(e.exprValue, doc)
}

// processExpr recursively validates operators and expressions.
// Each array values and document fields are validated recursively.
//
// It returns CommandError if any validation fails.
func (e *expr) validateExpr(exprValue any) error {
	switch exprValue := exprValue.(type) {
	case *types.Document:
		if IsOperator(exprValue) {
			op, err := NewOperator(exprValue)
			if err != nil {
				return processExprOperatorErrors(err, e.errArgument)
			}

			_, err = op.Process(nil)
			if err != nil {
				// TODO https://github.com/FerretDB/FerretDB/issues/3129
				return processExprOperatorErrors(err, e.errArgument)
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

			if err = e.validateExpr(v); err != nil {
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

			if err = e.validateExpr(v); err != nil {
				return err
			}
		}
	case string:
		_, err := aggregations.NewExpression(exprValue, nil)
		var exprErr *aggregations.ExpressionError

		if errors.As(err, &exprErr) && exprErr.Code() == aggregations.ErrNotExpression {
			err = nil
		}

		if err != nil {
			return processExprOperatorErrors(err, e.errArgument)
		}
	}

	return nil
}

// processExpr recursively processes operators and expressions and returns processed `exprValue`.
//
// Each array values and document fields are processed recursively.
// String expression is evaluated if any, an evaluation error due to missing field returns Null.
// Any value that does not require processing, it returns the original value.
func (e *expr) processExpr(exprValue any, doc *types.Document) (any, error) {
	switch exprValue := exprValue.(type) {
	case *types.Document:
		if IsOperator(exprValue) {
			op, err := NewOperator(exprValue)
			if err != nil {
				// $expr was validated in NewExpr
				return nil, lazyerrors.Error(err)
			}

			v, err := op.Process(doc)
			if err != nil {
				// Process does not return error for existing operators
				return nil, lazyerrors.Error(err)
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

			processed, err := e.processExpr(v, doc)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			res.Set(k, processed)
		}

		return res, nil
	case *types.Array:
		iter := exprValue.Iterator()
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

			processed, err := e.processExpr(v, doc)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			res.Append(processed)
		}

		return res, nil
	case string:
		expression, err := aggregations.NewExpression(exprValue, nil)

		var exprErr *aggregations.ExpressionError
		if errors.As(err, &exprErr) && exprErr.Code() == aggregations.ErrNotExpression {
			// not an expression, return the original value
			return exprValue, nil
		}

		if err != nil {
			// expression error was validated in NewExpr
			return nil, lazyerrors.Error(err)
		}

		v, err := expression.Evaluate(doc)
		if err != nil {
			// missing field is set to null
			return types.Null, nil
		}

		return v, nil
	default:
		// nothing to process, return the original value
		return exprValue, nil
	}
}

// ProcessMatchStageError takes internal error related to operator evaluation and
// expression evaluation and returns CommandError.
func processExprOperatorErrors(err error, argument string) error {
	var opErr OperatorError
	var exErr *aggregations.ExpressionError

	switch {
	case errors.As(err, &opErr):
		switch opErr.Code() {
		case ErrTooManyFields:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrExpressionWrongLenOfFields,
				"An object representing an expression must have exactly one field",
				argument,
			)
		case ErrNotImplemented:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"Invalid $match :: caused by :: "+opErr.Error(),
				argument,
			)
		case ErrArgsInvalidLen:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrOperatorWrongLenOfArgs,
				opErr.Error(),
				argument,
			)
		case ErrInvalidExpression:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidPipelineOperator,
				fmt.Sprintf("Unrecognized expression '%s'", opErr.Name()),
				argument,
			)
		case ErrInvalidNestedExpression:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidPipelineOperator,
				opErr.Error(),
				argument,
			)
		}

	case errors.As(err, &exErr):
		switch exErr.Code() {
		case aggregations.ErrInvalidExpression:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				fmt.Sprintf("'%s' starts with an invalid character for a user variable name", exErr.Name()),
				argument,
			)
		case aggregations.ErrEmptyFieldPath:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrGroupInvalidFieldPath,
				"'$' by itself is not a valid FieldPath",
				argument,
			)
		case aggregations.ErrUndefinedVariable:
			// TODO https://github.com/FerretDB/FerretDB/issues/2275
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"Aggregation expression variables are not implemented yet",
				argument,
			)
		case aggregations.ErrEmptyVariable:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				"empty variable names are not allowed",
				argument,
			)
		case aggregations.ErrNotExpression:
			// handled by upstream and this should not be reachable for existing expression implementation
			fallthrough
		default:
		}
	}

	return lazyerrors.Error(err)
}

// check interfaces
var (
	_ Operator = (*expr)(nil)
)
