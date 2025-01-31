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
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// MsgServerStatus implements `serverStatus` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgServerStatus(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
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

	metricsDoc := wirebson.MakeDocument(0)

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

			d := must.NotFail(wirebson.NewDocument("total", int64(total), "failed", int64(failed)))
			must.NoError(metricsDoc.Add(command, d))
		}
	}

	info := version.Get()

	buildEnvironment := wirebson.MakeDocument(len(info.BuildEnvironment))
	for _, k := range slices.Sorted(maps.Keys(info.BuildEnvironment)) {
		must.NoError(buildEnvironment.Add(k, info.BuildEnvironment[k]))
	}

	state := h.StateProvider.Get()
	uptime := time.Since(state.Start)

	res := must.NotFail(wirebson.NewDocument(
		"host", host,
		"version", info.MongoDBVersion,
		"process", filepath.Base(exec),
		"pid", int64(os.Getpid()),
		"uptime", uptime.Seconds(),
		"uptimeMillis", uptime.Milliseconds(),
		"uptimeEstimate", int64(uptime.Seconds()),
		"localTime", time.Now(),
		"freeMonitoring", must.NotFail(wirebson.NewDocument(
			"state", state.TelemetryString(),
		)),
		"metrics", must.NotFail(wirebson.NewDocument(
			"commands", metricsDoc,
		)),
		"catalogStats", must.NotFail(wirebson.NewDocument(
			"collections", int32(0),
			"clustered", int32(0),
			"timeseries", int32(0),
			"views", int32(0),
			"internalCollections", int32(0),
			"internalViews", int32(0),
		)),

		// our extensions for easier bug reporting
		"ferretdb", must.NotFail(wirebson.NewDocument(
			"version", info.Version,
			"gitVersion", info.Commit,
			"buildEnvironment", buildEnvironment,
			"debug", info.DevBuild,
			"package", info.Package,
			"postgresql", state.PostgreSQLVersion,
			"documentdb", state.DocumentDBVersion,
		)),

		"ok", float64(1),
	))

	return wire.NewOpMsg(must.NotFail(res.Encode()))
}
