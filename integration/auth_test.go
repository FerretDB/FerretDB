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

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/FerretDB/wire/wireclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	xdgscram "github.com/xdg-go/scram"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx := s.Ctx
	collection := s.Collection
	db := collection.Database()

	testCases := map[string]struct { //nolint:vet // for readability
		username       string
		password       string
		updatePassword string // if set, the password will be updated to this one after the user is created.
		mechanisms     bson.A // mechanisms to use for creating user authentication
		authMechanism  string

		userNotFound  bool
		wrongPassword bool
		pingErr       string
	}{
		"ScramSHA256": {
			username:      "scramsha256",
			password:      "password",
			mechanisms:    bson.A{"SCRAM-SHA-256"},
			authMechanism: "SCRAM-SHA-256",
		},
		"MultipleScramSHA256": {
			username:      "scramsha256multi",
			password:      "password",
			mechanisms:    bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"},
			authMechanism: "SCRAM-SHA-256",
		},
		"ScramSHA256Updated": {
			username:       "scramsha256updated",
			password:       "pass123",
			updatePassword: "anotherpassword",
			mechanisms:     bson.A{"SCRAM-SHA-256"},
			authMechanism:  "SCRAM-SHA-256",
		},
		"NotFoundUser": {
			username:      "notfound",
			password:      "something",
			mechanisms:    bson.A{"SCRAM-SHA-256"},
			authMechanism: "SCRAM-SHA-256",
			userNotFound:  true,
			pingErr:       "Authentication failed.",
		},
		"BadPassword": {
			username:      "baduser",
			password:      "something",
			mechanisms:    bson.A{"SCRAM-SHA-256"},
			authMechanism: "SCRAM-SHA-256",
			wrongPassword: true,
			pingErr:       "Authentication failed.",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !tc.userNotFound {
				// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
				_ = db.RunCommand(ctx, bson.D{{"dropUser", tc.username}})

				// root role is only available in admin database, a role with sufficient privilege is used
				roles := bson.A{"readWrite"}
				if !setup.IsMongoDB(t) {
					// TODO https://github.com/FerretDB/FerretDB/issues/3974
					roles = bson.A{}
				}

				createPayload := bson.D{
					{"createUser", tc.username},
					{"roles", roles},
					{"pwd", tc.password},
					{"mechanisms", tc.mechanisms},
				}

				err := db.RunCommand(ctx, createPayload).Err()
				require.NoError(t, err, "cannot create user")
			}

			if tc.updatePassword != "" {
				updatePayload := bson.D{
					{"updateUser", tc.username},
					{"pwd", tc.updatePassword},
				}

				err := db.RunCommand(ctx, updatePayload).Err()
				require.NoError(t, err, "cannot update user")
			}

			password := tc.password
			if tc.updatePassword != "" {
				password = tc.updatePassword
			}

			if tc.wrongPassword {
				password = "wrongpassword"
			}

			credential := options.Credential{
				AuthMechanism: tc.authMechanism,
				AuthSource:    db.Name(),
				Username:      tc.username,
				Password:      password,
			}

			opts := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)

			client, err := mongo.Connect(ctx, opts)
			require.NoError(t, err)

			t.Cleanup(func() {
				require.NoError(t, client.Disconnect(ctx))
			})

			// Ping to force connection to be established and tested.
			err = client.Ping(ctx, nil)

			if tc.pingErr != "" {
				require.ErrorContains(t, err, tc.pingErr)
				return
			}

			connCollection := client.Database(db.Name()).Collection(collection.Name())

			require.NotNil(t, connCollection, "cannot get collection")

			r, err := connCollection.InsertOne(ctx, bson.D{{"ping", "pong"}})
			require.NoError(t, err, "cannot insert document")
			id := r.InsertedID.(primitive.ObjectID)
			require.NotEmpty(t, id)

			var result bson.D
			err = connCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&result)
			require.NoError(t, err, "cannot find document")
			require.Equal(t, bson.D{{"_id", id}, {"ping", "pong"}}, result)
		})
	}
}

