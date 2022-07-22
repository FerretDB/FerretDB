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
	"net"
	"os"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/version"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgExplain implements HandlerInterface.
func (h *Handler) MsgExplain(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	sp, err := h.parseExplainUserInput(ctx, document)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resDocs := make([]*types.Document, 0, 16)
	err = h.pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		fetchedChan, err := h.pgPool.QueryDocuments(ctx, tx, sp)
		if err != nil {
			return err
		}
		defer func() {
			// Drain the channel to prevent leaking goroutines.
			// TODO Offer a better design instead of channels: https://github.com/FerretDB/FerretDB/issues/898.
			for range fetchedChan {
			}
		}()
		for fetchedItem := range fetchedChan {
			if fetchedItem.Err != nil {
				return fetchedItem.Err
			}
			resDocs = append(resDocs, fetchedItem.Docs...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return h.buildExplainResult(ctx, document, resDocs)
}

// parseExplainUserInput validates user input and returns pgdb.SQLParam.
func (h *Handler) parseExplainUserInput(ctx context.Context, document *types.Document) (pgdb.SQLParam, error) {
	common.Ignored(document, h.l, "verbosity")
	var err error
	var sp pgdb.SQLParam
	if sp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return sp, lazyerrors.Error(err)
	}
	commandParam, err := document.Get(document.Command())
	if err != nil {
		return sp, lazyerrors.Error(err)
	}
	var command *types.Document
	var ok bool
	if command, ok = commandParam.(*types.Document); !ok {
		return sp, common.NewErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("has invalid type %s", common.AliasFromType(commandParam)),
		)
	}
	switch command.Command() {
	case "count", "findAndModify", "find":
		// ok
	default:
		return sp, common.NewErrorMsg(
			common.ErrNotImplemented,
			fmt.Sprintf("explain for %s s not supported", command.Command()),
		)
	}
	collectionParam, err := command.Get(command.Command())
	if err != nil {
		return sp, lazyerrors.Error(err)
	}
	if sp.Collection, ok = collectionParam.(string); !ok {
		return sp, common.NewErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
		)
	}
	sp.Explain = true
	return sp, nil
}

// buildExplainResult build explain response.
func (h *Handler) buildExplainResult(ctx context.Context, document *types.Document, resDocs []*types.Document) (*wire.OpMsg, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	var port int
	connInfo := conninfo.GetConnInfo(ctx)
	if connInfo.PeerAddr != nil {
		port = connInfo.PeerAddr.(*net.TCPAddr).Port
	}

	serverInfo := must.NotFail(types.NewDocument(
		"host", hostname,
		"port", int32(port),
		"version", version.MongoDBVersion,
		"gitVersion", version.Get().Commit,
		"ferretdbVersion", version.Get().Version,
	))

	queryPlanner := types.MakeArray(len(resDocs))
	for _, item := range resDocs {
		must.NoError(queryPlanner.Append(item))
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"queryPlanner", queryPlanner,
			"explainVersion", int32(1),
			"command", document,
			"serverInfo", serverInfo,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return &reply, nil
}
