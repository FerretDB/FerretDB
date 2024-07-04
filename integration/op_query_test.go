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
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestOpQuery(t *testing.T) {
	t.Parallel()

	ctx, conn := setup.SetupDriver(t)

	t.Run("CollectionNameWithout.$cmd", func(t *testing.T) {
		q := must.NotFail(wire.NewOpQuery(must.NotFail(bson.NewDocument("unknown", int32(1)))))
		q.FullCollectionName = "invalid"
		q.NumberToReturn = -1

		_, resBody, err := conn.Request(ctx, &wire.MsgHeader{OpCode: wire.OpCodeQuery}, q)
		require.NoError(t, err)

		resMsg, err := resBody.(*wire.OpReply).Document()
		require.NoError(t, err)

		expected := must.NotFail(types.NewDocument(
			"$err", "OP_QUERY is no longer supported. The client driver may require an upgrade.",
			"code", int32(5739101),
			"ok", float64(0),
		))

		assert.Equal(t, expected, resMsg)
	})

	t.Run("UnknownOpQuery", func(t *testing.T) {
		q := must.NotFail(wire.NewOpQuery(must.NotFail(bson.NewDocument("unknown", int32(1)))))
		q.FullCollectionName = "admin.$cmd"
		q.NumberToReturn = -1

		_, resBody, err := conn.Request(ctx, &wire.MsgHeader{OpCode: wire.OpCodeQuery}, q)
		require.NoError(t, err)

		resMsg, err := resBody.(*wire.OpReply).Document()
		require.NoError(t, err)

		expected := must.NotFail(types.NewDocument(
			"$err", "OP_QUERY is no longer supported. The client driver may require an upgrade.",
			"code", int32(352),
			"ok", float64(0),
		))

		assert.Equal(t, expected, resMsg)
	})
}

