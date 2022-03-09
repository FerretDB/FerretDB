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

package handlers

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgServerStatus OpMsg used to get a server status.
func (h *Handler) MsgServerStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var db string
	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	host, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	exec, err := os.Executable()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	uptime := time.Since(h.startTime)

	stats, err := h.pgPool.SchemaStats(ctx, db)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"host", host,
			"version", versionValue,
			"process", filepath.Base(exec),
			"pid", int64(os.Getpid()),
			"uptime", int64(uptime.Seconds()),
			"uptimeMillis", uptime.Milliseconds(),
			"uptimeEstimate", int64(uptime.Seconds()),
			"localTime", time.Now(),
			"catalogStats", must.NotFail(types.NewDocument(
				"collections", stats.CountTables,
				"capped", int32(0),
				"timeseries", int32(0),
				"views", int32(0),
				"internalCollections", int32(0),
				"internalViews", int32(0),
			)),
			"freeMonitoring", must.NotFail(types.NewDocument(
				"state", "disabled",
			)),
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
