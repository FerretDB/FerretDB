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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		flag          string
		dnt           string
		execName      string
		expectedState *bool
		expectedErr   error
	}{
		"default": {},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var f Flag
			err := f.UnmarshalText([]byte(tc.flag))
			require.NoError(t, err)

			actualState, actualErr := State(&f, tc.dnt, tc.execName, nil)
			assert.Equal(t, tc.expectedState, actualState)
			assert.Equal(t, tc.expectedErr, actualErr)
		})
	}
}
