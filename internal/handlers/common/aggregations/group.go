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
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// newAccumFunc is a type for a function that creates a new group accumulationFields.
type newAccumFunc func(expression *types.Document) (Accumulator, error)

// Accumulator is a common interface for accumulation.
type Accumulator interface {
	// Accumulate documents and returns the result of accumulation.
	Accumulate(ctx context.Context, in []*types.Document) (any, error)
}

// accumulators maps all supported $group accumulators.
var accumulators = map[string]newAccumFunc{
	// sorted alphabetically
	"$count": newCountAccumulator,
}

// group represents $group stage.
//
//	{ $group: {
//			_id: groupKey,
//			accumulationFields[0].field: {accumulationFields[0].accumulator: accumulationFields[0].expression},
//	}}
type group struct {
	groupKey           any
	accumulationFields []accumulationFields
	// groupKey is the output _id of the document.
}

// accumulationFields represents accumulation to apply on the group.
type accumulationFields struct {
	expression any
	accumulate func(ctx context.Context, in []*types.Document) (any, error)
	field      string
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

	var g group
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
			g.groupKey = v
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

		a := accumulationFields{
			field:      field,
			accumulate: accumulator.Accumulate,
			expression: expression,
		}

		g.accumulationFields = append(g.accumulationFields, a)
	}

	if g.groupKey == nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupMissingID,
			"a group specification must include an _id",
			"$group",
		)
	}

	return &g, nil
}

type allGroups struct {
	docs []groupedDocuments
}

// add the docs to the groupKey.
func (gs *allGroups) add(groupKey any, docs ...*types.Document) {
	for i, g := range gs.docs {
		if types.Compare(groupKey, g.groupKey) == types.Equal {
			gs.docs[i].documents = append(gs.docs[i].documents, docs...)
			return
		}
	}

	gs.docs = append(gs.docs, groupedDocuments{
		groupKey:  groupKey,
		documents: docs,
	})
}

// getByGroupKey gets grouped documents for the given group key.
func (gs *allGroups) getByGroupKey(groupKey any) groupedDocuments {
	for _, g := range gs.docs {
		if types.Compare(groupKey, g.groupKey) == types.Equal {
			return g
		}
	}

	panic("must not fail")
}

// Process implements Stage interface.
func (g *group) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	gs := new(allGroups)
	var res []*types.Document

	switch groupKey := g.groupKey.(type) {
	case *types.Document,
		*types.Array,
		float64:
		// return a document that contains groupKey as _id.
		res = append(res, must.NotFail(types.NewDocument("_id", groupKey)))
		gs.add(groupKey, in...)
	case string:
		if !strings.HasPrefix(groupKey, "$") {
			res = append(res, must.NotFail(types.NewDocument("_id", groupKey)))
			gs.add(groupKey, in...)

			break
		}

		key := strings.TrimPrefix(groupKey, "$")
		if key == "" {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrGroupInvalidFieldPath,
				"'$' by itself is not a valid FieldPath",
				"$group",
			)
		}

		// find groups fields
		var err error
		gs.docs, err = groupDocumentsByKey(in, key)

		if err != nil {
			return nil, err
		}

		if len(gs.docs) == 0 {
			// return {"_id": null} when no group is found
			res = append(res, must.NotFail(types.NewDocument("_id", types.Null)))
			gs.add(types.Null, in...)

			break
		}

		for _, groupV := range gs.docs {
			res = append(res, must.NotFail(types.NewDocument("_id", groupV.groupKey)))
		}

	case types.Binary,
		types.ObjectID,
		bool,
		time.Time,
		types.NullType,
		types.Regex,
		int32,
		types.Timestamp,
		int64:
		// return a document that contains groupKey as _id.
		res = append(res, must.NotFail(types.NewDocument("_id", groupKey)))
		gs.add(groupKey, in...)

	default:
		panic(fmt.Sprintf("group: unexpected type %[1]T (%#[1]v)", groupKey))
	}

	for _, accumulation := range g.accumulationFields {
		for _, r := range res {
			accumulatedGroup, err := r.Get("_id")
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			groupedDocs := gs.getByGroupKey(accumulatedGroup)

			out, err := accumulation.accumulate(ctx, groupedDocs.documents)
			if err != nil {
				return nil, err
			}

			if r.Has(accumulation.field) {
				// duplicate key
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrDuplicateField,
					fmt.Sprintf("duplicate field: %s", accumulation.field),
					"$count",
				)
			}

			r.Set(accumulation.field, out)
		}
	}

	return res, nil
}

// groupedDocuments contains group key and the documents for that group.
type groupedDocuments struct {
	groupKey  any
	documents []*types.Document
}

// groupDocumentsByKey returns group key and documents value pair.
// The key is formatted as string to allow values such as Binary to be the key of the map.
// If the key is not found in the document, the document is ignored.
func groupDocumentsByKey(docs []*types.Document, key string) ([]groupedDocuments, error) {
	var group allGroups

	for _, doc := range docs {
		path, err := types.NewPathFromString(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		v, err := doc.GetByPath(path)
		if err != nil {
			group.add(types.Null, doc)
			continue
		}

		group.add(v, doc)
	}

	return group.docs, nil
}

// check interfaces
var (
	_ Stage = (*group)(nil)
)
