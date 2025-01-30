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
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func TestAggregateCommandCollStats(tt *testing.T) {
	tt.Parallel()

	ctx, collection := setup.Setup(tt, shareddata.ArrayDocuments)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		pipeline           bson.A // required
		expected           bson.D // optional, defaults to the standard response with storageStats
		expectedSizeIsZero bool   // optional, if true, the size field is expected to be 0 (due to scale factor)

		failsForFerretDB string
	}{
		"EmptyCollStats": {
			pipeline:         bson.A{bson.D{{"$collStats", bson.D{}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/534",
			expected: bson.D{
				{
					"ns", collection.Database().Name() + "." + collection.Name(),
				},
				{"host", ""},
				{"localTime", primitive.DateTime(0)},
			},
		},
		"Count": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"count", bson.D{}}}}}},
			expected: bson.D{
				{
					"ns", collection.Database().Name() + "." + collection.Name(),
				},
				{"host", ""},
				{"localTime", primitive.DateTime(0)},
				{"count", int32(4)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/534",
		},
		"StorageStats": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}}},
			expected: bson.D{
				{"ns", collection.Database().Name() + "." + collection.Name()},
				{"host", ""},
				{"localTime", primitive.DateTime(0)},
				{"storageStats", bson.D{
					{"size", int32(0)},
					{"count", int32(0)},
					{"avgObjSize", int32(0)},
					{"numOrphanDocs", int32(0)},
					{"storageSize", int32(0)},
					{"freeStorageSize", int32(0)},
					{"capped", false},
					{"nindexes", int32(0)},
					{"indexDetails", bson.D{}},
					{"indexBuilds", bson.A{}},
					{"totalIndexSize", int32(0)},
					{"indexSizes", bson.D{}},
					{"totalSize", int32(0)},
					{"scaleFactor", int32(1)},
				}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/534",
		},
		"StorageStatsWithScale": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", 1000}}}}}}},
			expected: bson.D{
				{"ns", collection.Database().Name() + "." + collection.Name()},
				{"host", ""},
				{"localTime", primitive.DateTime(0)},
				{"storageStats", bson.D{
					{"size", int32(0)},
					{"count", int32(0)},
					{"avgObjSize", int32(0)},
					{"numOrphanDocs", int32(0)},
					{"storageSize", int32(0)},
					{"freeStorageSize", int32(0)},
					{"capped", false},
					{"nindexes", int32(0)},
					{"indexDetails", bson.D{}},
					{"indexBuilds", bson.A{}},
					{"totalIndexSize", int32(0)},
					{"indexSizes", bson.D{}},
					{"totalSize", int32(0)},
					{"scaleFactor", int32(1000)},
				}},
			},
			expectedSizeIsZero: true,
			failsForFerretDB:   "https://github.com/FerretDB/FerretDB-DocumentDB/issues/534",
		},
		"StorageStatsFloatScale": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", 42.42}}}}}}},
			expected: bson.D{
				{"ns", collection.Database().Name() + "." + collection.Name()},
				{"host", ""},
				{"localTime", primitive.DateTime(0)},
				{"storageStats", bson.D{
					{"size", int32(0)},
					{"count", int32(0)},
					{"avgObjSize", int32(0)},
					{"numOrphanDocs", int32(0)},
					{"storageSize", int32(0)},
					{"freeStorageSize", int32(0)},
					{"capped", false},
					{"nindexes", int32(0)},
					{"indexDetails", bson.D{}},
					{"indexBuilds", bson.A{}},
					{"totalIndexSize", int32(0)},
					{"indexSizes", bson.D{}},
					{"totalSize", int32(0)},
					{"scaleFactor", int32(42)},
				}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/534",
		},
		"CountAndStorageStats": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"count", bson.D{}}, {"storageStats", bson.D{}}}}}},
			expected: bson.D{
				{"ns", collection.Database().Name() + "." + collection.Name()},
				{"host", ""},
				{"localTime", primitive.DateTime(0)},
				{"storageStats", bson.D{
					{"size", int32(0)},
					{"count", int32(0)},
					{"avgObjSize", int32(0)},
					{"numOrphanDocs", int32(0)},
					{"storageSize", int32(0)},
					{"freeStorageSize", int32(0)},
					{"capped", false},
					{"nindexes", int32(0)},
					{"indexDetails", bson.D{}},
					{"indexBuilds", bson.A{}},
					{"totalIndexSize", int32(0)},
					{"indexSizes", bson.D{}},
					{"totalSize", int32(0)},
					{"scaleFactor", int32(1)},
				}},
				{"count", int32(4)},
			},

			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/534",
		},
		"CollStatsCount": {
			pipeline: bson.A{
				bson.D{{"$collStats", bson.D{{"count", bson.D{}}, {"storageStats", bson.D{}}}}},
				bson.D{{"$count", "after"}},
			},
			expected: bson.D{{"after", int32(1)}},
		},
	} {
		tt.Run(name, func(tt *testing.T) {
			var t testing.TB = tt

			tt.Parallel()

			require.NotNil(tt, tc.pipeline, "pipeline must be set")

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			// call validate to updated statistics
			err := collection.Database().RunCommand(ctx, bson.D{{"validate", collection.Name()}}).Err()
			require.NoError(t, err)

			cursor, err := collection.Aggregate(ctx, tc.pipeline)
			require.NoError(t, err)

			res := FetchAll(t, ctx, cursor)
			require.Len(t, res, 1)

			var actualComparable bson.D

			for _, field := range res[0] {
				switch field.Key {
				case "host":
					var port string
					_, port, err = net.SplitHostPort(field.Value.(string))
					require.NoError(t, err)
					assert.NotEmpty(t, port)

					actualComparable = append(actualComparable, bson.E{Key: field.Key, Value: ""})

				case "localTime":
					assert.IsType(t, primitive.DateTime(0), field.Value)
					actualComparable = append(actualComparable, bson.E{Key: field.Key, Value: primitive.DateTime(0)})

				case "storageStats":
					storageStats := field.Value.(bson.D)
					require.NotEmpty(t, storageStats)

					var statsComparable bson.D

					for _, stat := range storageStats {
						switch stat.Key {
						case "freeStorageSize":
							statsComparable = append(statsComparable, bson.E{Key: stat.Key, Value: int32(0)})

						case "size":
							// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/556
							size, ok := stat.Value.(int32)
							require.True(t, ok)

							if tc.expectedSizeIsZero {
								assert.Zero(t, size)
							} else {
								assert.Greater(t, size, int32(0))
							}

							statsComparable = append(statsComparable, bson.E{Key: stat.Key, Value: int32(0)})

						case "count", "avgObjSize", "storageSize", "nindexes", "totalIndexSize", "totalSize":
							// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/556
							val, ok := stat.Value.(int32)
							require.True(t, ok)

							assert.Greater(t, val, int32(0), "field %s", stat.Key)

							statsComparable = append(statsComparable, bson.E{Key: stat.Key, Value: int32(0)})

						case "indexSizes":
							assert.NotEmpty(t, stat.Value.(bson.D))

							for _, indexSize := range stat.Value.(bson.D) {
								assert.Greater(t, indexSize.Value, int32(0))
							}

							statsComparable = append(statsComparable, bson.E{Key: stat.Key, Value: bson.D{}})

						case "indexDetails":
							assert.NotEmpty(t, stat.Value.(bson.D))
							statsComparable = append(statsComparable, bson.E{Key: stat.Key, Value: bson.D{}})

						case "indexBuilds":
							assert.IsType(t, bson.A{}, stat.Value)
							statsComparable = append(statsComparable, bson.E{Key: stat.Key, Value: bson.A{}})

						case "wiredTiger":
							// exclusive to MongoDB

						default:
							statsComparable = append(statsComparable, stat)
						}
					}

					actualComparable = append(actualComparable, bson.E{Key: field.Key, Value: statsComparable})

				default:
					actualComparable = append(actualComparable, field)
				}
			}

			AssertEqualDocuments(t, tc.expected, actualComparable)
		})
	}
}

func TestAggregateCommandCollStatsErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		command  bson.D          // required, command to run
		database *mongo.Database // defaults to collection.Database()

		err              *mongo.CommandError // required
		altMessage       string              // optional, alternative error message
		failsForFerretDB string
	}{
		"NonExistentDatabase": {
			database: collection.Database().Client().Database("non-existent"),
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code: 26,
				Name: "NamespaceNotFound",
				Message: `PlanExecutor error during aggregation :: caused by :: ` +
					`Unable to retrieve storageStats in $collStats stage :: caused by :: ` +
					`Collection [non-existent.TestAggregateCommandCollStatsErrors] not found.`,
			},
			altMessage: "Collection [non-existent.TestAggregateCommandCollStatsErrors] not found.",
		},
		"NonExistentCollection": {
			command: bson.D{
				{"aggregate", "non-existent"},
				{"pipeline", bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code: 26,
				Name: "NamespaceNotFound",
				Message: `PlanExecutor error during aggregation :: caused by :: ` +
					`Unable to retrieve storageStats in $collStats stage :: caused by :: ` +
					`Collection [TestAggregateCommandCollStatsErrors.non-existent] not found.`,
			},
			altMessage: "Collection [TestAggregateCommandCollStatsErrors.non-existent] not found.",
		},
		"NilCollStats": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{bson.D{{"$collStats", nil}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code:    5447000,
				Name:    "Location5447000",
				Message: `$collStats must take a nested object but found: $collStats: null`,
			},
			altMessage:       `$collStats must take a nested object but found: { $collStats: null }`,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
		},
		"StorageStatsNegativeScale": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", -1000}}}}}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: `BSON field 'scale' value must be >= 1, actual value '-1000'`,
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/536",
		},
		"StorageStatsInvalidScale": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", "invalid"}}}}}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: `BSON field '$collStats.storageStats.scale' is the wrong type 'string', expected types '[long, int, decimal, double']`,
			},
			altMessage:       `BSON field '$collStats.storageStats.scale' is the wrong type 'string', expected types '[long, int, decimal, double]'`,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/536",
		},
		"CountCollStatsCount": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{
					bson.D{{"$count", "before"}},
					bson.D{{"$collStats", bson.D{{"count", bson.D{}}, {"storageStats", bson.D{}}}}},
					bson.D{{"$count", "after"}},
				}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code:    40602,
				Name:    "Location40602",
				Message: `$collStats is only valid as the first stage in a pipeline`,
			},
			altMessage: `$collStats is only valid as the first stage in the pipeline.`,
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.err, "err must not be nil")

			db := tc.database
			if db == nil {
				db = collection.Database()
			}

			var res bson.D
			err := db.RunCommand(ctx, tc.command).Decode(&res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
			require.Nil(t, res)
		})
	}
}

