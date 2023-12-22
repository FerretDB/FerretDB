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

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestDropAllUsersFromDatabase(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)
	db := collection.Database()
	client := collection.Database().Client()
	users := client.Database("admin").Collection("system.users")

	require.NoError(t, collection.Database().RunCommand(ctx, bson.D{
		{"dropAllUsersFromDatabase", 1},
	}).Err())

	quantity := 5 // Add some users to the database.
	for i := 1; i <= quantity; i++ {
		err := db.RunCommand(ctx, bson.D{
			{"createUser", fmt.Sprintf("user_%d", i)},
			{"roles", bson.A{}},
			{"pwd", "password"},
		}).Err()
		require.NoError(t, err)
	}

	assertDropAllUsersFromDatabase(t, ctx, db, users, quantity)

	// Run for the second time to check if it still succeeds when there aren't any users remaining,
	// instead of returning an error.
	assertDropAllUsersFromDatabase(t, ctx, db, users, 0)
}

func assertDropAllUsersFromDatabase(t *testing.T, ctx context.Context, db *mongo.Database, users *mongo.Collection, quantity int) {
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

	assert.Equal(t, mongo.ErrNoDocuments, users.FindOne(ctx, bson.D{{"db", db.Name()}}).Err())
}