func TestAuthAlreadyAuthenticated(tt *testing.T) {
	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/953")

	tt.Parallel()

	s := setup.SetupWithOpts(tt, nil)
	ctx, db := s.Ctx, s.Collection.Database()
	username, password, mechanism := "testuser", "testpass", "SCRAM-SHA-256"

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	_ = db.RunCommand(ctx, bson.D{{"dropUser", username}})

	err := db.RunCommand(ctx, bson.D{
		{"createUser", username},
		{"roles", bson.A{}},
		{"pwd", password},
		{"mechanisms", bson.A{mechanism}},
	}).Err()
	require.NoError(t, err, "cannot create user")

	credential := options.Credential{
		AuthMechanism: mechanism,
		AuthSource:    db.Name(),
		Username:      username,
		Password:      password,
	}

	opts := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)

	client, err := mongo.Connect(ctx, opts)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, client.Disconnect(ctx))
	})

	db = client.Database(db.Name())
	var res bson.D
	err = db.RunCommand(ctx, bson.D{{"connectionStatus", 1}}).Decode(&res)
	require.NoError(t, err)

	expected := bson.D{
		{"authInfo", bson.D{
			{"authenticatedUsers", bson.A{
				bson.D{
					{"user", username},
					{"db", db.Name()},
				},
			}},
			{"authenticatedUserRoles", bson.A{}},
		}},
		{"ok", float64(1)},
	}
	AssertEqualDocuments(t, expected, res)

	saslStart := bson.D{
		{"saslStart", 1},
		{"mechanism", mechanism},
		{"payload", []byte("n,,n=testuser,r=Y0iJqJu58tGDrUdtqS7+m0oMe4sau3f6")},
		{"autoAuthorize", 1},
		{"options", bson.D{{"skipEmptyExchange", true}}},
	}
	err = db.RunCommand(ctx, saslStart).Decode(&res)
	require.NoError(t, err)

	err = db.RunCommand(ctx, bson.D{{"connectionStatus", 1}}).Decode(&res)
	require.NoError(t, err)
	AssertEqualDocuments(t, expected, res)

	err = db.RunCommand(ctx, saslStart).Decode(&res)
	require.NoError(t, err)
}

func TestSASLContinueErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		WireConn: setup.WireConnNoAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	u, err := url.Parse(s.MongoDBURI)
	require.NoError(t, err)

	username := u.User.Username()
	password, _ := u.User.Password()

	client := must.NotFail(xdgscram.SHA256.NewClient(username, password, ""))

	t.Run("HandshakeFails", func(t *testing.T) {
		conv := client.NewConversation()
		payload := must.NotFail(conv.Step(""))

		msg := must.NotFail(wire.NewOpMsg(must.NotFail(must.NotFail(wirebson.NewDocument(
			"saslStart", int32(1),
			"mechanism", "SCRAM-SHA-256",
			"payload", wirebson.Binary{B: []byte(payload)},
			"$db", "admin",
		)).Encode())))

		var resBody wire.MsgBody
		_, resBody, err = conn.Request(ctx, msg)
		require.NoError(t, err)

		var res *wirebson.Document
		res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		p, ok := res.Get("payload").(wirebson.Binary)
		require.True(t, ok)

		require.NotEmpty(t, p.B)
		err = res.Replace("payload", wirebson.Binary{})
		require.NoError(t, err)

		fixCluster(t, res)

		expected := must.NotFail(wirebson.NewDocument(
			"conversationId", int32(1),
			"done", false,
			"payload", wirebson.Binary{},
			"ok", float64(1),
		))

		testutil.AssertEqual(t, expected, res)

		payload = "invalid"

		msg = must.NotFail(wire.NewOpMsg(must.NotFail(must.NotFail(wirebson.NewDocument(
			"saslContinue", int32(1),
			"conversationId", int32(1),
			"payload", wirebson.Binary{B: []byte(payload)},
			"$db", "admin",
		)).Encode())))

		_, resBody, err = conn.Request(ctx, msg)
		require.NoError(t, err)

		res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		fixCluster(t, res)

		expected = must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "Authentication failed.",
			"code", int32(18),
			"codeName", "AuthenticationFailed",
		))

		testutil.AssertEqual(t, expected, res)

		msg = must.NotFail(wire.NewOpMsg(must.NotFail(must.NotFail(wirebson.NewDocument(
			"saslContinue", int32(1),
			"conversationId", int32(1),
			"payload", wirebson.Binary{},
			"$db", "admin",
		)).Encode())))

		_, resBody, err = conn.Request(ctx, msg)
		require.NoError(t, err)

		res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
		require.NoError(t, err)

		fixCluster(t, res)

		expected = must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "No SASL session state found",
			"code", int32(17),
			"codeName", "ProtocolError",
		))

		testutil.AssertEqual(t, expected, res)
	})

	t.Run("NoAuthenticatedUser", func(t *testing.T) {
		msg := must.NotFail(wire.NewOpMsg(must.NotFail(must.NotFail(wirebson.NewDocument(
			"connectionStatus", int64(1),
			"$db", testutil.DatabaseName(t),
		)).Encode())))

		var resBody wire.MsgBody
		_, resBody, err = conn.Request(ctx, msg)
		require.NoError(t, err)

		var res *wirebson.Document
		res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
		require.NoError(t, err)

		fixCluster(t, res)

		expected := must.NotFail(wirebson.NewDocument(
			"authInfo", must.NotFail(wirebson.NewDocument(
				"authenticatedUsers", wirebson.MakeArray(0),
				"authenticatedUserRoles", wirebson.MakeArray(0),
			)),
			"ok", float64(1),
		))

		testutil.AssertEqual(t, expected, res)
	})

	t.Run("FindFails", func(t *testing.T) {
		msg := must.NotFail(wire.NewOpMsg(must.NotFail(must.NotFail(wirebson.NewDocument(
			"find", "values",
			"filter", must.NotFail(wirebson.NewDocument()),
			"$db", testutil.DatabaseName(t),
		)).Encode())))

		var resBody wire.MsgBody
		_, resBody, err = conn.Request(ctx, msg)
		require.NoError(t, err)

		var res *wirebson.Document
		res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
		require.NoError(t, err)

		fixCluster(t, res)

		expected := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "Command find requires authentication",
			"code", int32(13),
			"codeName", "Unauthorized",
		))

		testutil.AssertEqual(t, expected, res)
	})
}

