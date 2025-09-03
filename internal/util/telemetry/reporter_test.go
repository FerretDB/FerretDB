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
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestReporterLocked(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		f        *Flag
		dnt      string
		execName string
		t        *bool
		locked   bool
	}{
		"NoSet": {
			f: NewFlag(nil),
		},
		"FlagEnable": {
			f:      NewFlag(pointer.ToBool(true)),
			t:      pointer.ToBool(true),
			locked: true,
		},
		"FlagDisable": {
			f:      NewFlag(pointer.ToBool(false)),
			t:      pointer.ToBool(false),
			locked: true,
		},
		"DoNotTrack": {
			f:      NewFlag(nil),
			dnt:    "enable",
			t:      pointer.ToBool(false),
			locked: true,
		},
		"ExecName": {
			f:        NewFlag(nil),
			execName: "exec_donottrack",
			t:        pointer.ToBool(false),
			locked:   true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sp, err := state.NewProvider("")
			require.NoError(t, err)

			_, err = NewReporter(&NewReporterOpts{
				URL:      "http://127.0.0.1:1/",
				Dir:      t.TempDir(),
				F:        tc.f,
				DNT:      tc.dnt,
				ExecName: tc.execName,
				Metrics:  middleware.NewMetrics(),
				P:        sp,
				L:        testutil.Logger(t),
			})
			assert.NoError(t, err)

			s := sp.Get()
			assert.Equal(t, tc.t, s.Telemetry)
			assert.Equal(t, tc.locked, s.TelemetryLocked)
		})
	}
}

func TestMakeReport(t *testing.T) {
	t.Parallel()

	m := map[string]map[string]map[string]middleware.CommandMetrics{
		"OP_MSG": {
			"update": {
				"$set": middleware.CommandMetrics{
					Failures: map[string]int{
						"NotImplemented": 1,
						"panic":          1,
					},
					Total: 3,
				},
			},
			"find": {
				"unknown": middleware.CommandMetrics{
					Total: 1,
				},
			},
			"atlasVersion": {
				"unknown": middleware.CommandMetrics{
					Failures: map[string]int{
						"CommandNotFound": 1,
					},
					Total: 1,
				},
			},
		},
	}

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	tr, err := NewReporter(&NewReporterOpts{
		URL: "http://127.0.0.1:1/",
		Dir: t.TempDir(),
		F:   NewFlag(pointer.ToBool(false)),
		P:   sp,
		L:   testutil.Logger(t),
	})
	assert.NoError(t, err)

	expected := map[string]map[string]map[string]map[string]int{
		"OP_MSG": {
			"update": {
				"$set": map[string]int{
					"NotImplemented": 1,
					"panic":          1,
					"ok":             1,
				},
			},
			"find": {
				"unknown": {
					"ok": 1,
				},
			},
			"atlasVersion": {
				"unknown": {
					"CommandNotFound": 1,
				},
			},
		},
	}
	assert.Equal(t, expected, tr.makeReport(m).CommandMetrics)
}
