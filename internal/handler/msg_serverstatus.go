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

package handler

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgServerStatus implements `serverStatus` command.
func (h *Handler) MsgServerStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	if err := h.authenticate(ctx); err != nil {
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

	uptime := time.Since(h.StateProvider.Get().Start)

	metricsDoc := types.MakeDocument(0)

	metrics := h.ConnMetrics.GetResponses()
	for _, commands := range metrics {
		for command, arguments := range commands {
			var total, failed int
			for _, m := range arguments {
				total += m.Total

				for _, v := range m.Failures {
					failed += v
				}
			}

			d := must.NotFail(types.NewDocument("total", int64(total), "failed", int64(failed)))
			metricsDoc.Set(command, d)
		}
	}

	res := must.NotFail(types.NewDocument(
		"host", host,
		"version", version.Get().MongoDBVersion,
		"process", filepath.Base(exec),
		"pid", int64(os.Getpid()),
		"uptime", uptime.Seconds(),
		"uptimeMillis", uptime.Milliseconds(),
		"uptimeEstimate", int64(uptime.Seconds()),
		"localTime", time.Now(),
		"freeMonitoring", must.NotFail(types.NewDocument(
			"state", h.StateProvider.Get().TelemetryString(),
		)),
		"metrics", must.NotFail(types.NewDocument(
			"commands", metricsDoc,
		)),

		// our extensions
		"ferretdbVersion", version.Get().Version,

		"ok", float64(1),
	))

	stats, err := h.b.Status(ctx, new(backends.StatusParams))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res.Set("catalogStats", must.NotFail(types.NewDocument(
		"collections", stats.CountCollections,
		"capped", stats.CountCappedCollections,
		"clustered", int32(0),
		"timeseries", int32(0),
		"views", int32(0),
		"internalCollections", int32(0),
		"internalViews", int32(0),
	)))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.MakeOpMsgSection(
		res,
	)))

	return &reply, nil
}
