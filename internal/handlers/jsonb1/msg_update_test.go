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

package jsonb1

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	ctx, h, schema := setup(t)

	header := &wire.MsgHeader{
		OpCode: wire.OP_MSG,
	}

	for i := 1; i <= 3; i++ {
		msg := &wire.OpMsg{
			Documents: []types.Document{types.MakeDocument(
				"insert", "test",
				"documents", types.Array{
					types.MakeDocument(
						"_id", types.ObjectID{byte(i)},
						"description", "Test "+strconv.Itoa(i),
					),
				},
				"$db", schema,
			)},
		}

		_, _, err := h.Handle(ctx, header, msg)
		require.NoError(t, err)
	}

	msg := &wire.OpMsg{
		Documents: []types.Document{types.MakeDocument(
			"update", "test",
			"updates", types.Array{
				types.MakeDocument(
					"q", types.MakeDocument(
						"_id", types.ObjectID{byte(1)},
					),
					"u", types.MakeDocument(
						"$set", types.MakeDocument(
							"description", "Test 1 updated",
						),
					),
				),
			},
			"$db", schema,
		)},
	}

	_, _, err := h.Handle(ctx, header, msg)
	require.NoError(t, err)
}
