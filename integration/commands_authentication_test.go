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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCommandsAuthenticationLogout(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, db := s.Ctx, s.Collection.Database()

	opts := options.Client().ApplyURI(s.MongoDBURI)
	client, err := mongo.Connect(ctx, opts)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, client.Disconnect(ctx))
	})

	db = client.Database(db.Name())

	var res bson.D
	err = db.RunCommand(ctx, bson.D{{"logout", 1}}).Decode(&res)
	require.NoError(t, err)

	actual := ConvertDocument(t, res)
	actual.Remove("$clusterTime")
	actual.Remove("operationTime")

	expected := ConvertDocument(t, bson.D{{"ok", float64(1)}})
	testutil.AssertEqual(t, expected, actual)

	// the test user logs out again, it has no effect
	err = db.RunCommand(ctx, bson.D{{"logout", 1}}).Decode(&res)
	require.NoError(t, err)

	actual = ConvertDocument(t, res)
	actual.Remove("$clusterTime")
	actual.Remove("operationTime")

	testutil.AssertEqual(t, expected, actual)
}

func TestCommandsAuthenticationLogoutAuthenticatedUser(t *testing.T) {
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

	actualAuth, err := ConvertDocument(t, res).Get("authInfo")
	require.NoError(t, err)

	actualUsersV, err := actualAuth.(*types.Document).Get("authenticatedUsers")
	require.NoError(t, err)

	actualUsers := actualUsersV.(*types.Array)
	require.Equal(t, 1, actualUsers.Len())

	actualUser := must.NotFail(actualUsers.Get(0)).(*types.Document)
	user, err := actualUser.Get("user")
	require.NoError(t, err)
	require.Equal(t, username, user)

	err = db.RunCommand(ctx, bson.D{{"logout", 1}}).Decode(&res)
	require.NoError(t, err)

	actual := ConvertDocument(t, res)
	actual.Remove("$clusterTime")
	actual.Remove("operationTime")

	expected := ConvertDocument(t, bson.D{{"ok", float64(1)}})
	testutil.AssertEqual(t, expected, actual)

	err = db.RunCommand(ctx, bson.D{{"connectionStatus", 1}}).Decode(&res)
	require.NoError(t, err)

	actualAuth, err = ConvertDocument(t, res).Get("authInfo")
	require.NoError(t, err)

	actualUsersV, err = actualAuth.(*types.Document).Get("authenticatedUsers")
	require.NoError(t, err)
	require.Empty(t, actualUsersV)
}
