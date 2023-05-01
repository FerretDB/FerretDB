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
	"github.com/FerretDB/FerretDB/internal/util/testutil"
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
		telemetryReponse string
		latestVersion    string
		updateAvailable  bool
	}{
		"UpdateAvailable": {
			f:                &Flag{v: pointer.ToBool(true)},
			telemetryReponse: `{"update_available": true, "latest_version": "0.3.4"}`,
			latestVersion:    "0.3.4",
			updateAvailable:  true,
		},
		"UpdateUnavailable": {
			f:                &Flag{v: pointer.ToBool(true)},
			telemetryReponse: `{"update_available": false, "latest_version": "0.3.4"}`,
			latestVersion:    "",
			updateAvailable:  false,
		},
		"TelemetryDisabled": {
			f:               &Flag{v: pointer.ToBool(false)},
			latestVersion:   "",
			updateAvailable: false,
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// use httptest.NewServer to mock telemetry response for http POST.
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					w.WriteHeader(http.StatusCreated)
					fmt.Fprintln(w, tc.telemetryReponse)
				}
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

			// check initial state of provider, it has not called telemetry yet,
			// no update is available and unaware of the latest version.
			s := r.P.Get()
			require.False(t, s.UpdateAvailable())
			require.Equal(t, "", s.LatestVersion)

			// call report to update the state of provider from telemetry.
			r.report(testutil.Ctx(t))

			// get updated the state of provider.
			s = r.P.Get()
			require.Equal(t, tc.updateAvailable, s.UpdateAvailable())
			require.Equal(t, tc.latestVersion, s.LatestVersion)
		})
	}
}
