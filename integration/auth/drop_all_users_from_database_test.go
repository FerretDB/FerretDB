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

package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/integration"
	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestDropAllUsersFromDatabaseCommand(tt *testing.T) {
	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/864")

	tt.Parallel()

	s := setup.SetupWithOpts(tt, nil)
	ctx := s.Ctx
	db := s.Collection.Database()
	client := db.Client()

	quantity := 5 // Add some users to the database.
	for i := 1; i <= quantity; i++ {
		username := fmt.Sprintf("user_%d", i)

		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
		_ = db.RunCommand(ctx, bson.D{{"dropUser", username}})

		err := db.RunCommand(ctx, bson.D{
			{"createUser", username},
			{"roles", bson.A{}},
			{"pwd", "password"},
		}).Err()
		require.NoError(t, err)
	}

	// Dropping all users from another database shouldn't influence on the number of users remaining on the current database.
	// So this call should remove zero users as the database doesn't exist. The next one, "quantity" users.
	assertDropAllUsersFromDatabase(t, ctx, client.Database(t.Name()+"_another_database"), 0)

	assertDropAllUsersFromDatabase(t, ctx, db, quantity)

	// Run for the second time to check if it still succeeds when there aren't any users remaining,
	// instead of returning an error.
	assertDropAllUsersFromDatabase(t, ctx, db, 0)
}

func assertDropAllUsersFromDatabase(t testing.TB, ctx context.Context, db *mongo.Database, quantity int) {
	t.Helper()

	var res bson.D
	err := db.RunCommand(ctx, bson.D{
		{"dropAllUsersFromDatabase", 1},
	}).Decode(&res)

	require.NoError(t, err)

	expected := bson.D{{"n", int32(quantity)}, {"ok", float64(1)}}
	integration.AssertEqualDocuments(t, expected, res)

	var usersInfo bson.D
	err = db.RunCommand(ctx, bson.D{{"usersInfo", 1}}).Decode(&usersInfo)
	assert.NoError(t, err)

	expectedUsersInfo := bson.D{
		{"users", bson.A{}},
		{"ok", float64(1)},
	}

	integration.AssertEqualDocuments(t, expectedUsersInfo, usersInfo)
}
