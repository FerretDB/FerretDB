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

package fjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name string
	v    fjsontype
	j    string
}

func testJSON(t *testing.T, testCases []testCase, newFunc func() fjsontype) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.name, "name should not be empty")
			require.NotNil(t, tc.v, "v should not be nil")
			require.NotEmpty(t, tc.j, "j should not be empty")

			t.Parallel()

			t.Run("MarshalJSON", func(t *testing.T) {
				t.Parallel()

				actualJ, err := tc.v.MarshalJSON()
				require.NoError(t, err)
				assert.Equal(t, tc.j, string(actualJ))
			})

			t.Run("Marshal", func(t *testing.T) {
				t.Parallel()

				actualJ, err := Marshal(fromFJSON(tc.v))
				require.NoError(t, err)
				assert.Equal(t, tc.j, string(actualJ))
			})
		})
	}
}
