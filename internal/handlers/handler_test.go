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

package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/handlers/jsonb1"
	"github.com/FerretDB/FerretDB/internal/handlers/shared"
	"github.com/FerretDB/FerretDB/internal/handlers/sql"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/wire"
)

func setup(t *testing.T, poolOpts *testutil.PoolOpts) (context.Context, *Handler, *pg.Pool) {
	t.Helper()

	if poolOpts == nil {
		poolOpts = new(testutil.PoolOpts)
	}

	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t, poolOpts)
	l := zaptest.NewLogger(t)
	shared := shared.NewHandler(pool, "127.0.0.1:12345")
	sql := sql.NewStorage(pool, l.Sugar())
	jsonb1 := jsonb1.NewStorage(pool, l)
	handler := New(&NewOpts{
		PgPool:        pool,
		Logger:        l,
		SharedHandler: shared,
		SQLStorage:    sql,
		JSONB1Storage: jsonb1,
		Metrics:       NewMetrics(),
	})

	return ctx, handler, pool
}

func createDB(ctx context.Context, pool *pg.Pool, name string) error {
	_, err := pool.Exec(ctx, fmt.Sprintf(
		`CREATE SCHEMA %s`,
		pgx.Identifier{name}.Sanitize(),
	))
	return err
}

func dropDB(ctx context.Context, pool *pg.Pool, name string) error {
	_, err := pool.Exec(ctx, fmt.Sprintf(
		`DROP SCHEMA IF EXISTS %s CASCADE`,
		pgx.Identifier{name}.Sanitize(),
	))
	return err
}

func TestDropDatabase(t *testing.T) { //nolint:paralleltest,tparallel // affects a global list of databases
	ctx, handler, pool := setup(t, nil)

	type testCase struct {
		req  types.Document
		resp types.Document
	}

	notExistedDbName := "not_existed_db"
	dummyDbName := "dummy_db"
	testCases := map[string]testCase{
		"dropNotExistedDatabase": {
			req: types.MustMakeDocument(
				"dropDatabase", int32(1),
			),
			resp: types.MustMakeDocument(
				"dropped", notExistedDbName,
				"ok", float64(1),
			),
		},
		"dropDummyDatabase": {
			req: types.MustMakeDocument(
				"dropDatabase", int32(1),
			),
			resp: types.MustMakeDocument(
				"dropped", dummyDbName,
				"ok", float64(1),
			),
		},
	}

	dummyTableName := "dummy_table"
	require.NoError(t, createDB(ctx, pool, dummyDbName))
	require.NoError(t, shared.CreateCollection(ctx, pool, dummyDbName, dummyTableName))

	_, err := pool.Exec(ctx, fmt.Sprintf(
		`INSERT INTO %s(_jsonb) VALUES ('{"key": "value"}')`,
		pgx.Identifier{dummyDbName, dummyTableName}.Sanitize(),
	))
	require.NoError(t, err)

	for name, tc := range testCases { //nolint:paralleltest // false positive
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			schemaName := tc.resp.Map()["dropped"].(string)

			tc.req.Set("$db", schemaName)
			reqHeader := wire.MsgHeader{
				RequestID: 1,
				OpCode:    wire.OP_MSG,
			}

			var reqMsg wire.OpMsg
			err := reqMsg.SetSections(wire.OpMsgSection{
				Documents: []types.Document{tc.req},
			})
			require.NoError(t, err)

			_, resBody, closeConn := handler.Handle(ctx, &reqHeader, &reqMsg)
			require.False(t, closeConn, "%s", wire.DumpMsgBody(resBody))

			actual, err := resBody.(*wire.OpMsg).Document()
			require.NoError(t, err)

			expected := tc.resp

			assert.Equal(t, actual.Map(), expected.Map())

			var count int

			// checking tables
			query := `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = $1`
			err = pool.QueryRow(ctx, query, schemaName).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, count, 0)

			// checking schema
			query = `SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = $1`
			err = pool.QueryRow(ctx, query, schemaName).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, count, 0)
		})
	}
}

