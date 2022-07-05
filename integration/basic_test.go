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

	ctx, collection := Setup(t)

	cases := map[string]struct {
		Collection string
		Err        *mongo.CommandError
		Alt        string
	}{
		"TooLongForBoth": {
			Collection: "very_long_collection_name_that_fails_both_in_mongo_and_in_ferretdb_databases" +
				"_for_ferretdb_it_fails_because_it_is_more_than_119_characters__for_mongo_it_fails_because_it_is_more_than_255_charachters_" +
				"long_that_excludes_non_latin_letters_spaces_dots_dollars_dashes",
			Err: &mongo.CommandError{
				Code:    73,
				Message: `Fully qualified namespace is too long`,
			},
			Alt: "Fully qualified namespace is too long",
		},
		"WithADot": {
			Collection: "collection_name_with_a_dot.",
			Err: &mongo.CommandError{
				Code:    73,
				Message: `Namespace must not contain non-latin letters, spaces, dots, dollars, dashes.`,
			},
		},
		"WithADash": {
			Collection: "collection_name_with_a-dash",
			Err: &mongo.CommandError{
				Code:    73,
				Message: `Namespace must not contain non-latin letters, spaces, dots, dollars, dashes.`,
			},
		},
		"WithADollarSign": {
			Collection: "collection_name_with_a-$",
			Err: &mongo.CommandError{
				Code:    73,
				Message: `Namespace must not contain non-latin letters, spaces, dots, dollars, dashes.`,
			},
		},
		"Empty": {
			Collection: "",
			Err: &mongo.CommandError{
				Code:    73,
				Message: `Namespace is empty`,
			},
		},
		"64CharLong": {
			Collection: "very_long_collection_name_that_is_more_than_64_characters_long_but_still_valid",
		},
		"63CharLong": {
			Collection: "this_is_a_collection_name_that_is_63_characters_long_abcdefghij",
		},
	}

	for _, tc := range cases {
		err := collection.Database().CreateCollection(ctx, tc.Collection)
		if tc.Err == nil {
			require.NoError(t, err)
			continue
		}
		AssertEqualAltError(t, *tc.Err, tc.Alt, err)
	}

	names, err := collection.Database().ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	assert.Contains(t, names, cases["64CharLong"].Collection)
	assert.Contains(t, names, cases["63CharLong"].Collection)
}
