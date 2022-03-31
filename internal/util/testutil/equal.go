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
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// AssertEqual asserts that two BSON values are equal.
func AssertEqual[T types.Type](t testing.TB, expected, actual T) bool {
	t.Helper()

	if Equal(expected, actual) {
		return true
	}

	expectedS, actualS, diff := diffValues(t, expected, actual)
	msg := fmt.Sprintf("Not equal: \nexpected: %s\nactual  : %s\n%s", expectedS, actualS, diff)
	return assert.Fail(t, msg)
}

// AssertNotEqual asserts that two BSON values are not equal.
func AssertNotEqual[T types.Type](t testing.TB, expected, actual T) bool {
	t.Helper()

	if !Equal(expected, actual) {
		return true
	}

	// The diff of equal values should be empty, but produce it anyway to catch subtle bugs.
	expectedS, actualS, diff := diffValues(t, expected, actual)
	msg := fmt.Sprintf("Unexpected equal: \nexpected: %s\nactual  : %s\n%s", expectedS, actualS, diff)
	return assert.Fail(t, msg)
}

// diffValues returns a readable form of given values and the difference between them.
func diffValues[T types.Type](t testing.TB, expected, actual T) (expectedS string, actualS string, diff string) {
	expectedS = Dump(t, expected)
	actualS = Dump(t, actual)

	var err error
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

// Equal compares any BSON values.
func Equal[T types.Type](v1, v2 T) bool {
	return equal(v1, v2)
}

// equal compares any BSON values.
func equal(v1, v2 any) bool {
	switch v1 := v1.(type) {
	case *types.Document:
		d, ok := v2.(*types.Document)
		if !ok {
			return false
		}
		if !equalDocuments(v1, d) {
			return false
		}

	case *types.Array:
		a, ok := v2.(*types.Array)
		if !ok {
			return false
		}
		if !equalArrays(v1, a) {
			return false
		}

	default:
		if !equalScalars(v1, v2) {
			return false
		}
	}

	return true
}

// equalDocuments compares BSON documents. Nils are not allowed.
func equalDocuments(v1, v2 *types.Document) bool {
	if v1 == nil {
		panic("v1 is nil")
	}
	if v2 == nil {
		panic("v2 is nil")
	}

	keys := v1.Keys()
	if !slices.Equal(keys, v2.Keys()) {
		return false
	}

	for _, k := range keys {
		f1 := must.NotFail(v1.Get(k))
		f2 := must.NotFail(v2.Get(k))
		if !equal(f1, f2) {
			return false
		}
	}

	return true
}

// equalArrays compares BSON arrays. Nils are not allowed.
func equalArrays(v1, v2 *types.Array) bool {
	if v1 == nil {
		panic("v1 is nil")
	}
	if v2 == nil {
		panic("v2 is nil")
	}

	l := v1.Len()
	if l != v2.Len() {
		return false
	}

	for i := 0; i < l; i++ {
		el1 := must.NotFail(v1.Get(i))
		el2 := must.NotFail(v2.Get(i))
		if !equal(el1, el2) {
			return false
		}
	}

	return true
}

// equalScalars compares BSON scalar values in a way that is useful for tests:
//  * float64 NaNs are equal to each other;
//  * time.Time values are compared using Equal method.
func equalScalars(v1, v2 any) bool {
	switch s1 := v1.(type) {
	case float64:
		s2, ok := v2.(float64)
		if !ok {
			return false
		}
		if math.IsNaN(s1) {
			return math.IsNaN(s2)
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

	case types.CString:
		s2, ok := v2.(types.CString)
		if !ok {
			return false
		}
		return s1 == s2

	default:
		panic(fmt.Sprintf("unhandled types %T, %T", v1, v2))
	}
}
