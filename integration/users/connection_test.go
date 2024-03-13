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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

func TestAuthentication(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{SetupUser: true})
	ctx := s.Ctx
	collection := s.Collection
	db := collection.Database()

	testCases := map[string]struct { //nolint:vet // for readability
		username       string
		password       string
		updatePassword string   // if true, the password will be updated to this one after the user is created.
		mechanisms     []string // mechanisms to use for creating user authentication

		connectionMechanism string // if set, try to establish connection with this mechanism

		userNotFound  bool
		wrongPassword bool
		topologyError bool
		errorMessage  string
	}{
		"Success": {
			username:            "username", // when using the PLAIN mechanism we must use user "username"
			password:            "password",
			mechanisms:          []string{"PLAIN"},
			connectionMechanism: "PLAIN",
		},
		"ScramSHA1": {
			username:            "scramsha1",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-1"},
			connectionMechanism: "SCRAM-SHA-1",
		},
		"ScramSHA256": {
			username:            "scramsha256",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
		},
		"MultipleScramSHA1": {
			username:            "scramsha1multi",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-1", "SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-1",
		},
		"MultipleScramSHA256": {
			username:            "scramsha256multi",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-1", "SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
		},
		"ScramSHA1Updated": {
			username:            "scramsha1updated",
			password:            "pass123",
			updatePassword:      "anotherpassword",
			mechanisms:          []string{"SCRAM-SHA-1"},
			connectionMechanism: "SCRAM-SHA-1",
		},
		"ScramSHA256Updated": {
			username:            "scramsha256updated",
			password:            "pass123",
			updatePassword:      "anotherpassword",
			mechanisms:          []string{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
		},
		"NotFoundUser": {
			username:            "notfound",
			password:            "something",
			mechanisms:          []string{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
			userNotFound:        true,
			errorMessage:        "Authentication failed.",
			topologyError:       true,
		},
		"BadPassword": {
			username:            "baduser",
			password:            "something",
			mechanisms:          []string{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
			wrongPassword:       true,
			topologyError:       true,
			errorMessage:        "Authentication failed.",
		},
		"MechanismNotSet": {
			username:            "user_mechanism_not_set",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-256"},
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
				var (
					// Use default mechanism for MongoDB and SCRAM-SHA-256 for FerretDB as SHA-1 won't be supported as it's deprecated.
					mechanisms bson.A

					hasPlain bool
				)

				if tc.mechanisms == nil {
					if !setup.IsMongoDB(t) {
						mechanisms = append(mechanisms, "SCRAM-SHA-1", "SCRAM-SHA-256")
					}
				} else {
					mechanisms = bson.A{}

					for _, mechanism := range tc.mechanisms {
						switch mechanism {
						case "PLAIN":
							hasPlain = true
							fallthrough
						case "SCRAM-SHA-1", "SCRAM-SHA-256":
							mechanisms = append(mechanisms, mechanism)
						default:
							t.Fatalf("unimplemented mechanism %s", mechanism)
						}
					}
				}

				if hasPlain {
					setup.SkipForMongoDB(t, "PLAIN mechanism is not supported by MongoDB")
				}

				createPayload := bson.D{
					{"createUser", tc.username},
					{"roles", bson.A{}},
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

	s := setup.SetupWithOpts(t, &setup.SetupOpts{SetupUser: true})
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

func TestAuthenticationLocalhostException(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForMongoDB(tt, "MongoDB is not connected via localhost")

	s := setup.SetupWithOpts(t, &setup.SetupOpts{CollectionName: testutil.CollectionName(t)})
	ctx := s.Ctx
	collection := s.Collection
	db := collection.Database()

	opts := options.Client().ApplyURI(s.MongoDBURI)
	clientNoAuth, err := mongo.Connect(ctx, opts)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, clientNoAuth.Disconnect(ctx))
	})

	db = clientNoAuth.Database(db.Name())

	username, password, mechanism := "testuser", "testpass", "SCRAM-SHA-256"
	firstUser := bson.D{
		{"createUser", username},
		{"roles", bson.A{}},
		{"pwd", password},
		{"mechanisms", bson.A{mechanism}},
	}
	err = db.RunCommand(ctx, firstUser).Err()
	require.NoError(t, err, "cannot create user")

	secondUser := bson.D{
		{"createUser", "anotheruser"},
		{"roles", bson.A{}},
		{"pwd", "anotherpass"},
		{"mechanisms", bson.A{mechanism}},
	}
	err = db.RunCommand(ctx, secondUser).Err()
	integration.AssertEqualCommandError(
		t,
		mongo.CommandError{
			Code:    18,
			Name:    "AuthenticationFailed",
			Message: "Authentication failed",
		},
		err,
	)

	credential := options.Credential{
		AuthMechanism: mechanism,
		AuthSource:    db.Name(),
		Username:      username,
		Password:      password,
	}

	opts = options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)

	client, err := mongo.Connect(ctx, opts)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, client.Disconnect(ctx))
	})

	db = client.Database(db.Name())
	err = db.RunCommand(ctx, secondUser).Err()
	require.NoError(t, err, "cannot create user")
}

func TestAuthenticationEnableNewAuthPLAIN(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForMongoDB(tt, "PLAIN mechanism is not supported by MongoDB")

	s := setup.SetupWithOpts(t, &setup.SetupOpts{SetupUser: true})
	ctx, cName, db := s.Ctx, s.Collection.Name(), s.Collection.Database()

	err := db.RunCommand(ctx, bson.D{
		{"createUser", "plain-user"},
		{"roles", bson.A{}},
		{"pwd", "correct"},
		{"mechanisms", bson.A{"PLAIN"}},
	}).Err()
	require.NoError(t, err, "cannot create user")

	err = db.RunCommand(ctx, bson.D{
		{"createUser", "scram-user"},
		{"roles", bson.A{}},
		{"pwd", "correct"},
		{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
	}).Err()
	require.NoError(t, err, "cannot create user")

	testCases := map[string]struct {
		username  string
		password  string
		mechanism string

		err *mongo.CommandError
	}{
		"Success": {
			username:  "plain-user",
			password:  "correct",
			mechanism: "PLAIN",
		},
		"BadPassword": {
			username:  "plain-user",
			password:  "wrong",
			mechanism: "PLAIN",
			err: &mongo.CommandError{
				Code:    18,
				Name:    "AuthenticationFailed",
				Message: "Authentication failed",
			},
		},
		"NonExistentUser": {
			username:  "not-found-user",
			password:  "something",
			mechanism: "PLAIN",
			err: &mongo.CommandError{
				Code:    18,
				Name:    "AuthenticationFailed",
				Message: "Authentication failed",
			},
		},
		"NonPLAINUser": {
			username:  "scram-user",
			password:  "correct",
			mechanism: "PLAIN",
			err: &mongo.CommandError{
				Code:    334,
				Name:    "ErrMechanismUnavailable",
				Message: "Unable to use PLAIN based authentication for user without any PLAIN credentials registered",
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		tt.Run(name, func(t *testing.T) {
			t.Parallel()

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

			if tc.err != nil {
				integration.AssertEqualCommandError(t, *tc.err, err)
				return
			}

			require.NoError(t, err, "cannot insert document")
		})
	}
}