func TestSASLContinueNoConversation(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		WireConn: setup.WireConnNoAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	msg := must.NotFail(wire.NewOpMsg(must.NotFail(must.NotFail(wirebson.NewDocument(
		"saslContinue", int32(1),
		"conversationId", int32(1),
		"payload", wirebson.Binary{},
		"$db", "admin",
	)).Encode())))

	_, resBody, err := conn.Request(ctx, msg)
	require.NoError(t, err)

	res, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).DecodeDeep()
	require.NoError(t, err)

	fixCluster(t, res)

	expected := must.NotFail(wirebson.NewDocument(
		"ok", float64(0),
		"errmsg", "No SASL session state found",
		"code", int32(17),
		"codeName", "ProtocolError",
	))

	testutil.AssertEqual(t, expected, res)
}

func TestHelloIsMasterOpQuerySpeculative(t *testing.T) {
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
				WireConn: setup.WireConnNoAuth,
			})

			ctx, conn := s.Ctx, s.WireConn

			u, err := url.Parse(s.MongoDBURI)
			require.NoError(t, err)

			username := u.User.Username()
			password, _ := u.User.Password()

			client := must.NotFail(xdgscram.SHA256.NewClient(username, password, ""))
			conv := client.NewConversation()
			payload := must.NotFail(conv.Step(""))

			t.Run("SpeculativeAuthenticate", func(t *testing.T) {
				// as is request sent by C# driver
				q := must.NotFail(wire.NewOpQuery(must.NotFail(must.NotFail(wirebson.NewDocument(
					tc.command, int32(1),
					"helloOk", true,
					"client", must.NotFail(wirebson.NewDocument(
						"driver", must.NotFail(wirebson.NewDocument(
							"name", "mongo-go-driver",
							"version", "2.28.0",
						)),
						"os", must.NotFail(wirebson.NewDocument(
							"type", "Linux",
							"name", "Ubuntu 22.04.5 LTS",
							"architecture", "amd64",
							"version", "22.04.5",
						)),
						"platform", ".NET 8.0.8",
					)),
					"compression", must.NotFail(wirebson.NewArray()),
					"speculativeAuthenticate", must.NotFail(wirebson.NewDocument(
						"saslStart", int32(1),
						"mechanism", "SCRAM-SHA-256",
						"payload", wirebson.Binary{B: []byte(payload)},
						"options", must.NotFail(wirebson.NewDocument("skipEmptyExchange", true)),
						"db", "admin",
					)),
				)).Encode())))
				q.FullCollectionName = "admin.$cmd"
				q.NumberToReturn = -1

				var resBody wire.MsgBody
				_, resBody, err = conn.Request(ctx, q)
				require.NoError(t, err)

				var res *wirebson.Document
				res, err = resBody.(*wire.OpReply).RawDocument().Decode()
				require.NoError(t, err)

				ok := res.Get("ok")
				require.Equal(t, float64(1), ok)

				speculativeAuthenticateV := res.Get("speculativeAuthenticate")
				require.NotNil(t, speculativeAuthenticateV)

				var speculativeAuthenticate *wirebson.Document
				speculativeAuthenticate, err = speculativeAuthenticateV.(wirebson.AnyDocument).Decode()
				require.NoError(t, err)

				done := speculativeAuthenticate.Get("done")
				require.Equal(t, false, done)

				payload, err = conv.Step(string(speculativeAuthenticate.Get("payload").(wirebson.Binary).B))
				require.NoError(t, err)
			})

			t.Run("SASLContinue", func(t *testing.T) {
				q := must.NotFail(wire.NewOpQuery(must.NotFail(must.NotFail(wirebson.NewDocument(
					"saslContinue", int32(1),
					"conversationId", int32(1),
					"payload", wirebson.Binary{B: []byte(payload)},
				)).Encode())))
				q.FullCollectionName = "admin.$cmd"
				q.NumberToReturn = -1

				var resBody wire.MsgBody
				_, resBody, err = conn.Request(ctx, q)
				require.NoError(t, err)

				var res *wirebson.Document
				res, err = resBody.(*wire.OpReply).RawDocument().Decode()
				require.NoError(t, err)

				serverPayload, ok := res.Get("payload").(wirebson.Binary)
				require.True(t, ok)

				fixCluster(t, res)

				expectedComparable := must.NotFail(wirebson.NewDocument(
					"conversationId", int32(1),
					"done", true,
					"payload", serverPayload,
					"ok", float64(1),
				))

				testutil.AssertEqual(t, expectedComparable, res)

				_, err = conv.Step(string(serverPayload.B))
				require.NoError(t, err)

				require.True(t, conv.Valid())
			})

			t.Run("SASLContinueTooMany", func(t *testing.T) {
				q := must.NotFail(wire.NewOpQuery(must.NotFail(must.NotFail(wirebson.NewDocument(
					"saslContinue", int32(1),
					"payload", wirebson.Binary{},
					"conversationId", int32(1),
				)).Encode())))
				q.FullCollectionName = "admin.$cmd"
				q.NumberToReturn = -1

				var resBody wire.MsgBody
				_, resBody, err = conn.Request(ctx, q)
				require.NoError(t, err)

				var res *wirebson.Document
				res, err = resBody.(*wire.OpReply).RawDocument().Decode()
				require.NoError(t, err)

				fixCluster(t, res)

				expected := must.NotFail(wirebson.NewDocument(
					"ok", float64(0),
					"errmsg", "No SASL session state found",
					"code", int32(17),
					"codeName", "ProtocolError",
				))
				testutil.AssertEqual(t, expected, res)
			})
		})

		t.Run(name+"NotSuccess", func(t *testing.T) {
			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				WireConn: setup.WireConnNoAuth,
			})

			ctx, conn := s.Ctx, s.WireConn

			client := must.NotFail(xdgscram.SHA256.NewClient("user", "wrong", ""))
			conv := client.NewConversation()
			payload := must.NotFail(conv.Step(""))

			q := must.NotFail(wire.NewOpQuery(must.NotFail(must.NotFail(wirebson.NewDocument(
				tc.command, int32(1),
				"speculativeAuthenticate", must.NotFail(wirebson.NewDocument(
					"saslStart", int32(1),
					"mechanism", "SCRAM-SHA-256",
					"payload", wirebson.Binary{B: []byte(payload)},
					"db", "admin",
				)),
			)).Encode())))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			_, resBody, err := conn.Request(ctx, q)
			require.NoError(t, err)

			res, err := resBody.(*wire.OpReply).RawDocument().Decode()
			require.NoError(t, err)

			ok := res.Get("ok")
			require.Equal(t, float64(1), ok)

			speculativeAuthenticate := res.Get("speculativeAuthenticate")
			require.Nil(t, speculativeAuthenticate)
		})

		t.Run(name+"TypeMismatch", func(t *testing.T) {
			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				WireConn: setup.WireConnNoAuth,
			})

			ctx, conn := s.Ctx, s.WireConn

			q := must.NotFail(wire.NewOpQuery(must.NotFail(wirebson.NewDocument(
				tc.command, int32(1),
				"speculativeAuthenticate", int32(1),
			))))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			_, resBody, err := conn.Request(ctx, q)
			require.NoError(t, err)

			resMsg, err := resBody.(*wire.OpReply).RawDocument().Decode()
			require.NoError(t, err)

			ok := resMsg.Get("ok")
			assert.Equal(t, float64(0), ok)

			code := resMsg.Get("code")
			assert.Equal(t, int32(14), code)

			codeName := resMsg.Get("codeName")
			assert.Equal(t, "TypeMismatch", codeName)
		})

		t.Run(name+"NoDB", func(t *testing.T) {
			mechanism := "SCRAM-SHA-256"

			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				WireConn: setup.WireConnNoAuth,
			})

			ctx, conn := s.Ctx, s.WireConn

			u, err := url.Parse(s.MongoDBURI)
			require.NoError(t, err)

			username := u.User.Username()
			password, _ := u.User.Password()

			client := must.NotFail(xdgscram.SHA256.NewClient(username, password, ""))
			conv := client.NewConversation()
			payload := must.NotFail(conv.Step(""))

			q := must.NotFail(wire.NewOpQuery(must.NotFail(wirebson.NewDocument(
				tc.command, int32(1),
				"speculativeAuthenticate", must.NotFail(wirebson.NewDocument(
					"saslStart", int32(1),
					"mechanism", mechanism,
					"payload", wirebson.Binary{B: []byte(payload)},
				)),
			))))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			_, resBody, err := conn.Request(ctx, q)
			require.NoError(t, err)

			res, err := resBody.(*wire.OpReply).RawDocument().Decode()
			require.NoError(t, err)

			ok := res.Get("ok")
			assert.Equal(t, float64(1), ok)

			speculativeAuthenticate := res.Get("speculativeAuthenticate")
			require.Nil(t, speculativeAuthenticate)
		})
	}
}

