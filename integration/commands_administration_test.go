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
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
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
	ctx, collection, _ := setupWithOpts(t, &setupOpts{
		databaseName: "admin",
	})
	client := collection.Database().Client()
	name := collection.Name()

	// drop remnants of the previous failed run
	_ = client.Database(name).Drop(ctx)

	filter := bson.D{{
		"name", bson.D{{
			"$in", bson.A{name, "admin"},
		}},
	}}
	names, err := client.ListDatabaseNames(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, []string{"admin"}, names)

	actual, err := client.ListDatabases(ctx, filter)
	require.NoError(t, err)

	expectedBefore := mongo.ListDatabasesResult{
		Databases: []mongo.DatabaseSpecification{{
			Name: "admin",
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
			Name: name,
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
	ctx, collection, _ := setupWithOpts(t, &setupOpts{
		databaseName: "admin",
	})

	for name, tc := range map[string]struct {
		command    bson.D
		expected   map[string]any
		unexpected []string
		err        *mongo.CommandError
		altMessage string
	}{
		"GetParameter_Asterisk1": {
			command: bson.D{{"getParameter", "*"}},
			expected: map[string]any{
				"acceptApiVersion2": false,
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"GetParameter_Asterisk2": {
			command: bson.D{{"getParameter", "*"}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"acceptApiVersion2": false,
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"GetParameter_Asterisk3": {
			command: bson.D{{"getParameter", "*"}, {"quiet", 1}, {"quiet_other", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"acceptApiVersion2": false,
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"GetParameter_Asterisk4": {
			command: bson.D{{"getParameter", "*"}, {"quiet_other", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"acceptApiVersion2": false,
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"GetParameter_Int": {
			command: bson.D{{"getParameter", 1}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"GetParameter_Zero": {
			command: bson.D{{"getParameter", 0}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"GetParameter_NaN": {
			command: bson.D{{"getParameter", math.NaN()}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"GetParameter_Nil": {
			command: bson.D{{"getParameter", nil}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"GetParameter_String": {
			command: bson.D{{"getParameter", "1"}, {"quiet", 1}, {"comment", "getParameter test"}},
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
		"ShowDetailsTrue": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
			unexpected: []string{"acceptApiVersion2"},
		},
		"ShowDetailsFalse": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", false}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
			unexpected: []string{"acceptApiVersion2"},
		},
		"ShowDetails_NoParameter_1": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}}}},
			err:     &mongo.CommandError{Message: `no option found to get`},
		},
		"ShowDetails_NoParameter_2": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", false}}}},
			err:     &mongo.CommandError{Message: `no option found to get`},
		},
		"AllParametersTrue": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", true}}}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"tlsMode": bson.D{
					{"value", "disabled"},
					{"settableAtRuntime", true},
					{"settableAtStartup", false},
				},
				"ok": float64(1),
			},
		},
		"AllParametersFalse_MissingParameter": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", false}}}},
			err:     &mongo.CommandError{Message: `no option found to get`},
		},
		"AllParametersFalse_PresentParameter": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", false}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
			unexpected: []string{"acceptApiVersion2"},
		},
		"AllParametersFalse_NonexistentParameter": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", false}}}, {"quiet_other", true}},
			err:     &mongo.CommandError{Message: `no option found to get`},
		},
		"ShowDetailsFalse_AllParametersTrue": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", false}, {"allParameters", true}}}},
			expected: map[string]any{
				"acceptApiVersion2": false,
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"ShowDetailsFalse_AllParametersFalse_1": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", false}, {"allParameters", false}}}},
			err:     &mongo.CommandError{Message: `no option found to get`},
		},
		"ShowDetailsFalse_AllParametersFalse_2": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", false}, {"allParameters", false}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
			unexpected: []string{"acceptApiVersion2"},
		},
		"ShowDetails_NegativeInt": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", int64(-1)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
		},
		"ShowDetails_PositiveInt": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", int64(1)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
		},
		"ShowDetails_ZeroInt": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", int64(0)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
			unexpected: []string{"acceptApiVersion2"},
		},
		"ShowDetails_ZeroFloat": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", float64(0)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
			unexpected: []string{"acceptApiVersion2"},
		},
		"ShowDetails_SmallestNonzeroFloat": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", math.SmallestNonzeroFloat64}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
		},
		"ShowDetails_Nil": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", nil}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"ShowDetails_NaN": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", math.NaN()}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
		},
		"ShowDetails_NegatuveZero": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", math.Copysign(0, -1)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": false,
				"ok":    float64(1),
			},
		},
		"ShowDetails_String": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", "1"}}}, {"quiet", true}},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: `BSON field 'getParameter.showDetails' is the wrong type 'string', expected types '[bool, long, int, decimal, double']`,
			},
			altMessage: `BSON field 'showDetails' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'`,
		},
		"AllParameters_NegativeInt": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", int64(-1)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"tlsMode": bson.D{
					{"value", "disabled"},
					{"settableAtRuntime", true},
					{"settableAtStartup", false},
				},
				"ok": float64(1),
			},
		},
		"AllParameters_PositiveInt": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", int64(1)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"tlsMode": bson.D{
					{"value", "disabled"},
					{"settableAtRuntime", true},
					{"settableAtStartup", false},
				},
				"ok": float64(1),
			},
		},
		"AllParameters_ZeroInt": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", int64(0)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
		},
		"AllParameters_ZeroFloat": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", float64(0)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
		},
		"AllParameters_SmallestNonzeroFloat": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", math.SmallestNonzeroFloat64}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"tlsMode": bson.D{
					{"value", "disabled"},
					{"settableAtRuntime", true},
					{"settableAtStartup", false},
				},
				"ok": float64(1),
			},
		},
		"AllParameters_Nil": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", nil}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
			unexpected: []string{"acceptApiVersion2"},
		},
		"AllParameters_NaN": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", math.NaN()}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"tlsMode": bson.D{
					{"value", "disabled"},
					{"settableAtRuntime", true},
					{"settableAtStartup", false},
				},
				"ok": float64(1),
			},
		},
		"AllParameters_NegatuveZero": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", math.Copysign(0, -1)}}}, {"quiet", true}},
			expected: map[string]any{
				"quiet": bson.D{
					{"value", false},
					{"settableAtRuntime", true},
					{"settableAtStartup", true},
				},
				"ok": float64(1),
			},
			unexpected: []string{"acceptApiVersion2"},
		},
		"AllParameters_String": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", "1"}}}, {"quiet", true}},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: `BSON field 'getParameter.allParameters' is the wrong type 'string', expected types '[bool, long, int, decimal, double']`,
			},
			altMessage: `BSON field 'allParameters' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'`,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)

			if tc.err != nil {
				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
				return
			}
			require.NoError(t, err)

			m := actual.Map()
			keys := CollectKeys(t, actual)

			for k, item := range tc.expected {
				assert.Contains(t, keys, k)
				assert.IsType(t, item, m[k])
				if it, ok := item.(primitive.D); ok {
					z := m[k].(primitive.D)
					AssertEqualDocuments(t, it, z)
				} else {
					assert.Equal(t, m[k], item)
				}
			}

			for _, k := range tc.unexpected {
				assert.NotContains(t, keys, k)
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

func TestCommandsAdministrationCollStatsEmpty(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	var actual bson.D
	command := bson.D{{"collStats", collection.Name()}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	assert.Equal(t, collection.Database().Name()+"."+collection.Name(), must.NotFail(doc.Get("ns")))
	assert.Equal(t, int32(0), must.NotFail(doc.Get("count")))
	assert.Equal(t, int32(1), must.NotFail(doc.Get("scaleFactor")))

	assert.InDelta(t, float64(8012), must.NotFail(doc.Get("size")), 8_012)
	assert.InDelta(t, float64(4096), must.NotFail(doc.Get("storageSize")), 8_012)
	assert.InDelta(t, float64(4096), must.NotFail(doc.Get("totalIndexSize")), 8_012)
	assert.InDelta(t, float64(4096), must.NotFail(doc.Get("totalSize")), 8_012)
}

func TestCommandsAdministrationCollStats(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	var actual bson.D
	command := bson.D{{"collStats", collection.Name()}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	assert.Equal(t, int32(1), must.NotFail(doc.Get("scaleFactor")))
	assert.Equal(t, collection.Database().Name()+"."+collection.Name(), must.NotFail(doc.Get("ns")))

	assert.InDelta(t, float64(8012), must.NotFail(doc.Get("size")), 16_024)
	assert.InDelta(t, float64(4096), must.NotFail(doc.Get("storageSize")), 8_012)
	assert.InDelta(t, float64(4096), must.NotFail(doc.Get("totalIndexSize")), 8_012)
	assert.InDelta(t, float64(4096), must.NotFail(doc.Get("totalSize")), 16_024)
}

func TestCommandsAdministrationDataSize(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	var actual bson.D
	command := bson.D{{"dataSize", collection.Database().Name() + "." + collection.Name()}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

	assert.InDelta(t, float64(8012), must.NotFail(doc.Get("size")), 16_024)
	assert.InDelta(t, float64(100), must.NotFail(doc.Get("millis")), 100)
}

func TestCommandsAdministrationDataSizeCollectionNotExist(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	var actual bson.D
	command := bson.D{{"dataSize", "some-database.some-collection"}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	assert.Equal(t, int32(0), must.NotFail(doc.Get("size")))
	assert.Equal(t, int32(0), must.NotFail(doc.Get("numObjects")))

	assert.InDelta(t, float64(100), must.NotFail(doc.Get("millis")), 100)
}

func TestCommandsAdministrationDBStatsEmpty(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	var actual bson.D
	command := bson.D{{"dbStats", int32(1)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	assert.Equal(t, collection.Database().Name(), must.NotFail(doc.Get("db")))
	assert.Equal(t, int32(1), must.NotFail(doc.Get("collections")))
	assert.Equal(t, int32(0), must.NotFail(doc.Get("views")))
	assert.Equal(t, int32(0), must.NotFail(doc.Get("objects")))
	assert.Equal(t, float64(0), must.NotFail(doc.Get("avgObjSize")))
	assert.Equal(t, float64(0), must.NotFail(doc.Get("dataSize")))

	assert.InDelta(t, float64(1), must.NotFail(doc.Get("indexes")), 1)
	assert.InDelta(t, float64(4096), must.NotFail(doc.Get("indexSize")), 4_096)

	assert.InDelta(t, float64(1), must.NotFail(doc.Get("totalSize")), 8192)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("scaleFactor")))
}

func TestCommandsAdministrationDBStatsWithScale(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	var actual bson.D
	command := bson.D{{"dbStats", int32(1)}, {"scale", float64(1_000)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	assert.Equal(t, collection.Database().Name(), must.NotFail(doc.Get("db")))
	assert.Equal(t, int32(1), must.NotFail(doc.Get("collections")))
	assert.Equal(t, int32(0), must.NotFail(doc.Get("views")))
	assert.Equal(t, float64(1000), must.NotFail(doc.Get("scaleFactor")))

	assert.InDelta(t, float64(2.161), must.NotFail(doc.Get("dataSize")), 10)
	assert.InDelta(t, float64(1), must.NotFail(doc.Get("indexes")), 1)
	assert.InDelta(t, float64(0), must.NotFail(doc.Get("indexSize")), 4060)
}

func TestCommandsAdministrationServerStatus(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	var actual bson.D
	command := bson.D{{"serverStatus", int32(1)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

	freeMonitoring, ok := must.NotFail(doc.Get("freeMonitoring")).(*types.Document)
	assert.True(t, ok)
	assert.NotEmpty(t, must.NotFail(freeMonitoring.Get("state")))

	assert.NotEmpty(t, must.NotFail(doc.Get("host")))
	assert.Regexp(t, `^5\.0\.`, must.NotFail(doc.Get("version")))
	assert.NotEmpty(t, must.NotFail(doc.Get("process")))

	assert.GreaterOrEqual(t, must.NotFail(doc.Get("pid")), int64(1))
	assert.GreaterOrEqual(t, must.NotFail(doc.Get("uptime")), float64(0))
	assert.GreaterOrEqual(t, must.NotFail(doc.Get("uptimeMillis")), int64(0))
	assert.GreaterOrEqual(t, must.NotFail(doc.Get("uptimeEstimate")), int64(0))

	expectedLocalTime := ConvertDocument(t, bson.D{{"localTime", primitive.NewDateTimeFromTime(time.Now())}})
	testutil.CompareAndSetByPathTime(t, expectedLocalTime, doc, time.Duration(2*time.Second), types.NewPathFromString("localTime"))

	catalogStats, ok := must.NotFail(doc.Get("catalogStats")).(*types.Document)
	assert.True(t, ok)

	assert.InDelta(t, float64(51), must.NotFail(catalogStats.Get("collections")), 50)
	assert.InDelta(t, float64(3), must.NotFail(catalogStats.Get("internalCollections")), 3)

	assert.Equal(t, int32(0), must.NotFail(catalogStats.Get("capped")))
	assert.Equal(t, int32(0), must.NotFail(catalogStats.Get("timeseries")))
	assert.Equal(t, int32(0), must.NotFail(catalogStats.Get("views")))
	assert.Equal(t, int32(0), must.NotFail(catalogStats.Get("internalViews")))
}

// TestCommandsAdministrationWhatsMyURI tests the `whatsmyuri` command.
// It connects two clients to the same server and checks that `whatsmyuri` returns different ports for these clients.
func TestCommandsAdministrationWhatsMyURI(t *testing.T) {
	t.Parallel()
	ctx, collection1, port := setupWithOpts(t, new(setupOpts))
	client2 := setupClient(t, ctx, port)
	collection2 := client2.Database(collection1.Database().Name()).Collection(collection1.Name())

	var ports []string
	for _, collection := range []*mongo.Collection{collection1, collection2} {
		var actual bson.D
		command := bson.D{{"whatsmyuri", int32(1)}}
		err := collection.Database().RunCommand(ctx, command).Decode(&actual)
		require.NoError(t, err)

		doc := ConvertDocument(t, actual)
		assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

		// record ports to compare that they are not equal for two different clients.
		_, port, err := net.SplitHostPort(must.NotFail(doc.Get("you")).(string))
		require.NoError(t, err)
		ports = append(ports, port)
	}

	require.Equal(t, 2, len(ports))
	assert.NotEqual(t, ports[0], ports[1])
}
