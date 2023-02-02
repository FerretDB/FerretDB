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
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
)

// GetAllByPath returns a value by path - a sequence of indexes and keys.
func GetAllByPath[T types.CompositeTypeInterface](tb testing.TB, comp T, path types.Path) []any {
	tb.Helper()

	res, err := comp.GetAllByPath(path, false)
	require.NoError(tb, err)
	return res
}

// SetByPath sets the value by path - a sequence of indexes and keys.
//
// The path must exist.
func SetByPath[T types.CompositeTypeInterface](tb testing.TB, comp T, value any, path types.Path) {
	tb.Helper()

	l := path.Len()
	require.NotZero(tb, l, "path is empty")

	var next any = comp
	for i, p := range path.Slice() {
		last := i == l-1
		switch c := next.(type) {
		case *types.Document:
			var err error
			next, err = c.Get(p)
			require.NoError(tb, err)

			if last {
				c.Set(p, value)
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
func CompareAndSetByPathNum[T types.CompositeTypeInterface](tb testing.TB, expected, actual T, delta float64, path types.Path) {
	tb.Helper()

	expectedVs := GetAllByPath(tb, expected, path)
	actualVs := GetAllByPath(tb, actual, path)

	require.Len(tb, expectedVs, 1, "expected component must have exactly one element matching the path %q", path)
	require.Len(tb, actualVs, 1, "actual component must have exactly one element matching the path %q", path)

	expectedV, actualV := expectedVs[0], actualVs[0]

	assert.IsType(tb, expectedV, actualV)
	assert.InDelta(tb, expectedV, actualV, delta)

	SetByPath(tb, expected, actualV, path)
}

// CompareAndSetByPathTime asserts that two values with the same path in two objects (documents or arrays)
// are within a given time delta, then updates the expected object with the actual value.
//
//nolint:lll // will be fixed when CompositeTypeInterface is removed
func CompareAndSetByPathTime[T types.CompositeTypeInterface](tb testing.TB, expected, actual T, delta time.Duration, path types.Path) {
	tb.Helper()

	expectedVs := GetAllByPath(tb, expected, path)
	actualVs := GetAllByPath(tb, actual, path)

	require.Len(tb, expectedVs, 1, "expected component must have exactly one element matching the path %q", path)
	require.Len(tb, actualVs, 1, "actual component must have exactly one element matching the path %q", path)

	expectedV, actualV := expectedVs[0], actualVs[0]

	assert.IsType(tb, expectedV, actualV)

	switch actualV := actualV.(type) {
	case time.Time:
		assert.WithinDuration(tb, expectedV.(time.Time), actualV, delta)

	case types.Timestamp:
		expectedT := expectedV.(types.Timestamp).Time()
		actualT := actualV.Time()
		assert.WithinDuration(tb, expectedT, actualT, delta)

	default:
		assert.Fail(tb, fmt.Sprintf("expected time.Time or types.Timestamp, got %T %T", expected, actual))
	}

	SetByPath(tb, expected, actualV, path)
}
