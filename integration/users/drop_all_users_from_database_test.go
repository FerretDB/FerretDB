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

	if setup.IsMongoDB(t) {
		assert.NoError(t, collection.Database().RunCommand(ctx, bson.D{
			{"dropAllUsersFromDatabase", 1},
		}).Err())
	} else {
		// Erase any previously saved user in the database.
		_, err := users.DeleteMany(ctx, bson.D{{"db", db.Name()}})
		assert.NoError(t, err)
	}

	quantity := 5
	for i := 1; i <= quantity; i++ {
		err := db.RunCommand(ctx, bson.D{
			{"createUser", fmt.Sprintf("user_%d", i)},
			{"roles", bson.A{}},
			{"pwd", "password"},
		}).Err()
		assert.NoError(t, err)
	}

	assertDropAllUsersFromDatabase(t, ctx, db, users, quantity)

	// FIXME: calling assertDropAllUsersFromDatabase a second time with quantity = 0
	// for FerretDB is triggering a "socket was unexpectedly closed: EOF" error for some reason.
	// assertDropAllUsersFromDatabase(t, ctx, db, users, 0)
}

func assertDropAllUsersFromDatabase(t *testing.T, ctx context.Context, db *mongo.Database, users *mongo.Collection, quantity int) {
	var res bson.D
	err := db.RunCommand(ctx, bson.D{
		{"dropAllUsersFromDatabase", 1},
	}).Decode(&res)

	assert.NoError(t, err)

	actual := integration.ConvertDocument(t, res)
	actual.Remove("$clusterTime")
	actual.Remove("operationTime")

	expected := must.NotFail(types.NewDocument("n", int32(quantity), "ok", float64(1)))
	testutil.AssertEqual(t, expected, actual)

	assert.Equal(t, mongo.ErrNoDocuments, users.FindOne(ctx, bson.D{{"db", db.Name()}}).Err())
}
