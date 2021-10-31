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

func Ctx(t testing.TB) context.Context {
	// TODO
	return context.Background()
}

func Pool(ctx context.Context, t testing.TB) *pgconn.Pool {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	pgPool, err := pgconn.NewPool("postgres://postgres@127.0.0.1:5432/mangodb", zaptest.NewLogger(t))
	require.NoError(t, err)
	t.Cleanup(pgPool.Close)

	return pgPool
}

func Schema(ctx context.Context, t testing.TB, pool *pgconn.Pool) string {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	name := strings.ToLower(t.Name())

	pool.Exec(ctx, "DROP SCHEMA "+name+" CASCADE")

	_, err := pool.Exec(ctx, "CREATE SCHEMA "+name)
	require.NoError(t, err)
	t.Cleanup(func() {
		// keep schema around for debugging
		if t.Failed() {
			return
		}

		_, err = pool.Exec(ctx, "DROP SCHEMA "+name+" CASCADE")
		require.NoError(t, err)
	})

	return name
}
