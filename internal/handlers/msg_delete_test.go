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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// TODO Rework to make them closer to other tests.
//nolint:paralleltest // TODO
func TestDelete(t *testing.T) {
	ctx, h, pool := setup(t, nil)
	schema := testutil.Schema(ctx, t, pool)

	header := wire.MsgHeader{
		OpCode: wire.OP_MSG,
	}

	t.Run(schema, func(t *testing.T) {
		for i := 1; i <= 5; i++ {
			var msg wire.OpMsg
			err := msg.SetSections(wire.OpMsgSection{
				Documents: []*types.Document{types.MustNewDocument(
					"insert", "test",
					"documents", types.MustNewArray(
						types.MustNewDocument(
							"_id", types.ObjectID{byte(10 + i)},
							"colour", "red",
						),
					),
					"$db", schema,
				)},
			})
			require.NoError(t, err)

			_, _, closeConn := h.Handle(ctx, &header, &msg)
			require.False(t, closeConn)
		}

		for i := 1; i <= 5; i++ {
			var msg wire.OpMsg
			err := msg.SetSections(wire.OpMsgSection{
				Documents: []*types.Document{types.MustNewDocument(
					"insert", "test",
					"documents", types.MustNewArray(
						types.MustNewDocument(
							"_id", types.ObjectID{byte(i)},
							"animal", "cat",
						),
					),
					"$db", schema,
				)},
			})
			require.NoError(t, err)

			_, resBody, closeConn := h.Handle(ctx, &header, &msg)
			require.False(t, closeConn, "%s", resBody.String())
		}

		type testCase struct {
			req  *types.Document
			resp *types.Document
		}

		testCases := map[string]testCase{
			"NothingToDelete": {
				req: types.MustNewDocument(
					"delete", "test",
					"deletes", types.MustNewArray(
						types.MustNewDocument(
							"q", types.MustNewDocument(
								"colour", "blue",
							),
							"limit", int32(0),
						),
					),
				),
				resp: types.MustNewDocument(
					"n", int32(0),
					"ok", float64(1),
				),
			},
			"DeleteLimit1": {
				req: types.MustNewDocument(
					"delete", "test",
					"deletes", types.MustNewArray(
						types.MustNewDocument(
							"q", types.MustNewDocument(
								"colour", "red",
							),
							"limit", int32(1),
						),
					),
				),
				resp: types.MustNewDocument(
					"n", int32(1),
					"ok", float64(1),
				),
			},
			"DeleteLimit0": {
				req: types.MustNewDocument(
					"delete", "test",
					"deletes", types.MustNewArray(
						types.MustNewDocument(
							"q", types.MustNewDocument(
								"animal", "cat",
							),
							"limit", int32(0),
						),
					),
				),
				resp: types.MustNewDocument(
					"n", int32(5),
					"ok", float64(1),
				),
			},
		}

		for name, tc := range testCases {
			tc := tc
			t.Run(name, func(t *testing.T) {
				tc.req.Set("$db", schema)

				var reqMsg wire.OpMsg
				err := reqMsg.SetSections(wire.OpMsgSection{
					Documents: []*types.Document{tc.req},
				})
				require.NoError(t, err)

				_, resBody, closeConn := h.Handle(ctx, &header, &reqMsg)
				require.False(t, closeConn, "%s", resBody.String())

				actual, err := resBody.(*wire.OpMsg).Document()
				require.NoError(t, err)

				expected := tc.resp
				assert.Equal(t, expected, actual)
			})
		}
	})
}
