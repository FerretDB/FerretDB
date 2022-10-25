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
		flag          string
		dnt           string
		execName      string
		expectedState *bool
		expectedErr   string
	}{
		"default": {},
		"flag": {
			flag:          "disable",
			expectedState: pointer.ToBool(false),
		},
		"dnt": {
			dnt:           "1",
			expectedState: pointer.ToBool(false),
		},
		"conflict": {
			flag:        "enable",
			execName:    "DoNotTrack",
			expectedErr: "telemetry is disabled by DO_NOT_TRACK environment variable or executable name",
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var f Flag
			err := f.UnmarshalText([]byte(tc.flag))
			require.NoError(t, err)

			logger := testutil.Logger(t, zap.NewAtomicLevelAt(zap.DebugLevel))
			actualState, actualErr := State(&f, tc.dnt, tc.execName, logger)
			assert.Equal(t, tc.expectedState, actualState)
			if tc.expectedErr != "" {
				assert.EqualError(t, actualErr, tc.expectedErr)
				return
			}
			assert.NoError(t, actualErr)
		})
	}
}
