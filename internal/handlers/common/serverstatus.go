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

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

// ServerStatus returns a common part of serverStatus command response.
func ServerStatus(startTime time.Time, cm *connmetrics.ConnMetrics) (*types.Document, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	exec, err := os.Executable()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	uptime := time.Since(startTime)

	metricsDoc := types.MakeDocument(0)

	metrics := cm.GetResponses()
	for cmd, cmdMetrics := range metrics {
		var cmdDoc *types.Document
		switch cmdMetrics := cmdMetrics.(type) {
		case connmetrics.UpdateCommandMetrics:
			cmdDoc = must.NotFail(types.NewDocument(
				"arrayFilters", cmdMetrics.ArrayFilters,
				"failed", cmdMetrics.Failed,
				"pipeline", cmdMetrics.Pipeline,
				"total", cmdMetrics.Total,
			))
		case connmetrics.BasicCommandMetrics:
			cmdDoc = must.NotFail(types.NewDocument("total", cmdMetrics.Total, "failed", cmdMetrics.Failed))
		}

		metricsDoc.Set(cmd, cmdDoc)
	}

	res := must.NotFail(types.NewDocument(
		"host", host,
		"version", version.MongoDBVersion,
		"process", filepath.Base(exec),
		"pid", int64(os.Getpid()),
		"uptime", uptime.Seconds(),
		"uptimeMillis", uptime.Milliseconds(),
		"uptimeEstimate", int64(uptime.Seconds()),
		"localTime", time.Now(),
		"freeMonitoring", must.NotFail(types.NewDocument(
			"state", "disabled",
		)),
		"metrics", must.NotFail(types.NewDocument(
			"commands", metricsDoc,
		)),
		"ok", float64(1),
	))

	return res, nil
}
