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
	"bytes"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/types"
)

// AssertEqual asserts that two BSON values are equal.
func AssertEqual[T types.Type](tb testing.TB, expected, actual T) bool {
	tb.Helper()

	if equal(tb, expected, actual) {
		return true
	}

	expectedS, actualS, diff := diffValues(tb, expected, actual)
	msg := fmt.Sprintf("Not equal: \nexpected: %s\nactual  : %s\n%s", expectedS, actualS, diff)
	return assert.Fail(tb, msg)
}

// AssertEqualSlices asserts that two BSON slices are equal.
func AssertEqualSlices[T types.Type](tb testing.TB, expected, actual []T) bool {
	tb.Helper()

	allEqual := len(expected) == len(actual)
	if allEqual {
		for i, e := range expected {
			a := actual[i]
			if !equal(tb, e, a) {
				allEqual = false
				break
			}
		}
	}

	if allEqual {
		return true
	}

	expectedS, actualS, diff := diffSlices(tb, expected, actual)
	msg := fmt.Sprintf("Not equal: \nexpected: %s\nactual  : %s\n%s", expectedS, actualS, diff)
	return assert.Fail(tb, msg)
}

// EqualValue returns true when two BSON values are equal,
// even if the numbers are different types.
func EqualValue(tb testing.TB, expected, actual any) bool {
	return equalValue(tb, expected, actual)
}

// AssertNotEqual asserts that two BSON values are not equal.
func AssertNotEqual[T types.Type](tb testing.TB, expected, actual T) bool {
	tb.Helper()

	if !equal(tb, expected, actual) {
		return true
	}

	// The diff of equal values should be empty, but produce it anyway to catch subtle bugs.
	expectedS, actualS, diff := diffValues(tb, expected, actual)
	msg := fmt.Sprintf("Unexpected equal: \nexpected: %s\nactual  : %s\n%s", expectedS, actualS, diff)
	return assert.Fail(tb, msg)
}

// AssertNotEqualSlices asserts that two BSON slices are not equal.
func AssertNotEqualSlices[T types.Type](tb testing.TB, expected, actual []T) bool {
	tb.Helper()

	allEqual := len(expected) == len(actual)
	if allEqual {
		for i, e := range expected {
			a := actual[i]
			if !equal(tb, e, a) {
				allEqual = false
				break
			}
		}
	}

	if !allEqual {
		return true
	}

	expectedS, actualS, diff := diffSlices(tb, expected, actual)
	msg := fmt.Sprintf("Unexpected equal: \nexpected: %s\nactual  : %s\n%s", expectedS, actualS, diff)
	return assert.Fail(tb, msg)
}

// diffValues returns a readable form of given values and the difference between them.
func diffValues[T types.Type](tb testing.TB, expected, actual T) (expectedS string, actualS string, diff string) {
	expectedS = Dump(tb, expected)
	actualS = Dump(tb, actual)

	var err error
	diff, err = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expectedS),
		FromFile: "expected",
		B:        difflib.SplitLines(actualS),
		ToFile:   "actual",
		Context:  1,
	})
	require.NoError(tb, err)

	return
}

// diffSlices returns a readable form of given slices and the difference between them.
func diffSlices[T types.Type](tb testing.TB, expected, actual []T) (expectedS string, actualS string, diff string) {
	expectedS = DumpSlice(tb, expected)
	actualS = DumpSlice(tb, actual)

	var err error
	diff, err = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expectedS),
		FromFile: "expected",
		B:        difflib.SplitLines(actualS),
		ToFile:   "actual",
		Context:  1,
	})
	require.NoError(tb, err)

	return
}

// equal compares any BSON values in a way that is useful for tests:
//   - float64 NaNs are equal to each other;
//   - float64 zero values are compared with sign (math.Copysign(0, -1) != math.Copysign(0, +1));
//   - time.Time values are compared using Equal method.
//
// This function is for tests; it should not try to convert values to different types before comparing them.
//
// Compare and contrast with types.Compare function.
func equal(tb testing.TB, v1, v2 any) bool {
	tb.Helper()

	switch v1 := v1.(type) {
	case *types.Document:
		d, ok := v2.(*types.Document)
		if !ok {
			return false
		}
		return equalDocuments(tb, v1, d)

	case *types.Array:
		a, ok := v2.(*types.Array)
		if !ok {
			return false
		}
		return equalArrays(tb, v1, a)

	default:
		return equalScalars(tb, v1, v2)
	}
}

// equalDocuments compares BSON documents.
func equalDocuments(tb testing.TB, v1, v2 *types.Document) bool {
	tb.Helper()

	require.NotNil(tb, v1)
	require.NotNil(tb, v2)

	keys := v1.Keys()
	if !slices.Equal(keys, v2.Keys()) {
		return false
	}

	for _, k := range keys {
		f1, err := v1.Get(k)
		require.NoError(tb, err)

		f2, err := v2.Get(k)
		require.NoError(tb, err)

		if !equal(tb, f1, f2) {
			return false
		}
	}

	return true
}

// equalArrays compares BSON arrays.
func equalArrays(tb testing.TB, v1, v2 *types.Array) bool {
	tb.Helper()

	require.NotNil(tb, v1)
	require.NotNil(tb, v2)

	l := v1.Len()
	if l != v2.Len() {
		return false
	}

	for i := 0; i < l; i++ {
		el1, err := v1.Get(i)
		require.NoError(tb, err)

		el2, err := v2.Get(i)
		require.NoError(tb, err)

		if !equal(tb, el1, el2) {
			return false
		}
	}

	return true
}

