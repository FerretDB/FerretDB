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
	ctx, collection := Setup(t)
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
	ctx, collection := Setup(t)

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
	ctx, collection := Setup(t, providers...)

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
	ctx, collection := Setup(t, shareddata.Scalars)
	name := collection.Name()
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
	ctx, collection := Setup(t, shareddata.Scalars)
	name := collection.Name()
	databaseNames, err := collection.Database().Client().ListDatabaseNames(ctx, bson.D{})
	require.NoError(t, err)
	comment := "*/ 1; DROP SCHEMA " + name + " CASCADE -- "

	var doc bson.D
	err = collection.FindOne(ctx, bson.M{"_id": "string", "$comment": comment}).Decode(&doc)
	require.NoError(t, err)
	assert.Contains(t, databaseNames, name)
}

func TestCollectionName(t *testing.T) {
	t.Parallel()

	t.Run("Err", func(t *testing.T) {
		ctx, collection := Setup(t)

		cases := map[string]struct {
			collection string
			err        *mongo.CommandError
			alt        string
		}{
			"TooLongForBoth": {
				collection: "very_long_collection_name_that_fails_both_in_mongo_and_in_ferretdb_databases" +
					"_for_ferretdb_it_fails_because_it_is_more_than_119_characters__for_mongo_it_fails_because_it_is_more_than_255_charachters_" +
					"long_that_excludes_non_latin_letters_spaces_dots_dollars_dashes",
				err: &mongo.CommandError{
					Name: "InvalidNamespace",
					Code: 73,
					Message: "Fully qualified namespace is too long. Namespace: testcollectionname-err.very_long_collection_name_that_fails_both_in_mongo_" +
						"and_in_ferretdb_databases_for_ferretdb_it_fails_because_it_is_more_than_119_characters__for_mongo_it_fails_because_it_is_more_than_" +
						"255_charachters_long_that_excludes_non_latin_letters_spaces_dots_dollars_dashes Max: 255",
				},
				alt: "Collection must not contain non-latin letters, spaces, dots, dollars, dashes and be shorter than 119",
			},
			"WithADollarSign": {
				collection: "collection_name_with_a-$",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid collection name: collection_name_with_a-$`,
				},
				alt: "Collection must not contain non-latin letters, spaces, dots, dollars, dashes and be shorter than 119",
			},
			"Empty": {
				collection: "",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: "Invalid namespace specified 'testcollectionname-err.'",
				},
				alt: "Collection must not contain non-latin letters, spaces, dots, dollars, dashes and be shorter than 119",
			},
		}

		for name, tc := range cases {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				err := collection.Database().CreateCollection(ctx, tc.collection)
				assert.NotNil(t, tc.err)
				AssertEqualAltError(t, *tc.err, tc.alt, err)
			})
		}
	})

	t.Run("Ok", func(t *testing.T) {
		ctx, collection := Setup(t)

		longCollectionName := "very_long_collection_name_that_is_more_than_64_characters_long_but_still_valid"
		err := collection.Database().CreateCollection(ctx, longCollectionName)
		require.NoError(t, err)
		sixtyThreeCharsCollectionName := "this_is_a_collection_name_that_is_63_characters_long_abcdefghij"
		err = collection.Database().CreateCollection(ctx, sixtyThreeCharsCollectionName)
		require.NoError(t, err)

		names, err := collection.Database().ListCollectionNames(ctx, bson.D{})
		require.NoError(t, err)

		assert.Contains(t, names, longCollectionName)
		assert.Contains(t, names, sixtyThreeCharsCollectionName)
	})
}
