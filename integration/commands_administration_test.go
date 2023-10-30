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
	"runtime"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCommandsAdministrationCreateDropList(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t) // no providers there

	db := collection.Database()
	name := collection.Name()

	names, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	require.Empty(t, names, "setup should not create collection if no providers are given")

	// drop non-existing collection in non-existing database; error is consumed by the driver
	err = collection.Drop(ctx)
	require.NoError(t, err)
	err = db.Collection(name).Drop(ctx)
	require.NoError(t, err)

	// drop manually to check error
	var res bson.D
	err = db.RunCommand(ctx, bson.D{{"drop", name}}).Decode(&res)
	expectedErr := mongo.CommandError{
		Code:    26,
		Name:    "NamespaceNotFound",
		Message: `ns not found`,
	}
	AssertEqualCommandError(t, expectedErr, err)

	err = db.CreateCollection(ctx, name)
	require.NoError(t, err)

	err = db.CreateCollection(ctx, name)
	expectedErr = mongo.CommandError{
		Code:    48,
		Name:    "NamespaceExists",
		Message: `Collection TestCommandsAdministrationCreateDropList.TestCommandsAdministrationCreateDropList already exists.`,
	}
	AssertEqualCommandError(t, expectedErr, err)

	names, err = db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	assert.Contains(t, names, name)

	// drop existing collection
	err = collection.Drop(ctx)
	require.NoError(t, err)

	// drop manually to check error
	err = db.RunCommand(ctx, bson.D{{"drop", name}}).Decode(&res)
	expectedErr = mongo.CommandError{
		Code:    26,
		Name:    "NamespaceNotFound",
		Message: `ns not found`,
	}
	AssertEqualCommandError(t, expectedErr, err)
}

func TestCommandsAdministrationCreateDropListDatabases(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t) // no providers there

	db := collection.Database()
	name := db.Name()

	filter := bson.D{{
		"name", bson.D{{
			"$in", bson.A{name}, // skip admin, other tests databases, etc
		}},
	}}
	names, err := db.Client().ListDatabaseNames(ctx, filter)
	require.NoError(t, err)
	require.Empty(t, names, "setup should not create database if no providers are given")

	// drop non-existing database; error is consumed by the driver
	err = db.Drop(ctx)
	require.NoError(t, err)

	// drop manually to check error
	var res bson.D
	err = db.RunCommand(ctx, bson.D{{"dropDatabase", 1}}).Decode(&res)
	require.NoError(t, err)
	assert.Equal(t, bson.D{{"ok", 1.0}}, res)

	// there is no explicit command to create database, so create collection instead
	err = db.Client().Database(name).CreateCollection(ctx, collection.Name())
	require.NoError(t, err)

	names, err = db.Client().ListDatabaseNames(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, []string{name}, names)

	// drop existing database
	err = db.Drop(ctx)
	require.NoError(t, err)

	// drop manually to check error
	err = db.RunCommand(ctx, bson.D{{"dropDatabase", 1}}).Decode(&res)
	require.NoError(t, err)
	assert.Equal(t, bson.D{{"ok", 1.0}}, res)
}

func TestCommandsAdministrationListDatabases(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	db := collection.Database()
	name := db.Name()

	actual, err := db.Client().ListDatabases(ctx, bson.D{{"name", name}})
	require.NoError(t, err)
	require.Len(t, actual.Databases, 1)

	expected := mongo.ListDatabasesResult{
		Databases: []mongo.DatabaseSpecification{{
			Name:       name,
			SizeOnDisk: actual.Databases[0].SizeOnDisk,
			Empty:      actual.Databases[0].Empty,
		}},
		TotalSize: actual.TotalSize,
	}

	assert.Equal(t, expected, actual)

	assert.NotZero(t, actual.Databases[0].SizeOnDisk, "%s's SizeOnDisk should be non-zero", name)
	assert.False(t, actual.Databases[0].Empty, "%s's Empty should be false", name)
	assert.NotZero(t, actual.TotalSize, "TotalSize should be non-zero")
}

func TestCommandsAdministrationListCollections(t *testing.T) {
	t.Parallel()
	ctx, targetCollections, compatCollections := setup.SetupCompat(t)

	require.Greater(t, len(targetCollections), 2)

	filter := bson.D{{
		"name", bson.D{{
			"$in", bson.A{
				targetCollections[0].Name(),
				targetCollections[len(targetCollections)-1].Name(),
			},
		}},
	}}

	target, err := targetCollections[0].Database().ListCollectionNames(ctx, filter)
	require.NoError(t, err)

	compat, err := compatCollections[0].Database().ListCollectionNames(ctx, filter)
	require.NoError(t, err)

	assert.Len(t, target, 2)
	assert.Equal(t, compat, target)
}

