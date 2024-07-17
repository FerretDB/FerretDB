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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/wire"
)

func TestDriver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	ctx := testutil.Ctx(t)

	c := must.NotFail(Connect(ctx, "mongodb://127.0.0.1:47017/", testutil.SLogger(t)))
	t.Cleanup(func() { require.NoError(t, c.Close()) })

	dbName := testutil.DatabaseName(t)

	doc1 := must.NotFail(bson.NewDocument("_id", int32(0), "w", int32(2), "v", int32(1)))
	doc2 := must.NotFail(bson.NewDocument("_id", int32(1), "v", int32(2)))
	doc3 := must.NotFail(bson.NewDocument("_id", int32(2), "v", int32(3)))

	// TODO https://github.com/FerretDB/FerretDB/issues/4448
	// must.NoError(c.Authenticate(ctx))

	t.Run("Drop", func(t *testing.T) {
		dropCmd := must.NotFail(bson.NewDocument(
			"dropDatabase", int32(1),
			"$db", dbName,
		))

		resHeader, resBody, err := c.Request(ctx, nil, must.NotFail(wire.NewOpMsg(dropCmd)))
		require.NoError(t, err)
		assert.NotZero(t, resHeader.RequestID)

		resMsg, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		ok := resMsg.Get("ok").(float64)
		require.Equal(t, float64(1), ok)
	})

	t.Run("Insert", func(t *testing.T) {
		insertCmd := must.NotFail(bson.NewDocument(
			"insert", "values",
			"documents", must.NotFail(bson.NewArray(doc1, doc2, doc3)),
			"$db", dbName,
		))

		resHeader, resBody, err := c.Request(ctx, nil, must.NotFail(wire.NewOpMsg(insertCmd)))
		require.NoError(t, err)
		assert.NotZero(t, resHeader.RequestID)

		resMsg, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		ok := resMsg.Get("ok").(float64)
		require.Equal(t, float64(1), ok)

		n := resMsg.Get("n").(int32)
		require.Equal(t, int32(3), n)
	})

	var cursorID int64

	expectedBatches := []*types.Array{
		must.NotFail(types.NewArray(must.NotFail(doc1.Convert()))),
		must.NotFail(types.NewArray(must.NotFail(doc2.Convert()))),
		must.NotFail(types.NewArray(must.NotFail(doc3.Convert()))),
	}

	t.Run("Find", func(t *testing.T) {
		findCmd := must.NotFail(bson.NewDocument(
			"find", "values",
			"filter", must.NotFail(bson.NewDocument()),
			"sort", must.NotFail(bson.NewDocument("_id", int32(1))),
			"batchSize", int32(1),
			"$db", dbName,
		))

		resHeader, resBody, err := c.Request(ctx, nil, must.NotFail(wire.NewOpMsg(findCmd)))
		require.NoError(t, err)
		assert.NotZero(t, resHeader.RequestID)

		resMsg, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		cursor, err := resMsg.Get("cursor").(bson.AnyDocument).Decode()
		require.NoError(t, err)

		firstBatch, err := cursor.Get("firstBatch").(bson.AnyArray).Decode()
		require.NoError(t, err)
		cursorID = cursor.Get("id").(int64)

		testutil.AssertEqual(t, expectedBatches[0], must.NotFail(firstBatch.Convert()))
		require.NotZero(t, cursorID)
	})
}

func TestDriverAuthSource(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	ctx := testutil.Ctx(t)
	l := testutil.SLogger(t)

	for name, tc := range map[string]struct {
		uri      string
		expected string
	}{
		"Default": {
			uri:      "mongodb://username:password@127.0.0.1:47017/",
			expected: "admin",
		},
		"DefaultAuthDB": {
			uri:      "mongodb://username:password@127.0.0.1:47017/foo",
			expected: "foo",
		},
		"AuthSource": {
			uri:      "mongodb://username:password@127.0.0.1:47017/?authSource=bar",
			expected: "bar",
		},
		"DefaultAuthDBAuthSource": {
			uri:      "mongodb://username:password@127.0.0.1:47017/foo?authSource=bar",
			expected: "bar",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c, err := Connect(ctx, tc.uri, l)
			require.NoError(t, err)

			t.Cleanup(func() { require.NoError(t, c.Close()) })

			require.Equal(t, tc.expected, c.authDB)
		})
	}
}
