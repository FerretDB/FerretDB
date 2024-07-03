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

package users

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testfail"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestAuthentication(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx := s.Ctx
	collection := s.Collection
	db := collection.Database()

	testCases := map[string]struct { //nolint:vet // for readability
		username       string
		password       string
		updatePassword string // if true, the password will be updated to this one after the user is created.
		mechanisms     bson.A // mechanisms to use for creating a user

		connectionMechanism string // if set, try to establish connection with this mechanism

		userNotFound  bool
		wrongPassword bool
		topologyError bool
		errorMessage  string
	}{
		"FailPLAIN": {
			username:            "plain-user",
			password:            "password",
			connectionMechanism: "PLAIN",
			topologyError:       true,
			errorMessage:        `unable to authenticate using mechanism "PLAIN"`,
		},
		"ScramSHA1": {
			username:            "scramsha1",
			password:            "password",
			mechanisms:          bson.A{"SCRAM-SHA-1"},
			connectionMechanism: "SCRAM-SHA-1",
		},
		"ScramSHA256": {
			username:            "scramsha256",
			password:            "password",
			mechanisms:          bson.A{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
		},
		"MultipleScramSHA1": {
			username:            "scramsha1multi",
			password:            "password",
			mechanisms:          bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-1",
		},
		"MultipleScramSHA256": {
			username:            "scramsha256multi",
			password:            "password",
			mechanisms:          bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
		},
		"ScramSHA1Updated": {
			username:            "scramsha1updated",
			password:            "pass123",
			updatePassword:      "anotherpassword",
			mechanisms:          bson.A{"SCRAM-SHA-1"},
			connectionMechanism: "SCRAM-SHA-1",
		},
		"ScramSHA256Updated": {
			username:            "scramsha256updated",
			password:            "pass123",
			updatePassword:      "anotherpassword",
			mechanisms:          bson.A{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
		},
		"NotFoundUser": {
			username:            "notfound",
			password:            "something",
			mechanisms:          bson.A{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
			userNotFound:        true,
			errorMessage:        "Authentication failed.",
			topologyError:       true,
		},
		"BadPassword": {
			username:            "baduser",
			password:            "something",
			mechanisms:          bson.A{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
			wrongPassword:       true,
			topologyError:       true,
			errorMessage:        "Authentication failed.",
		},
		"MechanismNotSet": {
			username:            "user_mechanism_not_set",
			password:            "password",
			mechanisms:          bson.A{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-1",
			topologyError:       true,
			errorMessage:        "Unable to use SCRAM-SHA-1 based authentication for user without any SCRAM-SHA-1 credentials registered",
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testtb.TB = tt

			if !tc.userNotFound {
				mechanisms := tc.mechanisms
				if mechanisms == nil {
					mechanisms = bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}
				}

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
					{"mechanisms", mechanisms},
				}

				err := db.RunCommand(ctx, createPayload).Err()
				require.NoErrorf(t, err, "cannot create user")
			}

			if tc.updatePassword != "" {
				updatePayload := bson.D{
					{"updateUser", tc.username},
					{"pwd", tc.updatePassword},
				}

				err := db.RunCommand(ctx, updatePayload).Err()
				require.NoErrorf(t, err, "cannot update user")
			}

			password := tc.password
			if tc.updatePassword != "" {
				password = tc.updatePassword
			}
			if tc.wrongPassword {
				password = "wrongpassword"
			}

			connectionMechanism := tc.connectionMechanism

			credential := options.Credential{
				AuthMechanism: connectionMechanism,
				AuthSource:    db.Name(),
				Username:      tc.username,
				Password:      password,
			}

			opts := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)

			client, err := mongo.Connect(ctx, opts)
			require.NoError(t, err)

			// Ping to force connection to be established and tested.
			err = client.Ping(ctx, nil)

			if tc.errorMessage != "" {
				require.Contains(t, err.Error(), tc.errorMessage, "expected error message")
			}

			if tc.topologyError {
				var ce topology.ConnectionError
				require.ErrorAs(t, err, &ce, "expected a connection error")
				return
			}

			require.NoError(t, err)

			connCollection := client.Database(db.Name()).Collection(collection.Name())

			require.NotNil(t, connCollection, "cannot get collection")

			r, err := connCollection.InsertOne(ctx, bson.D{{"ping", "pong"}})
			require.NoError(t, err, "cannot insert document")
			id := r.InsertedID.(primitive.ObjectID)
			require.NotEmpty(t, id)

			var result bson.D
			err = connCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&result)
			require.NoError(t, err, "cannot find document")
			assert.Equal(t, bson.D{{"_id", id}, {"ping", "pong"}}, result)

			require.NoError(t, client.Disconnect(context.Background()))
		})
	}
}

