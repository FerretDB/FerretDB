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
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
)

// AssertEqual asserts that two BSON values are equal.
func AssertEqual[T types.Type](t testing.TB, expected, actual T) bool {
	t.Helper()

	if types.Equal(expected, actual) {
		return true
	}

	expectedS, actualS, diff := diffValues(t, expected, actual)
	msg := fmt.Sprintf("Not equal: \nexpected: %s\nactual  : %s\n%s", expectedS, actualS, diff)
	return assert.Fail(t, msg)
}

// AssertNotEqual asserts that two BSON values are not equal.
func AssertNotEqual[T types.Type](t testing.TB, expected, actual T) bool {
	t.Helper()

	if !types.Equal(expected, actual) {
		return true
	}

	// The diff of equal values should be empty, but produce it anyway to catch subtle bugs.
	expectedS, actualS, diff := diffValues(t, expected, actual)
	msg := fmt.Sprintf("Unexpected equal: \nexpected: %s\nactual  : %s\n%s", expectedS, actualS, diff)
	return assert.Fail(t, msg)
}

// diffValues returns a readable form of given values and the difference between them.
func diffValues[T types.Type](t testing.TB, expected, actual T) (expectedS string, actualS string, diff string) {
	// We might switch to spew or something else later.
	expectedB, err := fjson.Marshal(expected)
	require.NoError(t, err)
	expectedS = string(expectedB)

	actualB, err := fjson.Marshal(actual)
	require.NoError(t, err)
	actualS = string(actualB)

	diff, err = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expectedS),
		FromFile: "expected",
		B:        difflib.SplitLines(actualS),
		ToFile:   "actual",
		Context:  1,
	})
	require.NoError(t, err)

	return
}
