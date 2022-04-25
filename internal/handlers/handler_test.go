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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/version"
	"github.com/FerretDB/FerretDB/internal/wire"
)

type setupOpts struct {
	noLogging bool
	poolOpts  *testutil.PoolOpts
}

// setup creates shared objects for testing.
//
// Using shared objects helps us spot concurrency bugs.
// If some test is failing and the log output is confusing, and you are tempted to move setup call to subtest,
// instead run that single test with `go test -run test/name`.
func setup(t testing.TB, opts *setupOpts) (context.Context, *pg.Handler, *pgdb.Pool) {
	t.Helper()

	if opts == nil {
		opts = new(setupOpts)
	}

	var l *zap.Logger
	if opts.noLogging {
		l = zap.NewNop()
	} else {
		l = zaptest.NewLogger(t)
	}

	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t, opts.poolOpts, l)
	handler := pg.New(&pg.NewOpts{
		PgPool:   pool,
		L:        l,
		PeerAddr: "127.0.0.1:12345",
		Metrics:  pg.NewMetrics(),
	})

	return ctx, handler, pool
}

func handle(ctx context.Context, t *testing.T, handler *pg.Handler, req *types.Document) *types.Document {
	t.Helper()

	var reqMsg wire.OpMsg
	err := reqMsg.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{req},
	})
	require.NoError(t, err)

	b, err := reqMsg.MarshalBinary()
	require.NoError(t, err)

	reqHeader := wire.MsgHeader{
		MessageLength: int32(wire.MsgHeaderLen + len(b)),
		RequestID:     1,
		OpCode:        wire.OP_MSG,
	}

	// TODO
	// addToSeedCorpus(t, &reqHeader, &reqMsg)

	_, resBody, closeConn := handler.Handle(ctx, &reqHeader, &reqMsg)
	require.False(t, closeConn, "%s", resBody.String())

	actual, err := resBody.(*wire.OpMsg).Document()
	require.NoError(t, err)

	return actual
}

