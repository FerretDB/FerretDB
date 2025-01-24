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
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/v2/internal/util/observability"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func TestCommandCaseSensitive(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/9")
	ctx, collection := setup.Setup(tt)

	db := collection.Database()

	res := db.RunCommand(ctx, bson.D{{"listcollections", 1}})
	err := res.Err()
	require.Error(t, err)
	AssertEqualCommandError(t, mongo.CommandError{Code: 59, Name: "CommandNotFound", Message: `no such command: 'listcollections'`}, err)

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
	res = db.RunCommand(ctx, bson.D{{"dbstats", 1}})
	assert.NoError(t, res.Err())
	res = db.RunCommand(ctx, bson.D{{"dbStats", 1}})
	assert.NoError(t, res.Err())
	res = db.RunCommand(ctx, bson.D{{"findandmodify", collection.Name()}, {"update", bson.D{}}})
	assert.NoError(t, res.Err())
	res = db.RunCommand(ctx, bson.D{{"findAndModify", collection.Name()}, {"update", bson.D{}}})
	assert.NoError(t, res.Err())
}

func TestFindNothing(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var doc bson.D

	// FindOne sets limit parameter to 1, Find leaves it unset.
	err := collection.FindOne(ctx, bson.D{}).Decode(&doc)
	require.Equal(t, mongo.ErrNoDocuments, err, "actual: %s", err)
	assert.Equal(t, bson.D(nil), doc)
}

func TestInsertFind(t *testing.T) {
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

func TestFindOtelComment(t *testing.T) {
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	ctx, span := otel.Tracer("").Start(ctx, "TestOtelComment")
	defer span.End()

	comment, err := observability.CommentFromSpanContext(span.SpanContext())
	require.NoError(t, err)

	var doc bson.D
	opts := options.FindOne().SetComment(string(comment))
	err = collection.FindOne(ctx, bson.D{{"_id", "string"}}, opts).Decode(&doc)
	require.NoError(t, err)
}

//nolint:paralleltest // we test a global list of databases
func TestFindCommentMethod(t *testing.T) {
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

func TestFindEmptyKey(t *testing.T) {
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

func TestCreateCollection(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	collectionName300 := strings.Repeat("aB", 150)
	collectionName235 := strings.Repeat("a", 235)

	// use short database name to stay within 255 bytes namespace limit for using long collection name
	dbName := "short-db"

	t.Cleanup(func() {
		require.NoError(t, collection.Database().Client().Database(dbName).Drop(ctx))
	})

	testCases := map[string]struct {
		collection string // collection name, defaults to empty string

		err              *mongo.CommandError // optional, expected error from MongoDB
		altMessage       string              // optional, alternative error message for FerretDB, ignored if empty
		failsForFerretDB string
	}{
		"TooLongForBothDBs": {
			collection: collectionName300,
			err: &mongo.CommandError{
				Name: "InvalidNamespace",
				Code: 73,
				Message: fmt.Sprintf(
					"Fully qualified namespace is too long. Namespace: %s.%s Max: 255",
					dbName,
					collectionName300,
				),
			},
			altMessage: fmt.Sprintf("Invalid collection name: %s", collectionName300),
		},
		"LongEnough": {
			collection:       collectionName235,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/380",
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
				Message: fmt.Sprintf("Invalid namespace specified '%s.'", dbName),
			},
			altMessage: "Invalid collection name: ",
		},
		"Null": {
			collection: "\x00",
			err: &mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: "namespaces cannot have embedded null characters",
			},
			altMessage: "Invalid collection name: \x00",
		},
		"DotSurround": {
			collection: ".collection..",
			err: &mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: "Collection names cannot start with '.': .collection..",
			},
			altMessage: "Invalid collection name: .collection..",
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

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			db := collection.Database().Client().Database(dbName)

			err := db.CreateCollection(ctx, tc.collection)
			if tc.err != nil {
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			assert.NoError(t, err)

			names, err := db.ListCollectionNames(ctx, bson.D{})
			require.NoError(t, err)
			assert.Contains(t, names, tc.collection)

			newCollection := db.Collection(tc.collection)

			doc := bson.D{{"_id", "item"}}
			_, err = newCollection.InsertOne(ctx, doc)
			require.NoError(t, err)

			res := newCollection.FindOne(ctx, doc)
			require.NoError(t, res.Err())
		})
	}
}

