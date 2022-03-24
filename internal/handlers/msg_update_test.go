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

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// TODO Rework to make them closer to other tests.
//nolint:paralleltest // TODO
func TestUpdate(t *testing.T) {
	ctx, h, pool := setup(t, nil)
	schema := testutil.Schema(ctx, t, pool)

	header := &wire.MsgHeader{
		OpCode: wire.OP_MSG,
	}

	for i := 1; i <= 3; i++ {
		var msg wire.OpMsg
		err := msg.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{types.MustNewDocument(
				"insert", "test",
				"documents", types.MustNewArray(
					types.MustNewDocument(
						"_id", types.ObjectID{byte(i)},
						"description", "Test "+strconv.Itoa(i),
					),
				),
				"$db", schema,
			)},
		})
		require.NoError(t, err)

		_, resBody, closeConn := h.Handle(ctx, header, &msg)
		require.False(t, closeConn, "%s", resBody.String())
	}

	var msg wire.OpMsg
	err := msg.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"update", "test",
			"updates", types.MustNewArray(
				types.MustNewDocument(
					"q", types.MustNewDocument(
						"_id", types.ObjectID{byte(1)},
					),
					"u", types.MustNewDocument(
						"$set", types.MustNewDocument(
							"description", "Test 1 updated",
						),
					),
				),
			),
			"$db", schema,
		)},
	})
	require.NoError(t, err)

	_, resBody, closeConn := h.Handle(ctx, header, &msg)
	require.False(t, closeConn, "%s", resBody.String())
}
