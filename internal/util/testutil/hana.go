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
	"os"
	"testing"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
	"github.com/stretchr/testify/require"
)

// TestHanaURI returns a HANA Database URL for testing.
// HANATODO Create a Database per test run?
func TestHanaURI(tb testtb.TB, ctx context.Context, baseURI string) string {
	tb.Helper()

	if testing.Short() {
		tb.Skip("skipping in -short mode")
	}

	if baseURI == "" {
		baseURI = os.Getenv("FERRETDB_HANA_URL")
	}

	if baseURI == "" {
		tb.Skip("FERRETDB_HANA_URL is not set")
	}

	name := DirectoryName(tb)

	db, err := sql.Open("hdb", baseURI)
	defer db.Close()
	require.NoError(tb, err)

	q := fmt.Sprintf("DROP SCHEMA %q CASCADE", name)

	// Drop database (schema) if it exists.
	_, err = db.ExecContext(ctx, q)
	if err != nil {
		require.ErrorContains(tb, err,
			"SQL Error 362 - invalid schema name:")
	}

	return baseURI
}
