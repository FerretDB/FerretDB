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
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// Time format for local file.
const fileTimeFormat = "2006-01-02 15:04:05Z07:00"

// report represents telemetry data to report.
//
//nolint:vet // for readability
type report struct {
	Comment string `json:"_comment,omitempty"` // for local file only

	Version          string            `json:"version"`
	Commit           string            `json:"commit"`
	Branch           string            `json:"branch"`
	Dirty            bool              `json:"dirty"`
	Package          string            `json:"package"`
	Debug            bool              `json:"debug"`
	BuildEnvironment map[string]string `json:"build_environment"`
	OS               string            `json:"os"`
	Arch             string            `json:"arch"`

	PostgreSQLVersion string `json:"postgresql_version"`
	DocumentDBVersion string `json:"documentdb_version"`

	UUID   string        `json:"uuid"`
	Uptime time.Duration `json:"uptime"`

	// opcode (e.g. "OP_MSG", "OP_QUERY") ->
	// command (e.g. "find", "aggregate") ->
	// argument that caused an error (e.g. "sort", "$count (stage)"; or "unknown") ->
	// result (e.g. "NotImplemented", "InternalError"; or "ok") ->
	// count.
	CommandMetrics map[string]map[string]map[string]map[string]int `json:"command_metrics"`
}

// response represents Beacon's response.
type response struct {
	LatestVersion   string `json:"latest_version"`
	UpdateInfo      string `json:"update_info"`
	UpdateAvailable bool   `json:"update_available"`
}

// Reporter converts already collected data (such as metrics) to the report,
// sends it to the Beacon if telemetry reporting is enabled,
// and writes it to a local file in the state directory.
type Reporter struct {
	*NewReporterOpts
	c *http.Client
}

// NewReporterOpts represents reporter options.
type NewReporterOpts struct {
	URL            string
	File           string
	F              *Flag
	DNT            string
	ExecName       string
	P              *state.Provider
	ConnMetrics    *connmetrics.ConnMetrics
	L              *slog.Logger
	UndecidedDelay time.Duration
	ReportInterval time.Duration
}

// NewReporter creates a new reporter.
func NewReporter(opts *NewReporterOpts) (*Reporter, error) {
	if opts.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	if opts.File == "" {
		return nil, fmt.Errorf("File is required")
	}

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
	r.L.DebugContext(ctx, "Reporter started")
	defer r.L.DebugContext(ctx, "Reporter stopped")

	// no delay for decided state
	if r.P.Get().Telemetry == nil {
		r.firstReportDelay(ctx)
	}

	for context.Cause(ctx) == nil {
		report := r.makeReport()

		if s := r.P.Get(); s.Telemetry == nil || *s.Telemetry {
			r.sendReport(ctx, report)
		}

		r.writeReport(report)

		ctxutil.Sleep(ctx, r.ReportInterval)
	}

	report := r.makeReport()

	// send one last time before exiting only if explicitly enabled (not undecided)
	if s := r.P.Get(); s.Telemetry != nil && *s.Telemetry {
		r.sendReport(ctx, report)
	}

	r.writeReport(report)
}

// firstReportDelay waits until telemetry reporting state is decided,
// context is canceled, or UndecidedDelay is reached.
func (r *Reporter) firstReportDelay(ctx context.Context) {
	msg := fmt.Sprintf(
		"The telemetry state is undecided; the first report will be sent in %s. "+
			"Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.com.",
		r.UndecidedDelay,
	)
	r.L.InfoContext(ctx, msg)

	delayCtx, delayCancel := context.WithTimeout(ctx, r.UndecidedDelay)
	defer delayCancel()

	ch := r.P.Subscribe()

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

// makeReport converts runtime state, metrics, and build information to telemetry data.
func (r *Reporter) makeReport() *report {
	commandMetrics := map[string]map[string]map[string]map[string]int{}

	for opcode, commands := range r.ConnMetrics.GetResponses() {
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

	info := version.Get()
	s := r.P.Get()

	return &report{
		Version:          info.Version,
		Commit:           info.Commit,
		Branch:           info.Branch,
		Dirty:            info.Dirty,
		Package:          info.Package,
		Debug:            info.DevBuild,
		BuildEnvironment: info.BuildEnvironment,
		OS:               runtime.GOOS,
		Arch:             runtime.GOARCH,

		PostgreSQLVersion: s.PostgreSQLVersion,
		DocumentDBVersion: s.DocumentDBVersion,

		UUID:   s.UUID,
		Uptime: time.Since(s.Start),

		CommandMetrics: commandMetrics,
	}
}

// sendReport sends telemetry report to the Beacon.
// It always set report.Comment field.
//
// It receives information about available updates and updates the state.
// If update is available, it logs the message.
func (r *Reporter) sendReport(ctx context.Context, report *report) {
	r.L.InfoContext(ctx, "Sending telemetry report", slog.String("url", r.URL), slog.Any("data", report))
	b, err := json.Marshal(report)
	report.Comment = fmt.Sprintf("Failed to send to %s at %s.", r.URL, time.Now().Format(fileTimeFormat))
	if err != nil {
		r.L.ErrorContext(ctx, "Failed to marshal telemetry report", logging.Error(err))
		return
	}

	reqCtx, reqCancel := context.WithTimeout(ctx, 3*time.Second)
	defer reqCancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, r.URL, bytes.NewReader(b))
	if err != nil {
		r.L.ErrorContext(ctx, "Failed to create telemetry request", logging.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := r.c.Do(req)
	if err != nil {
		r.L.DebugContext(ctx, "Failed to send telemetry report", logging.Error(err))
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		r.L.DebugContext(ctx, "Failed to send telemetry report", slog.Int("status", res.StatusCode))
		return
	}

	var response response
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		r.L.DebugContext(ctx, "Failed to read telemetry response", logging.Error(err))
		return
	}

	r.L.DebugContext(ctx, "Read telemetry response", slog.Any("response", response))

	if response.UpdateInfo != "" || response.UpdateAvailable {
		msg := response.UpdateInfo
		if msg == "" {
			msg = "A new version available!"
		}

		r.L.InfoContext(
			ctx,
			msg,
			slog.String("current_version", report.Version),
			slog.String("latest_version", response.LatestVersion),
		)
	}

	if err = r.P.Update(func(s *state.State) {
		s.LatestVersion = response.LatestVersion
		s.UpdateInfo = response.UpdateInfo
		s.UpdateAvailable = response.UpdateAvailable
	}); err != nil {
		r.L.ErrorContext(ctx, "Failed to update state with latest version", logging.Error(err))
	}

	report.Comment = fmt.Sprintf("Sent to %s at %s.", r.URL, time.Now().Format(fileTimeFormat))
}

// writeReport writes telemetry report to the local files.
func (r *Reporter) writeReport(report *report) {
	if report.Comment == "" {
		report.Comment = fmt.Sprintf("Created at %s, not sent because reporting is disabled.", time.Now().Format(fileTimeFormat))
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		r.L.Error("Failed to marshal telemetry report", logging.Error(err))
		return
	}

	if err = os.WriteFile(r.File, b, 0o666); err != nil {
		r.L.Error("Failed to write telemetry report to local file", slog.String("file", r.File), logging.Error(err))
		return
	}

	r.L.Info("Wrote telemetry report to local file", slog.String("file", r.File))
}
