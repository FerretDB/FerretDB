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

	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

const (
	undecidedDelay = time.Hour
	reportDelay    = 24 * time.Hour
)

type request struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Branch  string `json:"branch"`
	Dirty   bool   `json:"dirty"`
	Debug   bool   `json:"debug"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`

	UUID   string        `json:"uuid"`
	Uptime time.Duration `json:"uptime"`
}

type response struct {
	LatestVersion string `json:"latest_version"`
}

type Reporter struct {
	url string
	p   *state.Provider
	l   *zap.Logger
	c   *http.Client
}

func NewReporter(url string, p *state.Provider, l *zap.Logger) (*Reporter, error) {
	return &Reporter{
		url: url,
		p:   p,
		l:   l,
		c:   http.DefaultClient,
	}, nil
}

func (r *Reporter) Run(ctx context.Context) {
	ch := r.p.Subscribe()

	r.firstReportDelay(ctx, ch)

	for ctx.Err() == nil {
		r.report(ctx)

		delayCtx, delayCancel := context.WithTimeout(ctx, reportDelay)
		<-delayCtx.Done()
		delayCancel()
	}

	// do one last report before exiting if telemetry is explicitly enabled
	if pointer.GetBool(r.p.Get().Telemetry) {
		r.report(ctx)
	}
}

func makeRequest(s *state.State) *request {
	v := version.Get()

	return &request{
		Version: v.Version,
		Commit:  v.Commit,
		Branch:  v.Branch,
		Dirty:   v.Dirty,
		Debug:   v.Debug,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,

		UUID:   s.UUID,
		Uptime: time.Since(s.Start),
	}
}

func (r *Reporter) report(ctx context.Context) {
	s := r.p.Get()
	if s.Telemetry != nil && !*s.Telemetry {
		r.l.Debug("Telemetry is disabled, skipping reporting.")
		return
	}

	request := makeRequest(s)
	r.l.Info("Reporting telemetry.", zap.Reflect("data", request))

	b, err := json.Marshal(request)
	if err != nil {
		r.l.Error("Failed to marshal telemetry request.", zap.Error(err))
		return
	}

	reqCtx, reqCancel := context.WithTimeout(ctx, 5*time.Second)
	defer reqCancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, r.url, bytes.NewReader(b))
	if err != nil {
		r.l.Error("Failed to create telemetry request.", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := r.c.Do(req)
	if err != nil {
		r.l.Debug("Failed to send telemetry request.", zap.Error(err))
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		r.l.Debug("Failed to send telemetry request.", zap.Int("status", res.StatusCode))
		return
	}

	var response response
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		r.l.Debug("Failed to read telemetry response.", zap.Error(err))
		return
	}

	if response.LatestVersion == "" {
		r.l.Debug("No latest version in telemetry response.")
		return
	}

	if response.LatestVersion == s.LatestVersion {
		r.l.Debug("Latest version is up to date.")
		return
	}

	r.l.Info(
		"New version available.",
		zap.String("current_version", request.Version), zap.String("latest_version", response.LatestVersion),
	)

	err = r.p.Update(func(s *state.State) { s.LatestVersion = response.LatestVersion })
	if err != nil {
		r.l.Error("Failed to update state with latest version.", zap.Error(err))
		return
	}
}

// firstReportDelay waits until telemetry reporting state is decided,
// main context is cancelled, or timeout is reached.
func (r *Reporter) firstReportDelay(ctx context.Context, ch <-chan struct{}) {
	if r.p.Get().Telemetry != nil {
		return
	}

	msg := fmt.Sprintf(
		"Telemetry state undecided, waiting %s before the first report. "+
			"Read more about FerretDB telemetry at https://beacon.ferretdb.io",
		undecidedDelay,
	)
	r.l.Warn(msg)

	delayCtx, delayCancel := context.WithTimeout(ctx, undecidedDelay)
	defer delayCancel()

	for {
		select {
		case <-delayCtx.Done():
			return
		case <-ch:
			if r.p.Get().Telemetry != nil {
				return
			}
		}
	}
}