func TestHelloSpeculative(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		WireConn: setup.WireConnNoAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	u, err := url.Parse(s.MongoDBURI)
	require.NoError(t, err)

	username := u.User.Username()
	password, _ := u.User.Password()

	client := must.NotFail(xdgscram.SHA256.NewClient(username, password, ""))
	conv := client.NewConversation()
	clientPayload := must.NotFail(conv.Step(""))

	t.Run("SpeculativeAuthenticate", func(t *testing.T) {
		// as is request sent by legacy Mongo shell
		msg := wire.MustOpMsg(
			"hello", int32(1),
			"speculativeAuthenticate", must.NotFail(wirebson.NewDocument(
				"saslStart", int32(1),
				"mechanism", "SCRAM-SHA-256",
				"payload", wirebson.Binary{B: []byte(clientPayload)},
				"db", "admin",
			)),
			"saslSupportedMechs", "admin."+username,
			"client", must.NotFail(wirebson.NewDocument(
				"application", must.NotFail(wirebson.NewDocument(
					"name", "MongoDB Shell",
				)),
				"driver", must.NotFail(wirebson.NewDocument(
					"name", "MongoDB Internal Client",
					"version", "7.0.8",
				)),
				"os", must.NotFail(wirebson.NewDocument(
					"type", "Linux",
					"name", "Ubuntu",
					"architecture", "x86_64",
					"version", "22.04",
				)),
			)),
			"$db", "admin",
		)

		var resBody wire.MsgBody
		_, resBody, err = conn.Request(ctx, msg)
		require.NoError(t, err)

		var res *wirebson.Document
		res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		speculativeAuthenticateV, ok := res.Get("speculativeAuthenticate").(wirebson.AnyDocument)
		require.True(t, ok)

		var speculativeAuthenticate *wirebson.Document
		speculativeAuthenticate, err = speculativeAuthenticateV.Decode()
		require.NoError(t, err)

		serverPayload, ok := speculativeAuthenticate.Get("payload").(wirebson.Binary)
		require.True(t, ok)

		connectionID := res.Get("connectionId")
		assert.IsType(t, int32(0), connectionID)

		localTime := res.Get("localTime")
		assert.IsType(t, time.Time{}, localTime)

		saslSupportedMechs := res.Get("saslSupportedMechs")
		require.NotNil(t, saslSupportedMechs)

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
			"isWritablePrimary", true,
			"maxBsonObjectSize", int32(16777216),
			"maxMessageSizeBytes", int32(48000000),
			"maxWriteBatchSize", int32(100000),
			"localTime", localTime,
			"logicalSessionTimeoutMinutes", int32(30),
			"connectionId", connectionID,
			"minWireVersion", int32(0),
			"maxWireVersion", int32(21),
			"readOnly", false,
			"saslSupportedMechs", saslSupportedMechs,
			"speculativeAuthenticate", must.NotFail(wirebson.NewDocument(
				"conversationId", int32(1),
				"done", false,
				"payload", serverPayload,
			)),
			"ok", float64(1),
		))

		testutil.AssertEqual(t, expectedComparable, res)

		clientPayload, err = conv.Step(string(serverPayload.B))
		require.NoError(t, err)
	})

	t.Run("SASLContinue", func(t *testing.T) {
		msg := wire.MustOpMsg(
			"saslContinue", int32(1),
			"payload", wirebson.Binary{B: []byte(clientPayload)},
			"conversationId", int32(1),
			"$db", "admin",
		)

		var resBody wire.MsgBody
		_, resBody, err = conn.Request(ctx, msg)
		require.NoError(t, err)

		var res *wirebson.Document
		res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		serverPayload, ok := res.Get("payload").(wirebson.Binary)
		require.True(t, ok, "got: "+res.LogMessageIndent())

		fixCluster(t, res)

		expectedComparable := must.NotFail(wirebson.NewDocument(
			"conversationId", int32(1),
			"done", false,
			"payload", serverPayload,
			"ok", float64(1),
		))

		testutil.AssertEqual(t, expectedComparable, res)

		_, err = conv.Step(string(serverPayload.B))
		require.NoError(t, err)

		require.True(t, conv.Valid())
	})

	t.Run("SASLContinueEmpty", func(t *testing.T) {
		msg := wire.MustOpMsg(
			"saslContinue", int32(1),
			"payload", wirebson.Binary{}, // client sends empty payload
			"conversationId", int32(1),
			"$db", "admin",
		)

		var resBody wire.MsgBody
		_, resBody, err = conn.Request(ctx, msg)
		require.NoError(t, err)

		var res *wirebson.Document
		res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		serverPayload, ok := res.Get("payload").(wirebson.Binary)
		require.True(t, ok, "got: "+res.LogMessageIndent())

		fixCluster(t, res)

		expectedComparable := must.NotFail(wirebson.NewDocument(
			"conversationId", int32(1),
			"done", true,
			"payload", serverPayload,
			"ok", float64(1),
		))

		testutil.AssertEqual(t, expectedComparable, res)
	})

	t.Run("SASLContinueTooMany", func(t *testing.T) {
		msg := wire.MustOpMsg(
			"saslContinue", int32(1),
			"payload", wirebson.Binary{},
			"conversationId", int32(1),
			"$db", "admin",
		)

		var resBody wire.MsgBody
		_, resBody, err = conn.Request(ctx, msg)
		require.NoError(t, err)

		var res *wirebson.Document
		res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		fixCluster(t, res)

		expected := must.NotFail(wirebson.NewDocument(
			"ok", float64(0),
			"errmsg", "No SASL session state found",
			"code", int32(17),
			"codeName", "ProtocolError",
		))
		testutil.AssertEqual(t, expected, res)
	})
}

