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
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
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

	for name, tc := range map[string]struct {
		command    bson.D
		expected   map[string]any
		unexpected []string
		err        *mongo.CommandError
	}{
		"AllParameters_1": {
			command: bson.D{{"getParameter", "*"}},
			expected: map[string]any{
				"acceptApiVersion2": false,
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"AllParameters_2": {
			command: bson.D{{"getParameter", "*"}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"acceptApiVersion2": false,
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"AllParameters_3": {
			command: bson.D{{"getParameter", "*"}, {"quiet", 1}, {"quiet_other", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"acceptApiVersion2": false,
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"AllParameters_4": {
			command: bson.D{{"getParameter", "*"}, {"quiet_other", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"acceptApiVersion2": false,
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"ExistingParameters": {
			command: bson.D{{"getParameter", 1}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"Zero": {
			command: bson.D{{"getParameter", 0}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"NaN": {
			command: bson.D{{"getParameter", math.NaN()}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"Nil": {
			command: bson.D{{"getParameter", nil}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"NonexistentParameters": {
			command: bson.D{{"getParameter", 1}, {"quiet", 1}, {"quiet_other", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
			unexpected: []string{"quiet_other"},
		},
		"EmptyParameters": {
			command: bson.D{{"getParameter", 1}, {"comment", "getParameter test"}},
			err:     &mongo.CommandError{Message: `no option found to get`},
		},
		"OnlyNonexistentParameters": {
			command: bson.D{{"getParameter", 1}, {"quiet_other", 1}, {"comment", "getParameter test"}},
			err:     &mongo.CommandError{Message: `no option found to get`},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			m := actual.Map()
			k := CollectKeys(t, actual)

			for key, item := range tc.expected {
				assert.Contains(t, k, key)
				assert.Equal(t, m[key], item)
			}
			for _, key := range tc.unexpected {
				assert.NotContains(t, k, key)
			}
		})
	}
}

func TestCommandsAdministrationBuildInfo(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	var actual bson.D
	command := bson.D{{"buildInfo", int32(1)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	assert.Regexp(t, `^5\.0\.`, must.NotFail(doc.Get("version")))
	assert.NotEmpty(t, must.NotFail(doc.Get("gitVersion")))

	_, ok := must.NotFail(doc.Get("modules")).(*types.Array)
	assert.True(t, ok)

	assert.Equal(t, "deprecated", must.NotFail(doc.Get("sysInfo")))

	versionArray, ok := must.NotFail(doc.Get("versionArray")).(*types.Array)
	assert.True(t, ok)
	assert.Equal(t, int32(5), must.NotFail(versionArray.Get(0)))
	assert.Equal(t, int32(0), must.NotFail(versionArray.Get(1)))

	assert.Equal(t, int32(strconv.IntSize), must.NotFail(doc.Get("bits")))
	assert.False(t, must.NotFail(doc.Get("debug")).(bool))

	assert.Equal(t, int32(16777216), must.NotFail(doc.Get("maxBsonObjectSize")))
	_, ok = must.NotFail(doc.Get("buildEnvironment")).(*types.Document)
	assert.True(t, ok)
}

func TestCommandsAdministrationCollStats(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	var actual bson.D
	command := bson.D{{"collStats", collection.Name()}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	assert.Equal(t, collection.Database().Name()+"."+collection.Name(), must.NotFail(doc.Get("ns")))
	assert.LessOrEqual(t, int32(0), must.NotFail(doc.Get("count")))
	assert.LessOrEqual(t, int32(0), must.NotFail(doc.Get("size")))
	assert.LessOrEqual(t, int32(0), must.NotFail(doc.Get("storageSize")))
	assert.LessOrEqual(t, int32(0), must.NotFail(doc.Get("totalIndexSize")))
	assert.LessOrEqual(t, int32(0), must.NotFail(doc.Get("totalSize")))
	assert.LessOrEqual(t, int32(1), must.NotFail(doc.Get("scaleFactor")))
}

func TestCommandsAdministrationDataSize(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	var actual bson.D
	command := bson.D{{"dataSize", collection.Database().Name() + "." + collection.Name()}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	assert.LessOrEqual(t, int32(0), must.NotFail(doc.Get("size")))
	assert.LessOrEqual(t, int32(0), must.NotFail(doc.Get("numObjects")))
	assert.LessOrEqual(t, int32(0), must.NotFail(doc.Get("millis")))
}

func TestCommandsAdministrationWhatsMyURI(t *testing.T) {
	t.Skip("TODO: https://github.com/FerretDB/FerretDB/issues/536")
	// TODO
}

func TestStatisticsCommands(t *testing.T) {
	t.Skip("TODO: https://github.com/FerretDB/FerretDB/issues/536")

	t.Parallel()
	ctx, collection := setup(t)

	for name, tc := range map[string]struct {
		command  any
		response bson.D
	}{
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
