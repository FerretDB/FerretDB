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

package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/AlekSi/pointer"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// request represents telemetry request.
type request struct {
	Version          string         `json:"version"`
	Commit           string         `json:"commit"`
	Branch           string         `json:"branch"`
	Dirty            bool           `json:"dirty"`
	Debug            bool           `json:"debug"`
	BuildEnvironment map[string]any `json:"build_environment"`
	OS               string         `json:"os"`
	Arch             string         `json:"arch"`

	HandlerVersion string `json:"handler_version"` // PostgreSQL, Tigris, etc version

	UUID   string        `json:"uuid"`
	Uptime time.Duration `json:"uptime"`

	// opcode (e.g. "OP_MSG") -> command (e.g. "update") -> argument (e.g. "$set") -> result (e.g. "ok") -> count
	CommandMetrics map[string]map[string]map[string]map[string]int `json:"command_metrics"`
}

// response represents telemetry response.
type response struct {
	LatestVersion string `json:"latest_version"`
}

// Reporter sends telemetry reports if telemetry is enabled.
type Reporter struct {
	*NewReporterOpts
	c *http.Client
}

// NewReporterOpts represents reporter options.
type NewReporterOpts struct {
	URL            string
	F              *Flag
	DNT            string
	ExecName       string
	P              *state.Provider
	ConnMetrics    *connmetrics.ConnMetrics
	L              *zap.Logger
	UndecidedDelay time.Duration
	ReportInterval time.Duration
	ReportTimeout  time.Duration
}

// NewReporter creates a new reporter.
func NewReporter(opts *NewReporterOpts) (*Reporter, error) {
	t, locked, err := initialState(opts.F, opts.DNT, opts.ExecName, opts.P.Get().Telemetry, opts.L)
	if err != nil {
		return nil, err
	}

	err = opts.P.Update(func(s *state.State) {
		s.Telemetry = t
		s.TelemetryLocked = locked
	})
	if err != nil {
		return nil, err
	}

	return &Reporter{
		NewReporterOpts: opts,
		c:               http.DefaultClient,
	}, nil
}

// Run runs reporter until context is canceled.
func (r *Reporter) Run(ctx context.Context) {
	r.L.Debug("Reporter started.")
	defer r.L.Debug("Reporter stopped.")

	ch := r.P.Subscribe()

	r.firstReportDelay(ctx, ch)

	for ctx.Err() == nil {
		r.report(ctx)

		ctxutil.Sleep(ctx, r.ReportInterval)
	}

	// do one last report before exiting if telemetry is explicitly enabled
	if pointer.GetBool(r.P.Get().Telemetry) {
		r.report(context.Background())
	}
}

// firstReportDelay waits until telemetry reporting state is decided,
// main context is cancelled, or timeout is reached.
func (r *Reporter) firstReportDelay(ctx context.Context, ch <-chan struct{}) {
	if r.P.Get().Telemetry != nil {
		return
	}

	msg := fmt.Sprintf(
		"The telemetry state is undecided; the first report will be sent in %s. "+
			"Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.",
		r.UndecidedDelay,
	)
	r.L.Info(msg)

	delayCtx, delayCancel := context.WithTimeout(ctx, r.UndecidedDelay)
	defer delayCancel()

	for {
		select {
		case <-delayCtx.Done():
			return
		case <-ch:
			if r.P.Get().Telemetry != nil {
				return
			}
		}
	}
}

// makeRequest creates a new telemetry request.
func makeRequest(s *state.State, m *connmetrics.ConnMetrics) *request {
	commandMetrics := map[string]map[string]map[string]map[string]int{}

	for opcode, commands := range m.GetResponses() {
		for command, arguments := range commands {
			for argument, m := range arguments {
				if _, ok := commandMetrics[opcode]; !ok {
					commandMetrics[opcode] = map[string]map[string]map[string]int{}
				}

				if _, ok := commandMetrics[opcode][command]; !ok {
					commandMetrics[opcode][command] = map[string]map[string]int{}
				}

				if _, ok := commandMetrics[opcode][command][argument]; !ok {
					commandMetrics[opcode][command][argument] = map[string]int{}
				}

				var failures int

				for result, c := range m.Failures {
					if result == "ok" {
						panic("result should not be ok")
					}
					commandMetrics[opcode][command][argument][result] = c
					failures += c
				}

				commandMetrics[opcode][command][argument]["ok"] = m.Total - failures
			}
		}
	}

	v := version.Get()

	return &request{
		Version:          v.Version,
		Commit:           v.Commit,
		Branch:           v.Branch,
		Dirty:            v.Dirty,
		Debug:            v.DebugBuild,
		BuildEnvironment: v.BuildEnvironment.Map(),
		OS:               runtime.GOOS,
		Arch:             runtime.GOARCH,

		HandlerVersion: s.HandlerVersion,

		UUID:   s.UUID,
		Uptime: time.Since(s.Start),

		CommandMetrics: commandMetrics,
	}
}

// report sends telemetry report unless telemetry is disabled.
func (r *Reporter) report(ctx context.Context) {
	s := r.P.Get()
	if s.Telemetry != nil && !*s.Telemetry {
		r.L.Debug("Telemetry is disabled, skipping reporting.")
		return
	}

	request := makeRequest(s, r.ConnMetrics)
	r.L.Info("Reporting telemetry.", zap.String("url", r.URL), zap.Any("data", request))

	b, err := json.Marshal(request)
	if err != nil {
		r.L.Error("Failed to marshal telemetry request.", zap.Error(err))
		return
	}

	reqCtx, reqCancel := context.WithTimeout(ctx, r.ReportTimeout)
	defer reqCancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, r.URL, bytes.NewReader(b))
	if err != nil {
		r.L.Error("Failed to create telemetry request.", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := r.c.Do(req)
	if err != nil {
		r.L.Debug("Failed to send telemetry request.", zap.Error(err))
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		r.L.Debug("Failed to send telemetry request.", zap.Int("status", res.StatusCode))
		return
	}

	var response response
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		r.L.Debug("Failed to read telemetry response.", zap.Error(err))
		return
	}

	if response.LatestVersion == "" {
		r.L.Debug("No latest version in telemetry response.")
		return
	}

	if response.LatestVersion == s.LatestVersion {
		r.L.Debug("Latest version is up to date.")
		return
	}

	r.L.Info(
		"New version available.",
		zap.String("current_version", request.Version), zap.String("latest_version", response.LatestVersion),
	)

	err = r.P.Update(func(s *state.State) { s.LatestVersion = response.LatestVersion })
	if err != nil {
		r.L.Error("Failed to update state with latest version.", zap.Error(err))
		return
	}
}
