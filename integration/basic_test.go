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

	var doc bson.D

	// FindOne sets limit parameter to 1, Find leaves it unset.
	err := collection.FindOne(ctx, bson.D{}).Decode(&doc)
	require.Equal(t, mongo.ErrNoDocuments, err)
	assert.Equal(t, bson.D(nil), doc)
}

func TestInsertFind(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	for _, expected := range shareddata.Docs(providers...) {
		expected := expected.(bson.D)
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

func TestDeleteCommentMethod(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	name := collection.Database().Name()
	databaseNames, err := collection.Database().Client().ListDatabaseNames(ctx, bson.D{})
	require.NoError(t, err)

	comment := "*/ 1; DROP SCHEMA " + name + " CASCADE -- "
	filter := bson.D{{"_id", "string"}}

	opts := options.Delete().SetComment(comment)
	res, err := collection.DeleteOne(ctx, filter, opts)
	require.NoError(t, err)

	expected := &mongo.DeleteResult{
		DeletedCount: 1,
	}

	assert.Contains(t, databaseNames, name)
	assert.Equal(t, expected, res)
}

func TestDeleteCommentQuery(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	name := collection.Database().Name()
	databaseNames, err := collection.Database().Client().ListDatabaseNames(ctx, bson.D{})
	require.NoError(t, err)

	comment := "*/ 1; DROP SCHEMA " + name + " CASCADE -- "

	res, err := collection.DeleteOne(ctx, bson.M{"_id": "string", "$comment": comment})
	require.NoError(t, err)

	expected := &mongo.DeleteResult{
		DeletedCount: 1,
	}

	assert.Contains(t, databaseNames, name)
	assert.Equal(t, expected, res)
}

func TestEmptyKey(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Tigris field name cannot be empty")

	t.Parallel()
	ctx, collection := setup.Setup(t)

	doc := bson.D{{"_id", "empty-key"}, {"", "foo"}}

	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)

	res, err := collection.Find(ctx, bson.D{{"", "foo"}})
	require.NoError(t, err)

	var actual []bson.D
	require.NoError(t, res.All(ctx, &actual))

	expected := []bson.D{doc}

	assert.Equal(t, expected, actual)
}

func TestFindAndModifyCommentMethod(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	name := collection.Database().Name()
	databaseNames, err := collection.Database().Client().ListDatabaseNames(ctx, bson.D{})
	require.NoError(t, err)

	comment := "*/ 1; DROP SCHEMA " + name + " CASCADE -- "
	filter := bson.D{{"_id", "string"}}

	opts := options.Delete().SetComment(comment)
	res, err := collection.DeleteOne(ctx, filter, opts)
	require.NoError(t, err)

	expected := &mongo.DeleteResult{
		DeletedCount: 1,
	}

	assert.Contains(t, databaseNames, name)
	assert.Equal(t, expected, res)
}

func TestFindAndModifyCommentQuery(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	name := collection.Database().Name()
	databaseNames, err := collection.Database().Client().ListDatabaseNames(ctx, bson.D{})
	require.NoError(t, err)

	comment := "*/ 1; DROP SCHEMA " + name + " CASCADE -- "
	request := bson.D{
		{"findAndModify", collection.Name()},
		{"query", bson.D{{"_id", "string"}, {"$comment", comment}}},
		{"update", bson.D{{"$set", bson.D{{"v", "bar"}}}}},
	}

	expectedLastErrObj := bson.D{
		{"n", int32(1)},
		{"updatedExisting", true},
	}

	var actual bson.D
	err = collection.Database().RunCommand(ctx, request).Decode(&actual)
	require.NoError(t, err)

	lastErrObj, ok := actual.Map()["lastErrorObject"].(bson.D)
	if !ok {
		t.Fatal(actual)
	}

	assert.Contains(t, databaseNames, name)
	AssertEqualDocuments(t, expectedLastErrObj, lastErrObj)
}

func TestCollectionName(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	ctx, collection := setup.Setup(t)

	collectionName300 := strings.Repeat("aB", 150)
	collectionName235 := strings.Repeat("a", 235)

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
					"Fully qualified namespace is too long. Namespace: testcollectionname.%s Max: 255",
					collectionName300,
				),
			},
			alt: fmt.Sprintf("Invalid collection name: 'testcollectionname.%s'", collectionName300),
		},
		"LongEnough": {
			collection: collectionName235,
		},
		"Short": {
			collection: "a",
		},
		"WithADollarSign": {
			collection: "collection_name_with_a-$",
			err: &mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: `Invalid collection name: collection_name_with_a-$`,
			},
			alt: `Invalid collection name: 'testcollectionname.collection_name_with_a-$'`,
		},
		"WithADash": {
			collection: "collection_name_with_a-",
		},
		"WithADashAtBeginning": {
			collection: "-collection_name",
		},
		"Empty": {
			collection: "",
			err: &mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: "Invalid namespace specified 'testcollectionname.'",
			},
			alt: "Invalid collection name: 'testcollectionname.'",
		},
		"Null": {
			collection: "\x00",
			err: &mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: "namespaces cannot have embedded null characters",
			},
			alt: "Invalid collection name: 'testcollectionname.\x00'",
		},
		"Dot": {
			collection: "collection.name",
		},
		"Space": {
			collection: " ",
		},
		"NonLatin": {
			collection: "コレクション",
		},
		"Number": {
			collection: "1",
		},
		"SpecialCharacters": {
			collection: "+-/*<>=~!@#%^&|`?()[],;:. ",
		},
		"Capital": {
			collection: "A",
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			err := collection.Database().CreateCollection(ctx, tc.collection)
			if tc.err != nil {
				AssertEqualAltError(t, *tc.err, tc.alt, err)
				return
			}

			assert.NoError(t, err)

			// check collection name is in the list.
			names, err := collection.Database().ListCollectionNames(ctx, bson.D{})
			require.NoError(t, err)
			assert.Contains(t, names, tc.collection)

			newCollection := collection.Database().Collection(tc.collection)

			// document can be inserted and found in the collection.
			doc := bson.D{{"_id", "item"}}
			_, err = newCollection.InsertOne(ctx, doc)
			require.NoError(t, err)

			res := newCollection.FindOne(ctx, doc)
			require.NoError(t, res.Err())
		})
	}
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
						"TestDatabaseName-Err",
					),
				},
				alt: fmt.Sprintf("Invalid namespace: %s.%s", dbName64, "TestDatabaseName-Err"),
			},
			"WithADollarSign": {
				db: "name_with_a-$",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace: name_with_a-$.TestDatabaseName-Err`,
				},
			},
		}

		for name, tc := range cases {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				// there is no explicit command to create database, so create collection instead
				err := collection.Database().Client().Database(tc.db).CreateCollection(ctx, collection.Name())
				AssertEqualAltError(t, *tc.err, tc.alt, err)
			})
		}
	})

	t.Run("Empty", func(t *testing.T) {
		ctx, collection := setup.Setup(t)

		err := collection.Database().Client().Database("").CreateCollection(ctx, collection.Name())
		expectedErr := driver.InvalidOperationError(driver.InvalidOperationError{MissingField: "Database"})
		assert.Equal(t, expectedErr, err)
	})

	t.Run("63ok", func(t *testing.T) {
		ctx, collection := setup.Setup(t)

		dbName63 := strings.Repeat("a", 63)
		err := collection.Database().Client().Database(dbName63).CreateCollection(ctx, collection.Name())
		require.NoError(t, err)
		collection.Database().Client().Database(dbName63).Drop(ctx)
	})
}