func TestOpQueryIsMaster(t *testing.T) {
	t.Parallel()

	ctx, conn := setup.SetupDriver(t)

	for name, tc := range map[string]struct {
		command string
	}{
		"IsMaster": {command: "isMaster"},
		"Ismaster": {command: "ismaster"},
	} {
		t.Run(name, func(t *testing.T) {
			q := must.NotFail(wire.NewOpQuery(must.NotFail(bson.NewDocument(tc.command, int32(1)))))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			resHeader, resBody, err := conn.Request(ctx, &wire.MsgHeader{OpCode: wire.OpCodeQuery}, q)
			require.NoError(t, err)
			assert.NotZero(t, resHeader.RequestID)

			res, err := resBody.(*wire.OpReply).Document()
			require.NoError(t, err)

			connectionID, _ := res.Get("connectionId")
			assert.IsType(t, int32(0), connectionID)
			res.Remove("connectionId")

			localTime, _ := res.Get("localTime")
			assert.IsType(t, time.Time{}, localTime)
			res.Remove("localTime")

			// fields are missing
			res.Remove("hosts")
			res.Remove("setName")
			res.Remove("topologyVersion")
			res.Remove("setVersion")
			res.Remove("secondary")
			res.Remove("primary")
			res.Remove("me")
			res.Remove("electionId")
			res.Remove("lastWrite")

			// fields specific for MongoDB running in a cluster
			res.Remove("$clusterTime")
			res.Remove("operationTime")

			expectedComparable := must.NotFail(types.NewDocument(
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
			assert.Equal(t, expectedComparable, res)
		})
	}
}

func TestOpQueryIsMasterHelloOk(t *testing.T) {
	t.Parallel()

	ctx, conn := setup.SetupDriver(t)

	for name, tc := range map[string]struct {
		command string
	}{
		"IsMaster": {command: "isMaster"},
		"Ismaster": {command: "ismaster"},
	} {
		t.Run(name, func(t *testing.T) {
			q := must.NotFail(wire.NewOpQuery(must.NotFail(bson.NewDocument(
				tc.command, int32(1),
				"helloOk", true,
			))))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			resHeader, resBody, err := conn.Request(ctx, &wire.MsgHeader{OpCode: wire.OpCodeQuery}, q)
			require.NoError(t, err)
			assert.NotZero(t, resHeader.RequestID)

			res, err := resBody.(*wire.OpReply).Document()
			require.NoError(t, err)

			connectionID, _ := res.Get("connectionId")
			assert.IsType(t, int32(0), connectionID)
			res.Remove("connectionId")

			localTime, _ := res.Get("localTime")
			assert.IsType(t, time.Time{}, localTime)
			res.Remove("localTime")

			// fields are missing
			res.Remove("hosts")
			res.Remove("setName")
			res.Remove("topologyVersion")
			res.Remove("setVersion")
			res.Remove("secondary")
			res.Remove("primary")
			res.Remove("me")
			res.Remove("electionId")
			res.Remove("lastWrite")

			// fields specific for MongoDB running in a cluster
			res.Remove("$clusterTime")
			res.Remove("operationTime")

			expectedComparable := must.NotFail(types.NewDocument(
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
			assert.Equal(t, expectedComparable, res)
		})
	}
}

func TestOpQueryHello(t *testing.T) {
	t.Parallel()

	ctx, conn := setup.SetupDriver(t)

	q := must.NotFail(wire.NewOpQuery(must.NotFail(bson.NewDocument(
		"hello", int32(1),
	))))
	q.FullCollectionName = "admin.$cmd"
	q.NumberToReturn = -1

	_, resBody, err := conn.Request(ctx, &wire.MsgHeader{OpCode: wire.OpCodeQuery}, q)
	require.NoError(t, err)

	res, err := resBody.(*wire.OpReply).Document()
	require.NoError(t, err)

	connectionID, _ := res.Get("connectionId")
	assert.IsType(t, int32(0), connectionID)
	res.Remove("connectionId")

	localTime, _ := res.Get("localTime")
	assert.IsType(t, time.Time{}, localTime)
	res.Remove("localTime")

	// fields are missing
	res.Remove("topologyVersion")
	res.Remove("hosts")
	res.Remove("setName")
	res.Remove("setVersion")
	res.Remove("secondary")
	res.Remove("primary")
	res.Remove("me")
	res.Remove("electionId")
	res.Remove("lastWrite")

	// fields specific for MongoDB running in a cluster
	res.Remove("$clusterTime")
	res.Remove("operationTime")

	expectedComparable := must.NotFail(types.NewDocument(
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
	assert.Equal(t, expectedComparable, res)
}

func TestOpQuerySpeculativeAuthenticate(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		command string
	}{
		"Hello":    {command: "hello"},
		"IsMaster": {command: "isMaster"},
		// avoid panic: Directory name "testopqueryspeculativeauthenticate_ismaster" already used by
		// another test "TestOpQuerySpeculativeAuthenticate/Ismaster".
		"IsmasterLower": {command: "ismaster"},
	} {
		t.Run(name, func(t *testing.T) {
			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				UseDriver:      true,
				DriverNoAuth:   true,
				ExtraOptions:   url.Values{"authMechanism": []string{"SCRAM-SHA-256"}},
				BackendOptions: &setup.BackendOpts{DisableTLS: true},
			})

			ctx, conn := s.Ctx, s.DriverConn

			authCreds, authMechanism := conn.AuthInfo()
			require.Equal(t, "SCRAM-SHA-256", authMechanism)

			username := authCreds.Username()
			password, _ := authCreds.Password()
			client := must.NotFail(scram.SHA256.NewClient(username, password, ""))
			conv := client.NewConversation()
			payload := must.NotFail(conv.Step(""))

			t.Run("SpeculativeAuthenticate", func(t *testing.T) {
				query := must.NotFail(wire.NewOpQuery(must.NotFail(must.NotFail(bson.NewDocument(
					tc.command, int32(1),
					"speculativeAuthenticate", must.NotFail(bson.NewDocument(
						"saslStart", int32(1),
						"mechanism", authMechanism,
						"payload", bson.Binary{B: []byte(payload)},
						"db", "admin",
					)),
				)).Encode())))
				query.FullCollectionName = "admin.$cmd"
				query.NumberToReturn = -1

				_, resBody, err := conn.Request(ctx, &wire.MsgHeader{OpCode: wire.OpCodeQuery}, query)
				require.NoError(t, err)

				res, err := resBody.(*wire.OpReply).Document()
				require.NoError(t, err)

				ok, _ := res.Get("ok")
				require.Equal(t, float64(1), ok)

				speculativeAuthenticateV, _ := res.Get("speculativeAuthenticate")
				require.NotNil(t, speculativeAuthenticateV)

				speculativeAuthenticate := speculativeAuthenticateV.(*types.Document)
				require.NoError(t, err)

				done, _ := speculativeAuthenticate.Get("done")
				require.Equal(t, false, done)

				payload, err = conv.Step(string(must.NotFail(speculativeAuthenticate.Get("payload")).(types.Binary).B))
				require.NoError(t, err)
			})

			t.Run("SaslContinueFirst", func(tt *testing.T) {
				msg := must.NotFail(wire.NewOpMsg(must.NotFail(must.NotFail(bson.NewDocument(
					"saslContinue", int32(1),
					"conversationId", int32(1),
					"payload", bson.Binary{B: []byte(payload)},
					"$db", "admin",
				)).Encode())))

				_, resBody, err := conn.Request(ctx, nil, msg)
				require.NoError(tt, err)

				res, err := resBody.(*wire.OpMsg).Document()
				require.NoError(tt, err)

				ok, _ := res.Get("ok")
				assert.Equal(tt, float64(1), ok)

				//	t := setup.FailsForFerretDB(tt, "FIXME")

				done, _ := res.Get("done")
				assert.Equal(t, false, done)

				payload, err = conv.Step(string(must.NotFail(res.Get("payload")).(types.Binary).B))
				assert.NoError(t, err)
			})

			t.Run("SaslContinueSecond", func(tt *testing.T) {
				//	t := setup.FailsForFerretDB(tt, "FIXME")

				msg := must.NotFail(wire.NewOpMsg(must.NotFail(must.NotFail(bson.NewDocument(
					"saslContinue", int32(1),
					"conversationId", int32(1),
					"payload", bson.Binary{B: []byte(payload)},
					"$db", "admin",
				)).Encode())))

				_, resBody, err := conn.Request(ctx, nil, msg)
				require.NoError(t, err)

				res, err := resBody.(*wire.OpMsg).Document()
				require.NoError(t, err)

				ok, _ := res.Get("ok")
				require.Equal(t, float64(1), ok)

				done, _ := res.Get("done")
				require.Equal(t, true, done)

				payloadV, _ := res.Get("payload")
				require.Empty(t, payloadV)
			})
		})

		t.Run(name+"NotSuccess", func(t *testing.T) {
			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				UseDriver:      true,
				DriverNoAuth:   true,
				ExtraOptions:   url.Values{"authMechanism": []string{"SCRAM-SHA-256"}},
				BackendOptions: &setup.BackendOpts{DisableTLS: true},
			})

			ctx, conn := s.Ctx, s.DriverConn
			_, authMechanism := conn.AuthInfo()
			require.Equal(t, "SCRAM-SHA-256", authMechanism)

			client := must.NotFail(scram.SHA256.NewClient("user", "wrong", ""))
			conv := client.NewConversation()
			payload := must.NotFail(conv.Step(""))

			body := must.NotFail(wire.NewOpQuery(must.NotFail(must.NotFail(bson.NewDocument(
				tc.command, int32(1),
				"speculativeAuthenticate", must.NotFail(bson.NewDocument(
					"saslStart", int32(1),
					"mechanism", authMechanism,
					"payload", bson.Binary{B: []byte(payload)},
					"db", "admin",
				)),
			)).Encode())))
			body.FullCollectionName = "admin.$cmd"
			body.NumberToReturn = -1

			_, resBody, err := conn.Request(ctx, &wire.MsgHeader{OpCode: wire.OpCodeQuery}, body)
			require.NoError(t, err)

			res, err := resBody.(*wire.OpReply).Document()
			require.NoError(t, err)

			ok, _ := res.Get("ok")
			require.Equal(t, float64(1), ok)

			speculativeAuthenticate, _ := res.Get("speculativeAuthenticate")
			require.Nil(t, speculativeAuthenticate)
		})

		t.Run(name+"TypeMismatch", func(t *testing.T) {
			ctx, conn := setup.SetupDriver(t)

			q := must.NotFail(wire.NewOpQuery(must.NotFail(bson.NewDocument(
				tc.command, int32(1),
				"speculativeAuthenticate", int32(1),
			))))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			_, resBody, err := conn.Request(ctx, &wire.MsgHeader{OpCode: wire.OpCodeQuery}, q)
			require.NoError(t, err)

			resMsg, err := resBody.(*wire.OpReply).Document()
			require.NoError(t, err)

			expected := must.NotFail(types.NewDocument(
				"ok", float64(0),
				"code", int32(14),
				"codeName", "TypeMismatch",
			))
			require.Equal(t, expected, resMsg)
		})

		t.Run(name+"NoDB", func(t *testing.T) {
			mechanism := "SCRAM-SHA-256"

			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				UseDriver:    true,
				DriverNoAuth: true,
				ExtraOptions: url.Values{"authMechanism": []string{mechanism}},
			})

			ctx, conn := s.Ctx, s.DriverConn
			authCreds, authMechanism := conn.AuthInfo()
			require.Equal(t, mechanism, authMechanism)

			username := authCreds.Username()
			password, _ := authCreds.Password()
			client := must.NotFail(scram.SHA256.NewClient(username, password, ""))
			conv := client.NewConversation()
			payload := must.NotFail(conv.Step(""))

			q := must.NotFail(wire.NewOpQuery(must.NotFail(bson.NewDocument(
				tc.command, int32(1),
				"speculativeAuthenticate", must.NotFail(bson.NewDocument(
					"saslStart", int32(1),
					"mechanism", mechanism,
					"payload", bson.Binary{B: []byte(payload)},
				)),
			))))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			_, resBody, err := conn.Request(ctx, &wire.MsgHeader{OpCode: wire.OpCodeQuery}, q)
			require.NoError(t, err)

			res, err := resBody.(*wire.OpReply).Document()
			require.NoError(t, err)

			ok, _ := res.Get("ok")
			assert.Equal(t, float64(1), ok)

			speculativeAuthenticate, _ := res.Get("speculativeAuthenticate")
			require.Nil(t, speculativeAuthenticate)
		})
	}
}