func TestFind(t *testing.T) {
	t.Parallel()
	ctx, handler, _ := setup(t, &setupOpts{
		poolOpts: &testutil.PoolOpts{
			ReadOnly: true,
		},
	})

	lastUpdate := time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local()

	type testCase struct {
		schemas []string
		req     *types.Document
		resp    *types.Array
		err     error
	}

	// Do not use sentences, spaces, or underscores in subtest names
	// to make it easier to run individual tests with `go test -run test/name` and for consistency.
	testCases := map[string]testCase{
		"ProjectionElemMatch": {
			schemas: []string{"values"},
			req: types.MustNewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument("name", "array-embedded")),
				"projection", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$elemMatch", must.NotFail(types.NewDocument("document", "jkl")),
					)),
				)),
			),
			resp: types.MustNewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x05, 0x00, 0x00, 0x04, 0x05},
					"value", must.NotFail(types.NewArray(
						must.NotFail(types.NewDocument("document", "jkl", "score", int32(24), "age", int32(1002))),
					)),
				)),
			),
		},
		"InLteGte": {
			schemas: []string{"monila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"last_name", types.MustNewDocument(
						"$in", types.MustNewArray("HOFFMAN"),
					),
					"actor_id", types.MustNewDocument(
						"$gte", int32(50),
						"$lte", int32(100),
					),
				),
			),
			resp: types.MustNewArray(
				types.MustNewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x4f, 0x00, 0x00, 0x00, 0x4f},
					"actor_id", int32(79),
					"first_name", "MAE",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			),
		},
		"NinEqNe": {
			schemas: []string{"monila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"last_name", types.MustNewDocument(
						"$nin", types.MustNewArray("NEESON"),
						"$ne", "AKROYD",
					),
					"first_name", types.MustNewDocument(
						"$eq", "CHRISTIAN",
					),
				),
			),
			resp: types.MustNewArray(
				types.MustNewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00, 0x0a},
					"actor_id", int32(10),
					"first_name", "CHRISTIAN",
					"last_name", "GABLE",
					"last_update", lastUpdate,
				),
			),
		},
		"Not": {
			schemas: []string{"monila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"last_name", types.MustNewDocument(
						"$not", types.MustNewDocument(
							"$eq", "GUINESS",
						),
					),
				),
				"sort", types.MustNewDocument(
					"actor_id", int32(1),
				),
				"limit", int32(1),
			),
			resp: types.MustNewArray(
				types.MustNewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				),
			),
		},
		"NestedNot": {
			schemas: []string{"monila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"last_name", types.MustNewDocument(
						"$not", types.MustNewDocument(
							"$not", types.MustNewDocument(
								"$not", types.MustNewDocument(
									"$eq", "GUINESS",
								),
							),
						),
					),
				),
				"sort", types.MustNewDocument(
					"actor_id", int32(1),
				),
				"limit", int64(1),
			),
			resp: types.MustNewArray(
				types.MustNewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				),
			),
		},
		"AndOr": {
			schemas: []string{"monila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"$and", types.MustNewArray(
						types.MustNewDocument(
							"first_name", "CHRISTIAN",
						),
						types.MustNewDocument(
							"$or", types.MustNewArray(
								types.MustNewDocument(
									"last_name", "GABLE",
								),
								types.MustNewDocument(
									"last_name", "NEESON",
								),
							),
						),
					),
				),
				"sort", types.MustNewDocument(
					"actor_id", int32(1),
				),
				"limit", float64(1),
			),
			resp: types.MustNewArray(
				types.MustNewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00, 0x0a},
					"actor_id", int32(10),
					"first_name", "CHRISTIAN",
					"last_name", "GABLE",
					"last_update", lastUpdate,
				),
			),
		},
		"Nor": {
			schemas: []string{"monila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"$nor", types.MustNewArray(
						types.MustNewDocument("actor_id", types.MustNewDocument("$gt", int32(2))),
						types.MustNewDocument("first_name", "PENELOPE"),
					),
				),
			),
			resp: types.MustNewArray(
				types.MustNewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02},
					"actor_id", int32(2),
					"first_name", "NICK",
					"last_name", "WAHLBERG",
					"last_update", lastUpdate,
				),
			),
		},
		"Regex": {
			schemas: []string{"monila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"last_name", types.MustNewDocument(
						"$regex", types.Regex{Pattern: "hoffman", Options: "i"},
					),
				),
				"sort", types.MustNewDocument(
					"actor_id", int32(1),
				),
				"limit", int64(1),
			),
			resp: types.MustNewArray(
				types.MustNewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			),
		},
		"RegexOptions": {
			schemas: []string{"monila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"last_name", types.MustNewDocument(
						"$regex", types.Regex{Pattern: "hoffman"},
						"$options", "i",
					),
				),
				"sort", types.MustNewDocument(
					"actor_id", int32(1),
				),
				"limit", float64(1),
			),
			resp: types.MustNewArray(
				types.MustNewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			),
		},
		"RegexStringOptions": {
			schemas: []string{"monila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"last_name", types.MustNewDocument(
						"$regex", "hoffman",
						"$options", "i",
					),
				),
				"sort", types.MustNewDocument(
					"actor_id", int32(1),
				),
				"limit", int32(1),
			),
			resp: types.MustNewArray(
				types.MustNewDocument(
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

			require.NotEmpty(t, tc.schemas)
			require.NotEmpty(t, tc.req)

			collection := must.NotFail(tc.req.Get(tc.req.Command())).(string)
			require.NotEmpty(t, collection)

			for _, schema := range tc.schemas {
				t.Run(schema, func(t *testing.T) {
					// not parallel because we modify tc

					tc.req.Set("$db", schema)

					var expected *types.Document
					if tc.err == nil {
						expected = types.MustNewDocument(
							"cursor", types.MustNewDocument(
								"firstBatch", tc.resp,
								"id", int64(0),
								"ns", schema+"."+collection,
							),
							"ok", float64(1),
						)
					} else {
						require.Nil(t, tc.resp)
						pErr, ok := common.ProtocolError(tc.err)
						require.True(t, ok)
						expected = pErr.Document()
					}

					actual := handle(ctx, t, handler, tc.req)
					testutil.AssertEqual(t, expected, actual)
				})
			}
		})
	}
}

