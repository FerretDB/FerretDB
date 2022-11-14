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
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestState(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		flag     string
		dnt      string
		execName string
		prev     *bool
		locked   bool
		state    *bool
		err      string
	}{
		"default": {},
		"prev": {
			prev:   pointer.ToBool(false),
			locked: false,
			state:  pointer.ToBool(false),
		},
		"flag": {
			flag:   "disable",
			prev:   pointer.ToBool(true),
			locked: true,
			state:  pointer.ToBool(false),
		},
		"dnt": {
			dnt:    "1",
			state:  pointer.ToBool(false),
			locked: true,
		},
		"invalidDnt": {
			dnt:    "foo",
			locked: false,
			err:    "failed to parse foo",
		},
		"conflict": {
			flag:     "enable",
			execName: "DoNotTrack",
			locked:   false,
			err:      "telemetry can't be enabled",
		},
	} {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var f Flag
			err := f.UnmarshalText([]byte(tc.flag))
			require.NoError(t, err)

			logger := testutil.Logger(t, zap.NewAtomicLevelAt(zap.DebugLevel))
			actualState, locked, actualErr := initialState(&f, tc.dnt, tc.execName, tc.prev, logger)
			assert.Equal(t, tc.state, actualState)
			assert.Equal(t, tc.locked, locked)
			if tc.err != "" {
				assert.EqualError(t, actualErr, tc.err)
				return
			}
			assert.NoError(t, actualErr)
		})
	}
}
