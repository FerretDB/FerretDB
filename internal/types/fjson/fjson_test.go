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
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name   string
	v      fjsontype
	j      string
	canonJ string // canonical form without extra object fields, zero values, etc.
	jErr   string // unwrapped
}

func testJSON(t *testing.T, testCases []testCase, newFunc func() fjsontype) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.name, "name should not be empty")
			require.NotEmpty(t, tc.j, "j should not be empty")

			t.Parallel()

			if tc.jErr == "" {
				var dst bytes.Buffer
				require.NoError(t, json.Compact(&dst, []byte(tc.j)))
				require.Equal(t, tc.j, dst.String(), "j should be compacted")
				if tc.canonJ != "" {
					dst.Reset()
					require.NoError(t, json.Compact(&dst, []byte(tc.canonJ)))
					require.Equal(t, tc.canonJ, dst.String(), "canonJ should be compacted")
				}
			}

			t.Run("MarshalJSON", func(t *testing.T) {
				if tc.v == nil {
					t.Skip("v is nil")
				}

				t.Parallel()

				actualJ, err := tc.v.MarshalJSON()
				require.NoError(t, err)
				expectedJ := tc.j
				if tc.canonJ != "" {
					expectedJ = tc.canonJ
				}
				assert.Equal(t, expectedJ, string(actualJ))
			})

			t.Run("Marshal", func(t *testing.T) {
				if tc.v == nil {
					t.Skip("v is nil")
				}

				t.Parallel()

				actualJ, err := Marshal(fromFJSON(tc.v))
				require.NoError(t, err)
				expectedJ := tc.j
				if tc.canonJ != "" {
					expectedJ = tc.canonJ
				}
				assert.Equal(t, expectedJ, string(actualJ))
			})
		})
	}
}
