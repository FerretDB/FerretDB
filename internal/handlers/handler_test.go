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
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handlers/jsonb1"
	"github.com/FerretDB/FerretDB/internal/handlers/shared"
	"github.com/FerretDB/FerretDB/internal/handlers/sql"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/version"
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

func handle(ctx context.Context, t *testing.T, handler *Handler, req types.Document) types.Document {
	t.Helper()

	reqHeader := wire.MsgHeader{
		RequestID: 1,
		OpCode:    wire.OP_MSG,
	}

	var reqMsg wire.OpMsg
	err := reqMsg.SetSections(wire.OpMsgSection{
		Documents: []types.Document{req},
	})
	require.NoError(t, err)

	_, resBody, closeConn := handler.Handle(ctx, &reqHeader, &reqMsg)
	require.False(t, closeConn, "%s", wire.DumpMsgBody(resBody))

	actual, err := resBody.(*wire.OpMsg).Document()
	require.NoError(t, err)

	return actual
}

func TestFind(t *testing.T) {
	t.Parallel()
	ctx, handler, _ := setup(t, &testutil.PoolOpts{
		ReadOnly: true,
	})

	lastUpdate := time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local()

	type testCase struct {
		req  types.Document
		resp *types.Array
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
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x4f, 0x00, 0x00, 0x00, 0x4f},
					"actor_id", int32(79),
					"first_name", "MAE",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			),
		},
		"InLteGte": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.MustMakeDocument(
						"$in", types.MustNewArray("HOFFMAN"),
					),
					"actor_id", types.MustMakeDocument(
						"$gte", int32(50),
						"$lte", int32(100),
					),
				),
			),
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x4f, 0x00, 0x00, 0x00, 0x4f},
					"actor_id", int32(79),
					"first_name", "MAE",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			),
		},
		"NinEqNe": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"last_name", types.MustMakeDocument(
						"$nin", types.MustNewArray("NEESON"),
						"$ne", "AKROYD",
					),
					"first_name", types.MustMakeDocument(
						"$eq", "CHRISTIAN",
					),
				),
			),
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00, 0x0a},
					"actor_id", int32(10),
					"first_name", "CHRISTIAN",
					"last_name", "GABLE",
					"last_update", lastUpdate,
				),
			),
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
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				),
			),
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
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				),
			),
		},
		"AndOr": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"$and", types.MustNewArray(
						types.MustMakeDocument(
							"first_name", "CHRISTIAN",
						),
						types.MustMakeDocument(
							"$or", types.MustNewArray(
								types.MustMakeDocument(
									"last_name", "GABLE",
								),
								types.MustMakeDocument(
									"last_name", "NEESON",
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
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00, 0x0a},
					"actor_id", int32(10),
					"first_name", "CHRISTIAN",
					"last_name", "GABLE",
					"last_update", lastUpdate,
				),
			),
		},
		"Nor": {
			req: types.MustMakeDocument(
				"find", "actor",
				"filter", types.MustMakeDocument(
					"$nor", types.MustNewArray(
						types.MustMakeDocument("actor_id", types.MustMakeDocument("$gt", int32(2))),
						types.MustMakeDocument("first_name", "PENELOPE"),
					),
				),
			),
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				),
			),
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
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			),
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
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			),
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
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			),
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
			resp: types.MustNewArray(
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			),
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
						for i := 0; i < tc.resp.Len(); i++ {
							doc, err := tc.resp.Get(i)
							require.NoError(t, err)
							d := doc.(types.Document)
							d.Remove("_id")
							err = tc.resp.Set(i, d)
							require.NoError(t, err)
						}
					}

					actual := handle(ctx, t, handler, tc.req)
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
		compareFunc func(t testing.TB, req types.Document, actual, expected types.CompositeType)
	}

	hostname, err := os.Hostname()
	require.NoError(t, err)

	testCases := map[string]testCase{
		"BuildInfo": {
			req: types.MustMakeDocument(
				"buildInfo", int32(1),
			),
			resp: types.MustMakeDocument(
				"version", "5.0.42",
				"gitVersion", version.Get().Commit,
				"versionArray", types.MustNewArray(int32(5), int32(0), int32(42), int32(0)),
				"bits", int32(strconv.IntSize),
				"maxBsonObjectSize", int32(bson.MaxDocumentLen),
				"ok", float64(1),
			),
		},
		"CollStats": {
			req: types.MustMakeDocument(
				"collstats", "film",
			),
			reqSetDB: true,
			resp: types.MustMakeDocument(
				"ns", "manila.film",
				"count", int64(1_000),
				"size", int64(704512),
				"storageSize", int64(450560),
				"totalIndexSize", int64(221184),
				"totalSize", int64(704512),
				"scaleFactor", int64(1),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, req types.Document, actual, expected types.CompositeType) {
				db, err := req.Get("$db")
				require.NoError(t, err)
				if db.(string) == "manila" {
					assert.Equal(t, expected, actual)
				}
			},
		},
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

		"GetLog": {
			req: types.MustMakeDocument(
				"getLog", "startupWarnings",
			),
			resp: types.MustMakeDocument(
				"totalLinesWritten", int32(2),
				// will be replaced with the real value during the test
				"log", types.MakeArray(2),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _ types.Document, actual, expected types.CompositeType) {
				// Just testing "ok" response, not the body of the response
				actualV := testutil.GetByPath(t, actual, "log")
				testutil.SetByPath(t, expected, actualV, "log")
				assert.Equal(t, expected, actual)
			},
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

		"IsMaster": {
			req: types.MustMakeDocument(
				"isMaster", int32(1),
			),
			resp: types.MustMakeDocument(
				"helloOk", true,
				"ismaster", true,
				"maxBsonObjectSize", int32(bson.MaxDocumentLen),
				"maxMessageSizeBytes", int32(wire.MaxMsgLen),
				"maxWriteBatchSize", int32(100000),
				"localTime", time.Now(),
				"minWireVersion", int32(13),
				"maxWireVersion", int32(13),
				"readOnly", false,
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _ types.Document, actual, expected types.CompositeType) {
				testutil.CompareAndSetByPathTime(t, expected, actual, time.Second, "localTime")
				assert.Equal(t, expected, actual)
			},
		},
		"Hello": {
			req: types.MustMakeDocument(
				"hello", int32(1),
			),
			resp: types.MustMakeDocument(
				"helloOk", true,
				"ismaster", true,
				"maxBsonObjectSize", int32(bson.MaxDocumentLen),
				"maxMessageSizeBytes", int32(wire.MaxMsgLen),
				"maxWriteBatchSize", int32(100000),
				"localTime", time.Now(),
				"minWireVersion", int32(13),
				"maxWireVersion", int32(13),
				"readOnly", false,
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _ types.Document, actual, expected types.CompositeType) {
				testutil.CompareAndSetByPathTime(t, expected, actual, time.Second, "localTime")
				assert.Equal(t, expected, actual)
			},
		},

		"HostInfo": {
			req: types.MustMakeDocument(
				"hostInfo", int32(1),
			),
			resp: types.MustMakeDocument(
				"system", types.MustMakeDocument(
					"currentTime", time.Now(),
					"hostname", hostname,
					"cpuAddrSize", int32(strconv.IntSize),
					"numCores", int32(runtime.NumCPU()),
					"cpuArch", runtime.GOARCH,
					"numaEnabled", false,
				),
				"os", types.MustMakeDocument(
					"type", strings.Title(runtime.GOOS),
				),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _ types.Document, actual, expected types.CompositeType) {
				testutil.CompareAndSetByPathTime(t, expected, actual, time.Second, "system", "currentTime")
				assert.Equal(t, expected, actual)
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
					// not parallel because we modify tc

					if tc.reqSetDB {
						tc.req.Set("$db", schema)
					}

					actual := handle(ctx, t, handler, tc.req)
					if tc.compareFunc == nil {
						assert.Equal(t, tc.resp, actual)
					} else {
						tc.compareFunc(t, tc.req, tc.resp, actual)
					}
				})
			}
		})
	}
}

