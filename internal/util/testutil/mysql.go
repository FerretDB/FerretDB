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
	"database/sql"
	"fmt"
	"net/url"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// TestMySQLURI returns MySQL URI with test-specific database.
// It will be created before test and dropped after unless test fails.
//
// Base URI may be empty.
func TestMySQLURI(tb testtb.TB, ctx context.Context, baseURI string) string {
	tb.Helper()

	if testing.Short() {
		tb.Skip("skipping in -short mode")
	}

	if baseURI == "" {
		baseURI = "mysql://username:password@127.0.0.1:3306/ferretdb"
	}

	u, err := url.Parse(baseURI)
	require.NoError(tb, err)

	require.True(tb, u.Path != "")
	require.True(tb, u.Opaque == "")

	password, _ := u.User.Password()

	values := u.Query()
	params := make(map[string]string, len(values))

	for k := range values {
		params[k] = values.Get(k)
	}

	// mysql url requires a specified format to work
	// For example: username:password@tcp(127.0.0.1:3306)/ferretdb
	cfg := mysql.Config{
		User:   u.User.Username(),
		Passwd: password,
		Addr:   u.Host,
		DBName: u.Path,
		Params: params,
	}
	mysqlURL := cfg.FormatDSN()

	name := DirectoryName(tb)
	u.Path = name
	res := u.String()

	db, err := sql.Open("mysql", mysqlURL)
	require.NoError(tb, err)

	q := fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)
	_, err = db.ExecContext(ctx, q)
	require.NoError(tb, err)

	q = fmt.Sprintf("CREATE DATABASE %s", name)
	_, err = db.ExecContext(ctx, q)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		defer func() {
			err = db.Close()
			require.NoError(tb, err)
		}()

		if tb.Failed() {
			tb.Logf("Keeping database %s (%s) for debugging.", name, res)
			return
		}

		q = fmt.Sprintf("DROP DATABASE %s", name)
		_, err = db.ExecContext(ctx, q)
		require.NoError(tb, err)
	})

	return res
}