func TestAuthenticationOnAuthenticatedConnection(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, db := s.Ctx, s.Collection.Database()
	username, password, mechanism := "testuser", "testpass", "SCRAM-SHA-256"

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

	actualAuth, err := integration.ConvertDocument(t, res).Get("authInfo")
	require.NoError(t, err)

	actualUsersV, err := actualAuth.(*types.Document).Get("authenticatedUsers")
	require.NoError(t, err)

	actualUsers := actualUsersV.(*types.Array)
	require.Equal(t, 1, actualUsers.Len())

	actualUser := must.NotFail(actualUsers.Get(0)).(*types.Document)
	user, err := actualUser.Get("user")
	require.NoError(t, err)
	require.Equal(t, username, user)

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

	actualAuth, err = integration.ConvertDocument(t, res).Get("authInfo")
	require.NoError(t, err)

	actualUsersV, err = actualAuth.(*types.Document).Get("authenticatedUsers")
	require.NoError(t, err)

	actualUsers = actualUsersV.(*types.Array)
	require.Equal(t, 1, actualUsers.Len())

	actualUser = must.NotFail(actualUsers.Get(0)).(*types.Document)
	user, err = actualUser.Get("user")
	require.NoError(t, err)
	require.Equal(t, username, user)

	err = db.RunCommand(ctx, saslStart).Decode(&res)
	require.NoError(t, err)
}

func TestAuthenticationAuthSource(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, db := s.Ctx, s.Collection.Database()

	for name, tc := range map[string]struct {
		baseURI          string // host and port are replaced and additional query values from MongoDBURI added
		authenticationDB string
	}{
		"Admin": {
			baseURI:          "mongodb://adminuser:adminpass@127.0.0.1/",
			authenticationDB: "admin",
		},
		"DefaultAuthDB": {
			baseURI:          "mongodb://user1:pass1@127.0.0.1/TestAuthenticationAuthSource",
			authenticationDB: t.Name(),
		},
		"AuthSource": {
			baseURI:          "mongodb://user2:pass2@127.0.0.1/?authSource=TestAuthenticationAuthSource",
			authenticationDB: t.Name(),
		},
		"AuthSourceWithDB": {
			baseURI:          "mongodb://user3:pass3@127.0.0.1/XXX?authSource=TestAuthenticationAuthSource",
			authenticationDB: t.Name(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			roles := bson.A{"readWrite"}
			if !setup.IsMongoDB(t) {
				// TODO https://github.com/FerretDB/FerretDB/issues/3974
				roles = bson.A{}
			}

			testURI, err := url.Parse(tc.baseURI)
			require.NoError(t, err)

			username := testURI.User.Username()
			password, _ := testURI.User.Password()

			err = db.Client().Database(tc.authenticationDB).RunCommand(ctx, bson.D{
				{"createUser", username},
				{"roles", roles},
				{"pwd", password},
			}).Err()
			require.NoError(t, err)

			t.Cleanup(func() {
				err = db.Client().Database(tc.authenticationDB).RunCommand(ctx, bson.D{{"dropUser", username}}).Err()
				require.NoError(t, err)
			})

			u, err := url.Parse(s.MongoDBURI)
			require.NoError(t, err)

			u.User = testURI.User
			u.Path = testURI.Path

			// tls related query values are necessary
			q := u.Query()
			for k, v := range testURI.Query() {
				q.Set(k, v[0])
			}

			u.RawQuery = q.Encode()

			opts := options.Client().ApplyURI(u.String())

			t.Log("connecting", u.String())

			client, err := mongo.Connect(ctx, opts)
			require.NoError(t, err)

			t.Cleanup(func() {
				require.NoError(t, client.Disconnect(ctx))
			})

			_, err = client.Database(tc.authenticationDB).Collection("test").Find(ctx, bson.D{})
			require.NoError(t, err)
		})
	}
}