//nolint:paralleltest // we test a global list of databases
func TestListDropDatabase(t *testing.T) {
	ctx, handler, pool := setup(t, nil)

	t.Run("existing", func(t *testing.T) {
		db := testutil.Schema(ctx, t, pool)

		actualList := handle(ctx, t, handler, types.MustMakeDocument(
			"listDatabases", int32(1),
		))
		expectedList := types.MustMakeDocument(
			"databases", types.MustNewArray(
				types.MustMakeDocument(
					"name", "monila",
					"sizeOnDisk", int64(13_631_488),
					"empty", false,
				),
				types.MustMakeDocument(
					"name", "pagila",
					"sizeOnDisk", int64(7_127_040),
					"empty", false,
				),
				types.MustMakeDocument(
					"name", "test",
					"sizeOnDisk", int64(0),
					"empty", true,
				),
				types.MustMakeDocument(
					"name", db,
					"sizeOnDisk", int64(0),
					"empty", true,
				),
			),
			"totalSize", int64(30_286_627),
			"totalSizeMb", int64(28),
			"ok", float64(1),
		)

		testutil.CompareAndSetByPathNum(t, expectedList, actualList, 2_000_000, "totalSize")
		testutil.CompareAndSetByPathNum(t, expectedList, actualList, 2, "totalSizeMb")

		expectedDBs := testutil.GetByPath(t, expectedList, "databases").(*types.Array)
		actualDBs := testutil.GetByPath(t, actualList, "databases").(*types.Array)
		require.Equal(t, expectedDBs.Len(), actualDBs.Len())
		for i := 0; i < actualDBs.Len(); i++ {
			actualDB, err := actualDBs.Get(i)
			require.NoError(t, err)
			expectedDB, err := expectedDBs.Get(i)
			require.NoError(t, err)
			testutil.CompareAndSetByPathNum(t, expectedDB.(types.Document), actualDB.(types.Document), 500_000, "sizeOnDisk")
		}

		assert.Equal(t, expectedList, actualList)

		actualDrop := handle(ctx, t, handler, types.MustMakeDocument(
			"dropDatabase", int32(1),
			"$db", db,
		))
		expectedDrop := types.MustMakeDocument(
			"dropped", db,
			"ok", float64(1),
		)
		assert.Equal(t, expectedDrop, actualDrop)

		databases := testutil.GetByPath(t, expectedList, "databases").(*types.Array)
		databases, err := databases.Subslice(0, databases.Len()-1)
		require.NoError(t, err)
		testutil.SetByPath(t, expectedList, databases, "databases")

		actualList = handle(ctx, t, handler, types.MustMakeDocument(
			"listDatabases", int32(1),
		))
		assert.Equal(t, expectedList, actualList)
	})

	t.Run("nonexisting", func(t *testing.T) {
		actual := handle(ctx, t, handler, types.MustMakeDocument(
			"dropDatabase", int32(1),
			"$db", "nonexisting",
		))
		expected := types.MustMakeDocument(
			// no $db
			"ok", float64(1),
		)
		assert.Equal(t, expected, actual)
	})
}

