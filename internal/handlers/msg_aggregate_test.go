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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/wire"
)

func TestAggregate(t *testing.T) {
	ctx, h, pool := setup(t, nil)
	schema := testutil.Schema(ctx, t, pool)

	header := &wire.MsgHeader{
		OpCode: wire.OP_MSG,
	}

	for i := 1; i <= 3; i++ {
		var msg wire.OpMsg
		err := msg.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{types.MustNewDocument(
				"insert", "aggtest",
				"documents", types.MustNewArray(
					types.MustNewDocument(
						"_id", types.ObjectID{byte(i)},
						"description", "Test "+strconv.Itoa(i),
						"value", strconv.Itoa(i),
					),
				),
				"$db", schema,
			)},
		})
		require.NoError(t, err)

		_, _, closeConn := h.Handle(ctx, header, &msg)
		require.False(t, closeConn)
	}

	var msg wire.OpMsg
	err := msg.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"aggregate", "aggtest",
			"pipeline", types.MustNewArray(
				types.MustNewDocument(
					"$group", types.MustNewDocument(
						"_id", types.Null,
						"count", types.MustNewDocument(
							"$sum", int32(1),
						),
					),
				),
			),
			"$db", schema,
		)},
	})
	require.NoError(t, err)

	_, resBody, closeConn := h.Handle(ctx, header, &msg)
	require.False(t, closeConn)

	actual, err := resBody.(*wire.OpMsg).Document()
	require.NoError(t, err)

	expected := types.MustNewDocument(
		"cursor", types.MustNewDocument(
			"firstBatch", types.MustNewArray(
				types.MustNewDocument(
					"count", int32(3),
				),
			),
			"id", int64(0),
			"ns", schema+".aggtest",
		),
		"ok", float64(1),
	)
	assert.Equal(t, expected, actual)

	err = msg.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"aggregate", "aggtest",
			"pipeline", types.MustNewArray(
				types.MustNewDocument(
					"$group", types.MustNewDocument(
						"_id", types.Null,
						"count", types.MustNewDocument(
							"$sum", int32(1),
						),
					),
				),
				types.MustNewDocument(
					"$match", types.MustNewDocument(
						"value", "3",
					),
				),
			),
			"$db", schema,
		)},
	})
	require.NoError(t, err)

	_, resBody, closeConn = h.Handle(ctx, header, &msg)
	require.False(t, closeConn)

	actual, err = resBody.(*wire.OpMsg).Document()
	require.NoError(t, err)

	expected = types.MustNewDocument(
		"cursor", types.MustNewDocument(
			"firstBatch", types.MustNewArray(types.MustNewDocument(
				"count", int32(1),
			)),
			"id", int64(0),
			"ns", schema+".aggtest",
		),
		"ok", float64(1),
	)
	assert.Equal(t, expected, actual)

	err = msg.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"aggregate", "aggtest",
			"pipeline", types.MustNewArray(
				types.MustNewDocument(
					"$group", types.MustNewDocument(
						"_id", "$value",
						"count", types.MustNewDocument(
							"$sum", int32(1),
						),
					),
				),
				types.MustNewDocument(
					"$sort", types.MustNewDocument(
						"_id", int32(1),
					),
				),
			),
			"$db", schema,
		)},
	})
	require.NoError(t, err)

	_, resBody, closeConn = h.Handle(ctx, header, &msg)
	require.False(t, closeConn)

	actual, err = resBody.(*wire.OpMsg).Document()
	require.NoError(t, err)

	expected = types.MustNewDocument(
		"cursor", types.MustNewDocument(
			"firstBatch", types.MustNewArray(types.MustNewDocument(
				"_id", "1",
				"count", int32(1),
			), types.MustNewDocument(
				"_id", "2",
				"count", int32(1),
			), types.MustNewDocument(
				"_id", "3",
				"count", int32(1),
			)),
			"id", int64(0),
			"ns", schema+".aggtest",
		),
		"ok", float64(1),
	)
	assert.Equal(t, expected, actual)
}