func TestAuthenticationDifferentDatabase(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, db := s.Ctx, s.Collection.Database()
	user, db1, pass1, db2, pass2 := "user", t.Name()+"db1", "pass1", t.Name()+"db2", "pass2"

	for dbName, pass := range map[string]string{
		db1: pass1,
		db2: pass2,
	} {
		roles := bson.A{"readWrite"}
		if !setup.IsMongoDB(t) {
			// TODO https://github.com/FerretDB/FerretDB/issues/3974
			roles = bson.A{}
		}

		err := db.Client().Database(dbName).RunCommand(ctx, bson.D{
			{"createUser", user},
			{"roles", roles},
			{"pwd", pass},
		}).Err()
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Client().Database(dbName).RunCommand(ctx, bson.D{{"dropAllUsersFromDatabase", int32(1)}}).Err()
			require.NoError(t, err)
		})
	}

	for name, tc := range map[string]struct {
		username   string
		password   string
		authSource string

		errMsg string
	}{
		"CorrectDatabase": {
			// user:password1 successfully authenticate in db1
			username:   user,
			password:   pass1,
			authSource: db1,
		},
		"WrongDatabase": {
			// user:password2 failed to authenticate in db1
			username:   user,
			password:   pass2,
			authSource: db1,
			errMsg:     "Authentication failed.",
		},
		"NonExistingDatabase": {
			username:   user,
			password:   pass1,
			authSource: "non-existing-db",
			errMsg:     "Authentication failed.",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			credential := options.Credential{
				AuthSource: tc.authSource,
				Username:   tc.username,
				Password:   tc.password,
			}

			client, err := mongo.Connect(ctx, options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential))
			require.NoError(t, err)

			t.Cleanup(func() {
				require.NoError(t, client.Disconnect(ctx))
			})

			_, err = client.Database(tc.authSource).Collection("test").Find(ctx, bson.D{})
			if tc.errMsg != "" {
				require.ErrorContains(t, err, tc.errMsg)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAuthenticationPLAIN(t *testing.T) {
	t.Parallel()

	opts := &setup.SetupOpts{BackendOptions: &setup.BackendOpts{DisableNewAuth: true}}
	s := setup.SetupWithOpts(t, opts)
	ctx, cName, db := s.Ctx, s.Collection.Name(), s.Collection.Database()

	testCases := map[string]struct {
		username  string
		password  string
		mechanism string

		errMessage     string
		failsForSQLite string
	}{
		"Success": {
			username:  "username", // when using the PLAIN mechanism we must use user "username"
			password:  "password",
			mechanism: "PLAIN",
		},
		"BadPassword": {
			// succeeds due to `POSTGRES_HOST_AUTH_METHOD=trust` for PostgreSQL, and no backend authentication for SQLite
			username:  "username",
			password:  "wrong",
			mechanism: "PLAIN",
		},
		"NonExistentUser": {
			username:       "not-found-user",
			password:       "something",
			mechanism:      "PLAIN",
			errMessage:     `role "not-found-user" does not exist`,
			failsForSQLite: "backend authentication is not supported by SQLite",
		},
		"NonPLAINUser": {
			username:   "scram-user",
			password:   "correct",
			mechanism:  "SCRAM-SHA-256",
			errMessage: `unable to authenticate using mechanism "SCRAM-SHA-256": (AuthenticationFailed) Unsupported authentication mechanism "SCRAM-SHA-256"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			t := setup.FailsForMongoDB(tt, "PLAIN mechanism is not supported by MongoDB")

			if setup.IsSQLite(t) && tc.failsForSQLite != "" {
				t = testfail.Expected(tt, tc.failsForSQLite)
			}

			credential := options.Credential{
				AuthMechanism: tc.mechanism,
				AuthSource:    db.Name(),
				Username:      tc.username,
				Password:      tc.password,
			}

			opts := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)

			client, err := mongo.Connect(ctx, opts)
			require.NoError(t, err)

			t.Cleanup(func() {
				require.NoError(t, client.Disconnect(ctx))
			})

			c := client.Database(db.Name()).Collection(cName)
			_, err = c.InsertOne(ctx, bson.D{{"ping", "pong"}})

			if tc.errMessage != "" {
				require.ErrorContains(t, err, tc.errMessage)
				return
			}

			require.NoError(t, err, "cannot insert document")
		})
	}
}
