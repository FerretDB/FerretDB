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

func TestDocumentIterator(t *testing.T) {
	t.Parallel()

	t.Run("Normal", func(t *testing.T) {
		t.Parallel()

		iter := must.NotFail(NewDocument("foo", "bar", "baz", "qux")).Iterator()
		defer iter.Close()

		k, v, err := iter.Next()
		require.NoError(t, err)
		assert.Equal(t, "foo", k)
		assert.Equal(t, "bar", v)

		k, v, err = iter.Next()
		require.NoError(t, err)
		assert.Equal(t, "baz", k)
		assert.Equal(t, "qux", v)

		k, v, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, k)
		assert.Nil(t, v)

		// still done
		k, v, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, k)
		assert.Nil(t, v)
	})

	t.Run("EarlyClose", func(t *testing.T) {
		t.Parallel()

		iter := must.NotFail(NewDocument("foo", "bar", "baz", "qux")).Iterator()
		defer iter.Close()

		iter.Close()

		k, v, err := iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, k)
		assert.Nil(t, v)

		// still done
		k, v, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, k)
		assert.Nil(t, v)
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		iter := must.NotFail(NewDocument()).Iterator()
		defer iter.Close()

		k, v, err := iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, k)
		assert.Nil(t, v)

		iter.Close()

		// still done after Close()
		k, v, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, k)
		assert.Nil(t, v)
	})

	t.Run("Nil", func(t *testing.T) {
		t.Parallel()

		iter := (*Document)(nil).Iterator()
		defer iter.Close()

		k, v, err := iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, k)
		assert.Nil(t, v)

		iter.Close()

		// still done after Close()
		k, v, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		assert.Zero(t, k)
		assert.Nil(t, v)
	})
}
