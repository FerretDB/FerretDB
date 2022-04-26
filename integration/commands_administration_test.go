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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestCommandsAdministrationCreateDropList(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)
	db := collection.Database()
	name := collection.Name()

	names, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	assert.Contains(t, names, name)

	err = collection.Drop(ctx)
	require.NoError(t, err)

	// error is consumed by the driver
	err = collection.Drop(ctx)
	require.NoError(t, err)
	err = db.Collection(name).Drop(ctx)
	require.NoError(t, err)

	// drop manually to check error
	var actual bson.D
	err = db.RunCommand(ctx, bson.D{{"drop", name}}).Decode(&actual)
	expectedErr := mongo.CommandError{
		Code:    26,
		Name:    "NamespaceNotFound",
		Message: `ns not found`,
	}
	AssertEqualError(t, expectedErr, err)

	names, err = db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	assert.NotContains(t, names, name)

	err = db.CreateCollection(ctx, name)
	require.NoError(t, err)

	err = db.CreateCollection(ctx, name)
	expectedErr = mongo.CommandError{
		Code: 48,
		Name: "NamespaceExists",
		Message: `Collection already exists. ` +
			`NS: testcommandsadministrationcreatedroplist.testcommandsadministrationcreatedroplist`,
	}
	AssertEqualError(t, expectedErr, err)

	names, err = db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	assert.Contains(t, names, name)
}

// assertDatabases checks that expected and actual listDatabases results are similar.
func assertDatabases(t *testing.T, expected, actual mongo.ListDatabasesResult) {
	t.Helper()

	var sizeSum int64
	assert.NotZero(t, actual.TotalSize)

	require.Len(t, actual.Databases, len(expected.Databases), "%+v", actual)
	for i, a := range actual.Databases {
		e := expected.Databases[i]
		require.Equal(t, e.Name, a.Name)

		assert.Zero(t, e.SizeOnDisk)

		if a.Empty {
			assert.Zero(t, a.SizeOnDisk, "%+v", a)
			continue
		}

		// to make comparison easier
		assert.NotZero(t, a.SizeOnDisk, "%+v", a)
		sizeSum += a.SizeOnDisk
		actual.Databases[i].SizeOnDisk = 0
	}

	// That's not true for PostgreSQL, where a sum of `pg_total_relation_size` result for all schemas
	// is not equal to `pg_database_size` for the whole database.
	// assert.Equal(t, sizeSum, actual.TotalSize)
	actual.TotalSize = sizeSum

	expected.TotalSize = sizeSum
	assert.Equal(t, expected, actual)
}

//nolint:paralleltest // we test a global list of databases
func TestCommandsAdministrationCreateDropListDatabases(t *testing.T) {
	ctx, collection := setupWithOpts(t, &setupOpts{
		databaseName: "admin",
	})
	client := collection.Database().Client()
	name := collection.Name()

	// drop remnants of the previous failed run
	_ = client.Database(name).Drop(ctx)

	filter := bson.D{{
		"name", bson.D{{
			"$in", bson.A{"monila", "values", "admin", name},
		}},
	}}
	names, err := client.ListDatabaseNames(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, []string{"admin", "monila", "values"}, names)

	actual, err := client.ListDatabases(ctx, filter)
	require.NoError(t, err)

	expectedBefore := mongo.ListDatabasesResult{
		Databases: []mongo.DatabaseSpecification{{
			Name: "admin",
		}, {
			Name: "monila",
		}, {
			Name: "values",
		}},
	}
	assertDatabases(t, expectedBefore, actual)

	// there is no explicit command to create database, so create collection instead
	err = client.Database(name).CreateCollection(ctx, name)
	require.NoError(t, err)

	actual, err = client.ListDatabases(ctx, filter)
	require.NoError(t, err)

	expectedAfter := mongo.ListDatabasesResult{
		Databases: []mongo.DatabaseSpecification{{
			Name: "admin",
		}, {
			Name: "monila",
		}, {
			Name: name,
		}, {
			Name: "values",
		}},
	}
	assertDatabases(t, expectedAfter, actual)

	err = client.Database(name).Drop(ctx)
	require.NoError(t, err)

	// drop manually to check error
	var res bson.D
	err = client.Database(name).RunCommand(ctx, bson.D{{"dropDatabase", 1}}).Decode(&res)
	require.NoError(t, err)

	actual, err = client.ListDatabases(ctx, filter)
	require.NoError(t, err)
	assertDatabases(t, expectedBefore, actual)
}

