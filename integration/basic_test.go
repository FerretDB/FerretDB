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
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/util/testutil"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestMostCommandsAreCaseSensitive(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

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
	require.Equal(t, mongo.ErrNoDocuments, err)
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

func TestOtelComment(t *testing.T) {
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	ctx, span := otel.Tracer("").Start(ctx, "TestOtelComment")
	defer span.End()

	traceData := struct {
		TraceID [16]byte `json:"ferretTraceID"`
		SpanID  [8]byte  `json:"ferretSpanID"`
	}{
		TraceID: [16]byte(span.SpanContext().TraceID()),
		SpanID:  [8]byte(span.SpanContext().SpanID()),
	}

	comment, err := json.Marshal(traceData)
	require.NoError(t, err)

	traceDataDoc := bson.D{
		{"ferretTraceID", span.SpanContext().TraceID()},
		{"ferretSpanID", span.SpanContext().SpanID()},
	}

	var doc bson.D
	opts := options.FindOne().SetComment(string(comment))
	err = collection.FindOne(ctx, bson.D{{"_id", "string"}, {"$comment", traceDataDoc}}, opts).Decode(&doc)
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

func TestEmptyKey(t *testing.T) {
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

func TestCollectionName(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	collectionName300 := strings.Repeat("aB", 150)
	collectionName235 := strings.Repeat("a", 235)

	cases := map[string]struct {
		collection string // collection name, defaults to empty string

		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"TooLongForBothDBs": {
			collection: collectionName300,
			err: &mongo.CommandError{
				Name: "InvalidNamespace",
				Code: 73,
				Message: fmt.Sprintf(
					"Fully qualified namespace is too long. Namespace: TestCollectionName.%s Max: 255",
					collectionName300,
				),
			},
			altMessage: fmt.Sprintf("Invalid collection name: %s", collectionName300),
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
				Message: "Invalid namespace specified 'TestCollectionName.'",
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
		"Sqlite": {
			collection: "sqlite_",
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			err := collection.Database().CreateCollection(ctx, tc.collection)
			if tc.err != nil {
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			assert.NoError(t, err)

			names, err := collection.Database().ListCollectionNames(ctx, bson.D{})
			require.NoError(t, err)
			assert.Contains(t, names, tc.collection)

			newCollection := collection.Database().Collection(tc.collection)

			doc := bson.D{{"_id", "item"}}
			_, err = newCollection.InsertOne(ctx, doc)
			require.NoError(t, err)

			res := newCollection.FindOne(ctx, doc)
			require.NoError(t, res.Err())
		})
	}
}

func TestDatabaseName(t *testing.T) {
	t.Parallel()

	t.Run("NoErr", func(t *testing.T) {
		ctx, collection := setup.Setup(t)
		for name, tc := range map[string]struct {
			db   string // database name, defaults to empty string
			skip string // optional, skip test with a specified reason
		}{
			"Dash": {
				db: "--",
			},
			"Underscore": {
				db: "__",
			},
			"Sqlite": {
				db: "sqlite_",
			},
			"Number": {
				db: "0prefix",
			},
			"63ok": {
				db: strings.Repeat("a", 63),
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				if tc.skip != "" {
					t.Skip(tc.skip)
				}

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

		cases := map[string]struct {
			db string // database name, defaults to empty string

			err        *mongo.CommandError // required, expected error from MongoDB
			altMessage string              // optional, alternative error message for FerretDB, ignored if empty
			skip       string              // optional, skip test with a specified reason
		}{
			"TooLongForBothDBs": {
				db: dbName64,
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: "db name must be at most 63 characters, found: 64",
				},
				altMessage: fmt.Sprintf("Invalid namespace specified '%s.TestDatabaseName-Err'", dbName64),
			},
			"WithASlash": {
				db: "/",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace specified '/.TestDatabaseName-Err'`,
				},
			},

			"WithABackslash": {
				db: `\`,
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace specified '\.TestDatabaseName-Err'`,
				},
			},
			"WithADollarSign": {
				db: "name_with_a-$",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace: name_with_a-$.TestDatabaseName-Err`,
				},
				altMessage: `Invalid namespace specified 'name_with_a-$.TestDatabaseName-Err'`,
			},
			"WithSpace": {
				db: "data base",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `Invalid namespace specified 'data base.TestDatabaseName-Err'`,
				},
			},
			"WithDot": {
				db: "database.test",
				err: &mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: `'.' is an invalid character in a db name: database.test`,
				},
				altMessage: `Invalid namespace specified 'database.test.TestDatabaseName-Err'`,
			},
		}

		for name, tc := range cases {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				if tc.skip != "" {
					t.Skip(tc.skip)
				}

				t.Parallel()

				require.NotNil(t, tc.err, "err must not be nil")

				// there is no explicit command to create database, so create collection instead
				err := collection.Database().Client().Database(tc.db).CreateCollection(ctx, collection.Name())
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
			})
		}
	})
}

func TestDebugError(t *testing.T) {
	setup.SkipForMongoDB(t, "FerretDB-specific command")

	t.Parallel()

	ctx, collection := setup.Setup(t)
	db := collection.Database()

	// TODO https://github.com/FerretDB/FerretDB/issues/2412

	t.Run("ValidationError", func(t *testing.T) {
		t.Parallel()

		err := db.RunCommand(ctx, bson.D{{"debugError", bson.D{{"NaN", math.NaN()}}}}).Err()
		expected := mongo.CommandError{
			Code: 2,
			Name: "BadValue",
		}
		AssertMatchesCommandError(t, expected, err)
		assert.ErrorContains(t, err, "NaN is not supported")

		require.NoError(t, db.Client().Ping(ctx, nil), "validation errors should not close connection")
	})

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

	expected := ConvertDocument(t, bson.D{{"ok", float64(1)}})

	t.Run("Multiple", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 5; i++ {
			res := db.RunCommand(ctx, bson.D{{"ping", int32(1)}})

			var actualRes bson.D
			err := res.Decode(&actualRes)
			require.NoError(t, err)

			actual := ConvertDocument(t, actualRes)
			actual.Remove("$clusterTime")
			actual.Remove("operationTime")

			testutil.AssertEqual(t, expected, actual)
		}
	})

	t.Run("NonExistentDB", func(t *testing.T) {
		t.Parallel()

		dbName := "NonExistentDatabase"

		expectedDatabases, err := db.Client().ListDatabases(ctx, bson.D{{"name", dbName}})
		require.NoError(t, err)
		require.Empty(t, expectedDatabases.Databases)

		res := db.Client().Database(dbName).RunCommand(ctx, bson.D{{"ping", int32(1)}})

		var actualRes bson.D
		err = res.Decode(&actualRes)
		require.NoError(t, err)

		actual := ConvertDocument(t, actualRes)
		actual.Remove("$clusterTime")
		actual.Remove("operationTime")

		expected := ConvertDocument(t, bson.D{{"ok", float64(1)}})

		testutil.AssertEqual(t, expected, actual)

		// Ensure that we don't create database on ping
		// This also means that no collection is created during ping.
		actualDatabases, err := db.Client().ListDatabases(ctx, bson.D{{"name", dbName}})
		require.NoError(t, err)
		require.Empty(t, actualDatabases.Databases)
	})
}

func TestMutatingClientMetadata(t *testing.T) {
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
