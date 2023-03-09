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

// group represents $group stage.
//
//	{ $group: {
//		_id: <groupExpression>,
//		<groupBy[0].outputField>: {<groupBy[0].accumulator>: <groupBy[0].expression>},
//		...
//		<groupBy[N].outputField>: {<groupBy[N].accumulator>: <groupBy[N].expression>},
//	}}
type groupStage struct {
	groupExpression any
	groupBy         []groupBy
}

// groupBy represents accumulation to apply on the group.
type groupBy struct {
	expression  any
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
			"$group",
		)
	}

	if fields.Len() == 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupMissingID,
			"a group specification must include an _id",
			"$group",
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
				expression:  v,
			})

			continue
		}

		accumulation, ok := v.(*types.Document)
		if !ok || accumulation.Len() == 0 {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageGroupInvalidField,
				fmt.Sprintf("The field '%s' must be an accumulator object", field),
				"$group",
			)
		}

		// document contains only one.
		if accumulation.Len() > 1 {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageGroupOneAccumulator,
				fmt.Sprintf("The field '%s' must specify one accumulator", field),
				"$group",
			)
		}

		operator, expression, err := accumulation.Iterator().Next()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		newAccumulator, ok := accumulators[operator]
		if !ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"unimplemented",
				"$group",
			)
		}

		accumulator, err := newAccumulator(accumulation)
		if err != nil {
			return nil, err
		}

		groups = append(groups, groupBy{
			outputField: field,
			accumulate:  accumulator.Accumulate,
			expression:  expression,
		})
	}

	if groupKey == nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupMissingID,
			"a group specification must include an _id",
			"$group",
		)
	}

	return &groupStage{
		groupExpression: groupKey,
		groupBy:         groups,
	}, nil
}

// Process implements Stage interface.
func (g *groupStage) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	grouped, err := g.groupDocuments(ctx, in)
	if err != nil {
		return nil, err
	}

	var res []*types.Document

	for _, groupedDocument := range grouped {
		doc := new(types.Document)

		for _, accumulation := range g.groupBy {
			out, err := accumulation.accumulate(ctx, groupedDocument.documents)
			if err != nil {
				return nil, err
			}

			if doc.Has(accumulation.outputField) {
				// duplicate key
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrDuplicateField,
					fmt.Sprintf("duplicate outputField: %s", accumulation.outputField),
					"$group",
				)
			}

			doc.Set(accumulation.outputField, out)
		}

		res = append(res, doc)
	}

	return res, nil
}

// idAccumulator accumulates _id outputField.
type idAccumulator struct {
	expression any
}

// Accumulate implements Accumulator interface.
func (a *idAccumulator) Accumulate(ctx context.Context, in []*types.Document) (any, error) {
	if len(in) == 0 {
		return types.Null, nil
	}

	groupKey, ok := a.expression.(string)
	if !ok {
		return a.expression, nil
	}

	if !strings.HasPrefix(groupKey, "$") {
		return groupKey, nil
	}

	key := strings.TrimPrefix(groupKey, "$")
	if key == "" {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrGroupInvalidFieldPath,
			"'$' by itself is not a valid FieldPath",
			"$group",
		)
	}

	path, err := types.NewPathFromString(key)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

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
		return []groupedDocuments{{
			groupKey:  groupKey,
			documents: in,
		}}, nil
	}

	if !strings.HasPrefix(groupKey, "$") {
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
			"$group",
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
			// if the path does not exist, use null for grouping
			group.addOrAppend(types.Null, doc)
			continue
		}

		group.addOrAppend(v, doc)
	}

	if len(group.docs) == 0 {
		return []groupedDocuments{{
			groupKey:  types.Null,
			documents: in,
		}}, nil
	}

	return group.docs, nil
}

// groupedDocuments contains group key and the documents for that group.
type groupedDocuments struct {
	groupKey  any
	documents []*types.Document
}

// groupMap holds groups of documents using unique groupKey and slide of documents pair.
type groupMap struct {
	docs []groupedDocuments
}

// addOrAppend adds a groupKey documents pair if the groupKey does not exist,
// if the groupKey exists it appends the documents to the slice.
func (m *groupMap) addOrAppend(groupKey any, docs ...*types.Document) {
	for i, g := range m.docs {
		// Compare is used to check if the key exists in the group.
		// groupKey used as the key can be any BSON type including array and Binary,
		// so we cannot use structure like map.
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
	_ Stage = (*groupStage)(nil)
)
