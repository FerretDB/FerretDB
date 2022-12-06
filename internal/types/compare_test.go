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
	"github.com/FerretDB/FerretDB/internal/util/must"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCompare tests edge cases of the comparison.
func TestCompare(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		skip     string
		a        any
		b        any
		expected CompareResult
	}{
		"NullDocumentCompareDocument": {
			skip:     "https://github.com/FerretDB/FerretDB/issues/1023",
			a:        must.NotFail(NewDocument("foo", "bar")),
			b:        must.NotFail(NewDocument("foo", nil)),
			expected: Greater,
		},
		"DocumentCompareNullDocument": {
			skip:     "https://github.com/FerretDB/FerretDB/issues/1023",
			a:        must.NotFail(NewDocument("foo", nil)),
			b:        must.NotFail(NewDocument("foo", "bar")),
			expected: Less,
		},
		"UnsetCompareNullFieldDocument": {
			skip:     "https://github.com/FerretDB/FerretDB/issues/1023",
			a:        must.NotFail(NewDocument()),
			b:        must.NotFail(NewDocument("foo", nil)),
			expected: Equal,
		},
		"NullFieldCompareUnsetDocument": {
			skip:     "https://github.com/FerretDB/FerretDB/issues/1023",
			a:        must.NotFail(NewDocument("foo", nil)),
			b:        must.NotFail(NewDocument()),
			expected: Equal,
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
