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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestMostCommandsAreCaseSensitive(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t)
	db := collection.Database()

	res := db.RunCommand(ctx, bson.D{{"listcollections", 1}})
	err := res.Err()
	require.Error(t, err)
	AssertEqualError(t, mongo.CommandError{Code: 59, Name: "CommandNotFound", Message: `no such command: 'listcollections'`}, err)

	res = db.RunCommand(ctx, bson.D{{"listCollections", 1}})
	assert.NoError(t, res.Err())

	// special cases from the old `mongo` shell
	res = db.RunCommand(ctx, bson.D{{"ismaster", 1}})
	assert.NoError(t, res.Err())
	res = db.RunCommand(ctx, bson.D{{"isMaster", 1}})
	assert.NoError(t, res.Err())
	res = db.RunCommand(ctx, bson.D{{"buildinfo", 1}})
	assert.NoError(t, res.Err())
	res = db.RunCommand(ctx, bson.D{{"buildInfo", 1}})
	assert.NoError(t, res.Err())
}

func TestFindNothing(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

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
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	var docs []bson.D
	for _, provider := range providers {
		docs = append(docs, provider.Docs()...)
	}

	for _, expected := range docs {
		expected := expected
		id, ok := expected.Map()["_id"]
		require.True(t, ok)

		t.Run(fmt.Sprint(id), func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, bson.D{{"_id", id}}, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			require.Len(t, actual, 1)
			AssertEqualDocuments(t, expected, actual[0])
		})
	}
}

//nolint:paralleltest // we test a global list of databases
func TestFindCommentMethod(t *testing.T) {
	setup.SkipForTigris(t)

	ctx, collection := setup.Setup(t, shareddata.Scalars)
	name := collection.Database().Name()
	databaseNames, err := collection.Database().Client().ListDatabaseNames(ctx, bson.D{})
	require.NoError(t, err)
	comment := "*/ 1; DROP SCHEMA " + name + " CASCADE -- "

	var doc bson.D
	opts := options.FindOne().SetComment(comment)
	err = collection.FindOne(ctx, bson.D{{"_id", "string"}}, opts).Decode(&doc)
	require.NoError(t, err)
	assert.Contains(t, databaseNames, name)
}

//nolint:paralleltest // we test a global list of databases
func TestFindCommentQuery(t *testing.T) {
	setup.SkipForTigris(t)

	ctx, collection := setup.Setup(t, shareddata.Scalars)
	name := collection.Database().Name()
	databaseNames, err := collection.Database().Client().ListDatabaseNames(ctx, bson.D{})
	require.NoError(t, err)
	comment := "*/ 1; DROP SCHEMA " + name + " CASCADE -- "

	var doc bson.D
	err = collection.FindOne(ctx, bson.M{"_id": "string", "$comment": comment}).Decode(&doc)
	require.NoError(t, err)
	assert.Contains(t, databaseNames, name)
}

func TestUpdateCommentMethod(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)
	name := collection.Database().Name()
	databaseNames, err := collection.Database().Client().ListDatabaseNames(ctx, bson.D{})
	require.NoError(t, err)

	comment := "*/ 1; DROP SCHEMA " + name + " CASCADE -- "
	filter := bson.D{{"_id", "string"}}
	update := bson.D{{"$set", bson.D{{"v", "bar"}}}}

	opts := options.Update().SetComment(comment)
	res, err := collection.UpdateOne(ctx, filter, update, opts)
	require.NoError(t, err)

	expected := &mongo.UpdateResult{
		MatchedCount:  1,
		ModifiedCount: 1,
	}

	assert.Contains(t, databaseNames, name)
	assert.Equal(t, expected, res)
}

func TestUpdateCommentQuery(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)
	name := collection.Database().Name()
	databaseNames, err := collection.Database().Client().ListDatabaseNames(ctx, bson.D{})
	require.NoError(t, err)

	comment := "*/ 1; DROP SCHEMA " + name + " CASCADE -- "

	res, err := collection.UpdateOne(ctx, bson.M{"_id": "string", "$comment": comment}, bson.M{"$set": bson.M{"v": "bar"}})
	require.NoError(t, err)

	expected := &mongo.UpdateResult{
		MatchedCount:  1,
		ModifiedCount: 1,
	}

	assert.Contains(t, databaseNames, name)
	assert.Equal(t, expected, res)
}