func TestCommandsAdministrationGetParameter(t *testing.T) {
	t.Parallel()
	ctx, collection := setupWithOpts(t, &setupOpts{
		databaseName: "admin",
	})

	var actual bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"getParameter", "*"}}).Decode(&actual)
	require.NoError(t, err)

	m := actual.Map()
	t.Log(m)

	assert.Equal(t, 1.0, m["ok"])

	keys := CollectKeys(t, actual)
	assert.Contains(t, keys, "quiet")
	assert.Equal(t, false, m["quiet"])
}

func TestStatisticsCommands(t *testing.T) {
	t.Skip("TODO: https://github.com/FerretDB/FerretDB/issues/536")
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command  any
		response bson.D
	}{
		"BuildInfo": {
			command: bson.D{{"buildInfo", int32(1)}},
			response: bson.D{
				{"version", "5.0.42"},
				{"gitVersion", "123"},
				{"modules", primitive.A{}},
				{"sysInfo", "deprecated"},
				{"versionArray", primitive.A{int32(5), int32(0), int32(42), int32(0)}},
				{"bits", int32(strconv.IntSize)},
				{"debug", false},
				{"maxBsonObjectSize", int32(16777216)},
				{"buildEnvironment", bson.D{}},
				{"ok", 1.0},
			},
		},
		"CollStats": {
			command: bson.D{{"collStats", collection.Name()}},
			response: bson.D{
				{"ns", collection.Database().Name() + "." + collection.Name()},
				{"count", int32(43)},
				{"size", int32(16384)},
				{"storageSize", int32(8192)},
				{"totalIndexSize", int32(0)},
				{"totalSize", int32(16384)},
				{"scaleFactor", int32(1)},
				{"ok", 1.0},
			},
		},
		"DataSize": {
			command: bson.D{{"dataSize", collection.Database().Name() + "." + collection.Name()}},
			response: bson.D{
				{"estimate", false},
				{"size", int32(106_496)},
				{"numObjects", int32(210)},
				{"millis", int32(20)},
				{"ok", float64(1)},
			},
		},
		"DataSizeCollectionNotExist": {
			command: bson.D{{"dataSize", "some-database.some-collection"}},
			response: bson.D{
				{"size", int32(0)},
				{"numObjects", int32(0)},
				{"millis", int32(20)},
				{"ok", float64(1)},
			},
		},
		"DBStats": {
			command: bson.D{{"dbStats", int32(1)}},
			response: bson.D{
				{"db", collection.Database().Name()},
				{"collections", int32(1)},
				{"views", int32(0)},
				{"objects", int32(43)},
				{"avgObjSize", 481.88235294117646},
				{"dataSize", float64(8192)},
				{"indexes", int32(0)},
				{"indexSize", float64(0)},
				{"totalSize", float64(16384)},
				{"scaleFactor", float64(1)},
				{"ok", float64(1)},
			},
		},
		"DBStatsWithScale": {
			command: bson.D{{"dbStats", int32(1)}, {"scale", float64(1_000)}},
			response: bson.D{
				{"db", collection.Database().Name()},
				{"collections", int32(1)},
				{"views", int32(0)},
				{"objects", int32(43)},
				{"avgObjSize", 433.0},
				{"dataSize", 8.192},
				{"indexes", int32(0)},
				{"indexSize", float64(0)},
				{"totalSize", 16.384},
				{"scaleFactor", float64(1_000)},
				{"ok", float64(1)},
			},
		},
		"ServerStatus": {
			command: bson.D{{"serverStatus", int32(1)}},
			response: bson.D{
				{"host", ""},
				{"version", "5.0.42"},
				{"process", "handlers.test"},
				{"pid", int64(0)},
				{"uptime", int64(0)},
				{"uptimeMillis", int64(0)},
				{"uptimeEstimate", int64(0)},
				{"localTime", primitive.DateTime(time.Now().Unix())},
				{"catalogStats", bson.D{
					{"collections", int32(1)},
					{"capped", int32(0)},
					{"timeseries", int32(0)},
					{"views", int32(0)},
					{"internalCollections", int32(0)},
					{"internalViews", int32(0)},
				}},
				{"freeMonitoring", bson.D{{"state", "disabled"}}},
				{"ok", float64(1)},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.NoError(t, err)

			AssertEqualDocuments(t, tc.response, actual)
		})
	}
}
