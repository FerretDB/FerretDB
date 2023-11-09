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
	"fmt"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

func TestPostgreSQLURI(tb testtb.TB, ctx context.Context, baseURI string) string {
	if testing.Short() {
		tb.Skip("skipping in -short mode")
	}

	if baseURI == "" {
		baseURI = "postgres://username:password@127.0.0.1:5432/ferretdb"
	}

	u, err := url.Parse(baseURI)
	require.NoError(tb, err)

	name := DirectoryName(tb)
	u.Path = name

	p, err := pgxpool.New(ctx, baseURI)
	require.NoError(tb, err)

	q := fmt.Sprintf("DROP DATABASE IF EXISTS %s", pgx.Identifier{name}.Sanitize())
	_, err = p.Exec(ctx, q)
	require.NoError(tb, err)

	q = fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{name}.Sanitize())
	_, err = p.Exec(ctx, q)
	require.NoError(tb, err)

	p.Reset()

	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping database %s (%s) for debugging.", name, u.String())
			return
		}

		q = fmt.Sprintf("DROP DATABASE %s", pgx.Identifier{name}.Sanitize())
		_, err = p.Exec(context.Background(), q)
		require.NoError(tb, err)
	})

	return u.String()
}
