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

package sessions

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestStartSessionCommand(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/1554")

	s := setup.SetupWithOpts(t, nil)
	ctx := s.Ctx
	db := s.Collection.Database()

	optsNoAuth := options.Client().ApplyURI(s.MongoDBURI)
	clientNoAuth, err := mongo.Connect(ctx, optsNoAuth)
	require.NoError(t, err)
	require.NoError(t, clientNoAuth.Ping(ctx, nil))

	username1 := "sessionUser"
	username2 := "sessionUser2"

	err = db.RunCommand(ctx, bson.D{{"dropAllUsersFromDatabase", 1}}).Err()
	require.NoError(t, err)

	err = db.RunCommand(ctx, bson.D{
		{"createUser", username1},
		{"roles", bson.A{}},
		{"pwd", "password"},
		{"mechanisms", bson.A{"SCRAM-SHA-256"}},
	}).Err()
	require.NoError(t, err)

	err = db.RunCommand(ctx, bson.D{
		{"createUser", username2},
		{"roles", bson.A{}},
		{"pwd", "password2"},
		{"mechanisms", bson.A{"SCRAM-SHA-256"}},
	}).Err()
	require.NoError(t, err)

	credential := options.Credential{ // as created in setup.setupUser
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    db.Name(),
		Username:      username1,
		Password:      "password",
	}

	optsAuth := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)
	clientAuth, err := mongo.Connect(ctx, optsAuth)
	require.NoError(t, err)
	require.NoError(t, clientAuth.Ping(ctx, nil))

	credential = options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    db.Name(),
		Username:      username2,
		Password:      "password2",
	}
	optsAuth2 := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)
	clientAuth2, err := mongo.Connect(ctx, optsAuth2)
	require.NoError(t, err)
	require.NoError(t, clientAuth2.Ping(ctx, nil))

	for name, tc := range map[string]struct {
		clientToStartSession *mongo.Client
		clientToKillSession  *mongo.Client
		shouldBeAvailable    bool // whether the session should be available after an attempt to kill it
	}{
		"noAuth": {
			clientToStartSession: clientNoAuth,
			clientToKillSession:  clientNoAuth,
			shouldBeAvailable:    false,
		},
		/*	"auth": {
				clientToStartSession: clientAuth,
				clientToKillSession:  clientAuth,
				shouldBeAvailable:    false,
			},
			"otherClientEndsAuth": {
				clientToStartSession: clientAuth,
				clientToKillSession:  clientAuth2,
				shouldBeAvailable:    true,
			},
			"authEndsNoAuth": {
				clientToStartSession: clientNoAuth,
				clientToKillSession:  clientAuth,
				shouldBeAvailable:    true,
			},
			"noAuthEndsAuth": {
				clientToStartSession: clientAuth,
				clientToKillSession:  clientNoAuth,
				shouldBeAvailable:    true,
			},*/
	} {
		name, tc := name, tc

		tt.Run(name, func(t *testing.T) {
			t.Parallel()

			sessionID := startSession(t, ctx, tc.clientToStartSession.Database(db.Name()))

			var res bson.D
			killSessionsCommand := bson.D{{"killSessions", bson.A{bson.D{{"id", sessionID}}}}}
			err := tc.clientToKillSession.Database(db.Name()).RunCommand(ctx, killSessionsCommand).Decode(&res)
			require.NoError(t, err)

			doc := integration.ConvertDocument(t, res)
			assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

			id := types.Binary{
				Subtype: types.BinarySubtype(sessionID.Subtype),
				B:       sessionID.Data,
			}

			if tc.shouldBeAvailable {
				assert.True(t, localSessionExists(t, ctx, tc.clientToStartSession.Database(db.Name()), id))
			} else {
				assert.False(t, localSessionExists(t, ctx, tc.clientToStartSession.Database(db.Name()), id))
			}
		})
	}
}

func startSession(t *testing.T, ctx context.Context, db *mongo.Database) *primitive.Binary {
	var res bson.D
	err := db.RunCommand(ctx, bson.D{{"startSession", 1}}).Decode(&res)
	require.NoError(t, err)

	doc := integration.ConvertDocument(t, res)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	idDoc := must.NotFail(doc.Get("id")).(*types.Document)
	id := must.NotFail(idDoc.Get("id")).(types.Binary)
	assert.Len(t, id.B, 16)
	assert.Equal(t, types.BinaryUUID, id.Subtype)
	assert.Equal(t, int32(30), must.NotFail(doc.Get("timeoutMinutes")))

	return &idForFilter
}
