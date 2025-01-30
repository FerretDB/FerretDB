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

package documentdb

import (
	"context"
	"fmt"
	"testing"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

// testPool tries to create a new pool of PostgreSQL connections and use it.
// First error returned is newPgxPool's error, the second is Ping's error.
func testPool(t testing.TB, ctx context.Context, uri string, sp *state.Provider) (error, error) {
	t.Helper()

	pool, err := newPgxPool(uri, testutil.Logger(t), sp)
	if err != nil {
		return err, nil
	}

	require.NotNil(t, pool)
	defer pool.Close()

	return err, pool.Ping(ctx)
}

func TestNewPool(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	t.Run("Normal", func(t *testing.T) {
		t.Parallel()

		const uri = "postgres://username:password@127.0.0.1:5432/postgres"

		newErr, pingErr := testPool(t, ctx, uri, sp)
		assert.NoError(t, newErr)
		assert.NoError(t, pingErr)

		assert.Equal(t, version.DocumentDB, sp.Get().DocumentDBVersion, "version.DocumentDB wasn't updated")
	})

	t.Run("Invalid", func(t *testing.T) {
		t.Parallel()

		// nothing is listening on that port
		const uri = "postgres://127.0.0.1:56789/postgres"

		newErr, pingErr := testPool(t, ctx, uri, sp)
		assert.NoError(t, newErr)
		assert.Error(t, pingErr)
	})
}

func TestError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	const uri = "postgres://username:password@127.0.0.1:5432/postgres"

	l := testutil.Logger(t)

	pool, err := NewPool(uri, l, sp)
	require.NoError(t, err)
	defer pool.Close()

	var res wirebson.RawDocument

	err = pool.WithConn(func(conn *pgx.Conn) error {
		b := must.NotFail(wirebson.MustDocument(
			"delete", testutil.CollectionName(t),
			"deletes", wirebson.MustArray(wirebson.MustDocument(
				"q", wirebson.MustDocument(),
				"limit", int32(-1),
			)),
		).Encode())

		res, _, err = documentdb_api.Delete(ctx, conn, l, testutil.DatabaseName(t), b, nil)

		return err
	})

	require.Nil(t, res)
	require.Error(t, err)

	err = mongoerrors.Make(ctx, err, "", l)

	expected := "FailedToParse (9): The limit field in delete objects must be 0 or 1. Got -1"
	assert.Equal(t, expected, fmt.Sprintf("%s", err))
	assert.Equal(t, expected, fmt.Sprintf("%v", err))
	assert.Equal(t, expected, fmt.Sprintf("%+v", err))

	expected = "&mongoerrors.Error{" +
		"Code: 9, " +
		"Name: `FailedToParse`, " +
		"Message: `The limit field in delete objects must be 0 or 1. Got -1`, " +
		"Argument: `documentdb_api.delete`, " +
		"Wrapped: &pgconn.PgError{" +
		`Severity:"ERROR", SeverityUnlocalized:"ERROR", Code:"M0003", ` +
		`Message:"The limit field in delete objects must be 0 or 1. Got -1", Detail:"", Hint:"", ` +
		`Position:0, InternalPosition:0, InternalQuery:"", Where:"", SchemaName:"", TableName:"", ColumnName:"", ` +
		`DataTypeName:"", ConstraintName:"", File:"delete.c", Line:479, Routine:"BuildDeletionSpec"}}`
	assert.Equal(t, expected, fmt.Sprintf("%#v", err))
}
