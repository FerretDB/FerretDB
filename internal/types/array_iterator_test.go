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

		iter := must.NotFail(NewArray(int32(1), int32(2))).Iterator()
		defer iter.Close()

		n, v, err := iter.Next()
		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.Equal(t, int32(1), v)

		n, v, err = iter.Next()
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Equal(t, int32(2), v)

		n, v, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, n)
		assert.Nil(t, v)

		// still done
		n, v, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, n)
		assert.Nil(t, v)
	})

	t.Run("EarlyClose", func(t *testing.T) {
		t.Parallel()

		iter := must.NotFail(NewArray(int32(1), int32(2))).Iterator()
		defer iter.Close()

		iter.Close()

		n, v, err := iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, n)
		assert.Nil(t, v)

		// still done
		n, v, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, n)
		assert.Nil(t, v)
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		iter := must.NotFail(NewArray()).Iterator()
		defer iter.Close()

		n, v, err := iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, n)
		assert.Nil(t, v)

		iter.Close()

		// still done after Close()
		n, v, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, n)
		assert.Nil(t, v)
	})
}
