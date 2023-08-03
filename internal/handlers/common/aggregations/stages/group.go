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

package stages

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/operators"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/operators/accumulators"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// group represents $group stage.
//
//	{ $group: {
//		_id: <groupExpression>,
//		<groupBy[0].outputField>: {accumulator0: expression0},
//		...
//		<groupBy[N].outputField>: {accumulatorN: expressionN},
//	}}
//
// $group uses group expression to group documents that have the same evaluated expression.
// The evaluated expression becomes the _id for that group of documents.
// For each group of documents, accumulators are applied.
type group struct {
	groupExpression any
	groupBy         []groupBy
}

// groupBy represents accumulation to apply on the group.
type groupBy struct {
	accumulate  func(iter types.DocumentsIterator) (any, error)
	outputField string
}

// newGroup creates a new $group stage.
func newGroup(stage *types.Document) (aggregations.Stage, error) {
	fields, err := common.GetRequiredParam[*types.Document](stage, "$group")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupInvalidFields,
			"a group's fields must be specified in an object",
			"$group (stage)",
		)
	}

	var groupKey any
	var groups []groupBy

	iter := fields.Iterator()

	defer iter.Close()

	for {
		field, v, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if field == "_id" {
			if err = validateGroupKey(v); err != nil {
				return nil, err
			}

			groupKey = v
			continue
		}

		accumulator, err := accumulators.NewAccumulator("$group", field, v)
		if err != nil {
			return nil, processOperatorError(err)
		}

		groups = append(groups, groupBy{
			outputField: field,
			accumulate:  accumulator.Accumulate,
		})
	}

	if groupKey == nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupMissingID,
			"a group specification must include an _id",
			"$group (stage)",
		)
	}

	return &group{
		groupExpression: groupKey,
		groupBy:         groups,
	}, nil
}

// Process implements Stage interface.
func (g *group) Process(ctx context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error) { //nolint:lll // for readability
	// TODO https://github.com/FerretDB/FerretDB/issues/2863
	docs, err := iterator.ConsumeValues(iterator.Interface[struct{}, *types.Document](iter))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	groupedDocuments, err := g.groupDocuments(docs)
	if err != nil {
		return nil, err
	}

	var res []*types.Document

	for _, groupedDocument := range groupedDocuments {
		doc := must.NotFail(types.NewDocument("_id", groupedDocument.groupID))

		groupIter := iterator.Values(iterator.ForSlice(groupedDocument.documents))
		defer groupIter.Close()

		for _, accumulation := range g.groupBy {
			out, err := accumulation.accumulate(groupIter)
			if err != nil {
				return nil, processOperatorError(err)
			}

			if doc.Has(accumulation.outputField) {
				// document has duplicate key
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrDuplicateField,
					fmt.Sprintf("duplicate field: %s", accumulation.outputField),
					"$group (stage)",
				)
			}

			doc.Set(accumulation.outputField, out)
		}

		res = append(res, doc)
	}

	iter = iterator.Values(iterator.ForSlice(res))
	closer.Add(iter)

	return iter, nil
}

// validateGroupKey returns error on invalid group key.
// If group key is a document, it recursively validates operator and expression.
func validateGroupKey(groupKey any) error {
	doc, ok := groupKey.(*types.Document)
	if !ok {
		return nil
	}

	if operators.IsOperator(doc) {
		op, err := operators.NewOperator(doc)
		if err != nil {
			return processOperatorError(err)
		}

		_, err = op.Process(nil)
		if err != nil {
			// TODO https://github.com/FerretDB/FerretDB/issues/3129
			return processOperatorError(err)
		}

		return nil
	}

	iter := doc.Iterator()
	defer iter.Close()

	fields := make(map[string]struct{}, doc.Len())

	for {
		k, v, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		if _, ok := fields[k]; ok {
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrGroupDuplicateFieldName,
				fmt.Sprintf("duplicate field name specified in object literal: %s", types.FormatAnyValue(doc)),
				"$group (stage)",
			)
		}
		fields[k] = struct{}{}

		switch v := v.(type) {
		case *types.Document:
			return validateGroupKey(v)
		case string:
			_, err := aggregations.NewExpression(v)
			var exprErr *aggregations.ExpressionError

			if errors.As(err, &exprErr) && exprErr.Code() == aggregations.ErrNotExpression {
				err = nil
			}

			if err != nil {
				return processOperatorError(err)
			}
		}
	}

	return nil
}

// groupDocuments groups documents by group expression.
func (g *group) groupDocuments(in []*types.Document) ([]groupedDocuments, error) {
	switch groupKey := g.groupExpression.(type) {
	case *types.Document:
		if operators.IsOperator(groupKey) {
			return groupByOperator(groupKey, in)
		}

		return groupByDocumentKey(groupKey, in)
	case *types.Array, float64, types.Binary, types.ObjectID, bool, time.Time, types.NullType,
		types.Regex, int32, types.Timestamp, int64:
		// non-string or document key aggregates values of all `in` documents into one aggregated document.

	case string:
		expression, err := aggregations.NewExpression(groupKey)
		if err != nil {
			var exprErr *aggregations.ExpressionError
			if !errors.As(err, &exprErr) {
				return nil, lazyerrors.Error(err)
			}

			switch exprErr.Code() {
			case aggregations.ErrNotExpression:
				// constant value aggregates values of all `in` documents into one aggregated document.
				return []groupedDocuments{{
					groupID:   groupKey,
					documents: in,
				}}, nil
			case aggregations.ErrEmptyFieldPath:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					// TODO
					commonerrors.ErrGroupInvalidFieldPath,
					"'$' by itself is not a valid Expression",
					"$group (stage)",
				)
			case aggregations.ErrInvalidExpression:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFailedToParse,
					fmt.Sprintf("'%s' starts with an invalid character for a user variable name", types.FormatAnyValue(groupKey)),
					"$group (stage)",
				)
			case aggregations.ErrEmptyVariable:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFailedToParse,
					"empty variable names are not allowed",
					"$group (stage)",
				)
			// TODO https://github.com/FerretDB/FerretDB/issues/2275
			case aggregations.ErrUndefinedVariable:
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrGroupUndefinedVariable,
					fmt.Sprintf("Use of undefined variable: %s", types.FormatAnyValue(groupKey)),
					"$group (stage)",
				)
			default:
				panic(fmt.Sprintf("unhandled field path error %s", exprErr.Error()))
			}
		}

		var group groupMap

		for _, doc := range in {
			val, err := expression.Evaluate(doc)
			if err != nil {
				// $group treats non-existent fields as nulls
				val = types.Null
			}

			group.addOrAppend(val, doc)
		}

		return group.docs, nil

	default:
		panic(fmt.Sprintf("unexpected type %[1]T (%#[1]v)", groupKey))
	}

	return []groupedDocuments{{
		groupID:   g.groupExpression,
		documents: in,
	}}, nil
}