//nolint:paralleltest // we test a global list of collections
func TestCreateListDropCollection(t *testing.T) {
	ctx, handler, pool := setup(t, nil)
	db := testutil.Schema(ctx, t, pool)

	t.Run("nonexisting", func(t *testing.T) {
		collection := testutil.TableName(t)

		actual := handle(ctx, t, handler, types.MustMakeDocument(
			"create", collection,
			"$db", db,
		))
		expected := types.MustMakeDocument(
			"ok", float64(1),
		)
		assert.Equal(t, expected, actual)

		// TODO test listCollections command once we have better cursor support
		// https://github.com/FerretDB/FerretDB/issues/79

		tables, err := pool.Tables(ctx, db)
		require.NoError(t, err)
		assert.Equal(t, []string{collection}, tables)

		actual = handle(ctx, t, handler, types.MustMakeDocument(
			"drop", collection,
			"$db", db,
		))
		expected = types.MustMakeDocument(
			"nIndexesWas", int32(1),
			"ns", db+"."+collection,
			"ok", float64(1),
		)
		assert.Equal(t, expected, actual)

		actual = handle(ctx, t, handler, types.MustMakeDocument(
			"drop", collection,
			"$db", db,
		))
		expected = types.MustMakeDocument(
			"ok", float64(0),
			"errmsg", "ns not found",
			"code", int32(26),
			"codeName", "NamespaceNotFound",
		)
		assert.Equal(t, expected, actual)
	})

	t.Run("existing", func(t *testing.T) {
		collection := testutil.CreateTable(ctx, t, pool, db)

		actual := handle(ctx, t, handler, types.MustMakeDocument(
			"create", collection,
			"$db", db,
		))
		expected := types.MustMakeDocument(
			"ok", float64(0),
			"errmsg", "Collection already exists. NS: testcreatelistdropcollection.testcreatelistdropcollection_existing",
			"code", int32(48),
			"codeName", "NamespaceExists",
		)
		assert.Equal(t, expected, actual)
	})
}
