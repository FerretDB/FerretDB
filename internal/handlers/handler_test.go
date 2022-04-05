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
		"ValueLtGt": {
			schemas: []string{"monila"},
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
		"String": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", "foo",
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x02, 0x01},
					"name", "string",
					"value", "foo",
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
			)),
		},
		"EmptyString": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", "",
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x02, 0x00, 0x00, 0x02, 0x02},
					"name", "string-empty",
					"value", "",
				)),
			)),
		},
		"Double": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", 42.13,
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x01},
					"name", "double",
					"value", 42.13,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x03, 0x00, 0x00, 0x04, 0x03},
					"name", "array-one",
					"value", must.NotFail(types.NewArray(42.13)),
				)),
			)),
		},
		"DoubleNegativeInfinity": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", math.Inf(-1),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x06, 0x00, 0x00, 0x01, 0x06},
					"name", "double-negative-infinity",
					"value", math.Inf(-1),
				)),
			)),
		},
		"DoublePositiveInfinity": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", math.Inf(+1),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x05, 0x00, 0x00, 0x01, 0x05},
					"name", "double-positive-infinity",
					"value", math.Inf(+1),
				)),
			)),
		},
		"DoubleMax": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", math.MaxFloat64,
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x01, 0x03},
					"name", "double-max",
					"value", math.MaxFloat64,
				)),
			)),
		},
		"DoubleSmallest": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", math.SmallestNonzeroFloat64,
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x04, 0x00, 0x00, 0x01, 0x04},
					"name", "double-smallest",
					"value", math.SmallestNonzeroFloat64,
				)),
			)),
		},
		"Binary": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", types.Binary{Subtype: types.BinaryUser, B: []byte{42, 0, 13}},
				)),
			)),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x05, 0x01, 0x00, 0x00, 0x05, 0x01},
					"name", "binary",
					"value", types.Binary{Subtype: types.BinaryUser, B: []byte{42, 0, 13}},
				)),
			)),
		},
		"EmptyBinary": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", types.Binary{Subtype: 0, B: []byte{}},
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x05, 0x02, 0x00, 0x00, 0x05, 0x02},
					"name", "binary-empty",
					"value", types.Binary{Subtype: 0, B: []byte{}},
				)),
			)),
		},
		"BoolFalse": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", false,
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x08, 0x01},
					"name", "bool-false",
					"value", false,
				)),
			)),
		},
		"BoolTrue": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", true,
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x08, 0x02, 0x00, 0x00, 0x08, 0x02},
					"name", "bool-true",
					"value", true,
				)),
			)),
		},
		"Int32": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", int32(42),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", "array",
					"value", must.NotFail(types.NewArray("array", int32(42))),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x01, 0x00, 0x00, 0x12, 0x01},
					"name", "int64",
					"value", int64(42),
				)),
			)),
		},
		"Int32Zero": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", int32(0),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x01, 0x02},
					"name", "double-zero",
					"value", 0.0,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x10, 0x02},
					"name", "int32-zero",
					"value", int32(0),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x02, 0x00, 0x00, 0x12, 0x02},
					"name", "int64-zero",
					"value", int64(0),
				)),
			)),
		},
		"Int32Max": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", int32(2147483647),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x03, 0x00, 0x00, 0x10, 0x03},
					"name", "int32-max",
					"value", int32(2147483647),
				)),
			)),
		},
		"Int32Min": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", int32(-2147483648),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x04, 0x00, 0x00, 0x10, 0x04},
					"name", "int32-min",
					"value", int32(-2147483648),
				)),
			)),
		},
		"Int64": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", int64(42),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", "array",
					"value", must.NotFail(types.NewArray("array", int32(42))),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x01, 0x00, 0x00, 0x12, 0x01},
					"name", "int64",
					"value", int64(42),
				)),
			)),
		},
		"Int64Zero": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", int64(0),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x01, 0x02},
					"name", "double-zero",
					"value", 0.0,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x10, 0x02},
					"name", "int32-zero",
					"value", int32(0),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x02, 0x00, 0x00, 0x12, 0x02},
					"name", "int64-zero",
					"value", int64(0),
				)),
			)),
		},
		"Int64Max": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", int64(9223372036854775807),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x03, 0x00, 0x00, 0x12, 0x03},
					"name", "int64-max",
					"value", int64(9223372036854775807),
				)),
			)),
		},
		"Int64Min": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", int64(-9223372036854775808),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x04, 0x00, 0x00, 0x12, 0x04},
					"name", "int64-min",
					"value", int64(-9223372036854775808),
				)),
			)),
		},
		"DateTime": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC).Local(),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x01, 0x00, 0x00, 0x09, 0x01},
					"name", "datetime",
					"value", time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC).Local(),
				)),
			)),
		},
		"DateEpoch": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", time.Unix(0, 0),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x02, 0x00, 0x00, 0x09, 0x02},
					"name", "datetime-epoch",
					"value", time.Unix(0, 0),
				)),
			)),
		},
		"DateTimeMinYear": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x03, 0x00, 0x00, 0x09, 0x03},
					"name", "datetime-year-min",
					"value", time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
				)),
			)),
		},
		"DateTimeMaxYear": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC).Local(),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x04, 0x00, 0x00, 0x09, 0x04},
					"name", "datetime-year-max",
					"value", time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC).Local(),
				)),
			)),
		},
		"Timestamp": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", types.Timestamp(180388626445),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x11, 0x01, 0x00, 0x00, 0x11, 0x01},
					"name", "timestamp",
					"value", types.Timestamp(180388626445),
				)),
			)),
		},
		"Nil": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", types.Null,
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x0a, 0x01, 0x00, 0x00, 0x0a, 0x01},
					"name", "null",
					"value", types.Null,
				)),
			)),
		},
		"ValueRegex": {
			schemas: []string{"monila"},
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
		"FindManyRegexWithOption": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", types.Regex{Pattern: "foo", Options: "i"},
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x02, 0x01},
					"name", "string",
					"value", "foo",
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x0b, 0x01, 0x00, 0x00, 0x0b, 0x01},
					"name", "regex",
					"value", types.Regex{Pattern: "foo", Options: "i"},
				)),
			)),
		},
		"FindManyRegexWithoutOption": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", types.Regex{Pattern: "foo"},
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x02, 0x01},
					"name", "string",
					"value", "foo",
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
			)),
		},

		"EqString": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", "foo",
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x02, 0x01},
					"name", "string",
					"value", "foo",
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
			)),
		},
		"EqEmptyString": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", "",
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x02, 0x00, 0x00, 0x02, 0x02},
					"name", "string-empty",
					"value", "",
				)),
			)),
		},
		"EqDouble": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", 42.13,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x01},
					"name", "double",
					"value", 42.13,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x03, 0x00, 0x00, 0x04, 0x03},
					"name", "array-one",
					"value", must.NotFail(types.NewArray(42.13)),
				)),
			)),
		},
		"EqDoubleNegativeInfinity": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", math.Inf(-1),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x06, 0x00, 0x00, 0x01, 0x06},
					"name", "double-negative-infinity",
					"value", math.Inf(-1),
				)),
			)),
		},
		"EqDoublePositiveInfinity": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", math.Inf(+1),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x05, 0x00, 0x00, 0x01, 0x05},
					"name", "double-positive-infinity",
					"value", math.Inf(+1),
				)),
			)),
		},
		"EqDoubleZero": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", 0.0,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x01, 0x02},
					"name", "double-zero",
					"value", 0.0,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x10, 0x02},
					"name", "int32-zero",
					"value", int32(0),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x02, 0x00, 0x00, 0x12, 0x02},
					"name", "int64-zero",
					"value", int64(0),
				)),
			)),
		},
		"EqDoubleMax": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", math.MaxFloat64,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x01, 0x03},
					"name", "double-max",
					"value", math.MaxFloat64,
				)),
			)),
		},
		"EqDoubleSmallest": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", math.SmallestNonzeroFloat64,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x04, 0x00, 0x00, 0x01, 0x04},
					"name", "double-smallest",
					"value", math.SmallestNonzeroFloat64,
				)),
			)),
		},
		"EqBinary": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", types.Binary{Subtype: types.BinaryUser, B: []byte{42, 0, 13}},
					)),
				)),
			)),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x05, 0x01, 0x00, 0x00, 0x05, 0x01},
					"name", "binary",
					"value", types.Binary{Subtype: types.BinaryUser, B: []byte{42, 0, 13}},
				)),
			)),
		},
		"EqEmptyBinary": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", types.Binary{Subtype: 0, B: []byte{}},
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x05, 0x02, 0x00, 0x00, 0x05, 0x02},
					"name", "binary-empty",
					"value", types.Binary{Subtype: 0, B: []byte{}},
				)),
			)),
		},
		"EqBoolFalse": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", false,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x08, 0x01},
					"name", "bool-false",
					"value", false,
				)),
			)),
		},
		"EqBoolTrue": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", true,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x08, 0x02, 0x00, 0x00, 0x08, 0x02},
					"name", "bool-true",
					"value", true,
				)),
			)),
		},
		"EqInt32": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", int32(42),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", "array",
					"value", must.NotFail(types.NewArray("array", int32(42))),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x01, 0x00, 0x00, 0x12, 0x01},
					"name", "int64",
					"value", int64(42),
				)),
			)),
		},
		"EqInt32Zero": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", int32(0),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x01, 0x02},
					"name", "double-zero",
					"value", 0.0,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x10, 0x02},
					"name", "int32-zero",
					"value", int32(0),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x02, 0x00, 0x00, 0x12, 0x02},
					"name", "int64-zero",
					"value", int64(0),
				)),
			)),
		},
		"EqInt32Max": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", int32(2147483647),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x03, 0x00, 0x00, 0x10, 0x03},
					"name", "int32-max",
					"value", int32(2147483647),
				)),
			)),
		},
		"EqInt32Min": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", int32(-2147483648),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x04, 0x00, 0x00, 0x10, 0x04},
					"name", "int32-min",
					"value", int32(-2147483648),
				)),
			)),
		},
		"EqInt64": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", int64(42),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", "array",
					"value", must.NotFail(types.NewArray("array", int32(42))),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x01, 0x00, 0x00, 0x12, 0x01},
					"name", "int64",
					"value", int64(42),
				)),
			)),
		},
		"EqInt64Zero": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", int64(0),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x01, 0x02},
					"name", "double-zero",
					"value", 0.0,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x10, 0x02},
					"name", "int32-zero",
					"value", int32(0),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x02, 0x00, 0x00, 0x12, 0x02},
					"name", "int64-zero",
					"value", int64(0),
				)),
			)),
		},
		"EqInt64Max": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", int64(9223372036854775807),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x03, 0x00, 0x00, 0x12, 0x03},
					"name", "int64-max",
					"value", int64(9223372036854775807),
				)),
			)),
		},
		"EqInt64Min": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", int64(-9223372036854775808),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x04, 0x00, 0x00, 0x12, 0x04},
					"name", "int64-min",
					"value", int64(-9223372036854775808),
				)),
			)),
		},
		"EqDateTime": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC).Local(),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x01, 0x00, 0x00, 0x09, 0x01},
					"name", "datetime",
					"value", time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC).Local(),
				)),
			)),
		},
		"EqDateEpoch": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", time.Unix(0, 0),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x02, 0x00, 0x00, 0x09, 0x02},
					"name", "datetime-epoch",
					"value", time.Unix(0, 0),
				)),
			)),
		},
		"EqDateTimeMinYear": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x03, 0x00, 0x00, 0x09, 0x03},
					"name", "datetime-year-min",
					"value", time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Local(),
				)),
			)),
		},
		"EqDateTimeMaxYear": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC).Local(),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x04, 0x00, 0x00, 0x09, 0x04},
					"name", "datetime-year-max",
					"value", time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC).Local(),
				)),
			)),
		},
		"EqTimestamp": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", types.Timestamp(180388626445),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x11, 0x01, 0x00, 0x00, 0x11, 0x01},
					"name", "timestamp",
					"value", types.Timestamp(180388626445),
				)),
			)),
		},
		"EqNil": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", types.NullType{},
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x0a, 0x01, 0x00, 0x00, 0x0a, 0x01},
					"name", "null",
					"value", types.Null,
				)),
			)),
		},
		"EqRegexWithOption": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", types.Regex{Pattern: "foo", Options: "i"},
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x0b, 0x01, 0x00, 0x00, 0x0b, 0x01},
					"name", "regex",
					"value", types.Regex{Pattern: "foo", Options: "i"},
				)),
			)),
		},
		"EqRegexWithoutOption": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$eq", types.Regex{Pattern: "foo"},
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray()),
		},

		"GtString": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$gt", "boo",
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x02, 0x01},
					"name", "string",
					"value", "foo",
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x03, 0x00, 0x00, 0x02, 0x03},
					"name", "string-shorter",
					"value", "z",
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
			)),
		},
		"GtDouble": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$gt", 42.12,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x01},
					"name", "double",
					"value", 42.13,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x01, 0x03},
					"name", "double-max",
					"value", math.MaxFloat64,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x05, 0x00, 0x00, 0x01, 0x05},
					"name", "double-positive-infinity",
					"value", math.Inf(+1),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x03, 0x00, 0x00, 0x04, 0x03},
					"name", "array-one",
					"value", must.NotFail(types.NewArray(42.13)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x03, 0x00, 0x00, 0x10, 0x03},
					"name", "int32-max",
					"value", int32(2147483647),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x03, 0x00, 0x00, 0x12, 0x03},
					"name", "int64-max",
					"value", int64(9223372036854775807),
				)),
			)),
		},
		"GtInt32": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$gt", int32(41),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x01},
					"name", "double",
					"value", 42.13,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x01, 0x03},
					"name", "double-max",
					"value", math.MaxFloat64,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x05, 0x00, 0x00, 0x01, 0x05},
					"name", "double-positive-infinity",
					"value", math.Inf(+1),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", "array",
					"value", must.NotFail(types.NewArray("array", int32(42))),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x03, 0x00, 0x00, 0x04, 0x03},
					"name", "array-one",
					"value", must.NotFail(types.NewArray(42.13)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x03, 0x00, 0x00, 0x10, 0x03},
					"name", "int32-max",
					"value", int32(2147483647),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x01, 0x00, 0x00, 0x12, 0x01},
					"name", "int64",
					"value", int64(42),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x03, 0x00, 0x00, 0x12, 0x03},
					"name", "int64-max",
					"value", int64(9223372036854775807),
				)),
			)),
		},
		"GtInt64": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$gt", int64(41),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x01},
					"name", "double",
					"value", 42.13,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x01, 0x03},
					"name", "double-max",
					"value", math.MaxFloat64,
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x05, 0x00, 0x00, 0x01, 0x05},
					"name", "double-positive-infinity",
					"value", math.Inf(+1),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", "array",
					"value", must.NotFail(types.NewArray("array", int32(42))),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x03, 0x00, 0x00, 0x04, 0x03},
					"name", "array-one",
					"value", must.NotFail(types.NewArray(42.13)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x04, 0x00, 0x00, 0x04, 0x04},
					"name", "array-three",
					"value", must.NotFail(types.NewArray(int32(42), "foo", types.Null)),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x03, 0x00, 0x00, 0x10, 0x03},
					"name", "int32-max",
					"value", int32(2147483647),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x01, 0x00, 0x00, 0x12, 0x01},
					"name", "int64",
					"value", int64(42),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x03, 0x00, 0x00, 0x12, 0x03},
					"name", "int64-max",
					"value", int64(9223372036854775807),
				)),
			)),
		},
		"GtDateTime": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$gt", time.Date(2021, 11, 1, 10, 18, 42, 121000000, time.UTC).Local(),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x01, 0x00, 0x00, 0x09, 0x01},
					"name", "datetime",
					"value", time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC).Local(),
				)),
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x04, 0x00, 0x00, 0x09, 0x04},
					"name", "datetime-year-max",
					"value", time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC).Local(),
				)),
			)),
		},
		"GtTimestamp": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$gt", types.Timestamp(180388626444),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x11, 0x01, 0x00, 0x00, 0x11, 0x01},
					"name", "timestamp",
					"value", types.Timestamp(180388626445),
				)),
			)),
		},
		"GtNil": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"value", must.NotFail(types.NewDocument(
						"$gt", types.Null,
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray()),
		},

		"BitsAllClear": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAllClear", int32(21),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
			)),
		},
		"BitsAllClearEmptyResult": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAllClear", int32(53),
					)),
				)),
			)),
			resp: new(types.Array),
		},
		"BitsAllClearString": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAllClear", "123",
					)),
				)),
			)),
			err: common.NewErrorMsg(
				common.ErrBadValue,
				`value takes an Array, a number, or a BinData but received: $bitsAllClear: "123"`,
			),
		},
		"BitsAllClearFloat64": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAllClear", 1.2,
					)),
				)),
			)),
			err: common.NewErrorMsg(
				common.ErrFailedToParse,
				`Expected an integer: $bitsAllClear: 1.2`,
			),
		},
		"BitsAllClearNegativeNumber": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAllClear", int32(-1),
					)),
				)),
			)),
			err: common.NewErrorMsg(
				common.ErrFailedToParse,
				`Expected a positive number in: $bitsAllClear: -1`,
			),
		},
		"BitsAllSet": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAllSet", int32(42),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
			)),
		},
		"BitsAnyClear": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAnyClear", int32(1),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
			)),
		},
		"BitsAnyClearEmptyResult": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAnyClear", int32(42),
					)),
				)),
			)),
			resp: new(types.Array),
		},
		"BitsAnyClearBigBinary": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "binary-big",
					"value", must.NotFail(types.NewDocument(
						"$bitsAnyClear", int64(0b1000_0000_0000_0000),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x05, 0x03, 0x00, 0x00, 0x05, 0x03},
					"name", "binary-big",
					"value", types.Binary{B: []byte{0, 0, 128}},
				)),
			)),
		},
		"BitsAnyClearBigBinaryEmptyResult": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "binary-big",
					"value", must.NotFail(types.NewDocument(
						"$bitsAnyClear", int64(0b1000_0000_0000_0000_0000_0000),
					)),
				)),
			)),
			resp: new(types.Array),
		},
		"BitsAnySet": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAnySet", int32(22),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", "int32",
					"value", int32(42),
				)),
			)),
		},
		"BitsAnySetEmptyResult": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "int32",
					"value", must.NotFail(types.NewDocument(
						"$bitsAnySet", int32(4),
					)),
				)),
			)),
			resp: new(types.Array),
		},
		"BitsAnySetBigBinary": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "binary-big",
					"value", must.NotFail(types.NewDocument(
						"$bitsAnySet", int64(0b1000_0000_0000_0000_0000_0000),
					)),
				)),
			)),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x05, 0x03, 0x00, 0x00, 0x05, 0x03},
					"name", "binary-big",
					"value", types.Binary{B: []byte{0, 0, 128}},
				)),
			)),
		},
		"BitsAnySetBigBinaryEmptyResult": {
			schemas: []string{"values"},
			req: must.NotFail(types.NewDocument(
				"find", "values",
				"filter", must.NotFail(types.NewDocument(
					"name", "binary-big",
					"value", must.NotFail(types.NewDocument(
						"$bitsAnySet", int64(0b1000_0000_0000_0000),
					)),
				)),
			)),
			resp: new(types.Array),
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

	t.Run("EqNanDoubleDataType", func(t *testing.T) {
		req := must.NotFail(types.NewDocument(
			"find", "values",
			"filter", must.NotFail(types.NewDocument(
				"value", math.NaN(),
			)),
			"$db", "values",
		))

		response := handle(ctx, t, handler, req)
		firstBatch, err := testutil.GetByPath(t, response, "cursor", "firstBatch").(*types.Array).Get(0)
		require.NoError(t, err)
		responseValue, err := firstBatch.(*types.Document).Get("value")
		require.NoError(t, err)
		if nan, ok := responseValue.(float64); ok {
			assert.True(t, math.IsNaN(nan))
		}
	})
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
				testutil.AssertEqual(t, expected, actual)
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
				testutil.CompareAndSetByPathTime(t, expected, actual, 2*time.Second, "system", "currentTime")
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
					"name", "public",
					"sizeOnDisk", int64(0),
					"empty", true,
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
			"totalSize", int64(25_000_000),
			"totalSizeMb", int64(30),
			"ok", float64(1),
		)

		testutil.CompareAndSetByPathNum(t, expectedList, actualList, 5_000_000, "totalSize")
		testutil.CompareAndSetByPathNum(t, expectedList, actualList, 10, "totalSizeMb")

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

		testutil.AssertEqual(t, expectedList, actualList)

		actualDrop := handle(ctx, t, handler, types.MustNewDocument(
			"dropDatabase", int32(1),
			"$db", db,
		))
		expectedDrop := types.MustNewDocument(
			"dropped", db,
			"ok", float64(1),
		)
		testutil.AssertEqual(t, expectedDrop, actualDrop)

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
		testutil.AssertEqual(t, expectedList, actualList)
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
		testutil.AssertEqual(t, expected, actual)
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
		testutil.AssertEqual(t, expected, actual)

		// TODO test listCollections command once we have better cursor support
		// https://github.com/FerretDB/FerretDB/issues/79

		tables, err := pool.Tables(ctx, db)
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
		testutil.AssertEqual(t, expected, actual)

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
		testutil.AssertEqual(t, expected, actual)
	})

	t.Run("existing", func(t *testing.T) {
		collection := testutil.Table(ctx, t, pool, db)

		actual := handle(ctx, t, handler, types.MustNewDocument(
			"create", collection,
			"$db", db,
		))
		expected := types.MustNewDocument(
			"ok", float64(0),
			"errmsg", "Collection already exists. NS: testcreatelistdropcollection.testcreatelistdropcollection-existing",
			"code", int32(48),
			"codeName", "NamespaceExists",
		)
		testutil.AssertEqual(t, expected, actual)
	})
}
