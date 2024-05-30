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
	"encoding/json"
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

			sp, err := state.NewProvider("")
			require.NoError(t, err)

			opts := NewReporterOpts{
				F:           tc.f,
				DNT:         tc.dnt,
				ExecName:    tc.execName,
				ConnMetrics: connmetrics.NewListenerMetrics().ConnMetrics,
				P:           sp,
				L:           zap.L(),
			}

			_, err = NewReporter(&opts)
			assert.NoError(t, err)

			s := sp.Get()
			assert.Equal(t, tc.t, s.Telemetry)
			assert.Equal(t, tc.locked, s.TelemetryLocked)
		})
	}
}

// beaconServer returns a httptest.Server that emulates beacon server.
func beaconServer(t *testing.T, calls *int, res *response) *httptest.Server {
	t.Helper()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*calls++

		w.WriteHeader(http.StatusCreated)
		require.NoError(t, json.NewEncoder(w).Encode(res))
	}))

	t.Cleanup(s.Close)

	return s
}

func TestReporterReport(t *testing.T) {
	t.Parallel()

	t.Run("TelemetryEnabled", func(t *testing.T) {
		t.Parallel()

		var serverCalled int
		telemetryResponse := response{
			LatestVersion:   "v1.2.1",
			UpdateAvailable: true,
		}
		bs := beaconServer(t, &serverCalled, &telemetryResponse)

		sp, err := state.NewProvider("")
		require.NoError(t, err)

		opts := NewReporterOpts{
			URL:           bs.URL,
			F:             &Flag{v: pointer.ToBool(true)},
			ConnMetrics:   connmetrics.NewListenerMetrics().ConnMetrics,
			P:             sp,
			L:             zap.L(),
			ReportTimeout: 1 * time.Minute,
		}

		r, err := NewReporter(&opts)
		require.NoError(t, err)

		s := r.P.Get()
		assert.False(t, s.UpdateAvailable)
		assert.Empty(t, s.LatestVersion)

		r.report(testutil.Ctx(t))
		assert.Equal(t, 1, serverCalled)

		s = r.P.Get()
		assert.True(t, s.UpdateAvailable)
		assert.Equal(t, "v1.2.1", s.LatestVersion)

		telemetryResponse.UpdateAvailable = false

		r.report(testutil.Ctx(t))
		assert.Equal(t, 2, serverCalled)

		s = r.P.Get()
		assert.False(t, s.UpdateAvailable)
		assert.Equal(t, "v1.2.1", s.LatestVersion)

		telemetryResponse.UpdateAvailable = true
		telemetryResponse.LatestVersion = "v1.2.0"

		r.report(testutil.Ctx(t))
		assert.Equal(t, 3, serverCalled)

		s = r.P.Get()
		assert.True(t, s.UpdateAvailable)
		assert.Equal(t, "v1.2.0", s.LatestVersion)

		require.NoError(t, sp.Update(func(s *state.State) { s.DisableTelemetry() }))

		r.report(testutil.Ctx(t))
		assert.Equal(t, 3, serverCalled)

		s = r.P.Get()
		assert.False(t, s.UpdateAvailable)
		assert.Empty(t, s.LatestVersion)

		require.NoError(t, sp.Update(func(s *state.State) { s.EnableTelemetry() }))
		telemetryResponse.LatestVersion = "v1.2.2"

		r.report(testutil.Ctx(t))
		assert.Equal(t, 4, serverCalled)

		s = r.P.Get()
		assert.True(t, s.UpdateAvailable)
		assert.Equal(t, "v1.2.2", s.LatestVersion)
	})

	t.Run("TelemetryDisabled", func(t *testing.T) {
		t.Parallel()

		var serverCalled int
		telemetryResponse := response{
			LatestVersion:   "v1.2.1",
			UpdateAvailable: true,
		}
		bs := beaconServer(t, &serverCalled, &telemetryResponse)

		sp, err := state.NewProvider("")
		require.NoError(t, err)

		opts := NewReporterOpts{
			URL:           bs.URL,
			F:             &Flag{v: pointer.ToBool(false)},
			ConnMetrics:   connmetrics.NewListenerMetrics().ConnMetrics,
			P:             sp,
			L:             zap.L(),
			ReportTimeout: 1 * time.Minute,
		}

		r, err := NewReporter(&opts)
		require.NoError(t, err)

		s := r.P.Get()
		assert.False(t, s.UpdateAvailable)
		assert.Empty(t, s.LatestVersion)

		r.report(testutil.Ctx(t))
		assert.Equal(t, 0, serverCalled)

		s = r.P.Get()
		assert.False(t, s.UpdateAvailable)
		assert.Empty(t, s.LatestVersion)
	})
}
