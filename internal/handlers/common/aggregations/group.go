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
	"fmt"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"strings"
	"time"
)

// newAccumFunc is a type for a function that creates a new group accumulationFields.
type newAccumFunc func(field string, expression *types.Document) (Accumulator, error)

// Accumulator is a common interface for accumulation.
type Accumulator interface {
	// Accumulate applies an accumulation on group of documents.
	Accumulate(ctx context.Context, grouped []*types.Document) (any, error)
}

// accumulators maps all supported group accumulators.
var accumulators = map[string]newAccumFunc{
	// sorted alphabetically
	"$count": newCountGroupAccumulator,
}

// group represents $group stage.
//
//	{ $group: {
//			_id: groupKey,
//			accumulationFields[0].field: {accumulationFields[0].accumulator: accumulationFields[0].expression},
//	}}
type group struct {
	// groupKey is the output _id of the document.
	groupKey           any
	accumulationFields []accumulationFields
}

// accumulationFields represents accumulationFields of a group.
type accumulationFields struct {
	field       string
	accumulator Accumulator
	expression  any
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

	var group group
	iter := fields.Iterator()
	defer iter.Close()

	for {
		field, v, err := iter.Next()
		if err == iterator.ErrIteratorDone {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if field == "_id" {
			group.groupKey = v
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

		accumulationOperator, accumulationExpression, err := accumulation.Iterator().Next()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		newAccumulator, ok := accumulators[accumulationOperator]
		if !ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"unimplemented",
				"$group",
			)
		}

		accumulator, err := newAccumulator(field, accumulation)
		if err != nil {
			return nil, err
		}

		a := accumulationFields{
			field:       field,
			accumulator: accumulator,
			expression:  accumulationExpression,
		}

		group.accumulationFields = append(group.accumulationFields, a)
	}

	if group.groupKey == nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupMissingID,
			"a group specification must include an _id",
			"$group",
		)
	}

	return &group, nil
}

// Process implements Stage interface.
func (g *group) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	var res []*types.Document
	var distinct map[any]groupedDocuments

	// use groupKey to group them
	switch groupKey := g.groupKey.(type) {
	case string:
		if !strings.HasPrefix(groupKey, "$") {
			res = append(res, must.NotFail(types.NewDocument("_id", groupKey)))
			distinct = map[any]groupedDocuments{
				types.FormatAnyValue(groupKey): {
					groupKey:  groupKey,
					documents: in,
				}}
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

		// find distinct fields
		var err error
		distinct, err = groupDocumentsByKey(in, key)
		if err != nil {
			return nil, err
		}

		if len(distinct) == 0 {
			// return {"_id": null} when no distinct is found
			res = append(res, must.NotFail(types.NewDocument("_id", types.Null)))
			distinct = map[any]groupedDocuments{
				types.FormatAnyValue(types.Null): {
					groupKey:  types.Null,
					documents: in,
				}}
			break
		}

		for _, groupV := range distinct {
			res = append(res, must.NotFail(types.NewDocument("_id", groupV.groupKey)))
		}

	case *types.Document,
		*types.Array,
		float64,
		types.Binary,
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
		distinct = map[any]groupedDocuments{
			types.FormatAnyValue(groupKey): {
				groupKey:  groupKey,
				documents: in,
			}}

	default:
		panic(fmt.Sprintf("group: unexpected type %[1]T (%#[1]v)", groupKey))
	}

	for _, accumulation := range g.accumulationFields {
		for _, r := range res {
			accumulatedGroup, err := r.Get("_id")
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			groupedDocs, ok := distinct[types.FormatAnyValue(accumulatedGroup)]
			if !ok {
				// cannot fail because it was built for each ID
				return nil, lazyerrors.Error(err)
			}

			out, err := accumulation.accumulator.Accumulate(ctx, groupedDocs.documents)
			if err != nil {
				return nil, err
			}
			r.Set(accumulation.field, out)
		}
	}

	return res, nil
}

// groupedDocuments contains group key and the documents for that group.
type groupedDocuments struct {
	documents []*types.Document
	groupKey  any
}

// groupDocumentsByKey returns group key and documents value pair.
// The key is formatted as string to allow values such as Binary to be the key of the map.
// If the key is not found in the document, the document is ignored.
func groupDocumentsByKey(docs []*types.Document, key string) (map[any]groupedDocuments, error) {
	group := map[any]groupedDocuments{}

	for _, doc := range docs {
		path, err := types.NewPathFromString(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		v, err := doc.GetByPath(path)
		if err != nil {
			formattedGroupKey := types.FormatAnyValue(types.Null)
			v, ok := group[formattedGroupKey]
			if !ok {
				group[formattedGroupKey] = groupedDocuments{
					groupKey:  types.Null,
					documents: []*types.Document{},
				}
			}
			v.documents = append(v.documents, doc)
			continue
		}

		switch val := v.(type) {
		default:
			formattedGroupKey := types.FormatAnyValue(val)
			v, ok := group[formattedGroupKey]
			if !ok {
				group[formattedGroupKey] = groupedDocuments{
					groupKey:  val,
					documents: []*types.Document{},
				}
			}
			v.documents = append(v.documents, doc)
		}
	}

	return group, nil
}

// check interfaces
var (
	_ Stage = (*group)(nil)
)
