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

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
)

func TestMultiplyLongSafely(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		v1, v2   int64
		expected *int64

		// overflow checks if at least one value returned errLongExceeded error.
		err error
	}{
		"Zero": {
			v1:       0,
			v2:       1000,
			expected: pointer.ToInt64(0),
		},
		"One": {
			v1:       42,
			v2:       1,
			expected: pointer.ToInt64(42),
		},
		"DoubleMaxPrecision": {
			v1: 1 << 53,
			v2: 42,
		},
		"DoubleMaxPrecisionPlus": {
			v1: (1 << 53) + 1,
			v2: 42,
		},
		"OverflowLarge": {
			v1:  1 << 60,
			v2:  42,
			err: errLongExceeded,
		},
		"OverflowMax": {
			v1:  math.MaxInt64,
			v2:  2,
			err: errLongExceeded,
		},
		"MaxMinusOne": {
			v1: math.MaxInt64,
			v2: -1,
		},
		"OverflowMaxMinusTwo": {
			v1:  math.MaxInt64,
			v2:  -2,
			err: errLongExceeded,
		},
		"OverflowMin": {
			v1:  math.MinInt64,
			v2:  2,
			err: errLongExceeded,
		},
		"OverflowMinMinusOne": {
			v1:  math.MinInt64,
			v2:  -1,
			err: errLongExceeded,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actualRes, err := multiplyLongSafely(tc.v1, tc.v2)
			assert.Equal(t, tc.err, err)

			if tc.expected != nil {
				assert.Equal(t, *tc.expected, actualRes)
			}
		})
	}
}
