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

package pgdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// getPool creates a new connection's connection pool for testing.
func getPool(ctx context.Context, tb testing.TB, l *zap.Logger) *pgdb.Pool {
	tb.Helper()

	pool, err := pgdb.NewPool(ctx, testutil.PostgreSQLURL(tb, nil), l, false)
	require.NoError(tb, err)
	tb.Cleanup(pool.Close)

	return pool
}
