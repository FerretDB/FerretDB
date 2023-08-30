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
	"os"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgExplain implements HandlerInterface.
func (h *Handler) MsgExplain(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetExplainParams(document, h.L)
	if err != nil {
		return nil, err
	}

	qp := pgdb.QueryParams{
		DB:         params.DB,
		Collection: params.Collection,
		Explain:    true,
		Filter:     params.Filter,
		Sort:       params.Sort,
	}

	if params.Aggregate {
		qp.Filter, qp.Sort = aggregations.GetPushdownQuery(params.StagesDocs)
	}

	if h.DisableFilterPushdown {
		qp.Filter = nil
	}

	if !h.EnableSortPushdown {
		qp.Sort = nil
	}

	// Limit pushdown is not applied if:
	//  - `filter` is set, it must fetch all documents to filter them in memory;
	//  - `sort` is set but `EnableSortPushdown` is not set, it must fetch all documents
	//  and sort them in memory;
	//  - `skip` is non-zero value, skip pushdown is not supported yet.
	// TODO https://github.com/FerretDB/FerretDB/issues/3016
	if params.Filter.Len() == 0 && (params.Sort.Len() == 0 || h.EnableSortPushdown) && params.Skip == 0 {
		qp.Limit = params.Limit
	}

	var queryPlanner *types.Document
	var results pgdb.QueryResults

	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var err error
		queryPlanner, results, err = pgdb.Explain(ctx, tx, &qp)
		return err
	})
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	serverInfo := must.NotFail(types.NewDocument(
		"host", hostname,
		"version", version.Get().MongoDBVersion,
		"gitVersion", version.Get().Commit,

		// our extensions
		"ferretdbVersion", version.Get().Version,
	))

	cmd := params.Command
	cmd.Set("$db", qp.DB)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"queryPlanner", queryPlanner,
			"explainVersion", "1",
			"command", cmd,
			"serverInfo", serverInfo,

			// our extensions
			"pushdown", results.FilterPushdown,
			"sortingPushdown", results.SortPushdown,
			"limitPushdown", results.LimitPushdown,

			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