// equalScalars compares BSON scalar values.
func equalScalars(tb testing.TB, v1, v2 any) bool {
	tb.Helper()

	require.NotNil(tb, v1)
	require.NotNil(tb, v2)

	switch s1 := v1.(type) {
	case float64:
		s2, ok := v2.(float64)
		if !ok {
			return false
		}
		if math.IsNaN(s1) {
			return math.IsNaN(s2)
		}
		if s1 == 0 && s2 == 0 {
			return math.Signbit(s1) == math.Signbit(s2)
		}
		return s1 == s2

	case string:
		s2, ok := v2.(string)
		if !ok {
			return false
		}
		return s1 == s2

	case types.Binary:
		s2, ok := v2.(types.Binary)
		if !ok {
			return false
		}
		return s1.Subtype == s2.Subtype && bytes.Equal(s1.B, s2.B)

	case types.ObjectID:
		s2, ok := v2.(types.ObjectID)
		if !ok {
			return false
		}
		return s1 == s2

	case bool:
		s2, ok := v2.(bool)
		if !ok {
			return false
		}
		return s1 == s2

	case time.Time:
		s2, ok := v2.(time.Time)
		if !ok {
			return false
		}
		return s1.Equal(s2)

	case types.NullType:
		_, ok := v2.(types.NullType)
		return ok

	case types.Regex:
		s2, ok := v2.(types.Regex)
		if !ok {
			return false
		}
		return s1.Pattern == s2.Pattern && s1.Options == s2.Options

	case int32:
		s2, ok := v2.(int32)
		if !ok {
			return false
		}
		return s1 == s2

	case types.Timestamp:
		s2, ok := v2.(types.Timestamp)
		if !ok {
			return false
		}
		return s1 == s2

	case int64:
		s2, ok := v2.(int64)
		if !ok {
			return false
		}
		return s1 == s2

	default:
		tb.Fatalf("unhandled types %T, %T", v1, v2)
		panic("not reached")
	}
}

// equalValue compares any BSON values in a way that is useful for tests:
//   - number types are converted so different type int32, int64, float64 are equal for the value.
//
// For more see equal.
func equalValue(tb testing.TB, v1, v2 any) bool {
	tb.Helper()

	switch v1 := v1.(type) {
	case *types.Document:
		d, ok := v2.(*types.Document)
		if !ok {
			return false
		}

		return equalValueDocuments(tb, v1, d)

	case *types.Array:
		a, ok := v2.(*types.Array)
		if !ok {
			return false
		}

		return equalValueArrays(tb, v1, a)

	default:
		return equalValueScalars(tb, v1, v2)
	}
}

// equalValueDocuments compares BSON documents.
func equalValueDocuments(tb testing.TB, v1, v2 *types.Document) bool {
	tb.Helper()

	require.NotNil(tb, v1)
	require.NotNil(tb, v2)

	keys := v1.Keys()
	if !slices.Equal(keys, v2.Keys()) {
		return false
	}

	for _, k := range keys {
		f1, err := v1.Get(k)
		require.NoError(tb, err)

		f2, err := v2.Get(k)
		require.NoError(tb, err)

		if !equalValue(tb, f1, f2) {
			return false
		}
	}

	return true
}

// equalValueArrays compares BSON arrays.
func equalValueArrays(tb testing.TB, v1, v2 *types.Array) bool {
	tb.Helper()

	require.NotNil(tb, v1)
	require.NotNil(tb, v2)

	l := v1.Len()
	if l != v2.Len() {
		return false
	}

	for i := 0; i < l; i++ {
		el1, err := v1.Get(i)
		require.NoError(tb, err)

		el2, err := v2.Get(i)
		require.NoError(tb, err)

		if !equalValue(tb, el1, el2) {
			return false
		}
	}

	return true
}

// equalValueScalars compares BSON scalar values.
func equalValueScalars(tb testing.TB, v1, v2 any) bool {
	tb.Helper()

	require.NotNil(tb, v1)
	require.NotNil(tb, v2)

	switch s1 := v1.(type) {
	case float64:
		switch s2 := v2.(type) {
		case float64:
			if math.IsNaN(s1) {
				return math.IsNaN(s2)
			}

			if s1 == 0 && s2 == 0 {
				return math.Signbit(s1) == math.Signbit(s2)
			}

			return s1 == s2
		case int32:
			return equal(tb, s1, float64(s2))
		case int64:
			return equal(tb, s1, float64(s2))
		default:
			return false
		}

	case string:
		return equal(tb, v1, v2)

	case types.Binary:
		return equal(tb, v1, v2)

	case types.ObjectID:
		return equal(tb, v1, v2)

	case bool:
		return equal(tb, v1, v2)

	case time.Time:
		return equal(tb, v1, v2)

	case types.NullType:
		return equal(tb, v1, v2)

	case types.Regex:
		return equal(tb, v1, v2)

	case int32:
		switch s2 := v2.(type) {
		case float64:
			if math.IsNaN(s2) {
				return false
			}

			return float64(s1) == s2
		case int32:
			return s1 == s2
		case int64:
			return int64(s1) == s2
		default:
			return false
		}

	case types.Timestamp:
		return equal(tb, v1, v2)

	case int64:
		switch s2 := v2.(type) {
		case float64:
			if math.IsNaN(s2) {
				return false
			}

			return float64(s1) == s2
		case int32:
			return s1 == int64(s2)
		case int64:
			return s1 == s2
		default:
			return false
		}
	default:
		tb.Fatalf("unhandled types %T, %T", v1, v2)
		panic("not reached")
	}
}
