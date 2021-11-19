// Copyright 2021 Baltoro OÜ.
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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/MangoDB-io/MangoDB/internal/handlers/jsonb1"
	"github.com/MangoDB-io/MangoDB/internal/handlers/shared"
	"github.com/MangoDB-io/MangoDB/internal/handlers/sql"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/testutil"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

func TestFind(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t)
	l := zaptest.NewLogger(t)
	shared := shared.NewHandler(pool, "127.0.0.1:12345")
	sql := sql.NewStorage(pool, l.Sugar())
	jsonb1 := jsonb1.NewStorage(pool, l)
	handler := New(pool, l, shared, sql, jsonb1)

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

					_, respMsg, err := handler.Handle(ctx, &reqHeader, &reqMsg)
					require.NoError(t, err)

					actual, err := respMsg.(*wire.OpMsg).Document()
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
