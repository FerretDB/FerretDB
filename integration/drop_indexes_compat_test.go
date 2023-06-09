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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestIndexesDrop(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes are not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		dropIndexName string                   // name of a single index to drop
		dropAll       bool                     // set true for drop all indexes, if true dropIndexName must be empty.
		resultType    compatTestCaseResultType // defaults to nonEmptyResult
		toCreate      []mongo.IndexModel       // optional, if not nil create indexes before dropping
	}{
		"DropAllCommand": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", 1}}},
				{Keys: bson.D{{"foo", -1}}},
				{Keys: bson.D{{"bar", 1}}},
				{Keys: bson.D{{"pam.pam", -1}}},
			},
			dropAll: true,
		},
		"ID": {
			dropIndexName: "_id_",
			resultType:    emptyResult,
		},
		"AscendingValue": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", 1}}},
			},
			dropIndexName: "v_1",
		},
		"DescendingValue": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
			dropIndexName: "v_-1",
		},
		"AsteriskWithDropOne": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
			dropIndexName: "*",
			resultType:    emptyResult,
		},
		"NonExistent": {
			dropIndexName: "nonexistent_1",
			resultType:    emptyResult,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()
			t.Parallel()

			if tc.dropAll {
				require.Empty(t, tc.dropIndexName, "index name must be empty when dropping all indexes")
			}

			// It's enough to use a single provider for drop indexes test as indexes work the same for different collections.
			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.Composites},
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					if tc.toCreate != nil {
						_, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.toCreate)
						_, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.toCreate)
						require.NoError(t, compatErr)
						require.NoError(t, targetErr)
					}

					var targetRes, compatRes bson.Raw
					var targetErr, compatErr error

					if tc.dropAll {
						targetRes, targetErr = targetCollection.Indexes().DropAll(ctx)
						compatRes, compatErr = compatCollection.Indexes().DropAll(ctx)
					} else {
						targetRes, targetErr = targetCollection.Indexes().DropOne(ctx, tc.dropIndexName)
						compatRes, compatErr = compatCollection.Indexes().DropOne(ctx, tc.dropIndexName)
					}

					require.Equal(t, compatErr, targetErr)
					require.Equal(t, compatRes, targetRes)

					if targetErr == nil {
						nonEmptyResults = true
					}

					// List indexes to see they are identical after drop.
					targetCur, targetErr := targetCollection.Indexes().List(ctx)
					compatCur, compatErr := compatCollection.Indexes().List(ctx)

					require.NoError(t, compatErr)
					require.Equal(t, compatErr, targetErr)

					targetIndexes := FetchAll(t, ctx, targetCur)
					compatIndexes := FetchAll(t, ctx, compatCur)

					require.Equal(t, compatIndexes, targetIndexes)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				require.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				require.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestIndexesDropRunCommand(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes are not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		toCreate    []mongo.IndexModel       // optional, if set, create the given indexes before drop is called
		toDrop      any                      // index to drop
		resultType  compatTestCaseResultType // defaults to nonEmptyResult
		command     bson.D                   // optional, if set it runs this command instead of dropping toDrop
		altErrorMsg string                   // optional, alternative error message in case of error
		skip        string                   // optional, skip test with a specified reason
	}{
		"InvalidType": {
			toDrop:      true,
			resultType:  emptyResult,
			altErrorMsg: `BSON field 'dropIndexes.index' is the wrong type 'bool', expected types '[string, object]'`,
		},
		"MultipleIndexesByName": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"v", 1}, {"foo", 1}}},
				{Keys: bson.D{{"v.foo", -1}}},
			},
			toDrop: bson.A{"v_-1", "v_1_foo_1"},
		},
		"MultipleIndexesByKey": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"v.foo", -1}}},
			},
			toDrop:      bson.A{bson.D{{"v", -1}}, bson.D{{"v.foo", -1}}},
			resultType:  emptyResult,
			altErrorMsg: `BSON field 'dropIndexes.index' is the wrong type 'array', expected types '[string, object]'`,
		},
		"NonExistentMultipleIndexes": {
			toDrop:     bson.A{"non-existent", "invalid"},
			resultType: emptyResult,
		},
		"InvalidMultipleIndexType": {
			toDrop:      bson.A{1},
			resultType:  emptyResult,
			altErrorMsg: `BSON field 'dropIndexes.index' is the wrong type 'array', expected types '[string, object]'`,
		},
		"DocumentIndex": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
			toDrop: bson.D{{"v", -1}},
		},
		"InvalidDocumentIndex": {
			toDrop:     bson.D{{"invalid", "invalid"}},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/2311",
		},
		"NonExistentKey": {
			toDrop:     bson.D{{"non-existent", 1}},
			resultType: emptyResult,
		},
		"DocumentIndexID": {
			toDrop:     bson.D{{"_id", 1}},
			resultType: emptyResult,
		},
		"DropAllExpression": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"foo.bar", 1}}},
				{Keys: bson.D{{"foo", 1}, {"bar", 1}}},
			},
			toDrop: "*",
		},
		"MissingIndexField": {
			command: bson.D{
				{"dropIndexes", "collection"},
			},
			resultType: emptyResult,
		},
		"NonExistentDescendingID": {
			toDrop:     bson.D{{"_id", -1}},
			resultType: emptyResult,
		},
		"MultipleKeyIndex": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"_id", -1}, {"v", 1}}},
			},
			toDrop: bson.D{
				{"_id", -1},
				{"v", 1},
			},
		},
		"NonExistentMultipleKeyIndex": {
			toDrop: bson.D{
				{"non-existent1", -1},
				{"non-existent2", -1},
			},
			resultType: emptyResult,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Helper()
			t.Parallel()

			if tc.command != nil {
				require.Nil(t, tc.toDrop, "toDrop name must be nil when using command")
			}

			// It's enough to use a single provider for drop indexes test as indexes work the same for different collections.
			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.Composites},
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					if tc.toCreate != nil {
						_, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.toCreate)
						_, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.toCreate)
						require.NoError(t, compatErr)
						require.NoError(t, targetErr)

						// List indexes to see they are identical after creation.
						targetCur, targetListErr := targetCollection.Indexes().List(ctx)
						compatCur, compatListErr := compatCollection.Indexes().List(ctx)

						require.NoError(t, compatListErr)
						require.NoError(t, targetListErr)

						targetList := FetchAll(t, ctx, targetCur)
						compatList := FetchAll(t, ctx, compatCur)

						require.Equal(t, compatList, targetList)
					}

					targetCommand := bson.D{
						{"dropIndexes", targetCollection.Name()},
						{"index", tc.toDrop},
					}

					compatCommand := bson.D{
						{"dropIndexes", compatCollection.Name()},
						{"index", tc.toDrop},
					}

					if tc.command != nil {
						targetCommand = tc.command
						compatCommand = tc.command
					}

					var targetRes bson.D
					targetErr := targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetRes)

					var compatRes bson.D
					compatErr := compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatRes)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					if tc.resultType == emptyResult {
						require.Nil(t, targetRes)
						require.Nil(t, compatRes)
					}

					require.Equal(t, compatRes, targetRes)

					if compatErr == nil {
						nonEmptyResults = true
					}

					// List indexes to see they are identical after deletion.
					targetCur, targetListErr := targetCollection.Indexes().List(ctx)
					compatCur, compatListErr := compatCollection.Indexes().List(ctx)

					require.NoError(t, compatListErr)
					assert.Equal(t, compatListErr, targetListErr)

					targetList := FetchAll(t, ctx, targetCur)
					compatList := FetchAll(t, ctx, compatCur)

					assert.Equal(t, compatList, targetList)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				require.True(t, nonEmptyResults, "expected non-empty results (some indexes should be deleted)")
			case emptyResult:
				require.False(t, nonEmptyResults, "expected empty results (no indexes should be deleted)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}