func TestFind(t *testing.T) {
	t.Parallel()
	ctx, handler, _ := setup(t, &testutil.PoolOpts{
		ReadOnly: true,
	})

	lastUpdate := time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local()

	type testCase struct {
		req  types.Document
		resp types.Array
	}

	testCases := map[string]testCase{
		"ValueLtGt": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", "HOFFMAN",
					"actor_id", types.MustMakeDocument(
						"$gt", int32(50),
						"$lt", int32(100),
					),
				),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x4f, 0x00, 0x00, 0x00, 0x4f},
					"actor_id", int32(79),
					"first_name", "MAE",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			},
		},
		"InLteGte": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.MustMakeDocument(
						"$in", types.Array{"HOFFMAN"},
					),
					"actor_id", types.MustMakeDocument(
						"$gte", int32(50),
						"$lte", int32(100),
					),
				),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x4f, 0x00, 0x00, 0x00, 0x4f},
					"actor_id", int32(79),
					"first_name", "MAE",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			},
		},
		"NinEqNe": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.MustMakeDocument(
						"$nin", types.Array{"NEESON"},
						"$ne", "AKROYD",
					),
					"first_name", types.MustMakeDocument(
						"$eq", "CHRISTIAN",
					),
				),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00, 0x0a},
					"actor_id", int32(10),
					"first_name", "CHRISTIAN",
					"last_name", "GABLE",
					"last_update", lastUpdate,
				),
			},
		},
		"Not": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.MustMakeDocument(
						"$not", types.MustMakeDocument(
							"$eq", "GUINESS",
						),
					),
				),
				"sort", types.MustMakeDocument(
					"actor_id", int32(1),
				),
				"limit", int32(1),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				),
			},
		},
		"NestedNot": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.MustMakeDocument(
						"$not", types.MustMakeDocument(
							"$not", types.MustMakeDocument(
								"$not", types.MustMakeDocument(
									"$eq", "GUINESS",
								),
							),
						),
					),
				),
				"sort", types.MustMakeDocument(
					"actor_id", int32(1),
				),
				"limit", int32(1),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				),
			},
		},
		"AndOr": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"$and", types.Array{
						types.MustMakeDocument(
							"first_name", "CHRISTIAN",
						),
						types.MustMakeDocument(
							"$or", types.Array{
								types.MustMakeDocument(
									"last_name", "GABLE",
								),
								types.MustMakeDocument(
									"last_name", "NEESON",
								),
							},
						),
					},
				),
				"sort", types.MustMakeDocument(
					"actor_id", int32(1),
				),
				"limit", int32(1),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00, 0x0a},
					"actor_id", int32(10),
					"first_name", "CHRISTIAN",
					"last_name", "GABLE",
					"last_update", lastUpdate,
				),
			},
		},
		"Nor": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"$nor", types.Array{
						types.MustMakeDocument("actor_id", types.MustMakeDocument("$gt", int32(2))),
						types.MustMakeDocument("first_name", "PENELOPE"),
					},
				),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				),
			},
		},
		"ValueRegex": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.Regex{Pattern: "hoffman", Options: "i"},
				),
				"sort", types.MustMakeDocument(
					"actor_id", int32(1),
				),
				"limit", int32(1),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			},
		},
		"Regex": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.MustMakeDocument(
						"$regex", types.Regex{Pattern: "hoffman", Options: "i"},
					),
				),
				"sort", types.MustMakeDocument(
					"actor_id", int32(1),
				),
				"limit", int32(1),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			},
		},
		"RegexOptions": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.MustMakeDocument(
						"$regex", types.Regex{Pattern: "hoffman"},
						"$options", "i",
					),
				),
				"sort", types.MustMakeDocument(
					"actor_id", int32(1),
				),
				"limit", int32(1),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			},
		},
		"RegexStringOptions": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.MustMakeDocument(
						"$regex", "hoffman",
						"$options", "i",
					),
				),
				"sort", types.MustMakeDocument(
					"actor_id", int32(1),
				),
				"limit", int32(1),
			),
			resp: types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			},
		},
	}

	for name, tc := range testCases { //nolint:paralleltest // false positive
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, schema := range []string{"monila", "pagila"} {
				t.Run(schema, func(t *testing.T) {
					// not parallel because we modify tc

					tc.req.Set("$db", schema)

					if schema == "pagila" {
						for i, doc := range tc.resp {
							d := doc.(types.Document)
							d.Remove("_id")
							tc.resp[i] = d
						}
					}

					reqHeader := wire.MsgHeader{
						RequestID: 1,
						OpCode:    wire.OP_MSG,
					}

					var reqMsg wire.OpMsg
					err := reqMsg.SetSections(wire.OpMsgSection{
						Documents: []types.Document{tc.req},
					})
					require.NoError(t, err)

					_, resBody, closeConn := handler.Handle(ctx, &reqHeader, &reqMsg)
					require.False(t, closeConn, "%s", wire.DumpMsgBody(resBody))

					actual, err := resBody.(*wire.OpMsg).Document()
					require.NoError(t, err)

					expected := types.MustMakeDocument(
						"cursor", types.MustMakeDocument(
							"firstBatch", tc.resp,
							"id", int64(0),
							"ns", schema+".actor",
						),
						"ok", float64(1),
					)
					assert.Equal(t, expected, actual)
				})
			}
		})
	}
}

