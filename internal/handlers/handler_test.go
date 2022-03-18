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
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/jsonb1"
	"github.com/FerretDB/FerretDB/internal/pg"
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
func setup(t testing.TB, opts *setupOpts) (context.Context, *Handler, *pg.Pool) {
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
	jsonb1 := jsonb1.NewStorage(pool, l)
	handler := New(&NewOpts{
		PgPool:        pool,
		Logger:        l,
		PeerAddr:      "127.0.0.1:12345",
		JSONB1Storage: jsonb1,
		Metrics:       NewMetrics(),
	})

	return ctx, handler, pool
}

func handle(ctx context.Context, t *testing.T, handler *Handler, req *types.Document) *types.Document {
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

	addToSeedCorpus(t, &reqHeader, &reqMsg)

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
		"ValueLtGt": {
			schemas: []string{"monila", "pagila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"last_name", "HOFFMAN",
					"actor_id", types.MustNewDocument(
						"$gt", int32(50),
						"$lt", int32(100),
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
		"InLteGte": {
			schemas: []string{"monila", "pagila"},
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
			schemas: []string{"monila", "pagila"},
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
			schemas: []string{"monila", "pagila"},
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
			schemas: []string{"monila", "pagila"},
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
		"AndOr": {
			schemas: []string{"monila", "pagila"},
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
				"limit", int32(1),
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
			schemas: []string{"monila", "pagila"},
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
		"ValueRegex": {
			schemas: []string{"monila", "pagila"},
			req: types.MustNewDocument(
				"find", "actor",
				"filter", types.MustNewDocument(
					"last_name", types.Regex{Pattern: "hoffman", Options: "i"},
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
		"Regex": {
			schemas: []string{"monila", "pagila"},
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
		"RegexOptions": {
			schemas: []string{"monila", "pagila"},
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
		"RegexStringOptions": {
			schemas: []string{"monila", "pagila"},
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
		"SizeInt32": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", int32(2),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", "array",
					"value", must.NotFail(types.NewArray("array", int32(42))),
				)),
			)),
		},
		"SizeInt64": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", int64(2),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", "array",
					"value", must.NotFail(types.NewArray("array", int32(42))),
				)),
			)),
		},
		"SizeDouble": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", 2.0,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", "array",
					"value", must.NotFail(types.NewArray("array", int32(42))),
				)),
			)),
		},
		"SizeNotFound": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", int32(4),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray()),
		},
		"SizeZero": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", 0.0,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x02, 0x00, 0x00, 0x04, 0x02},
					"name", "array-empty",
					"value", must.NotFail(types.NewArray()),
				)),
			)),
		},
		"SizeInvalidType": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", must.NotFail(types.NewDocument("$gt", int32(1))),
					)),
				)),
			)),
			err: common.NewErrorMsg(common.ErrBadValue, "$size needs a number"),
		},
		"SizeNonWhole": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", 2.1,
					)),
				)),
			)),
			err: common.NewErrorMsg(common.ErrBadValue, "$size must be a whole number"),
		},
		"SizeNaN": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", math.NaN(),
					)),
				)),
			)),
			err: common.NewErrorMsg(common.ErrBadValue, "$size must be a whole number"),
		},
		"SizeInfinity": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", math.Inf(1),
					)),
				)),
			)),
			err: common.NewErrorMsg(common.ErrBadValue, "$size must be a whole number"),
		},
		"SizeNegative": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$size", int32(-1),
					)),
				)),
			)),
			err: common.NewErrorMsg(common.ErrBadValue, "$size may not be negative"),
		},
		"SizeInvalid": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"$size", int32(2),
				)),
			)),
			err: common.NewErrorMsg(
				common.ErrBadValue,
				`unknown top level operator: $size. `+
					`If you have a field name that starts with a '$' symbol, consider using $getField or $setField.`,
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

					// remove _id fields that are not present in pagila
					if schema == "pagila" {
						for i := 0; i < tc.resp.Len(); i++ {
							doc, err := tc.resp.Get(i)
							require.NoError(t, err)
							d := doc.(*types.Document)
							d.Remove("_id")
							err = tc.resp.Set(i, d)
							require.NoError(t, err)
						}
					}

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
					assert.Equal(t, expected, actual)
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

	hostname, err := os.Hostname()
	require.NoError(t, err)

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
				"size", int32(1_236_992),
				"storageSize", int32(1_204_224),
				"totalIndexSize", int32(0),
				"totalSize", int32(1_236_992),
				"scaleFactor", int32(1),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, req, expected, actual *types.Document) {
				db, err := req.Get("$db")
				require.NoError(t, err)
				if db.(string) == "monila" {
					testutil.CompareAndSetByPathNum(t, expected, actual, 30_000, "size")
					testutil.CompareAndSetByPathNum(t, expected, actual, 30_000, "storageSize")
					testutil.CompareAndSetByPathNum(t, expected, actual, 30_000, "totalSize")
					assert.Equal(t, expected, actual)
				}
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
				"size", int32(114_688),
				"numObjects", int32(200),
				"millis", int32(20),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, req, expected, actual *types.Document) {
				db, err := req.Get("$db")
				require.NoError(t, err)
				if db.(string) == "monila" {
					testutil.CompareAndSetByPathNum(t, expected, actual, 50, "millis")
					testutil.CompareAndSetByPathNum(t, expected, actual, 30_000, "size")
					assert.Equal(t, expected, actual)
				}
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
			compareFunc: func(t testing.TB, req, expected, actual *types.Document) {
				db, err := req.Get("$db")
				require.NoError(t, err)
				if db.(string) == "monila" {
					testutil.CompareAndSetByPathNum(t, expected, actual, 30, "millis")
					assert.Equal(t, expected, actual)
				}
			},
		},

		"DBStats": {
			req: types.MustNewDocument(
				"dbstats", int32(1),
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"db", "monila",
				"collections", int32(14),
				"views", int32(0),
				"objects", int32(30224),
				"avgObjSize", 437.7342509264161,
				"dataSize", 1.323008e+07,
				"indexes", int32(0),
				"indexSize", float64(0),
				"totalSize", 1.3615104e+07,
				"scaleFactor", float64(1),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, req, expected, actual *types.Document) {
				db, err := req.Get("$db")
				require.NoError(t, err)
				if db.(string) == "monila" {
					testutil.CompareAndSetByPathNum(t, expected, actual, 20, "avgObjSize")
					testutil.CompareAndSetByPathNum(t, expected, actual, 400_000, "dataSize")
					testutil.CompareAndSetByPathNum(t, expected, actual, 400_000, "totalSize")
					assert.Equal(t, expected, actual)
				}
			},
		},
		"DBStatsWithScale": {
			req: types.MustNewDocument(
				"dbstats", int32(1),
				"scale", float64(1_000),
			),
			reqSetDB: true,
			resp: types.MustNewDocument(
				"db", "monila",
				"collections", int32(14),
				"views", int32(0),
				"objects", int32(30224),
				"avgObjSize", 437.7342509264161,
				"dataSize", 13_230.08,
				"indexes", int32(0),
				"indexSize", float64(0),
				"totalSize", 13_615.104,
				"scaleFactor", float64(1_000),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, req, expected, actual *types.Document) {
				db, err := req.Get("$db")
				require.NoError(t, err)
				if db.(string) == "monila" {
					testutil.CompareAndSetByPathNum(t, expected, actual, 20, "avgObjSize")
					testutil.CompareAndSetByPathNum(t, expected, actual, 400, "dataSize")
					testutil.CompareAndSetByPathNum(t, expected, actual, 400, "totalSize")
					assert.Equal(t, expected, actual)
				}
			},
		},

		"FindProjectionActorsFirstAndLastName": {
			req: types.MustNewDocument(
				"find", "actor",
				"projection", types.MustNewDocument(
					"first_name", int32(1),
					"last_name", int32(1),
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
							"first_name", "WOODY",
							"last_name", "HOFFMAN",
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
				assert.Equal(t, expected, actual)
			},
		},

		"GetLog": {
			req: types.MustNewDocument(
				"getLog", "startupWarnings",
			),
			resp: types.MustNewDocument(
				"totalLinesWritten", int32(2),
				// will be replaced with the real value during the test
				"log", types.MakeArray(2),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _ *types.Document, actual, expected *types.Document) {
				// Just testing "ok" response, not the body of the response
				actualV := testutil.GetByPath(t, actual, "log")
				testutil.SetByPath(t, expected, actualV, "log")
				assert.Equal(t, expected, actual)
			},
		},

		"GetParameter": {
			req: types.MustNewDocument(
				"getParameter", int32(1),
			),
			resp: types.MustNewDocument(
				"version", "5.0.42",
				"ok", float64(1),
			),
		},

		"ListCommands": {
			req: types.MustNewDocument(
				"listCommands", int32(1),
			),
			resp: types.MustNewDocument(
				"commands", types.MustNewDocument(),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _ *types.Document, actual, expected *types.Document) {
				actualV := testutil.GetByPath(t, actual, "commands")
				testutil.SetByPath(t, expected, actualV, "commands")
				assert.Equal(t, expected, actual)
			},
		},

		"IsMaster": {
			req: types.MustNewDocument(
				"isMaster", int32(1),
			),
			resp: types.MustNewDocument(
				"helloOk", true,
				"ismaster", true,
				"maxBsonObjectSize", int32(16777216),
				"maxMessageSizeBytes", int32(wire.MaxMsgLen),
				"maxWriteBatchSize", int32(100000),
				"localTime", time.Now(),
				"minWireVersion", int32(13),
				"maxWireVersion", int32(13),
				"readOnly", false,
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _ *types.Document, actual, expected *types.Document) {
				testutil.CompareAndSetByPathTime(t, expected, actual, time.Second, "localTime")
				assert.Equal(t, expected, actual)
			},
		},
		"Hello": {
			req: types.MustNewDocument(
				"hello", int32(1),
			),
			resp: types.MustNewDocument(
				"helloOk", true,
				"ismaster", true,
				"maxBsonObjectSize", int32(16777216),
				"maxMessageSizeBytes", int32(wire.MaxMsgLen),
				"maxWriteBatchSize", int32(100000),
				"localTime", time.Now(),
				"minWireVersion", int32(13),
				"maxWireVersion", int32(13),
				"readOnly", false,
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _ *types.Document, actual, expected *types.Document) {
				testutil.CompareAndSetByPathTime(t, expected, actual, time.Second, "localTime")
				assert.Equal(t, expected, actual)
			},
		},

		"HostInfo": {
			req: types.MustNewDocument(
				"hostInfo", int32(1),
			),
			resp: types.MustNewDocument(
				"system", types.MustNewDocument(
					"currentTime", time.Now(),
					"hostname", hostname,
					"cpuAddrSize", int32(strconv.IntSize),
					"numCores", int32(runtime.NumCPU()),
					"cpuArch", runtime.GOARCH,
					"numaEnabled", false,
				),
				"os", types.MustNewDocument(
					"type", strings.Title(runtime.GOOS),
				),
				"ok", float64(1),
			),
			compareFunc: func(t testing.TB, _ *types.Document, actual, expected *types.Document) {
				testutil.CompareAndSetByPathTime(t, expected, actual, time.Second, "system", "currentTime")
				assert.Equal(t, expected, actual)
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
				testutil.CompareAndSetByPathTime(t, expected, actual, time.Second, "localTime")
				assert.Equal(t, expected, actual)
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

		actualList := handle(ctx, t, handler, types.MustNewDocument(
			"listDatabases", int32(1),
		))
		expectedList := types.MustNewDocument(
			"databases", types.MustNewArray(
				types.MustNewDocument(
					"name", "monila",
					"sizeOnDisk", int64(13_631_488),
					"empty", false,
				),
				types.MustNewDocument(
					"name", "pagila",
					"sizeOnDisk", int64(7_127_040),
					"empty", false,
				),
				types.MustNewDocument(
					"name", "test",
					"sizeOnDisk", int64(0),
					"empty", true,
				),
				types.MustNewDocument(
					"name", db,
					"sizeOnDisk", int64(0),
					"empty", true,
				),
				types.MustNewDocument(
					"name", "values",
					"sizeOnDisk", int64(16_384),
					"empty", false,
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
		require.Equal(t, expectedDBs.Len(), actualDBs.Len(), "actual:\n%s", testutil.Dump(t, actualList))
		for i := 0; i < actualDBs.Len(); i++ {
			actualDB, err := actualDBs.Get(i)
			require.NoError(t, err)
			expectedDB, err := expectedDBs.Get(i)
			require.NoError(t, err)
			testutil.CompareAndSetByPathNum(t, expectedDB.(*types.Document), actualDB.(*types.Document), 500_000, "sizeOnDisk")
		}

		assert.Equal(t, expectedList, actualList)

		actualDrop := handle(ctx, t, handler, types.MustNewDocument(
			"dropDatabase", int32(1),
			"$db", db,
		))
		expectedDrop := types.MustNewDocument(
			"dropped", db,
			"ok", float64(1),
		)
		assert.Equal(t, expectedDrop, actualDrop)

		// cut dropped db from the expected list
		databases := testutil.GetByPath(t, expectedList, "databases").(*types.Array)
		newDatabases, err := databases.Subslice(0, databases.Len()-2)
		require.NoError(t, err)
		valuesDB, err := databases.Get(databases.Len() - 1)
		require.NoError(t, err)
		err = newDatabases.Append(valuesDB)
		require.NoError(t, err)
		testutil.SetByPath(t, expectedList, newDatabases, "databases")

		actualList = handle(ctx, t, handler, types.MustNewDocument(
			"listDatabases", int32(1),
		))
		assert.Equal(t, expectedList, actualList)
	})

	t.Run("nonexisting", func(t *testing.T) {
		actual := handle(ctx, t, handler, types.MustNewDocument(
			"dropDatabase", int32(1),
			"$db", "nonexisting",
		))
		expected := types.MustNewDocument(
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

		actual := handle(ctx, t, handler, types.MustNewDocument(
			"create", collection,
			"$db", db,
		))
		expected := types.MustNewDocument(
			"ok", float64(1),
		)
		assert.Equal(t, expected, actual)

		// TODO test listCollections command once we have better cursor support
		// https://github.com/FerretDB/FerretDB/issues/79

		tables, _, err := pool.Tables(ctx, db)
		require.NoError(t, err)
		assert.Equal(t, []string{collection}, tables)

		actual = handle(ctx, t, handler, types.MustNewDocument(
			"drop", collection,
			"$db", db,
		))
		expected = types.MustNewDocument(
			"nIndexesWas", int32(1),
			"ns", db+"."+collection,
			"ok", float64(1),
		)
		assert.Equal(t, expected, actual)

		actual = handle(ctx, t, handler, types.MustNewDocument(
			"drop", collection,
			"$db", db,
		))
		expected = types.MustNewDocument(
			"ok", float64(0),
			"errmsg", "ns not found",
			"code", int32(26),
			"codeName", "NamespaceNotFound",
		)
		assert.Equal(t, expected, actual)
	})

	t.Run("existing", func(t *testing.T) {
		collection := testutil.CreateTable(ctx, t, pool, db)

		actual := handle(ctx, t, handler, types.MustNewDocument(
			"create", collection,
			"$db", db,
		))
		expected := types.MustNewDocument(
			"ok", float64(0),
			"errmsg", "Collection already exists. NS: testcreatelistdropcollection.testcreatelistdropcollection_existing",
			"code", int32(48),
			"codeName", "NamespaceExists",
		)
		assert.Equal(t, expected, actual)
	})
}
