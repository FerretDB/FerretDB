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

package common

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiplyLongSafely(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		values []int64 // [v1,v2,v1+v2] or [v1,v2] if you don't want to check multiplication result

		// overflow checks if at least one value returned errLongExceeded error.
		overflow bool // defaults to false

		// valuesRange allows to execute test for the numbers next to v1.
		// if v1=42 and valuesRange=2, following numbers will be tested: [42,43,44]
		// It shouldn't be used with multiplication result, but rather to test overflows on large ranges.
		valuesRange int
	}{
		"Zero": {
			values: []int64{
				0, 1000, 0,
				1000, 0, 0,
				0, 0, 0,
			},
		},
		"One": {
			values: []int64{
				100, 1, 100,
				1, 42, 42,
				1, 1, 1,
				1, math.MaxInt64, math.MaxInt64,
			},
		},
		"DoublePrecision": {
			values:      []int64{1 << 53, 42},
			valuesRange: 1000,
		},
		"OverflowLarge": {
			values:   []int64{1 << 60, 42},
			overflow: true,
		},
		"OverflowMax": {
			values:   []int64{math.MaxInt64, 2},
			overflow: true,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			valuesLen := len(tc.values)
			resultEnabled := valuesLen%3 == 0

			require.GreaterOrEqual(t, valuesLen, 2, "values must be set with at least 2 values [v1,v2]")
			require.True(
				t,
				resultEnabled || valuesLen%2 == 0,
				"values must contain 2 values [v1,v2] for each subtest, or 3 [v1,v2,v1+v2] "+
					"if we test multiplication results",
			)

			if tc.valuesRange != 0 {
				require.Positive(t, tc.valuesRange)
			}

			inc := 2
			if resultEnabled {
				inc = 3
			}

			var v1, v2, v3 int64
			var overflowActual bool

			for i := 0; i < valuesLen; i += inc {
				v1, v2 = tc.values[i], tc.values[i+1]

				for j := 0; j < 1+tc.valuesRange; j++ {
					// for easier debugging
					v1 := v1 + int64(j)

					actualRes, err := multiplyLongSafely(v1, v2)

					if err != nil && assert.ErrorIs(t, err, errLongExceeded) {
						overflowActual = true
					}

					if resultEnabled {
						v3 = tc.values[i+2]
						assert.Equal(t, v3, actualRes)
					}
				}
			}

			assert.Equal(t, tc.overflow, overflowActual)
		})
	}
}
