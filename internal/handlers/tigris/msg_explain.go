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

package tigris

import (
	"context"
	"net"
	"os"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
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

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command, err := common.GetRequiredParam[*types.Document](document, document.Command())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	explain, err := common.GetRequiredParam[*types.Document](document, "explain")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter, err := common.GetOptionalParam[*types.Document](explain, "filter", nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	queryPlanner := must.NotFail(types.NewDocument(
		"Filter", string(h.db.BuildFilter(filter)),
	))

	hostname, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var port int32

	connInfo := conninfo.GetConnInfo(ctx)
	if connInfo.PeerAddr != nil {
		if tcpAddr, ok := connInfo.PeerAddr.(*net.TCPAddr); ok {
			port = int32(tcpAddr.Port)
		}
	}

	serverInfo := must.NotFail(types.NewDocument(
		"host", hostname,
		"port", port,
		"version", version.MongoDBVersion,
		"gitVersion", version.Get().Commit,
		"ferretdbVersion", version.Get().Version,
	))

	cmd := command.DeepCopy()
	cmd.Set("$db", db)

	var reply wire.OpMsg

	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"queryPlanner", queryPlanner,
			"explainVersion", "1",
			"command", cmd,
			"serverInfo", serverInfo,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
