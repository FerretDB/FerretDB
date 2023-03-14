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

package aggregations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// newAccumulatorFunc is a type for a function that creates an accumulator.
type newAccumulatorFunc func(expression *types.Document) (Accumulator, error)

// Accumulator is a common interface for accumulation.
type Accumulator interface {
	// Accumulate documents and returns the result of accumulation.
	Accumulate(ctx context.Context, in []*types.Document) (any, error)
}

// accumulators maps all supported $group accumulators.
var accumulators = map[string]newAccumulatorFunc{
	// sorted alphabetically
	"$count": newCountAccumulator,
}

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
	accumulate  func(ctx context.Context, in []*types.Document) (any, error)
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

			idAccumulator := idAccumulator{
				expression: v,
			}
			groups = append(groups, groupBy{
				outputField: field,
				accumulate:  idAccumulator.Accumulate,
			})

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

		operator, _, err := accumulation.Iterator().Next()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		newAccumulator, ok := accumulators[operator]
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
		doc := new(types.Document)

		for _, accumulation := range g.groupBy {
			out, err := accumulation.accumulate(ctx, groupedDocument.documents)
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

// idAccumulator accumulates _id output field.
type idAccumulator struct {
	expression any
}

// Accumulate implements Accumulator interface.
//
//	{$group: {_id: "$v"}} sets the value of $v to _id, Accumulate returns the value found at key `v`.
//	{$group: {_id: null}} sets `null` to _id, Accumulate returns nil.
func (a *idAccumulator) Accumulate(ctx context.Context, in []*types.Document) (any, error) {
	groupKey, ok := a.expression.(string)
	if !ok {
		return a.expression, nil
	}

	if !strings.HasPrefix(groupKey, "$") {
		return groupKey, nil
	}

	key := strings.TrimPrefix(groupKey, "$")

	path, err := types.NewPathFromString(key)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// use the first element, it was already grouped by the groupKey,
	// so all `in` documents contain the same v.
	v, err := in[0].GetByPath(path)
	if err != nil {
		return types.Null, nil
	}

	return v, nil
}

// groupDocuments groups documents by group expression.
func (g *groupStage) groupDocuments(ctx context.Context, in []*types.Document) ([]groupedDocuments, error) {
	groupKey, ok := g.groupExpression.(string)
	if !ok {
		// non-string key aggregates values of all `in` documents into one aggregated document.
		return []groupedDocuments{{
			groupKey:  groupKey,
			documents: in,
		}}, nil
	}

	if !strings.HasPrefix(groupKey, "$") {
		// constant value aggregates values of all `in` documents into one aggregated document.
		return []groupedDocuments{{
			groupKey:  groupKey,
			documents: in,
		}}, nil
	}

	key := strings.TrimPrefix(groupKey, "$")
	if key == "" {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrGroupInvalidFieldPath,
			"'$' by itself is not a valid FieldPath",
			"$group (stage)",
		)
	}

	var group groupMap

	for _, doc := range in {
		path, err := types.NewPathFromString(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		v, err := doc.GetByPath(path)
		if err != nil {
			// if the path does not exist, use null for group key.
			group.addOrAppend(types.Null, doc)
			continue
		}

		group.addOrAppend(v, doc)
	}

	return group.docs, nil
}

// groupedDocuments contains group key and the documents for that group.
type groupedDocuments struct {
	groupKey  any
	documents []*types.Document
}

// groupMap holds groups of documents.
type groupMap struct {
	docs []groupedDocuments
}

// addOrAppend adds a groupKey documents pair if the groupKey does not exist,
// if the groupKey exists it appends the documents to the slice.
func (m *groupMap) addOrAppend(groupKey any, docs ...*types.Document) {
	for i, g := range m.docs {
		// groupKey is a distinct key and can be any BSON type including array and Binary,
		// so we cannot use structure like map.
		// Compare is used to check if groupKey exists in groupMap, because
		// numbers are grouped for the same value regardless of their number type.
		if types.Compare(groupKey, g.groupKey) == types.Equal {
			m.docs[i].documents = append(m.docs[i].documents, docs...)
			return
		}
	}

	m.docs = append(m.docs, groupedDocuments{
		groupKey:  groupKey,
		documents: docs,
	})
}

// check interfaces
var (
	_ Stage       = (*groupStage)(nil)
	_ Accumulator = (*countAccumulator)(nil)
)
