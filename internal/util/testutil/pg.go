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

package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
)

// PoolOpts represents options for creating a connection pool.
type PoolOpts struct {
	// If set, the pool will use read-only user.
	ReadOnly bool
}

// PoolConnString returns PostgreSQL connection string for testing.
func PoolConnString(tb testing.TB, opts *PoolOpts) string {
	tb.Helper()

	if testing.Short() {
		tb.Skip("skipping in -short mode")
	}

	if opts == nil {
		opts = new(PoolOpts)
	}

	username := "postgres"
	if opts.ReadOnly {
		username = "readonly"
	}

	return "postgres://" + username + "@127.0.0.1:5432/ferretdb?pool_min_conns=1"
}

// Pool creates a new connection connection pool for testing.
//
// TODO move to pg/pgdb tests.
//
// Deprecated: do not use in new code.
func Pool(ctx context.Context, tb testing.TB, opts *PoolOpts, l *zap.Logger) *pgdb.Pool {
	tb.Helper()

	pool, err := pgdb.NewPool(ctx, PoolConnString(tb, opts), l, false)
	require.NoError(tb, err)
	tb.Cleanup(pool.Close)

	return pool
}

// SchemaName should not be used.
//
// Deprecated: use DatabaseName instead.
func SchemaName(tb testing.TB) string {
	tb.Helper()

	return DatabaseName(tb)
}

// TableName should not be used.
//
// Deprecated: use CollectionName instead.
func TableName(tb testing.TB) string {
	tb.Helper()

	return CollectionName(tb)
}
