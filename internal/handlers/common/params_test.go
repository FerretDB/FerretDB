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
)

func TestMultiplyLongSafely(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		err              error
		v1, v2, expected int64
	}{
		"Zero": {
			v1:       0,
			v2:       1000,
			expected: 0,
		},
		"One": {
			v1:       42,
			v2:       1,
			expected: 42,
		},
		"DoubleMaxPrecision": {
			v1:       1 << 53,
			v2:       42,
			expected: 378302368699121664,
		},
		"DoubleMaxPrecisionPlus": {
			v1:       (1 << 53) + 1,
			v2:       42,
			expected: 378302368699121706,
		},
		"OverflowLarge": {
			v1:  1 << 60,
			v2:  42,
			err: errLongExceededPositive,
		},
		"OverflowMax": {
			v1:  math.MaxInt64,
			v2:  2,
			err: errLongExceededPositive,
		},
		"MaxMinusOne": {
			v1:       math.MaxInt64,
			v2:       -1,
			expected: -math.MaxInt64,
		},
		"OverflowMaxMinusTwo": {
			v1:  math.MaxInt64,
			v2:  -2,
			err: errLongExceededPositive,
		},
		"OverflowMin": {
			v1:  math.MinInt64,
			v2:  2,
			err: errLongExceededNegative,
		},
		"OverflowMinMinusOne": {
			v1:  math.MinInt64,
			v2:  -1,
			err: errLongExceededNegative,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actualRes, err := multiplyLongSafely(tc.v1, tc.v2)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, actualRes)
		})
	}
}
