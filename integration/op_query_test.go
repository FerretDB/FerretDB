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

package integration

import (
	"testing"
	"time"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestOpQuery(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		WireConn: setup.WireConnAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	t.Run("CollectionNameWithout.$cmd", func(t *testing.T) {
		q := must.NotFail(wire.NewOpQuery(must.NotFail(wirebson.NewDocument("unknown", int32(1)))))
		q.FullCollectionName = "invalid"
		q.NumberToReturn = -1

		_, resBody, err := conn.Request(ctx, q)
		require.NoError(t, err)

		resMsg, err := resBody.(*wire.OpReply).RawDocument().Decode()
		require.NoError(t, err)

		ok := resMsg.Get("ok")
		assert.Equal(t, float64(0), ok)

		code := resMsg.Get("code")
		assert.Equal(t, int32(5739101), code)

		expectedMsg := "OP_QUERY is no longer supported. The client driver may require an upgrade."
		resErr := resMsg.Get("$err")
		assert.Contains(t, resErr, expectedMsg)
	})

	t.Run("UnknownOpQuery", func(t *testing.T) {
		q := must.NotFail(wire.NewOpQuery(must.NotFail(wirebson.NewDocument("unknown", int32(1)))))
		q.FullCollectionName = "admin.$cmd"
		q.NumberToReturn = -1

		_, resBody, err := conn.Request(ctx, q)
		require.NoError(t, err)

		resMsg, err := resBody.(*wire.OpReply).RawDocument().Decode()
		require.NoError(t, err)

		ok := resMsg.Get("ok")
		assert.Equal(t, float64(0), ok)

		code := resMsg.Get("code")
		assert.Equal(t, int32(352), code)

		expectedMsg := "Unsupported OP_QUERY command: unknown. The client driver may require an upgrade."
		errMsg := resMsg.Get("errmsg")
		assert.Contains(t, errMsg, expectedMsg)
	})

	t.Run("BadNumberToReturn", func(t *testing.T) {
		q := must.NotFail(wire.NewOpQuery(must.NotFail(wirebson.NewDocument("ismaster", int32(1)))))
		q.FullCollectionName = "admin.$cmd"
		q.NumberToReturn = 0

		_, resBody, err := conn.Request(ctx, q)
		require.NoError(t, err)

		res, err := resBody.(*wire.OpReply).RawDocument().Decode()
		require.NoError(t, err)

		fixCluster(t, res)

		expected := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "Bad numberToReturn (0) for $cmd type ns - can only be 1 or -1",
			"code", int32(16979),
			"codeName", "Location16979",
		))

		testutil.AssertEqual(t, expected, res)
	})
}

func TestOpQueryIsMaster(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		WireConn: setup.WireConnAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	for name, tc := range map[string]struct {
		command string
	}{
		"IsMaster": {command: "isMaster"},
		"Ismaster": {command: "ismaster"},
	} {
		t.Run(name, func(tt *testing.T) {
			t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/955")

			q := must.NotFail(wire.NewOpQuery(must.NotFail(wirebson.NewDocument(tc.command, int32(1)))))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			resHeader, resBody, err := conn.Request(ctx, q)
			require.NoError(t, err)
			assert.NotZero(t, resHeader.RequestID)

			res, err := resBody.(*wire.OpReply).RawDocument().Decode()
			require.NoError(t, err)

			connectionID := res.Get("connectionId")
			assert.IsType(t, int32(0), connectionID)
			res.Remove("connectionId")

			localTime := res.Get("localTime")
			assert.IsType(t, time.Time{}, localTime)
			res.Remove("localTime")

			// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/566
			res.Remove("hosts")
			res.Remove("setName")
			res.Remove("topologyVersion")
			res.Remove("setVersion")
			res.Remove("secondary")
			res.Remove("primary")
			res.Remove("me")
			res.Remove("electionId")
			res.Remove("lastWrite")

			fixCluster(t, res)

			expectedComparable := must.NotFail(wirebson.NewDocument(
				"ismaster", true,
				"maxBsonObjectSize", int32(16777216),
				"maxMessageSizeBytes", int32(48000000),
				"maxWriteBatchSize", int32(100000),
				"logicalSessionTimeoutMinutes", int32(30),
				"minWireVersion", int32(0),
				"maxWireVersion", int32(21),
				"readOnly", false,
				"ok", float64(1),
			))
			testutil.AssertEqual(t, expectedComparable, res)
		})
	}
}

