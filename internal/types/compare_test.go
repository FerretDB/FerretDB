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

package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// TestCompare tests edge cases of the comparison.
func TestCompare(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		a        any
		b        any
		skip     string
		expected CompareResult
	}{
		"EmptyArrayCompareNullFieldArray": {
			a:        must.NotFail(NewArray()),
			b:        must.NotFail(NewArray(NullType{})),
			expected: Less,
		},
		"UnsetCompareNullTypeFieldDocument": {
			skip:     "https://github.com/FerretDB/FerretDB/issues/1023",
			a:        must.NotFail(NewDocument()),
			b:        must.NotFail(NewDocument("foo", NullType{})),
			expected: Equal,
		},
		"NullTypeFieldCompareUnsetDocument": {
			skip:     "https://github.com/FerretDB/FerretDB/issues/1023",
			a:        must.NotFail(NewDocument("foo", NullType{})),
			b:        must.NotFail(NewDocument()),
			expected: Equal,
		},
		"ArrayCompareNumber": {
			skip:     "https://github.com/FerretDB/FerretDB/issues/1522",
			a:        must.NotFail(NewArray(1)),
			b:        2,
			expected: Less,
		},
		"NumberCompareArray": {
			skip:     "https://github.com/FerretDB/FerretDB/issues/1522",
			a:        1,
			b:        must.NotFail(NewArray(2)),
			expected: Greater,
		},
		"EmptyArrayCompareEmptyDocument": {
			a:        must.NotFail(NewArray()),
			b:        must.NotFail(NewDocument()),
			expected: Less,
		},
		"EmptyDocumentCompareEmptyArray": {
			a:        must.NotFail(NewDocument()),
			b:        must.NotFail(NewArray()),
			expected: Less,
		},
		"DocumentCompareDocument": {
			a:        must.NotFail(NewDocument("a", must.NotFail(NewDocument()))),
			b:        must.NotFail(NewDocument("a", must.NotFail(NewDocument()))),
			expected: Equal,
		},
		"EmptyArrayCompareNull": {
			a:        must.NotFail(NewArray()),
			b:        NullType{},
			expected: Less,
		},
		"NullCompareEmptyArray": {
			skip:     "https://github.com/FerretDB/FerretDB/issues/1522",
			a:        NullType{},
			b:        must.NotFail(NewArray()),
			expected: Greater,
		},
		"DocumentCompareEmptyArray": {
			a:        must.NotFail(NewDocument("foo", "bar")),
			b:        must.NotFail(NewArray()),
			expected: Less,
		},
		"EmptyArrayCompareDocument": {
			a:        must.NotFail(NewArray()),
			b:        must.NotFail(NewDocument("foo", "bar")),
			expected: Less,
		},
		"DocumentCompareArray": {
			a:        must.NotFail(NewDocument("foo", "a")),
			b:        must.NotFail(NewArray(must.NotFail(NewDocument("foo", "z")))),
			expected: Less,
		},
		"ArrayCompareDocument": {
			a:        must.NotFail(NewArray(must.NotFail(NewDocument("foo", "baz")))),
			b:        must.NotFail(NewDocument("foo", "bar")),
			expected: Greater,
		},
		"ArrayCompareEqualDocument": {
			a: must.NotFail(NewArray(
				must.NotFail(NewDocument("foo", "a")),
				must.NotFail(NewDocument("foo", "b")),
				must.NotFail(NewDocument("foo", "bar")),
			)),
			b:        must.NotFail(NewDocument("foo", "bar")),
			expected: Equal,
		},
		"DocumentCompareGreaterArray": {
			a: must.NotFail(NewDocument("foo", "baz")),
			b: must.NotFail(NewArray(
				must.NotFail(NewDocument("foo", "a")),
				must.NotFail(NewDocument("foo", "b")),
				must.NotFail(NewDocument("foo", "bar")),
			)),
			expected: Less,
		},
		"ArrayCompareGreaterDocument": {
			a: must.NotFail(NewArray(
				must.NotFail(NewDocument("foo", "a")),
				must.NotFail(NewDocument("foo", "b")),
				must.NotFail(NewDocument("foo", "baz")),
			)),
			b:        must.NotFail(NewDocument("foo", "bar")),
			expected: Greater,
		},
		"DocumentCompareLessArray": {
			a: must.NotFail(NewDocument("foo", "bar")),
			b: must.NotFail(NewArray(
				must.NotFail(NewDocument("foo", "a")),
				must.NotFail(NewDocument("foo", "b")),
				must.NotFail(NewDocument("foo", "baz")),
			)),
			expected: Less,
		},
		"ArrayCompareLessDocument": {
			a: must.NotFail(NewArray(
				must.NotFail(NewDocument("foo", "a")),
				must.NotFail(NewDocument("foo", "b")),
				must.NotFail(NewDocument("foo", "bar")),
			)),
			b:        must.NotFail(NewDocument("foo", "baz")),
			expected: Less,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			res := Compare(tc.a, tc.b)
			require.Equal(t, tc.expected, res)
		})
	}
}
