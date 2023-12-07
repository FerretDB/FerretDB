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

package cursor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/iterator/testiterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCursor(t *testing.T) {
	t.Parallel()

	r := NewRegistry(testutil.Logger(t))
	t.Cleanup(r.Close)

	ctx := testutil.Ctx(t)

	doc1 := must.NotFail(types.NewDocument("v", int32(1)))
	doc2 := must.NotFail(types.NewDocument("v", int32(2)))
	doc3 := must.NotFail(types.NewDocument("v", int32(3)))

	doc1.SetRecordID(101)
	doc2.SetRecordID(102)
	doc3.SetRecordID(103)

	two := []*types.Document{doc1, doc2}
	all := []*types.Document{doc1, doc2, doc3}

	t.Run("Normal", func(t *testing.T) {
		t.Parallel()

		params := &NewParams{
			Type: Normal,
		}

		testiterator.TestIterator(t, func() iterator.Interface[struct{}, *types.Document] {
			return r.NewCursor(ctx, iterator.Values(iterator.ForSlice(all)), params)
		})

		t.Run("Consume", func(t *testing.T) {
			t.Parallel()

			c := r.NewCursor(ctx, iterator.Values(iterator.ForSlice(all)), params)

			actual, err := iterator.ConsumeValues(c)
			require.NoError(t, err)
			assert.Equal(t, all, actual)

			_, _, err = c.Next()
			assert.ErrorIs(t, err, iterator.ErrIteratorDone)

			assert.Nil(t, r.Get(c.ID), "cursor should be removed")
		})

		t.Run("Context", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(ctx)

			c := r.NewCursor(ctx, iterator.Values(iterator.ForSlice(all)), params)

			cancel()
			<-ctx.Done()
			time.Sleep(time.Second)

			_, _, err := c.Next()
			assert.ErrorIs(t, err, iterator.ErrIteratorDone)

			assert.Nil(t, r.Get(c.ID), "cursor should be removed")
		})

		t.Run("Reset", func(t *testing.T) {
			t.Parallel()

			c := r.NewCursor(ctx, iterator.Values(iterator.ForSlice(two)), params)

			actual, err := iterator.ConsumeValues(c)
			require.NoError(t, err)
			assert.Equal(t, two, actual)

			assert.PanicsWithValue(t, "Reset called on non-tailable cursor", func() {
				iter := iterator.Values(iterator.ForSlice(all))
				t.Cleanup(iter.Close)

				c.Reset(iter)
			})
		})
	})

	t.Run("Tailable", func(t *testing.T) {
		t.Parallel()

		params := &NewParams{
			Type: Tailable,
		}

		t.Run("Consume", func(t *testing.T) {
			t.Parallel()

			c := r.NewCursor(ctx, iterator.Values(iterator.ForSlice(all)), params)

			actual, err := iterator.ConsumeValues(c)
			require.NoError(t, err)
			assert.Equal(t, all, actual)

			_, _, err = c.Next()
			assert.ErrorIs(t, err, iterator.ErrIteratorDone)

			assert.Same(t, c, r.Get(c.ID), "cursor should not be removed")
		})

		t.Run("Context", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(ctx)

			c := r.NewCursor(ctx, iterator.Values(iterator.ForSlice(all)), params)

			cancel()
			<-ctx.Done()
			time.Sleep(time.Second)

			_, _, err := c.Next()
			assert.ErrorIs(t, err, iterator.ErrIteratorDone)

			assert.Nil(t, r.Get(c.ID), "cursor should be removed")
		})

		t.Run("Reset", func(t *testing.T) {
			t.Parallel()

			c := r.NewCursor(ctx, iterator.Values(iterator.ForSlice(two)), params)

			actual, err := iterator.ConsumeValues(c)
			require.NoError(t, err)
			assert.Equal(t, two, actual)

			err = c.Reset(iterator.Values(iterator.ForSlice(all)))
			require.NoError(t, err)

			assert.Same(t, c, r.Get(c.ID), "cursor should not be replaced")

			actual, err = iterator.ConsumeValues(c)
			require.NoError(t, err)
			assert.Equal(t, []*types.Document{doc3}, actual)
		})
	})
}
