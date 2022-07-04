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

package pg

import (
	"context"
	"fmt"
	"io"

	"github.com/FerretDB/FerretDB/internal/handlers/aggregate"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// $match
//
// Filters the document stream to allow only matching documents to pass
// unmodified into the next pipeline stage. $match uses standard MongoDB
// queries. For each input document, outputs either one document (a match) or
// zero documents (no match).

// $group
//
// Groups input documents by a specified identifier expression and applies the
// accumulator expression(s), if specified, to each group. Consumes all input
// documents and outputs one document per each distinct group. The output
// documents only contain the identifier field and, if specified, accumulated
// fields.

// $ sort
//
// Reorders the document stream by a specified sort key. Only the order
// changes; the documents remain unmodified. For each input document,
// outputs one document.

// $count
//
// Returns a count of the number of documents at this stage of the aggregation pipeline.

// Example:
//
// Match Stage -> (Filters Records) - Query1
//
//   SELECT _jsonb
//   FROM schema.table
//   WHERE _jsonb->'name' = 'foo'
//
// Group Stage -> (Filters Records) - Query2
//
//   SELECT
//     _jsonb->'name', COUNT(*) AS nameCount
//   FROM (Query1)
//   GROUP BY _jsonb->'name'
//
// Wrapping Stage -> (Wraps Records into JSON)
//
//   SELECT json_build_object(
//     '%k', json_build_array('name', 'nameCount'),
//     'name', name, 'nameCount', nameCount
//   )
//   FROM (Query3)

// MsgAggregate implements HandlerInterface.
func (h *Handler) MsgAggregate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	unimplementedFields := []string{
		"$project",
	}
	if err := common.Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
	}

	var sp sqlParam
	if sp.db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}
	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}
	var ok bool
	if sp.collection, ok = collectionParam.(string); !ok {
		return nil, common.NewErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
		)
	}

	var pipeline *types.Array
	if pipeline, err = common.GetOptionalParam(document, "pipeline", pipeline); err != nil {
		return nil, err
	}

	stages := []*aggregate.Stage{}

	for i := 0; i < pipeline.Len(); i++ {
		p := must.NotFail(pipeline.Get(i)).(*types.Document)
		for _, pipelineOp := range p.Keys() {
			switch pipelineOp {
			case "$match":
				match := must.NotFail(p.Get(pipelineOp)).(*types.Document)
				matchStage, err := aggregate.ParseMatchStage(match)
				if err != nil {
					return nil, err
				}

				stages = append(stages, matchStage)

			case "$count":
				count := must.NotFail(p.Get(pipelineOp))
				countStage, err := aggregate.ParseCountStage(count)
				if err != nil {
					return nil, err
				}

				stages = append(stages, countStage)

			case "$group":
				group := must.NotFail(p.Get(pipelineOp)).(*types.Document)
				groupStage, err := aggregate.ParseGroupStage(group)
				if err != nil {
					return nil, err
				}

				stages = append(stages, groupStage)

			case "$sort":
				sort := must.NotFail(p.Get(pipelineOp)).(*types.Document)
				err := aggregate.AddSortStage(&stages, sort)
				if err != nil {
					return nil, err
				}

			default:
				return nil, common.NewErrorMsg(common.ErrBadValue, fmt.Sprintf("unknown pipeline operator: %s", pipelineOp))
			}
		}
	}

	table := `"` + sp.db + `"."` + sp.collection + `"`

	sql, queryValues := aggregate.Wrap(table, stages)

	fmt.Printf(" *** SQL: %s %v %v\n", sql, queryValues, len(queryValues))

	rows, err := h.pgPool.Query(ctx, sql, queryValues...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	var resDocs []*types.Document
	for {
		doc, err := nextRow(rows)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		resDocs = append(resDocs, doc)
	}

	firstBatch := types.MakeArray(len(resDocs))
	for _, doc := range resDocs {
		if err = firstBatch.Append(doc); err != nil {
			return nil, err
		}
	}

	var reply wire.OpMsg

	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"firstBatch", firstBatch,
				"id", int64(0), // TODO
				"ns", sp.db+"."+sp.collection,
			)),
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
