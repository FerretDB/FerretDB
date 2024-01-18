// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestAuthentication(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{})
	ctx := s.Ctx
	collection := s.Collection
	db := collection.Database()

	testCases := map[string]struct { //nolint:vet // for readability
		username       string
		password       string
		updatePassword string // if true, the password will be updated to this one after the user is created.
		mechanism      string

		wrongPassword bool
	}{
		"Success": {
			username: "common",
			password: "password",
		},
		"Updated": {
			username:       "updated",
			password:       "pass123",
			updatePassword: "somethingelse",
		},
		"Scramsha256": {
			username:  "scramsha256",
			password:  "password",
			mechanism: "SCRAM-SHA-256",
		},
		"Scramsha256updated": {
			username:       "scramsha256updated",
			password:       "pass123",
			updatePassword: "anotherpassword",
			mechanism:      "SCRAM-SHA-256",
		},
		"BadPassword": {
			username:      "baduser",
			password:      "something",
			mechanism:     "SCRAM-SHA-256",
			wrongPassword: true,
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.mechanism == "PLAIN" {
				setup.SkipForMongoDB(t, "PLAIN credentials are only supported via LDAP (PLAIN) by MongoDB Enterprise")
			}

			createPayload := bson.D{
				{"createUser", tc.username},
				{"roles", bson.A{}},
				{"pwd", tc.password},
			}

			switch tc.mechanism {
			case "": // Use default mechanism for MongoDB and SCRAM-SHA-256 for FerretDB as SHA-1 won't be supported as it's deprecated.
				if !setup.IsMongoDB(t) {
					createPayload = append(createPayload, bson.E{"mechanisms", bson.A{"SCRAM-SHA-256"}})
				}
			case "SCRAM-SHA-256", "PLAIN":
				createPayload = append(createPayload, bson.E{"mechanisms", bson.A{tc.mechanism}})
			default:
				t.Fatalf("unimplemented mechanism %s", tc.mechanism)
			}

			err := db.RunCommand(ctx, createPayload).Err()
			require.NoErrorf(t, err, "cannot create user")

			if tc.updatePassword != "" {
				updatePayload := bson.D{
					{"updateUser", tc.username},
					{"pwd", tc.updatePassword},
				}

				err = db.RunCommand(ctx, updatePayload).Err()
				require.NoErrorf(t, err, "cannot update user")
			}

			serverAPI := options.ServerAPI(options.ServerAPIVersion1)

			password := tc.password
			if tc.updatePassword != "" {
				password = tc.updatePassword
			}
			if tc.wrongPassword {
				password = "wrongpassword"
			}

			credential := options.Credential{
				AuthMechanism: tc.mechanism,
				AuthSource:    db.Name(),
				Username:      tc.username,
				Password:      password,
			}

			opts := options.Client().ApplyURI(s.MongoDBURI).SetServerAPIOptions(serverAPI).SetAuth(credential)

			client, err := mongo.Connect(ctx, opts)
			require.NoError(t, err, "cannot connect to MongoDB")

			// Ping to force connection to be established and tested.
			err = client.Ping(ctx, nil)

			if tc.wrongPassword {
				var ce topology.ConnectionError
				require.ErrorAs(t, err, &ce, "expected a connection error")
				return
			}

			require.NoError(t, err, "cannot ping MongoDB")
			require.NoError(t, client.Disconnect(context.Background()))
		})
	}
}
