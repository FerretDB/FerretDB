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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestArraySorter(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		arr      *Array
		expected *Array
	}{
		"empty": {
			arr:      must.NotFail(NewArray()),
			expected: must.NotFail(NewArray()),
		},
		"int": {
			arr:      must.NotFail(NewArray(int64(2), int64(1), int64(3), int64(1))),
			expected: must.NotFail(NewArray(int64(1), int64(1), int64(2), int64(3))),
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sorter := newArraySort(tc.arr)
			sort.Sort(sorter)

			assert.Equal(t, tc.expected, sorter.Arr())
		})
	}
}
