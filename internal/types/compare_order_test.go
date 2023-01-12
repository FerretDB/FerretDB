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

// TestCompareOrderForSort tests edge cases only.
func TestCompareOrderForSort(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		a        any
		b        any
		order    SortType
		expected CompareResult
	}{
		"EmptyArrays": {
			a:        must.NotFail(NewArray()),
			b:        must.NotFail(NewArray()),
			order:    Ascending,
			expected: Equal,
		},
		"NonArrayAndEmptyArray": {
			a:        must.NotFail(NewDocument("foo", Null)),
			b:        must.NotFail(NewArray()),
			order:    Ascending,
			expected: Greater,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := CompareOrderForSort(tc.a, tc.b, tc.order)
			require.Equal(t, tc.expected, res)
		})
	}
}