// groupByOperator groups documents using operator found in group key.
func groupByOperator(groupKey *types.Document, docs []*types.Document) ([]groupedDocuments, error) {
	op, err := operators.NewOperator(groupKey)
	if err != nil {
		// operator error was validated in newGroup.
		return nil, lazyerrors.Error(err)
	}

	var m groupMap

	for _, doc := range docs {
		val, err := op.Process(doc)
		if err != nil {
			// existing operator errors are validated in newGroup
			return nil, processOperatorError(err)
		}

		m.addOrAppend(val, doc)
	}

	return m.docs, nil
}

// groupByDocumentKey groups documents by given group key. A group key could be an expression,
// that case it evaluates the expression and uses that as the group key.
func groupByDocumentKey(groupKey *types.Document, docs []*types.Document) ([]groupedDocuments, error) {
	var m groupMap

	for _, doc := range docs {
		iter := groupKey.Iterator()
		defer iter.Close()

		evaluatedGroupKey := new(types.Document)

		for {
			k, v, err := iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			switch v := v.(type) {
			case *types.Document:
				val, err := evaluateDocument(v, doc)
				if err != nil {
					evaluatedGroupKey.Set(k, new(types.Document))
					continue
				}

				evaluatedGroupKey.Set(k, val)

			case string:
				val, err := evaluateExpression(v, doc)
				if err != nil {
					if groupKey.Len() == 1 {
						// non-existent path is set to null only if _id contains one field
						evaluatedGroupKey.Set(k, types.Null)
					}

					continue
				}

				evaluatedGroupKey.Set(k, val)
			default:
				evaluatedGroupKey.Set(k, v)
			}
		}

		m.addOrAppend(evaluatedGroupKey, doc)
	}

	return m.docs, nil
}

// evaluateDocument recursively evaluates document's field expression and operator.
func evaluateDocument(expr, doc *types.Document) (any, error) {
	if operators.IsOperator(expr) {
		op, err := operators.NewOperator(expr)
		if err != nil {
			// operator error was validated in newGroup
			return nil, processOperatorError(err)
		}

		v, err := op.Process(doc)
		if err != nil {
			// existing operator errors are validated in newGroup
			return nil, processOperatorError(err)
		}

		return v, nil
	}

	iter := expr.Iterator()
	defer iter.Close()

	eval := new(types.Document)

	for {
		exprKey, exprVal, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		switch exprVal := exprVal.(type) {
		case *types.Document:
			val, err := evaluateDocument(exprVal, doc)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			eval.Set(exprKey, val)
		case string:
			val, err := evaluateExpression(exprVal, doc)
			if err != nil {
				if doc.Len() == 1 {
					// non-existent path is set to null if doc contains single field
					eval.Set(exprKey, types.Null)
				}

				return nil, err
			}

			eval.Set(exprKey, val)
		default:
			eval.Set(exprKey, exprVal)
		}
	}

	return eval, nil
}

// evaluateExpression evaluates string expression for the given document.
func evaluateExpression(expr string, doc *types.Document) (any, error) {
	expression, err := aggregations.NewExpression(expr)

	var exprErr *aggregations.ExpressionError
	if errors.As(err, &exprErr) && exprErr.Code() == aggregations.ErrNotExpression {
		return expr, nil
	}

	if err != nil {
		// expression error was validated in newGroup.
		return nil, lazyerrors.Error(err)
	}

	return expression.Evaluate(doc)
}

// groupedDocuments contains group key and the documents for that group.
type groupedDocuments struct {
	groupID   any
	documents []*types.Document
}

// groupMap holds groups of documents.
type groupMap struct {
	docs []groupedDocuments
}

// addOrAppend adds a groupID documents pair if the groupID does not exist,
// if the groupID exists it appends the documents to the slice.
func (m *groupMap) addOrAppend(groupKey any, docs ...*types.Document) {
	for i, g := range m.docs {
		// groupID is a distinct key and can be any BSON type including array and Binary,
		// so we cannot use structure like map.
		// Compare is used to check if groupID exists in groupMap, because
		// numbers are grouped for the same value regardless of their number type.
		if types.CompareForAggregation(groupKey, g.groupID) == types.Equal {
			m.docs[i].documents = append(m.docs[i].documents, docs...)
			return
		}
	}

	m.docs = append(m.docs, groupedDocuments{
		groupID:   groupKey,
		documents: docs,
	})
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

	return err
}

// check interfaces
var (
	_ aggregations.Stage = (*group)(nil)
)
