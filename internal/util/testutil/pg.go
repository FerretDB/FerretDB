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
	"net/url"
	"testing"
)

// PostgreSQLURLOpts represents PostgreSQLURL options.
type PostgreSQLURLOpts struct {
	// PostgreSQL database name, defaults to `ferretdb`.
	DatabaseName string

	// If set, the pool will use read-only user.
	ReadOnly bool

	// Extra query parameters.
	Params map[string]string
}

// PostgreSQLURL returns PostgreSQL URL for testing.
func PostgreSQLURL(tb testing.TB, opts *PostgreSQLURLOpts) string {
	tb.Helper()

	if testing.Short() {
		tb.Skip("skipping in -short mode")
	}

	if opts == nil {
		opts = new(PostgreSQLURLOpts)
	}

	databaseName := opts.DatabaseName
	if databaseName == "" {
		databaseName = "ferretdb"
	}

	username := "username"
	password := "password"
	if opts.ReadOnly {
		username = "readonly"
		password = "readonly_password"
	}

	q := url.Values{
		"pool_min_conns": []string{"1"},
	}
	for k, v := range opts.Params {
		q.Set(k, v)
	}

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(username, password),
		Host:     "127.0.0.1:5432",
		Path:     databaseName,
		RawQuery: q.Encode(),
	}

	return u.String()
}
