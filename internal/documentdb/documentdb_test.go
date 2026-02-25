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
	"runtime"
	"testing"

	"github.com/FerretDB/wire/wirebson"
	"github.com/FerretDB/wire/wiretest"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil/teststress"
)

// testPool tries to create a new pool of PostgreSQL connections and use it.
// First error returned is newPgxPool's error, the second is Ping's error.
func testPool(t testing.TB, ctx context.Context, uri string, sp *state.Provider) (error, error) {
	t.Helper()

	l := testutil.Logger(t)
	pool, err := newPgxPool(uri, l, newTracer(l), sp)
	if err != nil {
		return err, nil
	}

	require.NotNil(t, pool)
	defer pool.Close()

	return err, pool.Ping(ctx)
}

func TestNewPool(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	t.Run("Normal", func(t *testing.T) {
		uri := testutil.PostgreSQLURL(t)

		t.Parallel()

		newErr, pingErr := testPool(t, ctx, uri, sp)
		assert.NoError(t, newErr)
		assert.NoError(t, pingErr)

		assert.Equal(t, version.PostgreSQLTest, sp.Get().PostgreSQLVersion, "version.PostgreSQL wasn't updated")
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
	uri := testutil.PostgreSQLURL(t)

	t.Parallel()

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	l := testutil.Logger(t)

	pool, err := NewPool(uri, l, sp)
	require.NoError(t, err)
	defer pool.Close()

	var res wirebson.RawDocument

	err = pool.WithConn(ctx, func(conn *pgx.Conn) error {
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
		`DataTypeName:"", ConstraintName:"", File:"delete.c", Line:527, Routine:"BuildDeletionSpec"}}`
	assert.Equal(t, expected, fmt.Sprintf("%#v", err))
}

func TestSeqNoSeq(t *testing.T) {
	uri := testutil.PostgreSQLURL(t)

	t.Parallel()

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	l := testutil.Logger(t)

	pool, err := NewPool(uri, l, sp)
	require.NoError(t, err)

	dbName := testutil.DatabaseName(t)
	collName := testutil.CollectionName(t)

	defer func() {
		_ = pool.WithConn(ctx, func(conn *pgx.Conn) error {
			var drop bool
			drop, err = documentdb_api.DropCollection(ctx, conn, l, dbName, collName, nil, nil, false)
			require.NoError(t, err)
			assert.True(t, drop)

			return nil
		})

		pool.Close()
	}()

	var res *wirebson.Document

	// insert document using sequence from [wire.OpMsg.Sections]
	err = pool.WithConn(ctx, func(conn *pgx.Conn) error {
		b := must.NotFail(wirebson.MustDocument(
			"insert", collName,
		).Encode())

		seq := must.NotFail(wirebson.MustDocument("_id", int32(1)).Encode())
		seq = append(seq, must.NotFail(wirebson.MustDocument("_id", int32(2)).Encode())...)

		var raw wirebson.RawDocument

		raw, _, err = documentdb_api.Insert(ctx, conn, l, dbName, b, seq)
		if err == nil {
			res, err = raw.DecodeDeep()
		}

		return err
	})

	require.NoError(t, err)
	wiretest.AssertEqual(t, wirebson.MustDocument("n", int32(2), "ok", float64(1)), res)

	// insert document using single document from, for example, Data API
	err = pool.WithConn(ctx, func(conn *pgx.Conn) error {
		b := must.NotFail(wirebson.MustDocument(
			"insert", collName,
			"documents", wirebson.MustArray(
				wirebson.MustDocument("_id", int32(3)),
				wirebson.MustDocument("_id", int32(4)),
			),
		).Encode())

		var raw wirebson.RawDocument

		raw, _, err = documentdb_api.Insert(ctx, conn, l, dbName, b, nil)
		if err == nil {
			res, err = raw.DecodeDeep()
		}

		return err
	})

	require.NoError(t, err)
	wiretest.AssertEqual(t, wirebson.MustDocument("n", int32(2), "ok", float64(1)), res)

	spec := wirebson.MustDocument(
		"find", collName,
		"sort", wirebson.MustDocument("_id", int32(1)),
	)
	page, _, err := pool.Find(ctx, dbName, must.NotFail(spec.Encode()))

	require.NoError(t, err)

	res, err = page.DecodeDeep()
	require.NoError(t, err)

	expected := wirebson.MustArray(
		wirebson.MustDocument("_id", int32(1)),
		wirebson.MustDocument("_id", int32(2)),
		wirebson.MustDocument("_id", int32(3)),
		wirebson.MustDocument("_id", int32(4)),
	)
	wiretest.AssertEqual(t, expected, res.Get("cursor").(*wirebson.Document).Get("firstBatch"))
}

func TestWithConn(t *testing.T) {
	sp, err := state.NewProvider("")
	require.NoError(t, err)

	pool, err := NewPool(testutil.PostgreSQLURL(t), testutil.Logger(t), sp)
	require.NoError(t, err)

	defer pool.Close()

	// TODO https://github.com/FerretDB/FerretDB/issues/5446
	teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
		ready <- struct{}{}

		<-start

		_ = pool.WithConn(func(conn *pgx.Conn) error {
			must.NotBeZero(conn)

			for range 10 {
				runtime.GC()
				runtime.Gosched()
			}

			return nil
		})

		for range 10 {
			runtime.GC()
			runtime.Gosched()
		}
	})

	for range 10 {
		runtime.GC()
		runtime.Gosched()
	}
}
