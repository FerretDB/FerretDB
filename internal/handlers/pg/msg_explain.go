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

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
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

	var sp pgdb.SQLParam
	if sp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "verbosity")

	command, err := common.GetRequiredParam[*types.Document](document, document.Command())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if sp.Collection, err = common.GetRequiredParam[string](command, command.Command()); err != nil {
		return nil, lazyerrors.Error(err)
	}

	sp.Explain = true

	explain, err := common.GetRequiredParam[*types.Document](document, "explain")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sp.Filter, err = common.GetOptionalParam[*types.Document](explain, "filter", nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var queryPlanner *types.Document
	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var err error
		queryPlanner, err = pgdb.Explain(ctx, tx, &sp)
		return err
	})
	if err != nil {
		return nil, err
	}

	// if filter was set, it means that pushdown was done
	pushdown := queryPlanner.HasByPath(types.NewPath("Plan", "Filter"))

	hostname, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	serverInfo := must.NotFail(types.NewDocument(
		"host", hostname,
		"version", version.Get().MongoDBVersion,
		"gitVersion", version.Get().Commit,
		"ferretdbVersion", version.Get().Version,
	))

	cmd := command.DeepCopy()
	cmd.Set("$db", sp.DB)

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"queryPlanner", queryPlanner,
			"explainVersion", "1",
			"command", cmd,
			"pushdown", pushdown,
			"serverInfo", serverInfo,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return &reply, nil
}
