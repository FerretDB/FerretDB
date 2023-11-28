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
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCursorNormal(t *testing.T) {
	t.Parallel()

	r := NewRegistry(testutil.Logger(t))
	t.Cleanup(r.Close)

	ctx := testutil.Ctx(t)

	doc1 := must.NotFail(types.NewDocument("v", int32(1)))
	doc2 := must.NotFail(types.NewDocument("v", int32(2)))
	doc3 := must.NotFail(types.NewDocument("v", int32(3)))

	docs := []*types.Document{doc1, doc2, doc3}

	t.Run("Normal", func(t *testing.T) {
		t.Parallel()

		c := r.NewCursor(ctx, &NewCursorParams{
			Iter: iterator.Values(iterator.ForSlice(docs)),
		})

		actual, err := iterator.ConsumeValues(c)
		require.NoError(t, err)
		assert.Equal(t, docs, actual)
	})

	t.Run("ClosedByContext", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(ctx)

		c := r.NewCursor(ctx, &NewCursorParams{
			Iter: iterator.Values(iterator.ForSlice(docs)),
		})

		cancel()
		<-ctx.Done()

		time.Sleep(time.Second)

		_, _, err := c.Next()
		assert.Equal(t, context.Canceled, err)
	})
}
