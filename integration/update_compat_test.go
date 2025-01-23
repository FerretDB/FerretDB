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
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

// updateCompatTestCase describes update compatibility test case.
type updateCompatTestCase struct {
	update      bson.D                   // required if replace is nil
	replace     bson.D                   // required if update is nil
	filter      bson.D                   // defaults to bson.D{{"_id", id}}
	updateOpts  *options.UpdateOptions   // defaults to nil
	replaceOpts *options.ReplaceOptions  // defaults to nil
	resultType  compatTestCaseResultType // defaults to nonEmptyResult
	providers   []shareddata.Provider    // defaults to shareddata.AllProviders()

	skip             string // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1086
	failsForFerretDB string
	failsIDs         []struct {
		provider shareddata.Provider
		ids      []string // defaults to all IDs of the provider
	} // defaults to all providers
}

// testUpdateCompat tests update compatibility test cases.
func testUpdateCompat(t *testing.T, testCases map[string]updateCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			providers := shareddata.AllProviders()
			if tc.providers != nil {
				providers = tc.providers
			}

			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                providers,
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			update, replace := tc.update, tc.replace
			if update != nil {
				require.Nil(t, replace, "`replace` must be nil if `update` is set")
			} else {
				require.NotNil(t, replace, "`replace` must be set if `update` is nil")
			}

			failsProviders := make(map[string][]string)
			for _, p := range tc.failsIDs {
				failsProviders[p.provider.Name()] = p.ids
			}

			allProvidersFail := len(failsProviders) == 0

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				// a workaround to get provider name by using the part after last `_`,
				// e.g. `Binaries` from `TestUpdateFieldCompatBit-Bool_Binaries`
				str := strings.Split(targetCollection.Name(), "_")
				providerName := str[len(str)-1]

				failsIDs, providerFails := failsProviders[providerName]
				allIDsFail := providerFails && len(failsIDs) == 0

				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					allDocs := FindAll(t, ctx, targetCollection)

					for _, doc := range allDocs {
						id, ok := doc.Map()["_id"]
						require.True(t, ok)

						t.Run(fmt.Sprint(id), func(tt *testing.T) {
							tt.Helper()

							idFails := slices.Contains(failsIDs, fmt.Sprint(id))

							var t testing.TB = tt

							if tc.failsForFerretDB != "" && (allProvidersFail || allIDsFail || idFails) {
								t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
							}

							filter := tc.filter
							if tc.filter == nil {
								filter = bson.D{{"_id", id}}
							}

							var targetUpdateRes, compatUpdateRes *mongo.UpdateResult
							var targetErr, compatErr error

							// Replace with UpdateMany/ReplaceMany.
							// TODO https://github.com/FerretDB/FerretDB/issues/1507
							if update != nil {
								targetUpdateRes, targetErr = targetCollection.UpdateOne(ctx, filter, update, tc.updateOpts)
								compatUpdateRes, compatErr = compatCollection.UpdateOne(ctx, filter, update, tc.updateOpts)
							} else {
								targetUpdateRes, targetErr = targetCollection.ReplaceOne(ctx, filter, replace, tc.replaceOpts)
								compatUpdateRes, compatErr = compatCollection.ReplaceOne(ctx, filter, replace, tc.replaceOpts)
							}

							if targetErr != nil {
								t.Logf("Target error: %v", targetErr)

								if targetErr.Error() == "update document must have at least one element" {
									// mongo go driver sent error that the update document is empty.
									require.Equal(t, compatErr, targetErr)

									return
								}

								// AssertMatchesWriteError compares error types and codes, it does not compare messages.
								AssertMatchesWriteError(t, compatErr, targetErr)
							} else {
								require.NoError(t, compatErr, "compat error; target returned no error")
							}

							assert.Equal(t, compatUpdateRes, targetUpdateRes)

							if pointer.Get(targetUpdateRes).ModifiedCount > 0 || pointer.Get(targetUpdateRes).UpsertedCount > 0 {
								nonEmptyResults = true
							}

							var targetFindRes, compatFindRes bson.D
							require.NoError(t, targetCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&targetFindRes))
							require.NoError(t, compatCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&compatFindRes))
							AssertEqualDocuments(t, compatFindRes, targetFindRes)
						})
					}
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				if tc.failsForFerretDB != "" {
					return
				}

				assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				if tc.failsForFerretDB != "" {
					return
				}

				assert.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