func TestAggregateCommandCollStatsIndexSizes(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/538")

	ctx, collection := setup.Setup(tt, shareddata.DocumentsStrings)

	cursorNoScale, err := collection.Aggregate(ctx, bson.A{
		bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}},
	})
	require.NoError(t, err)

	defer cursorNoScale.Close(ctx)

	scale := int32(1000)
	cursor, err := collection.Aggregate(ctx, bson.A{
		bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", scale}}}}}},
	})
	require.NoError(t, err)

	defer cursor.Close(ctx)

	resNoScale := FetchAll(t, ctx, cursorNoScale)
	require.Equal(t, 1, len(resNoScale))

	res := FetchAll(t, ctx, cursor)
	require.Equal(t, 1, len(res))

	var resComparable bson.D

	for _, field := range res[0] {
		switch field.Key {
		case "localTime":
			assert.IsType(t, primitive.DateTime(0), field.Value)
			resComparable = append(resComparable, bson.E{Key: field.Key, Value: primitive.DateTime(0)})

		case "storageStats":
			storageStats := field.Value.(bson.D)
			require.NotNil(t, storageStats)

			var storageStatsComparable bson.D

			for _, stat := range storageStats {
				switch stat.Key {
				case "scaleFactor":
					assert.Equal(t, scale, stat.Value)
					storageStatsComparable = append(storageStatsComparable, bson.E{Key: stat.Key, Value: int32(1)})

				default:
					storageStatsComparable = append(storageStatsComparable, stat)
				}
			}

			resComparable = append(resComparable, bson.E{Key: field.Key, Value: storageStatsComparable})

		default:
			resComparable = append(resComparable, field)
		}
	}

	var resNoScaleComparable bson.D

	for _, field := range resNoScale[0] {
		switch field.Key {
		case "localTime":
			assert.IsType(t, primitive.DateTime(0), field.Value)
			resNoScaleComparable = append(resNoScaleComparable, bson.E{Key: field.Key, Value: primitive.DateTime(0)})

		case "storageStats":
			storageStats := field.Value.(bson.D)
			require.NotNil(t, storageStats)

			var storageStatsComparable bson.D

			for _, stat := range storageStats {
				switch stat.Key {
				case "size", "storageSize", "freeStorageSize", "totalIndexSize", "totalSize":
					// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/556
					sizeNoScale, ok := stat.Value.(int32)
					require.True(t, ok)

					storageStatsComparable = append(storageStatsComparable, bson.E{Key: stat.Key, Value: sizeNoScale / scale})

				case "indexSizes":
					var indexSizesComparable bson.D

					for _, indexSizeNoScale := range stat.Value.(bson.D) {
						// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/556
						sizeNoScale, ok := indexSizeNoScale.Value.(int32)
						require.True(t, ok)

						indexSizesComparable = append(indexSizesComparable, bson.E{Key: indexSizeNoScale.Key, Value: sizeNoScale / scale})
					}

					storageStatsComparable = append(storageStatsComparable, bson.E{Key: stat.Key, Value: indexSizesComparable})

				case "scaleFactor":
					assert.Equal(t, int32(1), stat.Value)
					storageStatsComparable = append(storageStatsComparable, stat)

				default:
					storageStatsComparable = append(storageStatsComparable, stat)
				}
			}

			resNoScaleComparable = append(resNoScaleComparable, bson.E{Key: field.Key, Value: storageStatsComparable})

		default:
			resNoScaleComparable = append(resNoScaleComparable, field)
		}
	}

	AssertEqualDocuments(t, resComparable, resNoScaleComparable)
}
