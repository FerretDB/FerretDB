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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgAggregate implements HandlerInterface.
func (h *Handler) MsgAggregate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	unimplementedFields := []string{
		"$group",
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

	// var aggregate string
	var pipeline *types.Array
	// if aggregate, err = common.GetOptionalParam(document, "aggregate", aggregate); err != nil {
	// 	return nil, err
	// }
	if pipeline, err = common.GetOptionalParam(document, "pipeline", pipeline); err != nil {
		return nil, err
	}

	fields := "_jsonb"
	sql := `FROM "` + sp.db + `"."` + sp.collection + `"`

	var queryValues []interface{}
	for i := 0; i < pipeline.Len(); i++ {
		p := must.NotFail(pipeline.Get(i)).(*types.Document)
		for _, pipelineOp := range p.Keys() {
			switch pipelineOp {
			case "$match":
				match := must.NotFail(p.Get(pipelineOp)).(*types.Document)
				where, values, err := common.AggregateMatch(match)
				if err != nil {
					return nil, err
				}

				sql += " WHERE " + *where
				queryValues = append(queryValues, values...)

			case "$count":
				count := must.NotFail(p.Get(pipelineOp)).(string)
				fields = common.AggregateCount(count)

			default:
				return nil, common.NewErrorMsg(common.ErrBadValue, fmt.Sprintf("unknown pipeline operator: %s", pipelineOp))
			}
		}
	}

	sql = "SELECT " + fields + " " + sql
	// fmt.Printf(" *** SQL: %s %v %v\n", sql, queryValues, len(queryValues))
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
