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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/operators"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// groupStage represents $group stage.
//
//	{ $group: {
//		_id: <groupExpression>,
//		<groupBy[0].outputField>: {accumulator0: expression0},
//		...
//		<groupBy[N].outputField>: {accumulatorN: expressionN},
//	}}
type groupStage struct {
	groupExpression any
	groupBy         []groupBy
}

// groupBy represents accumulation to apply on the group.
type groupBy struct {
	accumulate  func(ctx context.Context, groupID any, in []*types.Document) (any, error)
	outputField string
}

// newGroup creates a new $group stage.
func newGroup(stage *types.Document) (Stage, error) {
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
			groupKey = v
			continue
		}

		accumulation, ok := v.(*types.Document)
		if !ok || accumulation.Len() == 0 {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageGroupInvalidAccumulator,
				fmt.Sprintf("The field '%s' must be an accumulator object", field),
				"$group (stage)",
			)
		}

		// accumulation document contains only one field.
		if accumulation.Len() > 1 {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageGroupMultipleAccumulator,
				fmt.Sprintf("The field '%s' must specify one accumulator", field),
				"$group (stage)",
			)
		}

		operator := accumulation.Command()

		newAccumulator, ok := operators.GroupOperators[operator]
		if !ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("$group accumulator %q is not implemented yet", operator),
				operator+" (accumulator)",
			)
		}

		accumulator, err := newAccumulator(accumulation)
		if err != nil {
			return nil, err
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

	return &groupStage{
		groupExpression: groupKey,
		groupBy:         groups,
	}, nil
}

// Process implements Stage interface.
func (g *groupStage) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	groupedDocuments, err := g.groupDocuments(ctx, in)
	if err != nil {
		return nil, err
	}

	var res []*types.Document

	for _, groupedDocument := range groupedDocuments {
		doc := must.NotFail(types.NewDocument("_id", groupedDocument.groupID))

		for _, accumulation := range g.groupBy {
			out, err := accumulation.accumulate(ctx, groupedDocument.groupID, groupedDocument.documents)
			if err != nil {
				return nil, err
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

	return res, nil
}

// groupDocuments groups documents by group expression.
func (g *groupStage) groupDocuments(ctx context.Context, in []*types.Document) ([]groupedDocuments, error) {
	groupKey, ok := g.groupExpression.(string)
	if !ok {
		// non-string key aggregates values of all `in` documents into one aggregated document.
		return []groupedDocuments{{
			groupID:   g.groupExpression,
			documents: in,
		}}, nil
	}

	expression, err := types.NewExpression(groupKey)
	if err != nil {
		var fieldPathErr *types.FieldPathError
		if !errors.As(err, &fieldPathErr) {
			return nil, lazyerrors.Error(err)
		}

		switch fieldPathErr.Code() {
		case types.ErrNotFieldPath:
			// constant value aggregates values of all `in` documents into one aggregated document.
			return []groupedDocuments{{
				groupID:   groupKey,
				documents: in,
			}}, nil
		case types.ErrEmptyFieldPath:
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrGroupInvalidFieldPath,
				"'$' by itself is not a valid FieldPath",
				"$group (stage)",
			)
		case types.ErrInvalidFieldPath:
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				fmt.Sprintf("'%s' starts with an invalid character for a user variable name", types.FormatAnyValue(groupKey)),
				"$group (stage)",
			)
		case types.ErrEmptyVariable:
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				"empty variable names are not allowed",
				"$group (stage)",
			)
		case types.ErrUndefinedVariable:
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrGroupUndefinedVariable,
				fmt.Sprintf("Use of undefined variable: %s", types.FormatAnyValue(groupKey)),
				"$group (stage)",
			)
		default:
			panic(fmt.Sprintf("unhandled field path error %s", fieldPathErr.Error()))
		}
	}

	var group groupMap

	for _, doc := range in {
		val := expression.Evaluate(doc)
		group.addOrAppend(val, doc)
	}

	return group.docs, nil
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

// Type implements Stage interface.
func (g *groupStage) Type() StageType {
	return StageTypeDocuments
}

// check interfaces
var (
	_ Stage = (*groupStage)(nil)
)