func TestOpQueryIsMasterHelloOk(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		WireConn: setup.WireConnAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	for name, tc := range map[string]struct {
		command string
	}{
		"IsMaster": {command: "isMaster"},
		"Ismaster": {command: "ismaster"},
	} {
		t.Run(name, func(tt *testing.T) {
			t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/955")

			q := must.NotFail(wire.NewOpQuery(must.NotFail(wirebson.NewDocument(
				tc.command, int32(1),
				"helloOk", true,
			))))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			resHeader, resBody, err := conn.Request(ctx, q)
			require.NoError(t, err)
			assert.NotZero(t, resHeader.RequestID)

			res, err := resBody.(*wire.OpReply).RawDocument().Decode()
			require.NoError(t, err)

			connectionID := res.Get("connectionId")
			assert.IsType(t, int32(0), connectionID)
			res.Remove("connectionId")

			localTime := res.Get("localTime")
			assert.IsType(t, time.Time{}, localTime)
			res.Remove("localTime")

			// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/566
			res.Remove("hosts")
			res.Remove("setName")
			res.Remove("topologyVersion")
			res.Remove("setVersion")
			res.Remove("secondary")
			res.Remove("primary")
			res.Remove("me")
			res.Remove("electionId")
			res.Remove("lastWrite")

			fixCluster(t, res)

			expectedComparable := must.NotFail(wirebson.NewDocument(
				"helloOk", true,
				"ismaster", true,
				"maxBsonObjectSize", int32(16777216),
				"maxMessageSizeBytes", int32(48000000),
				"maxWriteBatchSize", int32(100000),
				"logicalSessionTimeoutMinutes", int32(30),
				"minWireVersion", int32(0),
				"maxWireVersion", int32(21),
				"readOnly", false,
				"ok", float64(1),
			))
			testutil.AssertEqual(t, expectedComparable, res)
		})
	}
}

func TestOpQueryHello(tt *testing.T) {
	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/955")

	tt.Parallel()

	s := setup.SetupWithOpts(tt, &setup.SetupOpts{
		WireConn: setup.WireConnAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	q := must.NotFail(wire.NewOpQuery(must.NotFail(wirebson.NewDocument(
		"hello", int32(1),
	))))
	q.FullCollectionName = "admin.$cmd"
	q.NumberToReturn = -1

	_, resBody, err := conn.Request(ctx, q)
	require.NoError(t, err)

	res, err := resBody.(*wire.OpReply).RawDocument().Decode()
	require.NoError(t, err)

	connectionID := res.Get("connectionId")
	assert.IsType(t, int32(0), connectionID)
	res.Remove("connectionId")

	localTime := res.Get("localTime")
	assert.IsType(t, time.Time{}, localTime)
	res.Remove("localTime")

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/566
	res.Remove("topologyVersion")
	res.Remove("hosts")
	res.Remove("setName")
	res.Remove("setVersion")
	res.Remove("secondary")
	res.Remove("primary")
	res.Remove("me")
	res.Remove("electionId")
	res.Remove("lastWrite")

	fixCluster(t, res)

	expectedComparable := must.NotFail(wirebson.NewDocument(
		"isWritablePrimary", true,
		"maxBsonObjectSize", int32(16777216),
		"maxMessageSizeBytes", int32(48000000),
		"maxWriteBatchSize", int32(100000),
		"logicalSessionTimeoutMinutes", int32(30),
		"minWireVersion", int32(0),
		"maxWireVersion", int32(21),
		"readOnly", false,
		"ok", float64(1),
	))
	testutil.AssertEqual(t, expectedComparable, res)
}
