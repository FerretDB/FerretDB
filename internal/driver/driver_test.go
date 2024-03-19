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

package driver

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/wire"
)

func TestDriver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	ctx := testutil.Ctx(t)

	c, err := Connect(ctx, "mongodb://127.0.0.1:47017/", testutil.SLogger(t))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, c.Close()) })

	t.Run("Insert", func(t *testing.T) {
		doc := must.NotFail(bson.NewDocument(
			"insert", "values",
			"documents", must.NotFail(bson.NewArray(
				must.NotFail(bson.NewDocument("v", int32(1), "_id", bson.ObjectID([]byte("65f83bddef2048e47170b641")))),
				must.NotFail(bson.NewDocument("v", int32(2), "_id", bson.ObjectID([]byte("65f83bddef2048e47170b642")))),
			)),
			"ordered", true,
			"lsid", int32(0),
			"txnNumber", int64(1),
			"$db", "test",
		))

		section, err := must.NotFail(bson.NewDocument(
			"Kind", int32(0),
			"Document", doc,
		)).Encode()

		require.NoError(t, err)

		body, err := wire.NewOpMsg(section)
		require.NoError(t, err)

		msgBin, err := body.MarshalBinary()
		require.NoError(t, err)

		// TODO verify header
		header := wire.MsgHeader{
			MessageLength: int32(len(msgBin) + wire.MsgHeaderLen),
			RequestID:     13,
			OpCode:        wire.OpCodeMsg,
		}

		//resHeader, resBody, err := c.Request(ctx, &header, body)
		_, _, err = c.Request(ctx, &header, body)
		require.NoError(t, err)

		//	assert.Equal(t, wire.MsgHeader{
		//		RequestID:     368,
		//		ResponseTo:    13,
		//		OpCode:        2013,
		//		MessageLength: 252,
		//	}, *resHeader)
		//	assert.Equal(t, wire.OpMsg{}, resBody)
	})

	t.Run("Find", func(t *testing.T) {
		doc := must.NotFail(bson.NewDocument(
			"find", "values",
			"filter", must.NotFail(bson.NewDocument()),
			"lsid", int32(0),
			"$db", "test",
		))

		section, err := must.NotFail(bson.NewDocument(
			"Kind", int32(0),
			"Document", doc,
		)).Encode()
		require.NoError(t, err)

		body, err := wire.NewOpMsg(section)
		require.NoError(t, err)

		msgBin, err := body.MarshalBinary()
		require.NoError(t, err)

		// TODO verify header
		header := wire.MsgHeader{
			MessageLength: int32(len(msgBin) + wire.MsgHeaderLen),
			RequestID:     13,
			OpCode:        wire.OpCodeMsg,
		}

		_, resBody, err := c.Request(ctx, &header, body)
		require.NoError(t, err)

	})
}
