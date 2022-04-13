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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestMostCommandsAreCaseSensitive(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)
	db := collection.Database()

	res := db.RunCommand(ctx, bson.D{{"listcollections", 1}})
	err := res.Err()
	require.Error(t, err)
	assertEqualError(t, mongo.CommandError{Code: 59, Name: "CommandNotFound", Message: `no such command: 'listcollections'`}, err)

	res = db.RunCommand(ctx, bson.D{{"listCollections", 1}})
	assert.NoError(t, res.Err())

	// special case
	res = db.RunCommand(ctx, bson.D{{"ismaster", 1}})
	assert.NoError(t, res.Err())
	res = db.RunCommand(ctx, bson.D{{"isMaster", 1}})
	assert.NoError(t, res.Err())
}

func TestFindNothing(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	cursor, err := collection.Find(ctx, bson.D{})
	require.NoError(t, err)

	var docs []bson.D
	err = cursor.All(ctx, &docs)
	require.NoError(t, err)
	assert.Equal(t, []bson.D(nil), docs)

	var doc bson.D
	err = collection.FindOne(ctx, bson.D{}).Decode(&doc)
	require.Equal(t, mongo.ErrNoDocuments, err)
	assert.Equal(t, bson.D(nil), doc)
}

func TestInsertFind(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	var docs []bson.D
	for _, provider := range providers {
		docs = append(docs, provider.Docs()...)
	}

	for _, expected := range docs {
		expected := expected
		id := expected.Map()["_id"]

		t.Run(fmt.Sprint(id), func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, bson.D{{"_id", id}}, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			require.Len(t, actual, 1)
			assertEqualDocuments(t, expected, actual[0])
		})
	}
}
