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

func removeIDs(docs types.Array) {
	for i, doc := range docs {
		d := doc.(types.Document)
		d.Remove("_id")
		docs[i] = d
	}
}

func TestQuery(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t)
	l := zaptest.NewLogger(t)
	shared := shared.NewHandler(pool, "127.0.0.1:12345")
	sql := sql.NewStorage(pool, l.Sugar())
	jsonb1 := jsonb1.NewStorage(pool, l)
	handler := New(pool, l, shared, sql, jsonb1)

	lastUpdate := time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local()

	for _, schema := range []string{"monila", "pagila"} {
		t.Run(schema, func(t *testing.T) {
			header := &wire.MsgHeader{
				RequestID: 1,
				OpCode:    wire.OP_MSG,
			}

			var msg wire.OpMsg
			err := msg.SetSections(wire.OpMsgSection{
				Documents: []types.Document{types.MustMakeDocument(
					"find", "actor",
					"filter", types.MustMakeDocument(
						"last_name", "HOFFMAN",
					),
					"$db", schema,
				)},
			})
			require.NoError(t, err)
			_, res, err := handler.Handle(ctx, header, &msg)
			require.NoError(t, err)

			expectedDocs := types.Array{
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x1c, 0x00, 0x00, 0x00, 0x1c},
					"actor_id", int32(28),
					"first_name", "WOODY",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0x4f, 0x00, 0x00, 0x00, 0x4f},
					"actor_id", int32(79),
					"first_name", "MAE",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
				types.MustMakeDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x00, 0xa9, 0x00, 0x00, 0x00, 0xa9},
					"actor_id", int32(169),
					"first_name", "KENNETH",
					"last_name", "HOFFMAN",
					"last_update", lastUpdate,
				),
			}
			if schema == "pagila" {
				removeIDs(expectedDocs)
			}

			actual, err := res.(*wire.OpMsg).Document()
			require.NoError(t, err)
			expected := types.MustMakeDocument(
				"cursor", types.MustMakeDocument(
					"firstBatch", expectedDocs,
					"id", int64(0),
					"ns", schema+".actor",
				),
				"ok", float64(1),
			)

			assert.Equal(t, expected, actual)
		})
	}
}
