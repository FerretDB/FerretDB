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

package tjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name   string
	v      any            // build-in
	j      string         // tigris
	s      map[string]any // schema
	canonJ string         // canonical form without extra object fields, zero values, etc.
	jErr   string         // unwrapped
}

// assertEqualWithNaN is assert.Equal that also can compare NaNs.
func assertEqualWithNaN(t testing.TB, expected, actual any) {
	t.Helper()

	if expectedD, ok := expected.(*doubleType); ok {
		require.IsType(t, expected, actual)
		actualD := actual.(*doubleType)
		if math.IsNaN(float64(*expectedD)) {
			assert.True(t, math.IsNaN(float64(*actualD)))
			return
		}
	}

	assert.Equal(t, expected, actual, "expected: %s\nactual  : %s", expected, actual)
}

// lastErr returns the last error in error chain.
func lastErr(err error) error {
	for {
		e := errors.Unwrap(err)
		if e == nil {
			return err
		}
		err = e
	}
}

func testJSON(t *testing.T, testCases []testCase, newFunc func() tjsontype) {
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

			t.Run("Marshal", func(t *testing.T) {
				t.Parallel()

				actualBuildIn, err := Marshal([]byte(tc.j), tc.s)

				if tc.jErr == "" {
					require.NoError(t, err)
					assertEqualWithNaN(t, tc.v, actualBuildIn)
					return
				}

				require.Error(t, err)
				require.Equal(t, tc.jErr, lastErr(err).Error())
			})

			t.Run("Unmarshal", func(t *testing.T) {
				if tc.v == nil {
					t.Skip("v is nil")
				}
				t.Parallel()

				actualJ, err := Unmarshal(tc.v, tc.s)
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

				actualF, err := Marshal([]byte(tc.j), tc.s)

				if tc.jErr != "" {
					require.Error(t, err)
					require.Equal(t, tc.jErr, lastErr(err).Error())
					return
				}

				require.NoError(t, err)
				expectedF := tc.v

				assert.Equal(t, expectedF, actualF)
			})
		})
	}
}

func fuzzJSON(f *testing.F, testCases []testCase, newFunc func() tjsontype) {
	for _, tc := range testCases {
		f.Add(tc.j)
		if tc.canonJ != "" {
			f.Add(tc.canonJ)
		}
	}

	f.Fuzz(func(t *testing.T, j string, s map[string]any) {
		t.Parallel()

		// raw "null" should never reach Unmarshal due to the way encoding/json works
		if j == "null" {
			t.Skip()
		}

		// j may not be a canonical form.
		// We can't compare it with Marshal() result directly.
		// Instead, we compare second results.

		v := newFunc()
		{
			err := v.Marshal([]byte(j), s)
			require.NoError(t, err)
		}

		// test Unmarshal
		{
			b, err := v.Unmarshal(s)
			require.NoError(t, err)
			j = string(b)
		}

		// test Marshal
		{
			actualV := newFunc()
			err := actualV.Marshal([]byte(j), s)
			require.NoError(t, err)
			assertEqualWithNaN(t, v, actualV)
		}
	})
}

func benchmark(b *testing.B, testCases []testCase, newFunc func() tjsontype) {
	for _, tc := range testCases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			b.Run("Marshal", func(b *testing.B) {
				data := []byte(tc.j)
				var v tjsontype
				var err error

				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					v = newFunc()
					err = v.Marshal(data, tc.s)
				}

				b.StopTimer()

				if tc.jErr == "" {
					require.NoError(b, err)
					assertEqualWithNaN(b, tc.v, v)
					return
				}

				require.Error(b, err)
				require.Equal(b, tc.jErr, lastErr(err).Error())
			})
		})
	}
}