func TestReadOnlyHandlers(t *testing.T) {
	t.Parallel()
	ctx, handler, _ := setup(t, &setupOpts{
		poolOpts: &testutil.PoolOpts{
			ReadOnly: true,
		},
	})

	type testCase struct {
		req         *types.Document
		reqSetDB    bool
		resp        *types.Document
		compareFunc func(t testing.TB, req, expected, actual *types.Document)
	}

	testCases := map[string]testCase{
		"BuildInfo": {
			req: types.MustNewDocument(
				"buildInfo", int32(1),
			),
			resp: types.MustNewDocument(
				"version", "5.0.42",
				"gitVersion", version.Get().Commit,
				"modules", must.NotFail(types.NewArray()),
				"sysInfo", "deprecated",
				"versionArray", must.NotFail(types.NewArray(int32(5), int32(0), int32(42), int32(0))),
				"bits", int32(strconv.IntSize),
				"debug", version.Get().Debug,
				"maxBsonObjectSize", int32(16777216),
				"buildEnvironment", must.NotFail(types.NewDocument()),
				"ok", float64(1),
			),
		},

		"CollStats": {
			req: types.MustNewDocument(
				"collStats", "film",
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"ns", "monila.film",
				"count", int32(1_000),
				"size", int32(1_228_800),
				"storageSize", int32(1_196_032),
				"totalIndexSize", int32(0),
				"totalSize", int32(1_228_800),
				"scaleFactor", int32(1),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _, expected, actual *types.Document) {
				testutil.CompareAndSetByPathNum(t, expected, actual, 300, "count") // that's not a number of rows
				testutil.CompareAndSetByPathNum(t, expected, actual, 32_768, "size")
				testutil.CompareAndSetByPathNum(t, expected, actual, 32_768, "storageSize")
				testutil.CompareAndSetByPathNum(t, expected, actual, 32_768, "totalSize")
				testutil.AssertEqual(t, expected, actual)
			},
		},

		"CountAllActors": {
			req: types.MustNewDocument(
				"count", "actor",
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"n", int32(200),
				"ok", float64(1),
			),
		},
		"CountExactlyOneActor": {
			req: types.MustNewDocument(
				"count", "actor",
				"query", types.MustNewDocument(
					"actor_id", int32(28),
				),
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"n", int32(1),
				"ok", float64(1),
			),
		},
		"CountLastNameHoffman": {
			req: types.MustNewDocument(
				"count", "actor",
				"query", types.MustNewDocument(
					"last_name", "HOFFMAN",
				),
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"n", int32(3),
				"ok", float64(1),
			),
		},
		"DataSize": {
			req: types.MustNewDocument(
				"dataSize", "monila.actor",
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"estimate", false,
				"size", int32(106_496),
				"numObjects", int32(210),
				"millis", int32(20),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _, expected, actual *types.Document) {
				testutil.CompareAndSetByPathNum(t, expected, actual, 10, "numObjects")
				testutil.CompareAndSetByPathNum(t, expected, actual, 50, "millis")
				testutil.CompareAndSetByPathNum(t, expected, actual, 32_768, "size")
				testutil.AssertEqual(t, expected, actual)
			},
		},
		"DataSizeCollectionNotExist": {
			req: types.MustNewDocument(
				"dataSize", "some-database.some-collection",
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"size", int32(0),
				"numObjects", int32(0),
				"millis", int32(20),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _, expected, actual *types.Document) {
				testutil.CompareAndSetByPathNum(t, expected, actual, 50, "millis")
				testutil.AssertEqual(t, expected, actual)
			},
		},

		"DBStats": {
			req: types.MustNewDocument(
				"dbStats", int32(1),
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"db", "monila",
				"collections", int32(14),
				"views", int32(0),
				"objects", int32(31_000),
				"avgObjSize", 433.0,
				"dataSize", 13_107_200.0,
				"indexes", int32(0),
				"indexSize", float64(0),
				"totalSize", 13_492_224.0,
				"scaleFactor", float64(1),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _, expected, actual *types.Document) {
				testutil.CompareAndSetByPathNum(t, expected, actual, 2_000, "objects")
				testutil.CompareAndSetByPathNum(t, expected, actual, 40, "avgObjSize")
				testutil.CompareAndSetByPathNum(t, expected, actual, 400_000, "dataSize")
				testutil.CompareAndSetByPathNum(t, expected, actual, 400_000, "totalSize")
				testutil.AssertEqual(t, expected, actual)
			},
		},
		"DBStatsWithScale": {
			req: types.MustNewDocument(
				"dbStats", int32(1),
				"scale", float64(1_000),
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"db", "monila",
				"collections", int32(14),
				"views", int32(0),
				"objects", int32(31_000),
				"avgObjSize", 433.0,
				"dataSize", 13_107.200,
				"indexes", int32(0),
				"indexSize", float64(0),
				"totalSize", 13_492.224,
				"scaleFactor", float64(1_000),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _, expected, actual *types.Document) {
				testutil.CompareAndSetByPathNum(t, expected, actual, 2_000, "objects")
				testutil.CompareAndSetByPathNum(t, expected, actual, 40, "avgObjSize")
				testutil.CompareAndSetByPathNum(t, expected, actual, 400, "dataSize")
				testutil.CompareAndSetByPathNum(t, expected, actual, 400, "totalSize")
				testutil.AssertEqual(t, expected, actual)
			},
		},

		"FindProjectionInclusions": {
			req: types.MustNewDocument(
				"find", "actor",
				"projection", types.MustNewDocument(
					"last_name", int32(1),
					"last_update", true,
				),
				"filter", types.MustNewDocument(
					"actor_id", int32(28),
				),
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"cursor", types.MustNewDocument(
					"firstBatch", types.MustNewArray(
						types.MustNewDocument(
							"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
							"last_name", "HOFFMAN",
							"last_update", time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local(),
						),
					),
					"id", int64(0),
					"ns", "", // set by compareFunc
				),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _, expected, actual *types.Document) {
				actualV := testutil.GetByPath(t, actual, "cursor", "ns")
				testutil.SetByPath(t, expected, actualV, "cursor", "ns")
				testutil.AssertEqual(t, expected, actual)
			},
		},

		"FindProjectionExclusions": {
			req: types.MustNewDocument(
				"find", "actor",
				"projection", types.MustNewDocument(
					"first_name", int32(0),
					"actor_id", false,
				),
				"filter", types.MustNewDocument(
					"actor_id", int32(28),
				),
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"cursor", types.MustNewDocument(
					"firstBatch", types.MustNewArray(
						types.MustNewDocument(
							"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
							"last_name", "HOFFMAN",
							"last_update", time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local(),
						),
					),
					"id", int64(0),
					"ns", "", // set by compareFunc
				),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _, expected, actual *types.Document) {
				actualV := testutil.GetByPath(t, actual, "cursor", "ns")
				testutil.SetByPath(t, expected, actualV, "cursor", "ns")
				testutil.AssertEqual(t, expected, actual)
			},
		},

		"FindProjectionIDInclusion": {
			req: types.MustNewDocument(
				"find", "actor",
				"projection", types.MustNewDocument(
					"_id", false,
					"actor_id", int32(1),
				),
				"filter", types.MustNewDocument(
					"actor_id", int32(28),
				),
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"cursor", types.MustNewDocument(
					"firstBatch", types.MustNewArray(
						types.MustNewDocument(
							"actor_id", int32(28),
						),
					),
					"id", int64(0),
					"ns", "", // set by compareFunc
				),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _, expected, actual *types.Document) {
				actualV := testutil.GetByPath(t, actual, "cursor", "ns")
				testutil.SetByPath(t, expected, actualV, "cursor", "ns")
				testutil.AssertEqual(t, expected, actual)
			},
		},

		"ServerStatus": {
			req: must.NotFail(types.NewDocument(
				"serverStatus", int32(1),
			)),
			reqSetDB: true,
			resp: must.NotFail(types.NewDocument(
				"host", "",
				"version", "5.0.42",
				"process", "handlers.test",
				"pid", int64(0),
				"uptime", int64(0),
				"uptimeMillis", int64(0),
				"uptimeEstimate", int64(0),
				"localTime", time.Now(),
				"catalogStats", must.NotFail(types.NewDocument(
					"collections", int32(28),
					"capped", int32(0),
					"timeseries", int32(0),
					"views", int32(0),
					"internalCollections", int32(0),
					"internalViews", int32(0),
				)),
				"freeMonitoring", must.NotFail(types.NewDocument(
					"state", "disabled",
				)),
				"ok", float64(1),
			)),
			compareFunc: func(t testing.TB, _ *types.Document, actual, expected *types.Document) {
				for _, key := range []string{"host", "pid", "uptime", "uptimeMillis", "uptimeEstimate"} {
					actualV := testutil.GetByPath(t, actual, key)
					testutil.SetByPath(t, expected, actualV, key)
				}
				testutil.CompareAndSetByPathNum(t, expected, actual, 20, "catalogStats", "collections")
				testutil.CompareAndSetByPathTime(t, expected, actual, 2*time.Second, "localTime")
				testutil.AssertEqual(t, expected, actual)
			},
		},
	}

	for name, tc := range testCases { //nolint:paralleltest // false positive
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, schema := range []string{"monila"} {
				t.Run(schema, func(t *testing.T) {
					// not parallel because we modify tc

					if tc.reqSetDB {
						tc.req.Set("$db", schema)
					}

					actual := handle(ctx, t, handler, tc.req)
					if tc.compareFunc == nil {
						testutil.AssertEqual(t, tc.resp, actual)
					} else {
						tc.compareFunc(t, tc.req, tc.resp, actual)
					}
				})
			}
		})
	}
}
