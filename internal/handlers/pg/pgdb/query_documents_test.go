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

// Use _test package to avoid import cycle with testutil.
package pgdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestQueryDocuments(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(testutil.Ctx(t))
	defer cancel()

	pool := testutil.Pool(ctx, t, nil, zaptest.NewLogger(t))
	dbName := testutil.Schema(ctx, t, pool)
	collectionName := testutil.Table(ctx, t, pool, dbName)

	// 0 docs
	// 1 doc
	// 2 docs
	// 3 docs

	// chan full

	doc, err := types.NewDocument("id", "1")
	require.NoError(t, err)

	err = pool.InsertDocument(ctx, dbName, collectionName, doc)
	require.NoError(t, err)

	fetchedChan, err := pool.QueryDocuments(ctx, dbName, collectionName, "")
	require.NoError(t, err)

	for {
		fetched, ok := <-fetchedChan
		if !ok {
			break
		}

		require.NoError(t, fetched.Err)
		require.Len(t, fetched.Docs, 1)
	}
}
