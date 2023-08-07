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

package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/operators"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// FilterIterator returns an iterator that filters out documents that don't match the filter.
// It will be added to the given closer.
//
// Next method returns the next document that matches the filter.
//
// Close method closes the underlying iterator.
func FilterIterator(iter types.DocumentsIterator, closer *iterator.MultiCloser, filter *types.Document) types.DocumentsIterator {
	res := &filterIterator{
		iter:   iter,
		filter: filter,
	}
	closer.Add(res)

	return res
}

// filterIterator is returned by FilterIterator.
type filterIterator struct {
	iter   types.DocumentsIterator
	filter *types.Document
}

// Next implements iterator.Interface. See FilterIterator for details.
func (iter *filterIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	for {
		_, doc, err := iter.iter.Next()
		if err != nil {
			return unused, nil, lazyerrors.Error(err)
		}

		var matches bool
		if iter.filter.Has("$expr") {
			matches, err = evaluateExpr(doc, iter.filter)
			if err != nil {
				return unused, nil, lazyerrors.Error(err)
			}

			if matches {
				return unused, doc, nil
			}

			continue
		}

		matches, err = FilterDocument(doc, iter.filter)
		if err != nil {
			return unused, nil, lazyerrors.Error(err)
		}

		if matches {
			return unused, doc, nil
		}
	}
}

// Close implements iterator.Interface. See FilterIterator for details.
func (iter *filterIterator) Close() {
	iter.iter.Close()
}

// evaluateExpr evaluates `$expr` operator and returns boolean indicating filter match.
func evaluateExpr(doc, filter *types.Document) (bool, error) {
	op, err := operators.NewExpr(filter)
	if err != nil {
		return false, processExprError(err)
	}

	v, err := op.Process(doc)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	switch v := v.(type) {
	case *types.Document:
		return true, nil
	case *types.Array:
		return true, nil
	case float64:
		if res := types.Compare(v, float64(0)); res == types.Equal {
			return false, nil
		}
	case string:
		return true, nil
	case types.Binary:
		return true, nil
	case types.ObjectID:
		return true, nil
	case bool:
		if res := types.Compare(v, false); res == types.Equal {
			return false, nil
		}
	case time.Time:
		return true, nil
	case types.NullType:
		return false, nil
	case types.Regex:
		return true, nil
	case int32:
		if res := types.Compare(v, int32(0)); res == types.Equal {
			return false, nil
		}
	case types.Timestamp:
		return true, nil
	case int64:
		if res := types.Compare(v, int64(0)); res == types.Equal {
			return false, nil
		}
	default:
		panic(fmt.Sprintf("common.evaluateExpr: unexpected type %[1]T (%#[1]v)", v))
	}

	return true, nil
}

// processExprError takes internal error related to operator evaluation and
// expression evaluation and returns CommandError.
func processExprError(err error) error {
	var opErr operators.OperatorError
	var exErr *aggregations.ExpressionError

	switch {
	case errors.As(err, &opErr):
		switch opErr.Code() {
		case operators.ErrTooManyFields:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrExpressionWrongLenOfFields,
				"An object representing an expression must have exactly one field",
				"$expr",
			)
		case operators.ErrNotImplemented:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"Invalid $match :: caused by :: "+opErr.Error(),
				"$expr",
			)
		case operators.ErrArgsInvalidLen:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrOperatorWrongLenOfArgs,
				opErr.Error(),
				"$expr",
			)
		case operators.ErrInvalidExpression:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidPipelineOperator,
				fmt.Sprintf("Unrecognized expression '%s'", opErr.Name()),
				"$expr",
			)
		case operators.ErrInvalidNestedExpression:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidPipelineOperator,
				opErr.Error(),
				"$expr",
			)
		}

	case errors.As(err, &exErr):
		switch exErr.Code() {
		case aggregations.ErrInvalidExpression:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				fmt.Sprintf("'%s' starts with an invalid character for a user variable name", exErr.Name()),
				"$expr",
			)
		case aggregations.ErrEmptyFieldPath:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageInvalidFieldPath,
				"'$' by itself is not a valid FieldPath",
				"$expr",
			)
		case aggregations.ErrUndefinedVariable:
			// TODO https://github.com/FerretDB/FerretDB/issues/2275
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"Aggregation expression variables are not implemented yet",
				"$expr",
			)
		case aggregations.ErrEmptyVariable:
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				"empty variable names are not allowed",
				"$expr",
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
	_ types.DocumentsIterator = (*filterIterator)(nil)
)
