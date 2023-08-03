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
			return nil, processGroupStageError(err)
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
				return nil, processGroupStageError(err)
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
			return processGroupStageError(err)
		}

		_, err = op.Process(nil)
		if err != nil {
			// TODO https://github.com/FerretDB/FerretDB/issues/3129
			return processGroupStageError(err)
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
				return processGroupStageError(err)
			}
		}
	}

	return nil
}

// groupDocuments groups documents into groups using group key. If group key contains expressions
// or operators, they are evaluated before using it as the group key of documents.
func (g *group) groupDocuments(in []*types.Document) ([]groupedDocuments, error) {
	var m groupMap

	for _, doc := range in {
		switch groupKey := g.groupExpression.(type) {
		case *types.Document:
			val, err := evaluateDocument(groupKey, doc, false)
			if err != nil {
				// operator and expression errors are validated in newGroup
				return nil, lazyerrors.Error(err)
			}

			m.addOrAppend(val, doc)
		case *types.Array, float64, types.Binary, types.ObjectID, bool, time.Time, types.NullType,
			types.Regex, int32, types.Timestamp, int64:
			m.addOrAppend(groupKey, doc)
		case string:
			expression, err := aggregations.NewExpression(groupKey)
			if err != nil {
				var exprErr *aggregations.ExpressionError
				if errors.As(err, &exprErr) {
					if exprErr.Code() == aggregations.ErrNotExpression {
						m.addOrAppend(groupKey, doc)
						continue
					}

					return nil, processGroupStageError(err)
				}

				return nil, lazyerrors.Error(err)
			}

			val, err := expression.Evaluate(doc)
			if err != nil {
				// $group treats non-existent fields as nulls
				val = types.Null
			}

			m.addOrAppend(val, doc)
		default:
			panic(fmt.Sprintf("unexpected type %[1]T (%#[1]v)", groupKey))
		}
	}

	return m.docs, nil
}

// evaluateDocument recursively evaluates document's field expressions and operators.
func evaluateDocument(expr, doc *types.Document, nestedField bool) (any, error) {
	if operators.IsOperator(expr) {
		op, err := operators.NewOperator(expr)
		if err != nil {
			// operator error was validated in newGroup
			return nil, processGroupStageError(err)
		}

		v, err := op.Process(doc)
		if err != nil {
			// operator and expression errors are validated in newGroup
			return nil, processGroupStageError(err)
		}

		return v, nil
	}

	iter := expr.Iterator()
	defer iter.Close()

	evaluatedDocument := new(types.Document)

	for {
		k, exprVal, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		switch exprVal := exprVal.(type) {
		case *types.Document:
			v, err := evaluateDocument(exprVal, doc, true)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			evaluatedDocument.Set(k, v)
		case string:
			expression, err := aggregations.NewExpression(exprVal)

			var exprErr *aggregations.ExpressionError
			if errors.As(err, &exprErr) && exprErr.Code() == aggregations.ErrNotExpression {
				evaluatedDocument.Set(k, exprVal)
				continue
			}

			if err != nil {
				// expression error was validated in newGroup.
				return nil, lazyerrors.Error(err)
			}

			v, err := expression.Evaluate(doc)
			if err != nil {
				if expr.Len() == 1 && !nestedField {
					// non-existent path is set to null if expression contains single field and not a nested document
					evaluatedDocument.Set(k, types.Null)
				}

				continue
			}

			evaluatedDocument.Set(k, v)
		default:
			evaluatedDocument.Set(k, exprVal)
		}
	}

	return evaluatedDocument, nil
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

// processGroupError takes internal error related to operator evaluation and
// expression evaluation and returns CommandError that can be returned by $group
// aggregation stage.
func processGroupStageError(err error) error {
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
