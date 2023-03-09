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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestFindAndModifyCompatSimple(t *testing.T) {
	testCases := map[string]findAndModifyCompatTestCase{
		"EmptyQueryRemove": {
			command: bson.D{
				{"query", bson.D{}},
				{"remove", true},
			},
		},
		"NewDoubleNonZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "double-smallest"}}},
				{"update", bson.D{{"_id", "double-smallest"}, {"v", float64(43)}}},
				{"new", float64(42)},
			},
		},
		"NewDoubleZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "double-zero"}}},
				{"update", bson.D{{"_id", "double-zero"}, {"v", 43.0}}},
				{"new", float64(0)},
			},
		},
		"NewIntNonZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
				{"new", int32(11)},
			},
		},
		"NewIntZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32-zero"}}},
				{"update", bson.D{{"_id", "int32-zero"}, {"v", int32(43)}}},
				{"new", int32(0)},
			},
		},
		"NewLongNonZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"_id", "int64"}, {"v", int64(43)}}},
				{"new", int64(11)},
			},
		},
		"NewLongZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64-zero"}}},
				{"update", bson.D{{"_id", "int64-zero"}, {"v", int64(43)}}},
				{"new", int64(0)},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatErrors(t *testing.T) {
	testCases := map[string]findAndModifyCompatTestCase{
		"NotEnoughParameters": {
			command: bson.D{},
		},
		"UpdateAndRemove": {
			command: bson.D{
				{"update", bson.D{}},
				{"remove", true},
			},
		},
		"NewAndRemove": {
			command: bson.D{
				{"new", true},
				{"remove", true},
			},
		},
		"BadUpdateType": {
			command: bson.D{
				{"query", bson.D{}},
				{"update", "123"},
			},
		},
		"BadMaxTimeMSTypeString": {
			command: bson.D{
				{"maxTimeMS", "string"},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUpdate(t *testing.T) {
	testCases := map[string]findAndModifyCompatTestCase{
		"Replace": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"_id", "int64"}, {"v", int64(43)}}},
			},
		},
		"ReplaceWithoutID": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"v", int64(43)}}},
			},
		},
		"ReplaceReturnNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
				{"new", true},
			},
		},
		"NotExistedIdInQuery": {
			command: bson.D{
				{"query", bson.D{{"_id", "no-such-id"}}},
				{"update", bson.D{{"v", int32(43)}}},
			},
		},
		"NotExistedIdNotInQuery": {
			command: bson.D{
				{"query", bson.D{{"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", 0}}}},
					bson.D{{"v", bson.D{{"$lt", 0}}}},
				}}}},
				{"update", bson.D{{"v", int32(43)}}},
			},
		},
		"UpdateOperatorSet": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"$set", bson.D{{"v", int64(43)}}}}},
			},
		},
		"UpdateOperatorSetReturnNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"$set", bson.D{{"v", int64(43)}}}}},
				{"new", true},
			},
		},
		"EmptyUpdate": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"v", bson.D{}}}},
			},
			skipForTigris: "schema validation would fail",
		},
	}

	testFindAndModifyCompat(t, testCases)
}

