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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestDropAllUsersFromDatabase(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{BackendOptions: &setup.BackendOpts{EnableNewAuth: true}})
	ctx := s.Ctx
	db := s.Collection.Database()
	client := db.Client()

	quantity := 5 // Add some users to the database.
	for i := 1; i <= quantity; i++ {
		err := db.RunCommand(ctx, bson.D{
			{"createUser", fmt.Sprintf("user_%d", i)},
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

func assertDropAllUsersFromDatabase(t *testing.T, ctx context.Context, db *mongo.Database, quantity int) {
	t.Helper()

	var res bson.D
	err := db.RunCommand(ctx, bson.D{
		{"dropAllUsersFromDatabase", 1},
	}).Decode(&res)

	require.NoError(t, err)

	actual := integration.ConvertDocument(t, res)
	actual.Remove("$clusterTime")
	actual.Remove("operationTime")

	expected := must.NotFail(types.NewDocument("n", int32(quantity), "ok", float64(1)))
	testutil.AssertEqual(t, expected, actual)

	var usersInfo bson.D
	err = db.RunCommand(ctx, bson.D{{"usersInfo", 1}}).Decode(&usersInfo)
	assert.NoError(t, err)

	expectedUsersInfo := must.NotFail(types.NewDocument(
		"users", new(types.Array),
		"ok", float64(1),
	))
	actualUsersInfo := integration.ConvertDocument(t, usersInfo)
	actualUsersInfo.Remove("$clusterTime")
	actualUsersInfo.Remove("operationTime")
	testutil.AssertEqual(t, expectedUsersInfo, actualUsersInfo)
}