func TestCollectionName(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	t.Run("Err", func(t *testing.T) {
		ctx, collection := setup.Setup(t)

		collectionName300 := strings.Repeat("aB", 150)
		cases := map[string]struct {
			collection string
			err        *mongo.CommandError
			alt        string
		}{
			"TooLongForBothDBs": {
				collection: collectionName300,
				err: &mongo.CommandError{
					Name: "InvalidNamespace",
					Code: 73,
					Message: fmt.Sprintf(
						"Fully qualified namespace is too long. Namespace: testcollectionname_err.%s Max: 255",
						collectionName300,
					),
				},
				alt: fmt.Sprintf("Invalid collection name: 'testcollectionname_err.%s'", collectionName300),
			},
			"WithADollarSign": {
				collection: "collection_name_with_a-$",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid collection name: collection_name_with_a-$`,
				},
				alt: `Invalid collection name: 'testcollectionname_err.collection_name_with_a-$'`,
			},
			"Empty": {
				collection: "",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: "Invalid namespace specified 'testcollectionname_err.'",
				},
				alt: "Invalid collection name: 'testcollectionname_err.'",
			},
		}

		for name, tc := range cases {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				err := collection.Database().CreateCollection(ctx, tc.collection)
				AssertEqualAltError(t, *tc.err, tc.alt, err)
			})
		}
	})

	t.Run("Ok", func(t *testing.T) {
		ctx, collection := setup.Setup(t)

		longCollectionName := strings.Repeat("a", 100)
		err := collection.Database().CreateCollection(ctx, longCollectionName)
		require.NoError(t, err)

		names, err := collection.Database().ListCollectionNames(ctx, bson.D{})
		require.NoError(t, err)

		assert.Contains(t, names, longCollectionName)
	})
}

func TestDatabaseName(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	t.Run("Err", func(t *testing.T) {
		ctx, collection := setup.Setup(t)

		dbName64 := strings.Repeat("a", 64)

		cases := map[string]struct {
			db  string
			err *mongo.CommandError
			alt string
		}{
			"TooLongForBothDBs": {
				db: dbName64,
				err: &mongo.CommandError{
					Name: "InvalidNamespace",
					Code: 73,
					Message: fmt.Sprintf(
						"Invalid namespace specified '%s.%s'",
						dbName64,
						"TestDatabaseName_Err_TooLongForBothDBs",
					),
				},
				alt: fmt.Sprintf("Invalid namespace: %s.%s", dbName64, "TestDatabaseName_Err_TooLongForBothDBs"),
			},
			"WithADollarSign": {
				db: "name_with_a-$",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace: name_with_a-$.TestDatabaseName_Err_WithADollarSign`,
				},
			},
		}

		for name, tc := range cases {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				// there is no explicit command to create database, so create collection instead
				err := collection.Database().Client().Database(tc.db).CreateCollection(ctx, testutil.CollectionName(t))
				AssertEqualAltError(t, *tc.err, tc.alt, err)
			})
		}
	})

	t.Run("Empty", func(t *testing.T) {
		ctx, collection := setup.Setup(t)

		err := collection.Database().Client().Database("").CreateCollection(ctx, testutil.CollectionName(t))
		expectedErr := driver.InvalidOperationError(driver.InvalidOperationError{MissingField: "Database"})
		assert.Equal(t, expectedErr, err)
	})

	t.Run("63ok", func(t *testing.T) {
		ctx, collection := setup.Setup(t)

		dbName63 := strings.Repeat("a", 63)
		err := collection.Database().Client().Database(dbName63).CreateCollection(ctx, testutil.CollectionName(t))
		require.NoError(t, err)
		collection.Database().Client().Database(dbName63).Drop(ctx)
	})
}
