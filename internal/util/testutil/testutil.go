// Copyright 2021 Baltoro OÜ.
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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/MangoDB-io/MangoDB/internal/pgconn"
)

func Ctx(tb testing.TB) context.Context {
	tb.Helper()

	// TODO
	return context.Background()
}

func Pool(ctx context.Context, tb testing.TB) *pgconn.Pool {
	tb.Helper()

	if testing.Short() {
		tb.Skip("skipping in -short mode")
	}

	pgPool, err := pgconn.NewPool("postgres://postgres@127.0.0.1:5432/mangodb", zaptest.NewLogger(tb))
	require.NoError(tb, err)
	tb.Cleanup(pgPool.Close)

	return pgPool
}

func Schema(ctx context.Context, tb testing.TB, pool *pgconn.Pool) string {
	tb.Helper()

	if testing.Short() {
		tb.Skip("skipping in -short mode")
	}

	name := strings.ToLower(tb.Name())

	pool.Exec(ctx, "DROP SCHEMA "+name+" CASCADE")

	_, err := pool.Exec(ctx, "CREATE SCHEMA "+name)
	require.NoError(tb, err)
	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping schema %q for debugging.", name)
			return
		}

		_, err = pool.Exec(ctx, "DROP SCHEMA "+name+" CASCADE")
		require.NoError(tb, err)
	})

	return name
}