func TestCommandsAdministrationGetParameter(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	for name, tc := range map[string]struct {
		command    bson.D         // required, command to run
		expected   map[string]any // optional, expected keys of response
		unexpected []string       // optional, unexpected keys of response

		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"GetParameter_Asterisk1": {
			command: bson.D{{"getParameter", "*"}},
			expected: map[string]any{
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"GetParameter_Asterisk2": {
			command: bson.D{{"getParameter", "*"}, {"quiet", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"GetParameter_Asterisk3": {
			command: bson.D{{"getParameter", "*"}, {"quiet", 1}, {"quiet_other", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
				"authSchemaVersion": int32(5),
				"quiet":             false,
				"ok":                float64(1),
			},
		},
		"GetParameter_Asterisk4": {
			command: bson.D{{"getParameter", "*"}, {"quiet_other", 1}, {"comment", "getParameter test"}},
			expected: map[string]any{
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
		"AllParameters_String": {
			command: bson.D{{"getParameter", bson.D{{"showDetails", true}, {"allParameters", "1"}}}, {"quiet", true}},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: `BSON field 'getParameter.allParameters' is the wrong type 'string', expected types '[bool, long, int, decimal, double']`,
			},
			altMessage: `BSON field 'allParameters' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'`,
		},
		"FeatureCompatibilityVersion": {
			command: bson.D{
				{"getParameter", bson.D{}},
				{"featureCompatibilityVersion", 1},
			},
			expected: map[string]any{
				"featureCompatibilityVersion": bson.D{{"version", "6.0"}},
				"ok":                          float64(1),
			},
		},
		"FeatureCompatibilityVersionShowDetails": {
			command: bson.D{
				{"getParameter", bson.D{{"showDetails", true}}},
				{"featureCompatibilityVersion", 1},
			},
			expected: map[string]any{
				"featureCompatibilityVersion": bson.D{
					{"value", bson.D{{"version", "6.0"}}},
					{"settableAtRuntime", false},
					{"settableAtStartup", false},
				},
				"ok": float64(1),
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.command, "command must not be nil")

			var actual bson.D
			err := s.Collection.Database().RunCommand(s.Ctx, tc.command).Decode(&actual)
			if tc.err != nil {
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
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

func TestGetParameterCommandAuthenticationMechanisms(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	t.Run("ShowDetails", func(t *testing.T) {
		var res bson.D
		err := s.Collection.Database().RunCommand(s.Ctx, bson.D{
			{"getParameter", bson.D{{"showDetails", true}}},
			{"authenticationMechanisms", 1},
		}).Decode(&res)
		require.NoError(t, err)

		doc := ConvertDocument(t, res)
		v, _ := doc.Get("authenticationMechanisms")
		require.NotNil(t, v)

		resOk, _ := doc.Get("ok")
		require.Equal(t, float64(1), resOk)

		authenticationMechanisms, ok := v.(*types.Document)
		require.True(t, ok)

		settableAtRuntime, _ := authenticationMechanisms.Get("settableAtRuntime")
		require.Equal(t, false, settableAtRuntime)

		settableAtStartup, _ := authenticationMechanisms.Get("settableAtStartup")
		require.Equal(t, true, settableAtStartup)
	})

	t.Run("Plain", func(t *testing.T) {
		setup.SkipForMongoDB(t, "PLAIN authentication mechanism is not support by MongoDB")

		var res bson.D
		err := s.Collection.Database().RunCommand(s.Ctx, bson.D{
			{"getParameter", bson.D{}},
			{"authenticationMechanisms", 1},
		}).Decode(&res)
		require.NoError(t, err)

		expected := bson.D{
			{"authenticationMechanisms", bson.A{"PLAIN"}},
			{"ok", float64(1)},
		}
		require.Equal(t, expected, res)
	})
}

func TestCommandsAdministrationBuildInfo(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var actual bson.D
	command := bson.D{{"buildInfo", int32(1)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	assert.Regexp(t, `^6\.0\.`, must.NotFail(doc.Get("version")))
	assert.NotEmpty(t, must.NotFail(doc.Get("gitVersion")))

	_, ok := must.NotFail(doc.Get("modules")).(*types.Array)
	assert.True(t, ok)

	assert.Equal(t, "deprecated", must.NotFail(doc.Get("sysInfo")))

	versionArray, ok := must.NotFail(doc.Get("versionArray")).(*types.Array)
	assert.True(t, ok)
	assert.Equal(t, int32(6), must.NotFail(versionArray.Get(0)))
	assert.Equal(t, int32(0), must.NotFail(versionArray.Get(1)))

	assert.Equal(t, int32(strconv.IntSize), must.NotFail(doc.Get("bits")))

	assert.Equal(t, int32(16777216), must.NotFail(doc.Get("maxBsonObjectSize")))
	_, ok = must.NotFail(doc.Get("buildEnvironment")).(*types.Document)
	assert.True(t, ok)
}

func TestCommandsAdministrationBuildInfoFerretdbExtensions(t *testing.T) {
	setup.SkipForMongoDB(t, "FerretDB-specific command's extensions")

	t.Parallel()
	ctx, collection := setup.Setup(t)

	var actual bson.D
	command := bson.D{{"buildInfo", int32(1)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	ferretdbFeatures, err := doc.Get("ferretdbFeatures")
	assert.NoError(t, err)
	ferretdbFeaturesDoc, ok := ferretdbFeatures.(*types.Document)
	assert.True(t, ok)
	assert.NotNil(t, ferretdbFeatures)
	aggregationStages, err := ferretdbFeaturesDoc.Get("aggregationStages")
	aggregationStagesArray, ok := aggregationStages.(*types.Array)
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.NotEmpty(t, aggregationStagesArray)
}

func TestCommandsAdministrationCollStatsEmpty(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var actual bson.D
	command := bson.D{{"collStats", collection.Name()}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	// Expected result is to have empty statistics (neither the database nor the collection exists)
	doc := ConvertDocument(t, actual)
	assert.Equal(t, collection.Database().Name()+"."+collection.Name(), must.NotFail(doc.Get("ns")))
	assert.EqualValues(t, 0, must.NotFail(doc.Get("size")))
	assert.EqualValues(t, 0, must.NotFail(doc.Get("count")))
	assert.EqualValues(t, 0, must.NotFail(doc.Get("storageSize")))
	assert.False(t, doc.Has("freeStorageSize"))
	assert.EqualValues(t, 0, must.NotFail(doc.Get("nindexes")))
	assert.EqualValues(t, 0, must.NotFail(doc.Get("totalIndexSize")))
	assert.EqualValues(t, 0, must.NotFail(doc.Get("totalSize")))
	assert.Empty(t, must.NotFail(doc.Get("indexSizes")))
	assert.Equal(t, int32(1), must.NotFail(doc.Get("scaleFactor")))
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
}

func TestCommandsAdministrationCollStats(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	var actual bson.D
	command := bson.D{{"collStats", collection.Name()}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	// Values are returned as "numbers" that could be int32 or int64.
	// FerretDB always returns int64 for simplicity.
	//
	// Set better expected results.
	// TODO https://github.com/FerretDB/FerretDB/issues/1771
	assert.Equal(t, collection.Database().Name()+"."+collection.Name(), must.NotFail(doc.Get("ns")))
	assert.EqualValues(t, 6, must.NotFail(doc.Get("count"))) // // Number of documents in DocumentsStrings
	assert.Equal(t, int32(1), must.NotFail(doc.Get("scaleFactor")))
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

	// Values are returned as "numbers" that could be int32 or int64.
	// FerretDB always returns int64 for simplicity.
	assert.InDelta(t, 40_000, must.NotFail(doc.Get("size")), 39_900)
	assert.InDelta(t, 2_400, must.NotFail(doc.Get("avgObjSize")), 2_370)
	assert.InDelta(t, 40_000, must.NotFail(doc.Get("storageSize")), 39_900)
	assert.EqualValues(t, 0, must.NotFail(doc.Get("freeStorageSize")))
	assert.EqualValues(t, 1, must.NotFail(doc.Get("nindexes")))
	assert.InDelta(t, 12_000, must.NotFail(doc.Get("totalIndexSize")), 11_000)
	assert.InDelta(t, 32_000, must.NotFail(doc.Get("totalSize")), 30_000)

	indexSizes := must.NotFail(doc.Get("indexSizes")).(*types.Document)
	assert.Equal(t, []string{"_id_"}, indexSizes.Keys())
	assert.NotZero(t, must.NotFail(indexSizes.Get("_id_")))

	capped, _ := doc.Get("capped")
	assert.Equal(t, false, capped)

	max, _ := doc.Get("max")
	assert.Nil(t, max)

	maxSize, _ := doc.Get("maxSize")
	assert.Nil(t, maxSize)
}

func TestCommandsAdministrationCollStatsWithScale(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	var actual bson.D
	command := bson.D{{"collStats", collection.Name()}, {"scale", float64(1_000)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	// Set better expected results.
	// TODO https://github.com/FerretDB/FerretDB/issues/1771
	doc := ConvertDocument(t, actual)
	assert.Equal(t, collection.Database().Name()+"."+collection.Name(), must.NotFail(doc.Get("ns")))
	assert.EqualValues(t, 6, must.NotFail(doc.Get("count"))) // Number of documents in DocumentsStrings
	assert.Equal(t, int32(1000), must.NotFail(doc.Get("scaleFactor")))
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

	assert.InDelta(t, 16, must.NotFail(doc.Get("size")), 16)
	assert.InDelta(t, 2_400, must.NotFail(doc.Get("avgObjSize")), 2_370)
	assert.InDelta(t, 24, must.NotFail(doc.Get("storageSize")), 24)
	assert.Zero(t, must.NotFail(doc.Get("freeStorageSize")))
	assert.EqualValues(t, 1, must.NotFail(doc.Get("nindexes")))
	assert.InDelta(t, 8, must.NotFail(doc.Get("totalIndexSize")), 8)
	assert.InDelta(t, 24, must.NotFail(doc.Get("totalSize")), 24)
}

// TestCommandsAdministrationCollStatsCount adds large number of documents and checks
// approximation used by backends returns the correct count of documents from collStats.
func TestCommandsAdministrationCollStatsCount(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	var n int32 = 1000
	docs, _ := generateDocuments(0, n)
	_, err := collection.InsertMany(ctx, docs)
	require.NoError(t, err)

	var actual bson.D
	command := bson.D{{"collStats", collection.Name()}}
	err = collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)
	assert.EqualValues(t, n, must.NotFail(doc.Get("count")))
}

func TestCommandsAdministrationCollStatsSizes(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/3582")

	ctx, collection := setup.Setup(t)

	maxInt32PlusSize := math.MaxInt32 + 1
	smallSize := 1 << 30

	largeCollection := testutil.CollectionName(t) + "maxInt32Plus"
	smallCollection := testutil.CollectionName(t) + "smallSize"

	opts := options.CreateCollection().SetCapped(true)

	err := collection.Database().CreateCollection(ctx, largeCollection, opts.SetSizeInBytes(int64(maxInt32PlusSize)))
	require.NoError(t, err)

	err = collection.Database().CreateCollection(ctx, smallCollection, opts.SetSizeInBytes(int64(smallSize)))
	require.NoError(t, err)

	var largeRes bson.D
	err = collection.Database().RunCommand(ctx, bson.D{{"collStats", largeCollection}}).Decode(&largeRes)
	require.NoError(t, err)

	var smallRes bson.D
	err = collection.Database().RunCommand(ctx, bson.D{{"collStats", smallCollection}}).Decode(&smallRes)
	require.NoError(t, err)

	largeDoc := ConvertDocument(t, largeRes)
	largeMaxSize, ok := must.NotFail(largeDoc.Get("maxSize")).(int64)
	assert.True(t, ok, "int64 is used for sizes greater than math.MaxInt32")
	assert.Equal(t, int64(maxInt32PlusSize), largeMaxSize)

	smallDoc := ConvertDocument(t, smallRes)
	smallMaxSize, ok := must.NotFail(smallDoc.Get("maxSize")).(int32)
	assert.True(t, ok, "int32 is used for sizes less than math.MaxInt32")
	assert.Equal(t, int32(smallSize), smallMaxSize)
}

func TestCommandsAdministrationCollStatsScaleIndexSizes(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	indexName := "custom-name"
	resIndexName, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{"foo", 1}, {"bar", -1}},
		Options: options.Index().SetName(indexName),
	})
	require.NoError(t, err)
	require.Equal(t, indexName, resIndexName)

	scale := int32(10)
	var resNoScale bson.D
	err = collection.Database().RunCommand(ctx, bson.D{{"collStats", collection.Name()}}).Decode(&resNoScale)
	require.NoError(t, err)

	var res bson.D
	err = collection.Database().RunCommand(ctx, bson.D{{"collStats", collection.Name()}, {"scale", scale}}).Decode(&res)
	require.NoError(t, err)

	docNoScale := ConvertDocument(t, resNoScale)
	doc := ConvertDocument(t, res)

	assert.Equal(t, float64(1), must.NotFail(docNoScale.Get("ok")))
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

	assert.EqualValues(t, 2, must.NotFail(docNoScale.Get("nindexes")))
	assert.EqualValues(t, 2, must.NotFail(doc.Get("nindexes")))

	indexSizesNoScale := must.NotFail(docNoScale.Get("indexSizes")).(*types.Document)
	indexSizes := must.NotFail(doc.Get("indexSizes")).(*types.Document)

	require.Equal(t, []string{"_id_", indexName}, indexSizesNoScale.Keys())
	require.Equal(t, []string{"_id_", indexName}, indexSizes.Keys())

	idIndexSize := must.NotFail(indexSizes.Get("_id_"))
	switch idIndexSizeNoScale := must.NotFail(indexSizesNoScale.Get("_id_")).(type) {
	case int32:
		assert.EqualValues(t, idIndexSizeNoScale/scale, idIndexSize)
	case int64:
		assert.EqualValues(t, idIndexSizeNoScale/int64(scale), idIndexSize)
	default:
		t.Fatalf("unknown type %v", idIndexSizeNoScale)
	}

	customIndexSize := must.NotFail(indexSizes.Get(indexName))
	switch customIndexSizeNoScale := must.NotFail(indexSizesNoScale.Get(indexName)).(type) {
	case int32:
		assert.EqualValues(t, customIndexSizeNoScale/scale, customIndexSize)
	case int64:
		assert.EqualValues(t, customIndexSizeNoScale/int64(scale), customIndexSize)
	default:
		t.Fatalf("unknown type %v", customIndexSizeNoScale)
	}
}

func TestCommandsAdministrationDataSize(t *testing.T) {
	t.Parallel()

	t.Run("Existing", func(t *testing.T) {
		t.Parallel()

		ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

		// call validate to updated statistics
		err := collection.Database().RunCommand(ctx, bson.D{{"validate", collection.Name()}}).Err()
		require.NoError(t, err)

		var actual bson.D
		command := bson.D{{"dataSize", collection.Database().Name() + "." + collection.Name()}}
		err = collection.Database().RunCommand(ctx, command).Decode(&actual)
		require.NoError(t, err)

		doc := ConvertDocument(t, actual)
		assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
		assert.InDelta(t, 24_576, must.NotFail(doc.Get("size")), 24_576)
		assert.EqualValues(t, len(shareddata.DocumentsStrings.Docs()), must.NotFail(doc.Get("numObjects")))
		assert.InDelta(t, 200, must.NotFail(doc.Get("millis")), 300)
	})

	t.Run("NonExistent", func(t *testing.T) {
		t.Parallel()

		ctx, collection := setup.Setup(t)

		var actual bson.D
		command := bson.D{{"dataSize", collection.Database().Name() + "." + collection.Name()}}
		err := collection.Database().RunCommand(ctx, command).Decode(&actual)
		require.NoError(t, err)

		doc := ConvertDocument(t, actual)
		assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
		assert.EqualValues(t, 0, must.NotFail(doc.Get("size")))
		assert.EqualValues(t, 0, must.NotFail(doc.Get("numObjects")))
		assert.InDelta(t, 159, must.NotFail(doc.Get("millis")), 159)
	})
}

func TestCommandsAdministrationDataSizeErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	for name, tc := range map[string]struct { //nolint:vet // for readability
		command bson.D // required, command to run

		err        *mongo.CommandError // required, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"InvalidNamespace": {
			command: bson.D{{"dataSize", "invalid"}},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Invalid namespace specified 'invalid'",
			},
		},
		"InvalidNamespaceTypeDocument": {
			command: bson.D{{"dataSize", bson.D{}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type object",
			},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.command, "command must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)

			assert.Nil(t, actual)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestCommandsAdministrationDBStats(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	var actual bson.D
	command := bson.D{{"dbStats", int32(1)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	// Values are returned as "numbers" that could be int32 or int64.
	// FerretDB always returns int64 for simplicity.
	//
	// Set better expected results.
	// TODO https://github.com/FerretDB/FerretDB/issues/1771
	doc := ConvertDocument(t, actual)
	assert.Equal(t, collection.Database().Name(), doc.Remove("db"))
	assert.EqualValues(t, 1, doc.Remove("collections"))
	assert.EqualValues(t, len(shareddata.DocumentsStrings.Docs()), doc.Remove("objects"))
	assert.Equal(t, float64(1), doc.Remove("scaleFactor"))
	assert.Equal(t, float64(1), doc.Remove("ok"))

	assert.InDelta(t, 37_500, doc.Remove("avgObjSize"), 37_460)
	assert.InDelta(t, 37_500, doc.Remove("dataSize"), 37_450)
	assert.InDelta(t, 37_500, doc.Remove("storageSize"), 37_450)
	assert.InDelta(t, 49_152, doc.Remove("totalSize"), 49_100)

	freeStorageSize, _ := doc.Get("freeStorageSize")
	assert.Nil(t, freeStorageSize)

	totalFreeStorageSize, _ := doc.Get("totalFreeStorageSize")
	assert.Nil(t, totalFreeStorageSize)

	assert.Equal(t, int32(0), doc.Remove("views"))
	assert.EqualValues(t, 1, doc.Remove("indexes"))
	assert.NotZero(t, doc.Remove("indexSize"))
}

func TestCommandsAdministrationDBStatsEmpty(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	var actual bson.D
	command := bson.D{{"dbStats", int32(1)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	assert.Equal(t, float64(1), doc.Remove("ok"))
	assert.Equal(t, collection.Database().Name(), doc.Remove("db"))

	// use Equal instead of EqualValues
	// TODO https://github.com/FerretDB/FerretDB/issues/3582
	assert.EqualValues(t, float64(1), doc.Remove("scaleFactor"))

	assert.InDelta(t, 1, doc.Remove("collections"), 1)
	assert.InDelta(t, 35500, doc.Remove("dataSize"), 35500)
	assert.InDelta(t, 16384, doc.Remove("totalSize"), 16384)

	assert.Equal(t, int32(0), doc.Remove("views"))
	assert.EqualValues(t, 0, doc.Remove("indexes"))
	assert.Zero(t, doc.Remove("indexSize"))
}

func TestCommandsAdministrationDBStatsWithScale(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	var actual bson.D
	command := bson.D{{"dbStats", int32(1)}, {"scale", float64(1_000)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	assert.Equal(t, float64(1), doc.Remove("ok"))
	assert.Equal(t, collection.Database().Name(), doc.Remove("db"))
	assert.Equal(t, float64(1000), doc.Remove("scaleFactor"))

	assert.InDelta(t, 1, doc.Remove("collections"), 1)
	assert.InDelta(t, 35500, doc.Remove("dataSize"), 35500)
	assert.InDelta(t, 16384, doc.Remove("totalSize"), 16384)

	assert.Equal(t, int32(0), doc.Remove("views"))
	assert.EqualValues(t, 1, doc.Remove("indexes"))
	assert.NotZero(t, doc.Remove("indexSize"))
}

func TestCommandsAdministrationDBStatsEmptyWithScale(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	var actual bson.D
	command := bson.D{{"dbStats", int32(1)}, {"scale", float64(1_000)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	assert.Equal(t, float64(1), doc.Remove("ok"))
	assert.Equal(t, collection.Database().Name(), doc.Remove("db"))

	// use Equal instead of EqualValues
	// TODO https://github.com/FerretDB/FerretDB/issues/3582
	assert.EqualValues(t, float64(1000), doc.Remove("scaleFactor"))

	assert.InDelta(t, 1, doc.Remove("collections"), 1)
	assert.InDelta(t, 35500, doc.Remove("dataSize"), 35500)
	assert.InDelta(t, 16384, doc.Remove("totalSize"), 16384)

	assert.Equal(t, int32(0), doc.Remove("views"))
	assert.EqualValues(t, 0, doc.Remove("indexes"))
	assert.Zero(t, doc.Remove("indexSize"))
}

func TestCommandsAdministrationDBStatsFreeStorage(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	var res bson.D
	command := bson.D{{"dbStats", int32(1)}, {"freeStorage", int32(1)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&res)
	require.NoError(t, err)

	doc := ConvertDocument(t, res)

	assert.Equal(t, float64(1), doc.Remove("scaleFactor"))
	assert.Equal(t, float64(1), doc.Remove("ok"))
	assert.Zero(t, must.NotFail(doc.Get("freeStorageSize")))
	assert.Zero(t, must.NotFail(doc.Get("totalFreeStorageSize")))
}

//nolint:paralleltest // we test a global server status
func TestCommandsAdministrationServerStatus(t *testing.T) {
	ctx, collection := setup.Setup(t)

	var actual bson.D
	command := bson.D{{"serverStatus", int32(1)}}
	err := collection.Database().RunCommand(ctx, command).Decode(&actual)
	require.NoError(t, err)

	doc := ConvertDocument(t, actual)

	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

	freeMonitoring, err := doc.Get("freeMonitoring")
	require.NoError(t, err)
	assert.NotEmpty(t, must.NotFail(freeMonitoring.(*types.Document).Get("state")))

	assert.NotEmpty(t, must.NotFail(doc.Get("host")))
	assert.Regexp(t, `^6\.0\.`, must.NotFail(doc.Get("version")))
	assert.NotEmpty(t, must.NotFail(doc.Get("process")))

	assert.GreaterOrEqual(t, must.NotFail(doc.Get("pid")), int64(1))
	assert.GreaterOrEqual(t, must.NotFail(doc.Get("uptime")), float64(0))
	assert.GreaterOrEqual(t, must.NotFail(doc.Get("uptimeMillis")), int64(0))
	assert.GreaterOrEqual(t, must.NotFail(doc.Get("uptimeEstimate")), int64(0))

	assert.WithinDuration(t, time.Now(), must.NotFail(doc.Get("localTime")).(time.Time), 2*time.Second)

	catalogStats, ok := must.NotFail(doc.Get("catalogStats")).(*types.Document)
	assert.True(t, ok)

	// catalogStats is calculated across all the databases, so there could be quite a lot of collections here.
	assert.InDelta(t, 632, must.NotFail(catalogStats.Get("collections")), 632)
	assert.InDelta(t, 3, must.NotFail(catalogStats.Get("internalCollections")), 3)

	assert.Equal(t, int32(0), must.NotFail(catalogStats.Get("capped")))
	assert.Equal(t, int32(0), must.NotFail(catalogStats.Get("timeseries")))
	assert.Equal(t, int32(0), must.NotFail(catalogStats.Get("views")))
	assert.Equal(t, int32(0), must.NotFail(catalogStats.Get("internalViews")))
	assert.Equal(t, int32(0), must.NotFail(catalogStats.Get("capped")))

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1000).SetMaxDocuments(10)
	err = collection.Database().CreateCollection(ctx, testutil.CollectionName(t), opts)
	require.NoError(t, err)

	err = collection.Database().RunCommand(ctx, bson.D{{"serverStatus", int32(1)}}).Decode(&actual)
	require.NoError(t, err)

	doc = ConvertDocument(t, actual)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

	catalogStats, ok = must.NotFail(doc.Get("catalogStats")).(*types.Document)
	assert.True(t, ok)

	assert.Equal(t, int32(1), must.NotFail(catalogStats.Get("capped")))
}

func TestCommandsAdministrationServerStatusMetrics(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		cmds            []bson.D
		metricsPath     types.Path
		expectedNonZero []string
	}{
		"BasicCmd": {
			cmds: []bson.D{
				{{"ping", int32(1)}},
			},
			metricsPath:     types.NewStaticPath("metrics", "commands", "ping"),
			expectedNonZero: []string{"total"},
		},
		"UpdateCmd": {
			cmds: []bson.D{
				{{"update", "values"}, {"updates", bson.A{bson.D{{"q", bson.D{{"v", "foo"}}}}}}},
			},
			metricsPath:     types.NewStaticPath("metrics", "commands", "update"),
			expectedNonZero: []string{"total"},
		},
		"UpdateCmdFailed": {
			cmds: []bson.D{
				{{"update", int32(1)}},
			},
			metricsPath:     types.NewStaticPath("metrics", "commands", "update"),
			expectedNonZero: []string{"failed", "total"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t)

			for _, cmd := range tc.cmds {
				collection.Database().RunCommand(ctx, cmd)
			}

			command := bson.D{{"serverStatus", int32(1)}}

			var actual bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&actual)
			require.NoError(t, err)

			actualMetric, err := ConvertDocument(t, actual).GetByPath(tc.metricsPath)
			assert.NoError(t, err)

			actualDoc, ok := actualMetric.(*types.Document)
			require.True(t, ok)

			actualFields := actualDoc.Keys()

			sort.Strings(actualFields)

			var actualNotZeros []string
			for key, value := range actualDoc.Map() {
				assert.IsType(t, int64(0), value)

				if value != 0 {
					actualNotZeros = append(actualNotZeros, key)
				}
			}

			for _, expectedName := range tc.expectedNonZero {
				assert.Contains(t, actualNotZeros, expectedName)
			}
		})
	}
}

func TestCommandsAdministrationServerStatusFreeMonitoring(t *testing.T) {
	// this test shouldn't be run in parallel, because it requires a specific state of the field which would be modified by the other tests.
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	for name, tc := range map[string]struct {
		command        bson.D // required, command to run
		expectedStatus string // optional
		skipForMongoDB string // optional, skip test for MongoDB backend with a specific reason
	}{
		"Enable": {
			command:        bson.D{{"setFreeMonitoring", 1}, {"action", "enable"}},
			expectedStatus: "enabled",
			skipForMongoDB: "MongoDB decommissioned enabling free monitoring",
		},
		"Disable": {
			command:        bson.D{{"setFreeMonitoring", 1}, {"action", "disable"}},
			expectedStatus: "disabled",
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			if tc.skipForMongoDB != "" {
				setup.SkipForMongoDB(t, tc.skipForMongoDB)
			}

			require.NotNil(t, tc.command, "command must not be nil")

			res := s.Collection.Database().RunCommand(s.Ctx, tc.command)
			require.NoError(t, res.Err())

			// MongoDB might be slow to update the status
			var status any
			var retry int64
			for i := 0; i < 3; i++ {
				var actual bson.D
				err := s.Collection.Database().RunCommand(s.Ctx, bson.D{{"serverStatus", 1}}).Decode(&actual)
				require.NoError(t, err)

				doc := ConvertDocument(t, actual)

				freeMonitoring, ok := must.NotFail(doc.Get("freeMonitoring")).(*types.Document)
				assert.True(t, ok)

				status, err = freeMonitoring.Get("state")
				assert.NoError(t, err)

				if status == tc.expectedStatus {
					break
				}

				retry++
				ctxutil.SleepWithJitter(s.Ctx, time.Second, retry)
			}

			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func TestCommandsAdministrationServerStatusStress(t *testing.T) {
	// It should be rewritten to use teststress.Stress.

	ctx, collection := setup.Setup(t) // no providers there, we will create collections concurrently
	client := collection.Database().Client()

	dbNum := runtime.GOMAXPROCS(-1) * 10

	ready := make(chan struct{}, dbNum)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < dbNum; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			ready <- struct{}{}

			<-start

			dbName := fmt.Sprintf("%s_stress_%d", collection.Database().Name(), i)
			db := client.Database(dbName)
			_ = db.Drop(ctx) // make sure DB doesn't exist (it will be created together with the collection)

			collName := fmt.Sprintf("stress_%d", i)

			err := db.CreateCollection(ctx, collName)
			assert.NoError(t, err)

			err = db.Drop(ctx)
			assert.NoError(t, err)

			command := bson.D{{"serverStatus", int32(1)}}
			var actual bson.D
			err = collection.Database().RunCommand(ctx, command).Decode(&actual)

			assert.NoError(t, err)
		}(i)
	}

	for i := 0; i < dbNum; i++ {
		<-ready
	}

	close(start)

	wg.Wait()
}

func TestCommandsAdministrationCompactForce(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
		Providers:    []shareddata.Provider{shareddata.DocumentsStrings},
	})

	for name, tc := range map[string]struct {
		force any // optional, defaults to unset

		err            *mongo.CommandError // optional
		altMessage     string              // optional, alternative error message
		skip           string              // optional, skip test with a specified reason
		skipForMongoDB string              // optional, skip test for mongoDB with a specific reason
	}{
		"True": {
			force: true,
		},
		"False": {
			force: false,
		},
		"Int32": {
			force: int32(1),
		},
		"Int32Zero": {
			force: int32(0),
		},
		"Int64": {
			force: int64(1),
		},
		"Int64Zero": {
			force: int64(0),
		},
		"Double": {
			force: float64(1),
		},
		"DoubleZero": {
			force: float64(0),
		},
		"Unset": {},
		"String": {
			force: "foo",
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'force' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
			},
			skipForMongoDB: "force is FerretDB specific field",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			if tc.skipForMongoDB != "" {
				setup.SkipForMongoDB(t, tc.skipForMongoDB)
			}

			t.Parallel()

			command := bson.D{{"compact", s.Collection.Name()}}
			if tc.force != nil {
				command = append(command, bson.E{Key: "force", Value: tc.force})
			}

			var res bson.D
			err := s.Collection.Database().RunCommand(
				s.Ctx,
				command,
			).Decode(&res)

			if tc.err != nil {
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoError(t, err)

			doc := ConvertDocument(t, res)
			assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
			assert.NotNil(t, must.NotFail(doc.Get("bytesFreed")))
		})
	}
}

func TestCommandsAdministrationCompactErrors(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		dbName string

		err        *mongo.CommandError // required
		altMessage string              // optional, alternative error message
		skip       string              // optional, skip test with a specified reason
	}{
		"NonExistentDB": {
			dbName: "non-existent",
			err: &mongo.CommandError{
				Code:    26,
				Name:    "NamespaceNotFound",
				Message: "database does not exist",
			},
			altMessage: "Invalid namespace specified 'non-existent.non-existent'",
		},
		"NonExistentCollection": {
			dbName: "admin",
			err: &mongo.CommandError{
				Code:    26,
				Name:    "NamespaceNotFound",
				Message: "collection does not exist",
			},
			altMessage: "Invalid namespace specified 'admin.non-existent'",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.err, "err must not be nil")

			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				DatabaseName: tc.dbName,
			})

			var res bson.D
			err := s.Collection.Database().RunCommand(
				s.Ctx,
				bson.D{{"compact", "non-existent"}},
			).Decode(&res)

			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestCommandsAdministrationCurrentOp(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})

	var res bson.D
	err := s.Collection.Database().RunCommand(
		s.Ctx,
		bson.D{{"currentOp", int32(1)}},
	).Decode(&res)
	require.NoError(t, err)

	doc := ConvertDocument(t, res)

	_, ok := must.NotFail(doc.Get("inprog")).(*types.Array)
	assert.True(t, ok)
}

func TestCommandsAdministrationKillCursors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Strings)

	// does not show up in cursorsAlive or anywhere else
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(1))
	require.NoError(t, err)
	require.True(t, cursor.Next(ctx))

	defer cursor.Close(ctx)

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		var a bson.D
		err := collection.Database().RunCommand(ctx, bson.D{
			{"killCursors", collection.Name()},
			{"cursors", bson.A{}},
		}).Decode(&a)
		require.NoError(t, err)

		actual := ConvertDocument(t, a)
		actual.Remove("$clusterTime")
		actual.Remove("operationTime")

		expected := ConvertDocument(t, bson.D{
			{"cursorsKilled", bson.A{}},
			{"cursorsNotFound", bson.A{}},
			{"cursorsAlive", bson.A{}},
			{"cursorsUnknown", bson.A{}},
			{"ok", float64(1)},
		})
		testutil.AssertEqual(t, expected, actual)
	})

	t.Run("WrongType", func(t *testing.T) {
		t.Parallel()

		c, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(1))
		require.NoError(t, err)
		require.True(t, c.Next(ctx))
		defer c.Close(ctx)

		var a bson.D
		err = collection.Database().RunCommand(ctx, bson.D{
			{"killCursors", collection.Name()},
			{"cursors", bson.A{c.ID(), int32(100500)}},
		}).Decode(&a)

		expectedErr := mongo.CommandError{
			Code:    14,
			Name:    "TypeMismatch",
			Message: "BSON field 'killCursors.cursors.1' is the wrong type 'int', expected type 'long'",
		}
		AssertEqualCommandError(t, expectedErr, err)

		assert.True(t, c.Next(ctx))
		assert.NoError(t, c.Err())
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		c, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(1))
		require.NoError(t, err)
		require.True(t, c.Next(ctx))
		defer c.Close(ctx)

		var a bson.D
		err = collection.Database().RunCommand(ctx, bson.D{
			{"killCursors", collection.Name()},
			{"cursors", bson.A{c.ID()}},
		}).Decode(&a)
		require.NoError(t, err)

		actual := ConvertDocument(t, a)
		actual.Remove("$clusterTime")
		actual.Remove("operationTime")

		expected := ConvertDocument(t, bson.D{
			{"cursorsKilled", bson.A{c.ID()}},
			{"cursorsNotFound", bson.A{}},
			{"cursorsAlive", bson.A{}},
			{"cursorsUnknown", bson.A{}},
			{"ok", float64(1)},
		})
		testutil.AssertEqual(t, expected, actual)

		assert.False(t, c.Next(ctx))
		expectedErr := mongo.CommandError{
			Code: 43,
			Name: "CursorNotFound",
		}
		AssertMatchesCommandError(t, expectedErr, c.Err())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		c, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(1))
		require.NoError(t, err)
		require.True(t, c.Next(ctx))
		defer c.Close(ctx)

		var a bson.D
		err = collection.Database().RunCommand(ctx, bson.D{
			{"killCursors", collection.Name()},
			{"cursors", bson.A{c.ID(), int64(100500)}},
		}).Decode(&a)
		require.NoError(t, err)

		actual := ConvertDocument(t, a)
		actual.Remove("$clusterTime")
		actual.Remove("operationTime")

		expected := ConvertDocument(t, bson.D{
			{"cursorsKilled", bson.A{c.ID()}},
			{"cursorsNotFound", bson.A{int64(100500)}},
			{"cursorsAlive", bson.A{}},
			{"cursorsUnknown", bson.A{}},
			{"ok", float64(1)},
		})
		testutil.AssertEqual(t, expected, actual)

		assert.False(t, c.Next(ctx))
		expectedErr := mongo.CommandError{
			Code: 43,
			Name: "CursorNotFound",
		}
		AssertMatchesCommandError(t, expectedErr, c.Err())
	})
}
