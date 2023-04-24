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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

func TestNewReporterLock(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		f        *Flag
		dnt      string
		execName string
		t        *bool
		locked   bool
	}{
		"NoSet": {
			f: new(Flag),
		},
		"FlagEnable": {
			f:      &Flag{v: pointer.ToBool(true)},
			t:      pointer.ToBool(true),
			locked: true,
		},
		"FlagDisable": {
			f:      &Flag{v: pointer.ToBool(false)},
			t:      pointer.ToBool(false),
			locked: true,
		},
		"DoNotTrack": {
			f:      new(Flag),
			dnt:    "enable",
			t:      pointer.ToBool(false),
			locked: true,
		},
		"ExecName": {
			f:        new(Flag),
			execName: "exec_donottrack",
			t:        pointer.ToBool(false),
			locked:   true,
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			provider, err := state.NewProvider("")
			require.NoError(t, err)

			opts := NewReporterOpts{
				F:           tc.f,
				DNT:         tc.dnt,
				ExecName:    tc.execName,
				ConnMetrics: connmetrics.NewListenerMetrics().ConnMetrics,
				P:           provider,
				L:           zap.L(),
			}

			_, err = NewReporter(&opts)
			assert.NoError(t, err)

			s := provider.Get()
			assert.Equal(t, tc.t, s.Telemetry)
			assert.Equal(t, tc.locked, s.TelemetryLocked)
		})
	}
}

func TestReporterReport(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		f                *Flag
		reporterResponse string
		latestVersion    string
		updateAvailable  bool
	}{
		"UpdateAvailable": {
			f:                &Flag{v: pointer.ToBool(true)},
			updateAvailable:  true,
			latestVersion:    "0.3.4",
			reporterResponse: `{"update_available": true, "latest_version": "0.3.4"}`,
		},
		"UpdateUnavailable": {
			f:                &Flag{v: pointer.ToBool(true)},
			updateAvailable:  false,
			latestVersion:    "",
			reporterResponse: `{"update_available": false, "latest_version": "0.3.4"}`,
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, tc.reporterResponse)
			}))
			defer ts.Close()

			provider, err := state.NewProvider("")
			require.NoError(t, err)

			opts := NewReporterOpts{
				URL:           ts.URL,
				F:             tc.f,
				ConnMetrics:   connmetrics.NewListenerMetrics().ConnMetrics,
				P:             provider,
				L:             zap.L(),
				ReportTimeout: 1 * time.Minute,
			}

			r, err := NewReporter(&opts)
			assert.NoError(t, err)

			ctx := context.Background()
			r.report(ctx)

			s := r.P.Get()

			require.Equal(t, s.UpdateAvailable(), s.UpdateAvailable())
			require.Equal(t, tc.latestVersion, s.LatestVersion)
		})
	}
}