func TestReadOnlyHandlers(t *testing.T) {
	t.Parallel()
	ctx, handler, _ := setup(t, &testutil.PoolOpts{
		ReadOnly: true,
	})

	type testCase struct {
		req         types.Document
		reqSetDB    bool
		resp        types.Document
		compareFunc func(t testing.TB, actual, expected any)
	}

	testCases := map[string]testCase{
		"CountAllActors": {
			req: types.MustMakeDocument(
				"count", "actor",
			),
			reqSetDB: true,
			resp: types.MustMakeDocument(
				"n", int32(200),
				"ok", float64(1),
			),
		},
		"CountExactlyOneActor": {
			req: types.MustMakeDocument(
				"count", "actor",
				"query", types.MustMakeDocument(
					"actor_id", int32(28),
				),
			),
			reqSetDB: true,
			resp: types.MustMakeDocument(
				"n", int32(1),
				"ok", float64(1),
			),
		},
		"CountLastNameHoffman": {
			req: types.MustMakeDocument(
				"count", "actor",
				"query", types.MustMakeDocument(
					"last_name", "HOFFMAN",
				),
			),
			reqSetDB: true,
			resp: types.MustMakeDocument(
				"n", int32(3),
				"ok", float64(1),
			),
		},

		"GetParameter": {
			req: types.MustMakeDocument(
				"getParameter", int32(1),
			),
			resp: types.MustMakeDocument(
				"version", "5.0.42",
				"ok", float64(1),
			),
		},

		"ListDatabases": {
			req: types.MustMakeDocument(
				"listDatabases", int32(1),
			),
			resp: types.MustMakeDocument(
				"databases", types.Array{
					types.MustMakeDocument(
						"name", "monila",
						"sizeOnDisk", int64(13516800),
						"empty", false,
					),
					types.MustMakeDocument(
						"name", "pagila",
						"sizeOnDisk", int64(7127040),
						"empty", false,
					),
					types.MustMakeDocument(
						"name", "test",
						"sizeOnDisk", int64(0),
						"empty", true,
					),
				},
				"totalSize", int64(30114595),
				"totalSizeMb", int64(28),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, expected, actual any) {
				expectedDoc, actualDoc := expected.(types.Document), actual.(types.Document)

				testutil.CompareAndSetByPath(t, expectedDoc, actualDoc, 1_000_000, "totalSize")
				testutil.CompareAndSetByPath(t, expectedDoc, actualDoc, 1, "totalSizeMb")

				expectedDBs := testutil.GetByPath(t, expectedDoc, "databases").(types.Array)
				actualDBs := testutil.GetByPath(t, actualDoc, "databases").(types.Array)
				require.Equal(t, len(expectedDBs), len(actualDBs))
				for i, actualDB := range actualDBs {
					testutil.CompareAndSetByPath(t, expectedDBs[i], actualDB, 200_000, "sizeOnDisk")
				}

				assert.Equal(t, expectedDoc, actualDoc)
			},
		},

		"ServerStatus": {
			req: types.MustMakeDocument(
				"serverStatus", int32(1),
			),
			resp: types.MustMakeDocument(
				"version", "5.0.42",
				"ok", float64(1),
			),
		},
	}

	for name, tc := range testCases { //nolint:paralleltest // false positive
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, schema := range []string{"monila", "pagila"} {
				t.Run(schema, func(t *testing.T) {
					if tc.reqSetDB {
						tc.req.Set("$db", schema)
					}

					reqHeader := wire.MsgHeader{
						RequestID: 1,
						OpCode:    wire.OP_MSG,
					}

					var reqMsg wire.OpMsg
					err := reqMsg.SetSections(wire.OpMsgSection{
						Documents: []types.Document{tc.req},
					})
					require.NoError(t, err)

					_, resBody, closeConn := handler.Handle(ctx, &reqHeader, &reqMsg)
					require.False(t, closeConn, "%s", wire.DumpMsgBody(resBody))

					actual, err := resBody.(*wire.OpMsg).Document()
					require.NoError(t, err)
					if tc.compareFunc == nil {
						assert.Equal(t, tc.resp, actual)
					} else {
						tc.compareFunc(t, tc.resp, actual)
					}
				})
			}
		})
	}
}
func TestCreate(t *testing.T) { //nolint:paralleltest,tparallel // affects a global list of databases
	ctx, handler, pool := setup(t, nil)

	type testCase struct {
		req     types.Document
		resp    types.Document
		success bool
	}

	dbName := "test_create_db"
	collectionNew := "new_collection"
	collectionExisted := "existed_collection"

	testCases := map[string]testCase{
		"new-collection": {
			req: types.MustMakeDocument(
				"create", collectionNew,
				"$db", dbName,
			),
			success: true,
			resp: types.MustMakeDocument(
				"created", fmt.Sprintf("%s.%s", dbName, collectionNew),
				"ok", float64(1),
			),
		},
		"existed-collection": {
			req: types.MustMakeDocument(
				"create", collectionExisted,
				"$db", dbName,
			),
			success: false,
			resp: types.MustMakeDocument(
				"ns", collectionExisted,
				"ok", float64(1),
			),
		},
	}

	require.NoError(t, dropDB(ctx, pool, dbName))
	require.NoError(t, createDB(ctx, pool, dbName))
	require.NoError(t, shared.CreateCollection(ctx, pool, dbName, collectionExisted))

	for name, tc := range testCases { //nolint:paralleltest // false positive
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reqHeader := wire.MsgHeader{
				RequestID: 1,
				OpCode:    wire.OP_MSG,
			}

			var reqMsg wire.OpMsg
			require.NoError(t, reqMsg.SetSections(wire.OpMsgSection{
				Documents: []types.Document{tc.req},
			}))

			_, resBody, closeConn := handler.Handle(ctx, &reqHeader, &reqMsg)
			if !tc.success {
				require.True(t, closeConn, "%s", wire.DumpMsgBody(resBody))
				return
			}

			require.False(t, closeConn, "%s", wire.DumpMsgBody(resBody))
			actual, err := resBody.(*wire.OpMsg).Document()
			require.NoError(t, err)

			assert.Equal(t, actual.Map(), tc.resp.Map())

			// checking tables
			query := `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = $1 and table_name = $2`
			var count int
			require.NoError(t, pool.QueryRow(ctx, query, dbName, collectionNew).Scan(&count))
			assert.Equal(t, count, 1)
		})
	}
}