func TestHelloOpQuerySASLSupportedMechs(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		WireConn: setup.WireConnNoAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	u, err := url.Parse(s.MongoDBURI)
	require.NoError(t, err)

	username := u.User.Username()
	password, _ := u.User.Password()

	client := must.NotFail(xdgscram.SHA256.NewClient(username, password, ""))
	conv := client.NewConversation()
	payload := must.NotFail(conv.Step(""))

	q := must.NotFail(wire.NewOpQuery(must.NotFail(must.NotFail(wirebson.NewDocument(
		"hello", int32(1),
		"saslSupportedMechs", "admin.username",
		"speculativeAuthenticate", must.NotFail(wirebson.NewDocument(
			"saslStart", int32(1),
			"mechanism", "SCRAM-SHA-256",
			"payload", wirebson.Binary{B: []byte(payload)},
			"db", "admin",
		)),
	)).Encode())))
	q.FullCollectionName = "admin.$cmd"
	q.NumberToReturn = -1

	var resBody wire.MsgBody
	_, resBody, err = conn.Request(ctx, q)
	require.NoError(t, err)

	var res *wirebson.Document
	res, err = resBody.(*wire.OpReply).RawDocument().DecodeDeep()
	require.NoError(t, err)

	ok := res.Get("ok")
	require.Equal(t, float64(1), ok)

	speculativeAuthenticateV := res.Get("speculativeAuthenticate")
	require.NotNil(t, speculativeAuthenticateV)

	payloadAny := speculativeAuthenticateV.(*wirebson.Document).Get("payload")

	connectionID := res.Get("connectionId")
	assert.IsType(t, int32(0), connectionID)
	res.Remove("connectionId")

	localTime := res.Get("localTime")
	assert.IsType(t, time.Time{}, localTime)
	res.Remove("localTime")

	saslSupportedMechs := res.Get("saslSupportedMechs")
	require.NotNil(t, saslSupportedMechs)

	mechs := saslSupportedMechs.(*wirebson.Array)

	var found bool

	for v := range mechs.Values() {
		if v == "SCRAM-SHA-256" {
			found = true
			break
		}
	}

	assert.True(t, found, "expected SCRAM-SHA-256")

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
		"isWritablePrimary", true,
		"maxBsonObjectSize", int32(16777216),
		"maxMessageSizeBytes", int32(48000000),
		"maxWriteBatchSize", int32(100000),
		"logicalSessionTimeoutMinutes", int32(30),
		"minWireVersion", int32(0),
		"maxWireVersion", int32(21),
		"readOnly", false,
		"saslSupportedMechs", saslSupportedMechs,
		"speculativeAuthenticate", must.NotFail(wirebson.NewDocument(
			"conversationId", int32(1),
			"done", false,
			"payload", payloadAny,
		)),
		"ok", float64(1),
	))

	testutil.AssertEqual(t, expectedComparable, res)
}

func TestSASLStartOpQueryErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		WireConn: setup.WireConnNoAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	u, err := url.Parse(s.MongoDBURI)
	require.NoError(t, err)

	username := u.User.Username()
	password, _ := u.User.Password()

	client := must.NotFail(xdgscram.SHA256.NewClient(username, password, ""))
	conv := client.NewConversation()
	payload := must.NotFail(conv.Step(""))

	for name, tc := range map[string]struct { //nolint:vet // test only
		query              *wirebson.Document
		fullCollectionName string

		reply            *wirebson.Document
		failsForFerretDB string
	}{
		"WrongAuthDB": {
			query: must.NotFail(wirebson.NewDocument(
				"saslStart", int32(1),
				"mechanism", "SCRAM-SHA-256",
				"payload", wirebson.Binary{B: []byte(payload)},
			)),
			fullCollectionName: "wrong-auth-db.$cmd",
			reply: must.NotFail(wirebson.NewDocument(
				"ok", float64(0),
				"errmsg", "Authentication failed.",
				"code", int32(18),
				"codeName", "AuthenticationFailed",
			)),
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/932",
		},
		"NotAllowed$db": {
			query: must.NotFail(wirebson.NewDocument(
				"saslStart", int32(1),
				"mechanism", "SCRAM-SHA-256",
				"payload", wirebson.Binary{B: []byte(payload)},
				"$db", "admin",
			)),
			fullCollectionName: "admin.$cmd",
			reply: must.NotFail(wirebson.NewDocument(
				"ok", float64(0),
				"errmsg", "$db is not allowed in OP_QUERY requests",
				"code", int32(40621),
				"codeName", "Location40621",
			)),
		},
		"InvalidPayload": {
			query: must.NotFail(wirebson.NewDocument(
				"saslStart", int32(1),
				"mechanism", "SCRAM-SHA-256",
				"payload", wirebson.Binary{B: []byte("invalid")},
			)),
			fullCollectionName: "admin.$cmd",
			reply: must.NotFail(wirebson.NewDocument(
				"ok", float64(0),
				"errmsg", "Authentication failed.",
				"code", int32(18),
				"codeName", "AuthenticationFailed",
			)),
		},
		"UnsupportedMechanism": {
			query: must.NotFail(wirebson.NewDocument(
				"saslStart", int32(1),
				"mechanism", "UNKNOWN-MECHANISM",
				"payload", wirebson.Binary{B: []byte(payload)},
			)),
			fullCollectionName: "admin.$cmd",
			reply: must.NotFail(wirebson.NewDocument(
				"ok", float64(0),
				"errmsg", "Received authentication for mechanism UNKNOWN-MECHANISM which is not enabled",
				"code", int32(334),
				"codeName", "MechanismUnavailable",
			)),
		},
	} {
		t.Run(name, func(tt *testing.T) {
			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(t, tc.failsForFerretDB)
			}

			q := must.NotFail(wire.NewOpQuery(must.NotFail(tc.query.Encode())))
			q.FullCollectionName = tc.fullCollectionName
			q.NumberToReturn = -1

			var resBody wire.MsgBody
			_, resBody, err = conn.Request(ctx, q)
			require.NoError(t, err)

			var res *wirebson.Document
			res, err = resBody.(*wire.OpReply).RawDocument().Decode()
			require.NoError(t, err)

			fixCluster(t, res)

			testutil.AssertEqual(t, tc.reply, res)
		})
	}
}

func TestSASLStartErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		WireConn: setup.WireConnNoAuth,
	})

	ctx, conn := s.Ctx, s.WireConn

	u, err := url.Parse(s.MongoDBURI)
	require.NoError(t, err)

	username := u.User.Username()
	password, _ := u.User.Password()

	client := must.NotFail(xdgscram.SHA256.NewClient(username, password, ""))
	conv := client.NewConversation()
	payload := must.NotFail(conv.Step(""))

	for name, tc := range map[string]struct {
		msg *wirebson.Document

		res *wirebson.Document
	}{
		"InvalidOptions": {
			msg: must.NotFail(wirebson.NewDocument(
				"saslStart", int32(1),
				"mechanism", "SCRAM-SHA-256",
				"payload", wirebson.Binary{B: []byte(payload)},
				"options", "invalid",
				"$db", "admin",
			)),
			res: must.NotFail(wirebson.NewDocument(
				"ok", float64(0),
				"errmsg", "BSON field 'saslStart.options' is the wrong type 'string', expected type 'object'",
				"code", int32(14),
				"codeName", "TypeMismatch",
			)),
		},
		"InvalidOptionsSkipEmptyExchange": {
			msg: must.NotFail(wirebson.NewDocument(
				"saslStart", int32(1),
				"mechanism", "SCRAM-SHA-256",
				"payload", wirebson.Binary{B: []byte(payload)},
				"options", wirebson.MustDocument("skipEmptyExchange", "invalid"),
				"$db", "admin",
			)),
			res: must.NotFail(wirebson.NewDocument(
				"ok", float64(0),
				"errmsg", "Authentication failed.",
				"code", int32(18),
				"codeName", "AuthenticationFailed",
			)),
		},
	} {
		t.Run(name, func(t *testing.T) {
			q := must.NotFail(wire.NewOpMsg(must.NotFail(tc.msg.Encode())))

			var resBody wire.MsgBody
			_, resBody, err = conn.Request(ctx, q)
			require.NoError(t, err)

			var res *wirebson.Document
			res, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
			require.NoError(t, err)

			fixCluster(t, res)

			testutil.AssertEqual(t, tc.res, res)
		})
	}
}

func TestSASLContinueOpQueryErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	ctx := s.Ctx

	for name, tc := range map[string]struct { //nolint:vet // test only
		query              *wirebson.Document
		fullCollectionName string

		reply *wirebson.Document
	}{
		"AuthenticationError": {
			query: must.NotFail(wirebson.NewDocument(
				"saslContinue", int32(1),
				"conversationId", int32(1),
				"payload", wirebson.Binary{},
			)),
			fullCollectionName: "wrong-auth-db.$cmd",
			reply: must.NotFail(wirebson.NewDocument(
				"ok", float64(0),
				"errmsg", "Authentication failed.",
				"code", int32(18),
				"codeName", "AuthenticationFailed",
			)),
		},
		"NotAllowed$db": {
			query: must.NotFail(wirebson.NewDocument(
				"saslContinue", int32(1),
				"conversationId", int32(1),
				"payload", wirebson.Binary{},
				"$db", "admin",
			)),
			fullCollectionName: "admin.$cmd",
			reply: must.NotFail(wirebson.NewDocument(
				"ok", float64(0),
				"errmsg", "$db is not allowed in OP_QUERY requests",
				"code", int32(40621),
				"codeName", "Location40621",
			)),
		},
	} {
		t.Run(name, func(t *testing.T) {
			u, err := url.Parse(s.MongoDBURI)
			require.NoError(t, err)

			username := u.User.Username()
			password, _ := u.User.Password()

			client := must.NotFail(xdgscram.SHA256.NewClient(username, password, ""))
			conv := client.NewConversation()
			payload := must.NotFail(conv.Step(""))

			q := must.NotFail(wire.NewOpQuery(must.NotFail(must.NotFail(wirebson.NewDocument(
				"saslStart", int32(1),
				"mechanism", "SCRAM-SHA-256",
				"payload", wirebson.Binary{B: []byte(payload)},
			)).Encode())))
			q.FullCollectionName = "admin.$cmd"
			q.NumberToReturn = -1

			conn, err := wireclient.Connect(ctx, s.MongoDBURI, testutil.Logger(t))
			require.NoError(t, err)

			_, resBody, err := conn.Request(ctx, q)
			require.NoError(t, err)

			res, err := resBody.(*wire.OpReply).RawDocument().Decode()
			require.NoError(t, err)

			fixCluster(t, res)

			err = res.Replace("payload", wirebson.Binary{})
			require.NoError(t, err)

			expectedComparable := must.NotFail(wirebson.NewDocument(
				"conversationId", int32(1),
				"done", false,
				"payload", wirebson.Binary{},
				"ok", float64(1),
			))

			testutil.AssertEqual(t, expectedComparable, res)

			q = must.NotFail(wire.NewOpQuery(must.NotFail(tc.query.Encode())))
			q.FullCollectionName = tc.fullCollectionName
			q.NumberToReturn = -1

			_, resBody, err = conn.Request(ctx, q)
			require.NoError(t, err)

			res, err = resBody.(*wire.OpReply).RawDocument().Decode()
			require.NoError(t, err)

			fixCluster(t, res)

			testutil.AssertEqual(t, tc.reply, res)
		})
	}
}