func TestCreateCollectionDatabaseName(t *testing.T) {
	t.Parallel()

	t.Run("NoErr", func(t *testing.T) {
		ctx, collection := setup.Setup(t)
		for name, tc := range map[string]struct {
			db string // database name, defaults to empty string
		}{
			"Dash": {
				db: "--",
			},
			"Underscore": {
				db: "__",
			},
			"Number": {
				db: "0prefix",
			},
			"63ok": {
				db: strings.Repeat("a", 63),
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				// there is no explicit command to create database, so create collection instead
				err := collection.Database().Client().Database(tc.db).CreateCollection(ctx, collection.Name())
				require.NoError(t, err)

				err = collection.Database().Client().Database(tc.db).Drop(ctx)
				require.NoError(t, err)
			})
		}
	})

	t.Run("Err", func(t *testing.T) {
		ctx, collection := setup.Setup(t)

		dbName64 := strings.Repeat("a", 64)

		testCases := map[string]struct {
			db string // database name, defaults to empty string

			err        *mongo.CommandError // required, expected error from MongoDB
			altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		}{
			"TooLongForBothDBs": {
				db: dbName64,
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: "db name must be at most 63 characters, found: 64",
				},
				altMessage: "database name is too long",
			},
			"WithASlash": {
				db: "/",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace specified '/.TestCreateCollectionDatabaseName-Err'`,
				},
				altMessage: "Database / has an invalid character /",
			},

			"WithABackslash": {
				db: `\`,
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace specified '\.TestCreateCollectionDatabaseName-Err'`,
				},
				altMessage: `Database \ has an invalid character \`,
			},
			"WithADollarSign": {
				db: "name_with_a-$",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace: name_with_a-$.TestCreateCollectionDatabaseName-Err`,
				},
				altMessage: "Database name_with_a-$ has an invalid character $",
			},
			"WithSpace": {
				db: "data base",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace specified 'data base.TestCreateCollectionDatabaseName-Err'`,
				},
				altMessage: "Database data base has an invalid character  ",
			},
			"WithDot": {
				db: "database.test",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `'.' is an invalid character in a db name: database.test`,
				},
				altMessage: "Database database.test has an invalid character .",
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				require.NotNil(t, tc.err, "err must not be nil")

				// there is no explicit command to create database, so create collection instead
				err := collection.Database().Client().Database(tc.db).CreateCollection(ctx, collection.Name())
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
			})
		}
	})
}

func TestDebugCommandErrors(t *testing.T) {
	setup.SkipForMongoDB(t, "FerretDB-specific command")

	t.Parallel()

	ctx, collection := setup.Setup(t)
	db := collection.Database()

	// TODO https://github.com/FerretDB/FerretDB/issues/2412

	t.Run("LazyError", func(t *testing.T) {
		t.Parallel()

		err := db.RunCommand(ctx, bson.D{{"debugError", "lazy error"}}).Err()
		expected := mongo.CommandError{
			Code: 1,
			Name: "InternalError",
		}
		AssertMatchesCommandError(t, expected, err)
		assert.Regexp(t, `msg_debugerror\.go.+MsgDebugError.+lazy error$`, err.Error())

		require.NoError(t, db.Client().Ping(ctx, nil), "lazy errors should not close connection")
	})

	t.Run("OtherError", func(t *testing.T) {
		t.Parallel()

		err := db.RunCommand(ctx, bson.D{{"debugError", "other error"}}).Err()
		expected := mongo.CommandError{
			Code: 1,
			Name: "InternalError",
		}
		AssertMatchesCommandError(t, expected, err)
		assert.ErrorContains(t, err, "other error")

		require.NoError(t, db.Client().Ping(ctx, nil), "other errors should not close connection")
	})
}

func TestPingCommand(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)
	db := collection.Database()

	t.Run("Multiple", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 5; i++ {
			var res bson.D
			err := db.RunCommand(ctx, bson.D{{"ping", int32(1)}}).Decode(&res)
			require.NoError(t, err)

			AssertEqualDocuments(t, bson.D{{"ok", float64(1)}}, res)
		}
	})

	t.Run("NonExistentDB", func(t *testing.T) {
		t.Parallel()

		dbName := "NonExistentDatabase"

		list, err := db.Client().ListDatabases(ctx, bson.D{{"name", dbName}})
		require.NoError(t, err)

		for _, dbSpec := range list.Databases {
			require.NotEqual(t, dbSpec.Name, dbName)
		}

		var res bson.D
		err = db.Client().Database(dbName).RunCommand(ctx, bson.D{{"ping", int32(1)}}).Decode(&res)
		require.NoError(t, err)

		AssertEqualDocuments(t, bson.D{{"ok", float64(1)}}, res)

		list, err = db.Client().ListDatabases(ctx, bson.D{{"name", dbName}})
		require.NoError(t, err)

		for _, dbSpec := range list.Databases {
			require.NotEqual(t, dbSpec.Name, dbName)
		}
	})
}

func TestHelloIsMasterCommandMutatingClientMetadata(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)
	db := collection.Database()

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		command bson.D
		err     *mongo.CommandError
	}{
		"NoMetadataHello": {
			command: bson.D{
				{"hello", int32(1)},
			},
		},
		"NoMetadataIsMaster": {
			command: bson.D{
				{"isMaster", int32(1)},
			},
		},
		"SomeMetadataHello": {
			command: bson.D{
				{"hello", int32(1)},
				{"client", bson.D{{"application", "foobar"}}},
			},
			err: &mongo.CommandError{
				Name:    "ClientMetadataCannotBeMutated",
				Code:    186,
				Message: "The client metadata document may only be sent in the first hello",
			},
		},
		"SomeMetadataIsMaster": {
			command: bson.D{
				{"isMaster", int32(1)},
				{"client", bson.D{{"application", "foobar"}}},
			},
			err: &mongo.CommandError{
				Name:    "ClientMetadataCannotBeMutated",
				Code:    186,
				Message: "The client metadata document may only be sent in the first hello",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var res bson.D

			err := db.RunCommand(ctx, tc.command).Decode(&res)
			if tc.err != nil {
				AssertEqualCommandError(t, *tc.err, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
		})
	}
}

func TestInsertNullStrings(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/420
	t.Parallel()

	ctx, collection := setup.Setup(t)

	_, err := collection.InsertOne(ctx, bson.D{
		{"_id", "document"},
		{"a", string([]byte{0})},
	})

	require.NoError(t, err)
}

func TestInsertUpdateNestedArrays(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/420
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()

		_, err := collection.InsertOne(ctx, bson.D{{"foo", bson.A{bson.A{"bar"}}}})
		require.NoError(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		_, err := collection.UpdateOne(ctx, bson.D{}, bson.D{{"$set", bson.D{{"foo", bson.A{bson.A{"bar"}}}}}})
		require.NoError(t, err)
	})
}

func TestInsertUpdateFindNegativeZero(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/420
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		insert bson.D
		update bson.D
		filter bson.D
	}{
		"Insert": {
			insert: bson.D{{"_id", "1"}, {"v", math.Copysign(0.0, -1)}},
			filter: bson.D{{"_id", "1"}},
		},
		"UpdateZeroMulNegative": {
			insert: bson.D{{"_id", "zero"}, {"v", int32(0)}},
			update: bson.D{{"$mul", bson.D{{"v", float64(-1)}}}},
			filter: bson.D{{"_id", "zero"}},
		},
		"UpdateNegativeMulZero": {
			insert: bson.D{{"_id", "negative"}, {"v", int64(-1)}},
			update: bson.D{{"$mul", bson.D{{"v", float64(0)}}}},
			filter: bson.D{{"_id", "negative"}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := collection.InsertOne(ctx, tc.insert)
			require.NoError(t, err)

			if tc.update != nil {
				_, err = collection.UpdateOne(ctx, tc.filter, tc.update)
				require.NoError(t, err)
			}

			var res bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&res)
			require.NoError(t, err)

			expected := bson.D{
				{"_id", tc.filter[0].Value},
				{"v", math.Copysign(0.0, -1)},
			}

			AssertEqualDocuments(t, expected, res)
		})
	}
}

func TestInsertDocumentValidation(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/420
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			doc bson.D
		}{
			"DollarSign": {
				doc: bson.D{{"$foo", "bar"}},
			},
			"DotSign": {
				doc: bson.D{{"foo.bar", "baz"}},
			},
			"Infinity": {
				doc: bson.D{{"foo", math.Inf(1)}},
			},
			"NegativeInfinity": {
				doc: bson.D{{"foo", math.Inf(-1)}},
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := collection.InsertOne(ctx, tc.doc)
				require.NoError(t, err)
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		_, err := collection.UpdateOne(ctx, bson.D{}, bson.D{{"$set", bson.D{{"foo", bson.D{{"bar.baz", "qaz"}}}}}}, nil)
		require.NoError(t, err)
	})
}

func TestUpdateProduceInfinity(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/420
	t.Parallel()

	ctx, collection := setup.Setup(t)
	_, err := collection.InsertOne(ctx, bson.D{{"_id", "number"}, {"v", int32(42)}})
	require.NoError(t, err)

	_, err = collection.UpdateOne(ctx, bson.D{{"_id", "number"}}, bson.D{{"$mul", bson.D{{"v", math.MaxFloat64}}}})
	require.NoError(t, err)
}

func TestCreateCollectionDatabaseNameNonLatin(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/420
	t.Parallel()

	ctx, collection := setup.Setup(t)

	dbName := "データベース"
	cName := testutil.CollectionName(t)

	err := collection.Database().Client().Database(dbName).CreateCollection(ctx, cName)
	require.NoError(t, err)

	err = collection.Database().Client().Database(dbName).Drop(ctx)
	require.NoError(t, err)
}