// testUpdateManyCompatTestCase describes update compatibility test case.
type testUpdateManyCompatTestCase struct { //nolint:vet // used for testing only
	update     bson.D                   // required if replace is nil
	filter     bson.D                   // defaults to bson.D{}
	updateOpts *options.UpdateOptions   // defaults to nil
	resultType compatTestCaseResultType // defaults to nonEmptyResult
	providers  []shareddata.Provider    // defaults to shareddata.AllProviders()

	skip string // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1086
}

// testUpdateManyCompat tests update compatibility test cases.
func testUpdateManyCompat(t *testing.T, testCases map[string]testUpdateManyCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			setupOpts := &setup.SetupCompatOpts{
				Providers: tc.providers,
			}

			if tc.providers == nil {
				setupOpts.Providers = shareddata.AllProviders()
				setupOpts.AddNonExistentCollection = true
			}

			s := setup.SetupCompatWithOpts(t, setupOpts)

			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			update := tc.update
			require.NotNil(t, update, "`update` must be set")

			filter := tc.filter
			if filter == nil {
				filter = bson.D{}
			}

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					var targetUpdateRes, compatUpdateRes *mongo.UpdateResult
					var targetErr, compatErr error

					targetUpdateRes, targetErr = targetCollection.UpdateMany(ctx, filter, update, tc.updateOpts)
					compatUpdateRes, compatErr = compatCollection.UpdateMany(ctx, filter, update, tc.updateOpts)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						AssertMatchesWriteError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					assert.Equal(t, compatUpdateRes, targetUpdateRes)

					if pointer.Get(targetUpdateRes).ModifiedCount > 0 || pointer.Get(targetUpdateRes).UpsertedCount > 0 {
						nonEmptyResults = true
					}

					opts := options.Find().SetSort(bson.D{{"_id", 1}})
					targetCursor, targetErr := targetCollection.Find(ctx, bson.D{}, opts)
					compatCursor, compatErr := compatCollection.Find(ctx, bson.D{}, opts)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					require.NoError(t, targetErr)
					require.NoError(t, compatErr)

					targetRes := FetchAll(t, ctx, targetCursor)
					compatRes := FetchAll(t, ctx, compatCursor)

					AssertEqualDocumentsSlice(t, compatRes, targetRes)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

// updateCommandCompatTestCase describes update command compatibility test case.
type updateCommandCompatTestCase struct {
	multi      any                      // defaults to false, if true updates multiple documents
	upsert     bool                     // defaults to false
	update     bson.D                   // required
	filter     bson.D                   // defaults to bson.D{{"_id", id}}
	resultType compatTestCaseResultType // defaults to nonEmptyResult

	failsForFerretDB string
	failsIDs         []struct {
		provider shareddata.Provider
		ids      []string // defaults to all IDs of the provider
	} // defaults to all providers
}

// testUpdateCommandCompat tests command compatibility test cases used for update.
// This is used for updating multiple documents and testing multi flag values.
func testUpdateCommandCompat(tt *testing.T, testCases map[string]updateCommandCompatTestCase) {
	tt.Helper()

	for name, tc := range testCases {
		tt.Run(name, func(tt *testing.T) {
			tt.Helper()

			tt.Parallel()

			// Use per-test setup because updates modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(tt)

			update := tc.update
			require.NotNil(tt, update, "`update` must be set")

			multi := tc.multi
			if multi == nil {
				multi = false
			}

			failsProviders := make(map[string][]string)
			for _, p := range tc.failsIDs {
				failsProviders[p.provider.Name()] = p.ids
			}

			allProvidersFail := len(failsProviders) == 0

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				// a workaround to get provider name by using the part after last `_`,
				// e.g. `Binaries` from `TestUpdateFieldCompatBit-Bool_Binaries`
				str := strings.Split(targetCollection.Name(), "_")
				providerName := str[len(str)-1]

				failsIDs, providerFails := failsProviders[providerName]
				allIDsFail := providerFails && len(failsIDs) == 0

				tt.Run(targetCollection.Name(), func(tt *testing.T) {
					tt.Helper()

					allDocs := FindAll(tt, ctx, targetCollection)

					for _, doc := range allDocs {
						id, ok := doc.Map()["_id"]
						require.True(tt, ok)

						tt.Run(fmt.Sprint(id), func(tt *testing.T) {
							idFails := slices.Contains(failsIDs, fmt.Sprint(id))

							var t testing.TB = tt

							if tc.failsForFerretDB != "" && (allProvidersFail || allIDsFail || idFails) {
								t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
							}

							t.Helper()

							filter := tc.filter
							if tc.filter == nil {
								filter = bson.D{{"_id", id}}
							}

							targetCommand := bson.D{
								{"update", targetCollection.Name()},
								{"updates", bson.A{bson.D{
									{"q", filter},
									{"u", update},
									{"multi", multi},
									{"upsert", tc.upsert},
								}}},
							}

							compatCommand := bson.D{
								{"update", compatCollection.Name()},
								{"updates", bson.A{bson.D{
									{"q", filter},
									{"u", update},
									{"multi", multi},
									{"upsert", tc.upsert},
								}}},
							}

							var targetUpdateRes, compatUpdateRes bson.D
							var targetErr, compatErr error

							targetErr = targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetUpdateRes)
							compatErr = compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatUpdateRes)

							if targetErr != nil {
								t.Logf("Target error: %v", targetErr)
								t.Logf("Compat error: %v", compatErr)

								// compare error type and error code
								AssertMatchesError(t, compatErr, targetErr)
							} else {
								require.NoError(t, compatErr, "compat error; target returned no error")
							}

							if targetUpdateRes != nil || compatUpdateRes != nil {
								nonEmptyResults = true
							}

							AssertEqualDocuments(t, compatUpdateRes, targetUpdateRes)

							if isMulti, ok := multi.(bool); ok && isMulti {
								// if multi == false, an item updated by compat and target are different.

								opts := options.Find().SetSort(bson.D{{"_id", 1}})
								targetCursor, targetErr := targetCollection.Find(ctx, filter, opts)
								compatCursor, compatErr := compatCollection.Find(ctx, filter, opts)

								if targetCursor != nil {
									defer targetCursor.Close(ctx)
								}
								if compatCursor != nil {
									defer compatCursor.Close(ctx)
								}

								if targetErr != nil {
									t.Logf("Target error: %v", targetErr)
									t.Logf("Compat error: %v", compatErr)

									// error messages are intentionally not compared
									AssertMatchesCommandError(t, compatErr, targetErr)

									return
								}
								require.NoError(t, compatErr, "compat error; target returned no error")

								targetRes := FetchAll(t, ctx, targetCursor)
								compatRes := FetchAll(t, ctx, compatCursor)

								t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatRes))
								t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetRes))
								AssertEqualDocumentsSlice(t, compatRes, targetRes)
							}
						})
					}
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				if tc.failsForFerretDB != "" {
					return
				}

				assert.True(tt, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				assert.False(tt, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				tt.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

// updateCurrentDateCompatTestCase describes update current date compatibility test case.
type updateCurrentDateCompatTestCase struct {
	keys       []string                 // keys to check after the update that we expect being updated by the operation
	update     bson.D                   // required
	filter     bson.D                   // defaults to bson.D{{"_id", id}}
	resultType compatTestCaseResultType // defaults to nonEmptyResult

	failsForFerretDB   string
	failsProvidersDocs []struct {
		provider shareddata.Provider
		ids      []string
	} // use only if failsForFerretDB is set, defaults to all providers and all documents
}

// testUpdateCurrentDateCompat tests update compatibility test cases for current date.
// It checks current date in compat and target are within acceptable difference.
func testUpdateCurrentDateCompat(tt *testing.T, testCases map[string]updateCurrentDateCompatTestCase) {
	tt.Helper()

	maxDifference := 2 * time.Minute

	for name, tc := range testCases {
		tt.Run(name, func(tt *testing.T) {
			tt.Helper()

			tt.Parallel()

			// Use per-test setup because updates modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(tt)

			update := tc.update
			require.NotNil(tt, update, "`update` must be set")

			keys := tc.keys

			failsProviders := make(map[string][]string)
			for _, p := range tc.failsProvidersDocs {
				failsProviders[p.provider.Name()] = p.ids
			}

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				tt.Run(targetCollection.Name(), func(tt *testing.T) {
					tt.Helper()

					allDocs := FindAll(tt, ctx, targetCollection)

					for _, doc := range allDocs {
						id, ok := doc.Map()["_id"]
						require.True(tt, ok)

						idstr := fmt.Sprint(id)

						tt.Run(idstr, func(tt *testing.T) {
							// a workaround to get provider name by using the part after last `_`,
							// e.g. `Doubles` from `TestQueryArrayCompatElemMatch_Doubles`
							str := strings.Split(targetCollection.Name(), "_")
							providerName := str[len(str)-1]

							failsForDoc := len(tc.failsProvidersDocs) == 0 || slices.Contains(failsProviders[providerName], idstr)

							var t testing.TB = tt

							if tc.failsForFerretDB != "" && failsForDoc {
								t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
							}

							t.Helper()

							filter := tc.filter
							if tc.filter == nil {
								filter = bson.D{{"_id", id}}
							}

							var targetUpdateRes, compatUpdateRes *mongo.UpdateResult
							var targetErr, compatErr error

							targetUpdateRes, targetErr = targetCollection.UpdateOne(ctx, filter, update)
							compatUpdateRes, compatErr = compatCollection.UpdateOne(ctx, filter, update)

							if targetErr != nil {
								t.Logf("Target error: %v", targetErr)
								// AssertMatchesWriteError compares error types and codes, it does not compare messages.
								AssertMatchesWriteError(t, compatErr, targetErr)
							} else {
								require.NoError(t, compatErr, "compat error; target returned no error")
							}

							if pointer.Get(targetUpdateRes).ModifiedCount > 0 || pointer.Get(compatUpdateRes).ModifiedCount > 0 {
								nonEmptyResults = true
							}

							assert.Equal(t, compatUpdateRes, targetUpdateRes)

							var targetFindRes, compatFindRes bson.D
							require.NoError(t, targetCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&targetFindRes))
							require.NoError(t, compatCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&compatFindRes))

							for _, k := range keys {
								var expectedV, actualV any

								for _, el := range compatFindRes {
									if el.Key == k {
										expectedV = el.Value
									}
								}

								for i, el := range targetFindRes {
									if el.Key == k {
										actualV = el.Value

										el.Value = expectedV
										targetFindRes[i] = el
									}
								}

								require.NotNil(t, expectedV)
								require.NotNil(t, actualV)
								require.IsType(t, expectedV, actualV)

								switch actualV := actualV.(type) {
								case primitive.DateTime:
									assert.WithinDuration(t, expectedV.(primitive.DateTime).Time(), actualV.Time(), maxDifference)

								case primitive.Timestamp:
									expectedT := expectedV.(primitive.Timestamp)
									assert.WithinDuration(t, time.Unix(int64(expectedT.T), 0), time.Unix(int64(actualV.T), 0), maxDifference)

								default:
									assert.Fail(t, fmt.Sprintf("expected primitive.DateTime or primitive.Timestamp, got %T %T", expectedV, actualV))
								}
							}

							AssertEqualDocuments(t, compatFindRes, targetFindRes)
						})
					}
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				if tc.failsForFerretDB != "" {
					return
				}

				assert.True(tt, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				assert.False(tt, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				tt.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestUpdateCommandCompat(t *testing.T) {
	_, collection := setup.Setup(t)
	nonExistentDB := collection.Database().Client().Database("non-existent")

	for name, tc := range map[string]struct {
		db    *mongo.Database // defaults to targetCollection.Database() and compatCollection.Database()
		cName string
	}{
		"NonExistentDB": {
			db:    nonExistentDB,
			cName: "nonExistentCollection",
		},
		"NonExistentCollection": {
			cName: "nonExistentCollection",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.Composites},
				AddNonExistentCollection: true,
			})
			ctx, targetCollection, compatCollection := s.Ctx, s.TargetCollections[0], s.CompatCollections[0]

			targetDB := targetCollection.Database()
			compatDB := compatCollection.Database()
			if tc.db != nil {
				targetDB = tc.db
				compatDB = tc.db
			}

			q := bson.D{{"_id", "array"}}
			command := bson.D{
				{"update", tc.cName},
				{"updates", bson.A{bson.D{
					{"q", q},
					{"u", bson.D{{"$set", bson.D{{"v", float64(1)}}}}},
				}}},
			}

			var targetRes, compatRes bson.D
			err := targetDB.RunCommand(ctx, command).Decode(&targetRes)
			require.NoError(t, err)

			err = compatDB.RunCommand(ctx, command).Decode(&compatRes)
			require.NoError(t, err)

			require.NoError(t, targetCollection.FindOne(ctx, q).Decode(&targetRes))
			require.NoError(t, compatCollection.FindOne(ctx, q).Decode(&compatRes))

			assert.Equal(t, compatRes, targetRes)
		})
	}
}

func TestUpdateCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"UpdateEmptyDocument": {
			update:     bson.D{},
			resultType: emptyResult,
		},
		"ReplaceSimple": {
			replace:          bson.D{{"v", "foo"}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/489",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Strings, ids: []string{"string", "string-duplicate"}},
				{provider: shareddata.Scalars, ids: []string{"string"}},
			},
		},
		"ReplaceEmpty": {
			replace:          bson.D{{"v", ""}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/489",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Strings, ids: []string{"string-empty"}},
				{provider: shareddata.Scalars, ids: []string{"string-empty"}},
			},
		},
		"ReplaceNull": {
			replace:          bson.D{{"v", nil}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/489",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Strings, ids: []string{"string-null"}},
				{provider: shareddata.Scalars, ids: []string{"null", "string-null"}},
				{provider: shareddata.Nulls, ids: []string{"null"}},
				{provider: shareddata.Bools, ids: []string{"bool-null"}},
				{provider: shareddata.Mixed, ids: []string{"null"}},
				{provider: shareddata.Timestamps, ids: []string{"timestamp-null"}},
				{provider: shareddata.DateTimes, ids: []string{"datetime-null"}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.Regexes, ids: []string{"regex-null"}},
				{provider: shareddata.Binaries, ids: []string{"binary-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.Doubles, ids: []string{"double-null"}},
				{provider: shareddata.ObjectIDs, ids: []string{"objectid-null"}},
			},
		},
		"ReplaceEmptyDocument": {
			replace: bson.D{},
		},
		"ReplaceNonExistentUpsert": {
			filter:      bson.D{{"non-existent", "no-match"}},
			replace:     bson.D{{"_id", "new"}},
			replaceOpts: options.Replace().SetUpsert(true),
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"decimal128", "decimal128-int", "decimal128-int-zero", "decimal128-zero", "decimal128-double", "decimal128-whole",
					"unset", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles, ids: []string{
					"double-whole", "double-zero", "double-smallest", "double-big", "double-big-plus", "double-big-minus", "double-prec-max",
					"double-prec-max-plus", "double-prec-max-plus-two", "double-prec-max-minus", "double-neg-big", "double-neg-big-plus",
					"double-neg-big-minus", "double-prec-min", "double-prec-min-plus", "double-prec-min-minus", "double-prec-min-minus-two",
					"double-null", "double-1", "double-2", "double-3", "double-4", "double-max-overflow", "double-min-overflow",
				}},
				{provider: shareddata.Decimal128s, ids: []string{
					"decimal128-int", "decimal128-int-zero", "decimal128-max-exp", "decimal128-max-exp-sig",
					"decimal128-max-sig", "decimal128-min-exp", "decimal128-min-exp-sig", "decimal128-whole", "decimal128-zero",
				}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-max"}},
				{provider: shareddata.SmallDoubles, ids: []string{"double-whole", "double-1", "double-2", "double-3"}},
				{provider: shareddata.Strings, ids: []string{"string-double", "string-whole", "string-empty", "string-duplicate", "string-null"}},
				{provider: shareddata.Binaries, ids: []string{"binary-empty", "binary-null"}},
				{provider: shareddata.Bools, ids: []string{"bool-true", "bool-null"}},
				{provider: shareddata.DateTimes, ids: []string{"datetime-epoch", "datetime-year-min", "datetime-year-max", "datetime-null"}},
				{provider: shareddata.Regexes, ids: []string{"regex-empty", "regex-null"}},
				{provider: shareddata.Int32s, ids: []string{"int32-zero", "int32-max", "int32-min", "int32-1", "int32-2", "int32-3"}},
				{provider: shareddata.Timestamps, ids: []string{"timestamp-i", "timestamp-null"}},
				{provider: shareddata.Int64s, ids: []string{
					"int64-zero", "int64-max", "int64-min", "int64-1", "int64-2", "int64-3", "int64-big", "int64-big-plus", "int64-big-minus",
					"int64-prec-max", "int64-prec-max-plus", "int64-prec-max-plus-two", "int64-prec-max-minus", "int64-neg-big", "int64-neg-big-plus",
					"int64-neg-big-minus", "int64-prec-min", "int64-prec-min-plus", "int64-prec-min-minus", "int64-prec-min-minus-two",
				}},
				{provider: shareddata.ObjectIDs, ids: []string{"objectid-empty", "objectid-null"}},
				{provider: shareddata.ObjectIDKeys, ids: []string{fmt.Sprint(
					primitive.ObjectID{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11},
				)}},
				{provider: shareddata.Composites, ids: []string{
					"array-documents", "array-empty", "document", "document-composite", "document-composite-numerical-field-name",
					"document-composite-reverse", "document-empty", "document-null", "array-composite", "array-null",
					"array-numbers-asc", "array-strings-desc", "array-three", "array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{
					"document-double-whole", "document-double-zero", "document-double-max", "document-double-smallest",
					"document-double-big", "document-double-empty", "document-double-null",
				}},
				{provider: shareddata.DocumentsStrings, ids: []string{
					"document-string-double", "document-string-whole", "document-string-empty-str", "document-string-empty", "document-string-nil",
				}},
				{provider: shareddata.DocumentsDeeplyNested, ids: []string{"four", "three", "two"}},
				{provider: shareddata.DocumentsDocuments, ids: []string{
					fmt.Sprint(primitive.ObjectID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
					fmt.Sprint(primitive.ObjectID{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02}),
					fmt.Sprint(primitive.ObjectID{0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03}),
					fmt.Sprint(primitive.ObjectID{0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04}),
				}},
				{provider: shareddata.ArrayStrings, ids: []string{
					"array-string-duplicate", "array-string-numbers", "array-string-with-nil", "array-string-empty",
				}},
				{provider: shareddata.ArrayDoubles, ids: []string{
					"array-double-desc", "array-double-duplicate", "array-double-empty",
					"array-double-big-plus", "array-double-prec-max", "array-double-prec-max-plus",
				}},
				{provider: shareddata.ArrayInt32s, ids: []string{
					"array-int32-one", "array-int32-two", "array-int32-three",
					"array-int32-six",
				}},
				{provider: shareddata.Mixed, ids: []string{"array-null", "null", "unset"}},
				{provider: shareddata.ArrayDocuments, ids: []string{
					"array-documents-nested-duplicate", "array-three-documents", "array-two-documents",
				}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents-two-fields", "document"}},
			},
		},
		"UpdateNonExistentUpsert": {
			filter:     bson.D{{"_id", "non-existent"}},
			update:     bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			updateOpts: options.Update().SetUpsert(true),
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateCompatArray(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"ReplaceDocumentFilter": {
			filter:  bson.D{{"v", bson.D{{"$eq", true}}}},
			replace: bson.D{{"foo", int32(1)}},
		},
		"ReplaceDotNotationFilter": {
			filter:  bson.D{{"v.array.0", bson.D{{"$eq", int32(42)}}}, {"_id", "document-composite"}},
			replace: bson.D{{"replacement-value", int32(1)}},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateCompatMultiFlagCommand(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCommandCompatTestCase{
		"False": {
			filter: bson.D{{"v", int32(42)}},
			update: bson.D{{"$set", bson.D{{"v", int32(43)}}}},
			multi:  false,
		},
		"True": {
			filter: bson.D{{"v", int32(42)}},
			update: bson.D{{"$set", bson.D{{"v", int32(43)}}}},
			multi:  true,
		},
		"String": {
			filter:     bson.D{{"v", int32(42)}},
			update:     bson.D{{"$set", bson.D{{"v", int32(43)}}}},
			multi:      "false",
			resultType: emptyResult,
		},
		"Int": {
			filter:     bson.D{{"v", int32(42)}},
			update:     bson.D{{"$set", bson.D{{"v", int32(43)}}}},
			multi:      int32(0),
			resultType: emptyResult,
		},
		"TrueEmptyDocument": {
			update:     bson.D{},
			multi:      true,
			resultType: emptyResult,
		},
		"FalseEmptyDocument": {
			update: bson.D{},
			multi:  false,
		},
	}

	testUpdateCommandCompat(t, testCases)
}

func TestReplaceKeepOrderCompat(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers: shareddata.Providers{shareddata.Int32s},
	})

	ctx := s.Ctx
	targetCollection := s.TargetCollections[0]
	compatCollection := s.CompatCollections[0]

	replace := bson.D{{"_id", "arr"}, {"c", int32(1)}, {"b", int32(2)}, {"a", int32(3)}}
	filter := bson.D{{"_id", "arr"}}

	_, err := targetCollection.InsertOne(ctx, filter)
	require.NoError(t, err)
	_, err = compatCollection.InsertOne(ctx, filter)
	require.NoError(t, err)

	_, err = targetCollection.ReplaceOne(ctx, filter, replace)
	require.NoError(t, err)
	_, err = compatCollection.ReplaceOne(ctx, filter, replace)
	require.NoError(t, err)

	targetResult := targetCollection.FindOne(ctx, filter)
	require.NoError(t, targetResult.Err())
	compatResult := compatCollection.FindOne(ctx, filter)
	require.NoError(t, compatResult.Err())

	var targetDoc, compatDoc bson.D
	require.NoError(t, targetResult.Decode(&targetDoc))
	require.NoError(t, compatResult.Decode(&compatDoc))

	assert.Equal(t, compatDoc, targetDoc)
}

func TestUpdateCompatReplacementDoc(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCommandCompatTestCase{
		"Basic": {
			update: bson.D{{"v", int32(43)}},
		},
		"EmptyDoc": {
			update: bson.D{},
		},
		"FilterAndUpsertTrue": {
			filter:           bson.D{{"_id", "non-existent"}},
			update:           bson.D{{"v", int32(43)}},
			upsert:           true,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/359",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"decimal128", "decimal128-int", "decimal128-int-zero", "decimal128-zero", "decimal128-double", "decimal128-whole",
					"unset", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Decimal128s, ids: []string{
					"decimal128-int", "decimal128-int-zero", "decimal128-max-exp", "decimal128-max-exp-sig",
					"decimal128-max-sig", "decimal128-min-exp", "decimal128-min-exp-sig", "decimal128-whole", "decimal128-zero",
				}},
				{provider: shareddata.Doubles, ids: []string{
					"double-whole", "double-zero", "double-smallest", "double-big", "double-big-plus", "double-big-minus", "double-prec-max",
					"double-prec-max-plus", "double-prec-max-plus-two", "double-prec-max-minus", "double-neg-big", "double-neg-big-plus",
					"double-neg-big-minus", "double-prec-min", "double-prec-min-plus", "double-prec-min-minus", "double-prec-min-minus-two",
					"double-null", "double-1", "double-2", "double-3", "double-4", "double-max-overflow", "double-min-overflow",
				}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-max"}},
				{provider: shareddata.SmallDoubles, ids: []string{"double-whole", "double-1", "double-2", "double-3"}},
				{provider: shareddata.Strings, ids: []string{"string-double", "string-whole", "string-empty", "string-duplicate", "string-null"}},
				{provider: shareddata.Binaries, ids: []string{"binary-empty", "binary-null"}},
				{provider: shareddata.Bools, ids: []string{"bool-true", "bool-null"}},
				{provider: shareddata.DateTimes, ids: []string{"datetime-epoch", "datetime-year-min", "datetime-year-max", "datetime-null"}},
				{provider: shareddata.Regexes, ids: []string{"regex-empty", "regex-null"}},
				{provider: shareddata.Int32s, ids: []string{"int32-zero", "int32-max", "int32-min", "int32-1", "int32-2", "int32-3"}},
				{provider: shareddata.Timestamps, ids: []string{"timestamp-i", "timestamp-null"}},
				{provider: shareddata.Int64s, ids: []string{
					"int64-zero", "int64-max", "int64-min", "int64-1", "int64-2", "int64-3", "int64-big", "int64-big-plus", "int64-big-minus",
					"int64-prec-max", "int64-prec-max-plus", "int64-prec-max-plus-two", "int64-prec-max-minus", "int64-neg-big", "int64-neg-big-plus",
					"int64-neg-big-minus", "int64-prec-min", "int64-prec-min-plus", "int64-prec-min-minus", "int64-prec-min-minus-two",
				}},
				{provider: shareddata.ObjectIDs, ids: []string{"objectid-empty", "objectid-null"}},
				{provider: shareddata.ObjectIDKeys, ids: []string{fmt.Sprint(primitive.ObjectID{
					0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11,
				})}},
				{provider: shareddata.Composites, ids: []string{
					"array-documents", "array-empty", "document", "document-composite", "document-composite-numerical-field-name",
					"document-composite-reverse", "document-empty", "document-null", "array-composite", "array-null",
					"array-numbers-asc", "array-strings-desc", "array-three", "array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{
					"document-double-whole", "document-double-zero", "document-double-max", "document-double-smallest",
					"document-double-big", "document-double-empty", "document-double-null",
				}},
				{provider: shareddata.DocumentsStrings, ids: []string{
					"document-string-double", "document-string-whole", "document-string-empty-str", "document-string-empty", "document-string-nil",
				}},
				{provider: shareddata.DocumentsDeeplyNested, ids: []string{"four", "three", "two"}},
				{provider: shareddata.DocumentsDocuments, ids: []string{
					fmt.Sprint(primitive.ObjectID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
					fmt.Sprint(primitive.ObjectID{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02}),
					fmt.Sprint(primitive.ObjectID{0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03}),
					fmt.Sprint(primitive.ObjectID{0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04}),
				}},
				{provider: shareddata.ArrayStrings, ids: []string{
					"array-string-duplicate", "array-string-numbers", "array-string-with-nil", "array-string-empty",
				}},
				{provider: shareddata.ArrayDoubles, ids: []string{
					"array-double-desc", "array-double-duplicate", "array-double-empty",
					"array-double-big-plus", "array-double-prec-max", "array-double-prec-max-plus",
				}},
				{provider: shareddata.ArrayInt32s, ids: []string{
					"array-int32-one", "array-int32-two", "array-int32-three",
					"array-int32-six",
				}},
				{provider: shareddata.Mixed, ids: []string{"array-null", "null", "unset"}},
				{provider: shareddata.ArrayDocuments, ids: []string{
					"array-documents-nested-duplicate", "array-three-documents", "array-two-documents",
				}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents-two-fields", "document"}},
			},
		},
		"WithUpdateOp": {
			update:     bson.D{{"v", int32(43)}, {"$set", bson.D{{"test", int32(0)}}}},
			resultType: emptyResult,
		},
		"SameId": {
			filter: bson.D{{"_id", "int32"}},
			update: bson.D{{"_id", "int32"}, {"v", int32(43)}},
		},
		"DifferentId": {
			filter: bson.D{{"_id", "int32"}},
			update: bson.D{{"_id", "non-existent"}, {"v", int32(43)}},
		},
	}

	testUpdateCommandCompat(t, testCases)
}
