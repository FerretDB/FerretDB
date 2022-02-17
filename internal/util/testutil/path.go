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

package testutil

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
)

// GetByPath returns a value by path - a sequence of indexes and keys.
func GetByPath[T types.CompositeTypeInterface](tb testing.TB, comp T, path ...string) any {
	tb.Helper()

	res, err := comp.GetByPath(path...)
	require.NoError(tb, err)
	return res
}

// SetByPath sets the value by path - a sequence of indexes and keys.
//
// The path must exist.
func SetByPath[T types.CompositeTypeInterface](tb testing.TB, comp T, value any, path ...string) {
	tb.Helper()

	l := len(path)
	require.NotZero(tb, l, "path is empty")

	var next any = comp
	for i, p := range path {
		last := i == l-1
		switch c := next.(type) {
		case *types.Document:
			var err error
			next, err = c.Get(p)
			require.NoError(tb, err)

			if last {
				err = c.Set(p, value)
				require.NoError(tb, err)
			}

		case *types.Array:
			index, err := strconv.Atoi(p)
			require.NoError(tb, err)

			next, err = c.Get(index)
			require.NoError(tb, err)

			if last {
				err = c.Set(index, value)
				require.NoError(tb, err)
			}

		default:
			tb.Fatalf("can't access %T by path %q", next, p)
		}
	}
}

// CompareAndSetByPathNum asserts that two values with the same path in two objects (documents or arrays)
// are within a given numerical delta, then updates the expected object with the actual value.
func CompareAndSetByPathNum[T types.CompositeTypeInterface](tb testing.TB, expected, actual T, delta float64, path ...string) {
	tb.Helper()

	expectedV := GetByPath(tb, expected, path...)
	actualV := GetByPath(tb, actual, path...)
	assert.IsType(tb, expectedV, actualV)
	assert.InDelta(tb, expectedV, actualV, delta)

	SetByPath(tb, expected, actualV, path...)
}

// CompareAndSetByPathTime asserts that two values with the same path in two objects (documents or arrays)
// are within a given time delta, then updates the expected object with the actual value.
//
//nolint:lll // will be fixed when CompositeTypeInterface is removed
func CompareAndSetByPathTime[T types.CompositeTypeInterface](tb testing.TB, expected, actual T, delta time.Duration, path ...string) {
	tb.Helper()

	expectedV := GetByPath(tb, expected, path...)
	actualV := GetByPath(tb, actual, path...)
	assert.IsType(tb, expectedV, actualV)
	require.IsType(tb, time.Time{}, actualV)
	assert.WithinDuration(tb, expectedV.(time.Time), actualV.(time.Time), delta)

	SetByPath(tb, expected, actualV, path...)
}
