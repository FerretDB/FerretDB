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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestArrayIterator(t *testing.T) {
	t.Parallel()

	t.Run("Normal", func(t *testing.T) {
		t.Parallel()

		iter := must.NotFail(NewArray(1, 2, 3)).Iterator()
		defer iter.Close()

		n, v, err := iter.Next()
		assert.Equal(t, 0, n)
		assert.Equal(t, nil, v)
		assert.Equal(t, iterator.ErrIteratorDone, err)
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		iter := must.NotFail(NewArray()).Iterator()
		defer iter.Close()

		n, v, err := iter.Next()
		assert.Equal(t, 0, n)
		assert.Equal(t, nil, v)
		assert.Equal(t, iterator.ErrIteratorDone, err)

		iter.Close()

		n, v, err = iter.Next()
		assert.Equal(t, 0, n)
		assert.Equal(t, nil, v)
		assert.Equal(t, iterator.ErrIteratorDone, err)
	})

	for name, tc := range map[string]struct {
		arr      *Array
		expected []any
	}{
		"empty": {
			arr:      must.NotFail(NewArray()),
			expected: []any{},
		},
		"three": {
			arr:      must.NotFail(NewArray(1, 2, 3)),
			expected: []any{1, 2, 3},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			iter := tc.arr.Iterator()
			defer iter.Close()

			for i := 0; i < len(tc.expected); i++ {
				n, v, err := iter.Next()
				require.NoError(t, err)

				assert.Equal(t, i, n)
				assert.Equal(t, tc.expected[i], v)
			}

			_, _, err := iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)

			// check that Next() can be called again
			_, _, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
		})
	}
}