// TestFindAndModifyCompatSort tests how various sort orders are handled.
func TestFindAndModifyCompatSort(t *testing.T) {
	testCases := map[string]findAndModifyCompatTestCase{
		"DotNotation": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$set", bson.D{{"v.0.foo.0.bar", "baz"}}}}},
				{"sort", bson.D{{"v.0.foo", 1}, {"_id", 1}}},
			},
		},
		"DotNotationIndex": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$set", bson.D{{"v.0.foo.0.bar", "baz"}}}}},
				{"sort", bson.D{{"v.0.foo.0.bar", 1}, {"_id", 1}}},
			},
		},
		"DotNotationNonExistent": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$set", bson.D{{"v.0.foo.0.bar", "baz"}}}}},
				{"sort", bson.D{{"invalid.foo", 1}, {"_id", 1}}},
			},
		},
		"DotNotationMissingField": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$set", bson.D{{"v.0.foo.0.bar", "baz"}}}}},
				{"sort", bson.D{{"v..foo", 1}, {"_id", 1}}},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUpsert(t *testing.T) {
	setup.SkipForTigrisWithReason(
		t,
		"Tigris' schema doesn't fit for most of providers, upsert for Tigris is tested in TestFindAndModifyUpsert.",
	)

	testCases := map[string]findAndModifyCompatTestCase{
		"Upsert": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1098",
		},
		"UpsertNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
				{"new", true},
			},
		},
		"UpsertNoSuchDocument": {
			command: bson.D{
				{"query", bson.D{{"_id", "no-such-doc"}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
				{"new", true},
			},
		},
		"UpsertNoSuchReplaceDocument": {
			command: bson.D{
				{"query", bson.D{{"_id", "no-such-doc"}}},
				{"update", bson.D{{"v", 43.13}}},
				{"upsert", true},
				{"new", true},
			},
		},
		"UpsertReplace": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"v", 43.13}}},
				{"upsert", true},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1098",
		},
		"UpsertReplaceReturnNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"v", 43.13}}},
				{"upsert", true},
				{"new", true},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatRemove(t *testing.T) {
	testCases := map[string]findAndModifyCompatTestCase{
		"Remove": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"remove", true},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1243",
		},
		"RemoveEmptyQueryResult": {
			command: bson.D{
				{
					"query",
					bson.D{{
						"$and",
						bson.A{
							bson.D{{"v", bson.D{{"$gt", 0}}}},
							bson.D{{"v", bson.D{{"$lt", 0}}}},
						},
					}},
				},
				{"remove", true},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1243",
		},
	}

	testFindAndModifyCompat(t, testCases)
}

// findAndModifyCompatTestCase describes findAndModify compatibility test case.
type findAndModifyCompatTestCase struct {
	command       bson.D
	skip          string // skips test if non-empty
	skipForTigris string // skips test for Tigris if non-empty
}

// testFindAndModifyCompat tests findAndModify compatibility test cases.
func testFindAndModifyCompat(t *testing.T, testCases map[string]findAndModifyCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}
			if tc.skipForTigris != "" {
				setup.SkipForTigrisWithReason(t, tc.skipForTigris)
			}

			t.Parallel()

			// Use per-test setup because findAndModify modifies data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			//	var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					targetCommand := bson.D{{"findAndModify", targetCollection.Name()}}
					targetCommand = append(targetCommand, tc.command...)
					if targetCommand.Map()["sort"] == nil {
						targetCommand = append(targetCommand, bson.D{{"sort", bson.D{{"_id", 1}}}}...)
					}

					compatCommand := bson.D{{"findAndModify", compatCollection.Name()}}
					compatCommand = append(compatCommand, tc.command...)
					if compatCommand.Map()["sort"] == nil {
						compatCommand = append(compatCommand, bson.D{{"sort", bson.D{{"_id", 1}}}}...)
					}

					var targetMod, compatMod bson.D
					var targetErr, compatErr error
					targetErr = targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetMod)
					compatErr = compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatMod)
					require.Equal(t, compatErr, targetErr)
					AssertEqualDocuments(t, compatMod, targetMod)

					// To make sure that the results of modification are equal,
					// find all the documents in target and compat collections and compare that they are the same
					opts := options.Find().SetSort(bson.D{{"_id", 1}})
					targetCursor, targetErr := targetCollection.Find(ctx, bson.D{}, opts)
					compatCursor, compatErr := compatCollection.Find(ctx, bson.D{}, opts)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						targetErr = UnsetRaw(t, targetErr)
						compatErr = UnsetRaw(t, compatErr)
						assert.Equal(t, compatErr, targetErr)
						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					var targetRes, compatRes []bson.D
					require.NoError(t, targetCursor.All(ctx, &targetRes))
					require.NoError(t, compatCursor.All(ctx, &compatRes))

					t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatRes))
					t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetRes))
					AssertEqualDocumentsSlice(t, compatRes, targetRes)
				})
			}
		})
	}
}
