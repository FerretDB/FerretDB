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

package common

import (
	"os"
	"path/filepath"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// ServerStatus returns a common part of serverStatus command response.
func ServerStatus(state *state.State, cm *connmetrics.ConnMetrics) (*types.Document, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	exec, err := os.Executable()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	uptime := time.Since(state.Start)

	metricsDoc := types.MakeDocument(0)

	metrics := cm.GetResponses()
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

	telemetryState := "disabled"

	switch {
	case state.Telemetry == nil:
		telemetryState = "undecided"
	case pointer.GetBool(state.Telemetry):
		telemetryState = "enabled"
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
			"state", telemetryState,
		)),
		"metrics", must.NotFail(types.NewDocument(
			"commands", metricsDoc,
		)),
		"ok", float64(1),
	))

	return res, nil
}
