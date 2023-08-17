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

package query_and_write_ops

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

type insertCompatTestCase struct {
	insert     []any                                // required, slice of bson.D to be insert
	ordered    bool                                 // defaults to false
	resultType integration.CompatTestCaseResultType // defaults to NonEmptyResult

	failsForSQLite string // optional, if set, the case is expected to fail for SQLite due to given issue
}

// testInsertCompat tests insert compatibility test cases.
func testInsertCompat(tt *testing.T, testCases map[string]insertCompatTestCase) {
	tt.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		tt.Run(name, func(tt *testing.T) {
			tt.Helper()
			tt.Parallel()

			tt.Run("InsertOne", func(tt *testing.T) {
				tt.Helper()
				tt.Parallel()

				ctx, targetCollections, compatCollections := setup.SetupCompat(tt)

				insert := tc.insert
				require.NotEmpty(tt, insert, "insert should be set")

				for i := range targetCollections {
					targetCollection := targetCollections[i]
					compatCollection := compatCollections[i]
					tt.Run(targetCollection.Name(), func(tt *testing.T) {
						tt.Helper()

						var t testtb.TB = tt
						if tc.failsForSQLite != "" {
							t = setup.FailsForSQLite(tt, tc.failsForSQLite)
						}

						for _, doc := range insert {
							targetInsertRes, targetErr := targetCollection.InsertOne(ctx, doc)
							compatInsertRes, compatErr := compatCollection.InsertOne(ctx, doc)

							if targetErr != nil {
								switch targetErr := targetErr.(type) { //nolint:errorlint // don't inspect error chain
								case mongo.WriteException:
									integration.AssertMatchesWriteError(t, compatErr, targetErr)
								case mongo.BulkWriteException:
									integration.AssertMatchesBulkException(t, compatErr, targetErr)
								default:
									assert.Equal(t, compatErr, targetErr)
								}

								continue
							}

							require.NoError(t, compatErr, "compat error; target returned no error")
							require.Equal(t, compatInsertRes, targetInsertRes)
						}

						targetFindRes := integration.FindAll(t, ctx, targetCollection)
						compatFindRes := integration.FindAll(t, ctx, compatCollection)

						require.Equal(t, len(compatFindRes), len(targetFindRes))

						for i := range compatFindRes {
							integration.AssertEqualDocuments(t, compatFindRes[i], targetFindRes[i])
						}
					})
				}
			})

			tt.Run("InsertMany", func(tt *testing.T) {
				tt.Helper()
				tt.Parallel()

				ctx, targetCollections, compatCollections := setup.SetupCompat(tt)

				insert := tc.insert
				require.NotEmpty(tt, insert, "insert should be set")

				var NonEmptyResults bool
				for i := range targetCollections {
					targetCollection := targetCollections[i]
					compatCollection := compatCollections[i]
					tt.Run(targetCollection.Name(), func(tt *testing.T) {
						tt.Helper()

						var t testtb.TB = tt
						if tc.failsForSQLite != "" {
							t = setup.FailsForSQLite(tt, tc.failsForSQLite)
						}

						opts := options.InsertMany().SetOrdered(tc.ordered)
						targetInsertRes, targetErr := targetCollection.InsertMany(ctx, insert, opts)
						compatInsertRes, compatErr := compatCollection.InsertMany(ctx, insert, opts)

						// If the result contains inserted ids, we consider the result non-empty.
						if (compatInsertRes != nil && len(compatInsertRes.InsertedIDs) > 0) ||
							(targetInsertRes != nil && len(targetInsertRes.InsertedIDs) > 0) {
							NonEmptyResults = true
						}

						if targetErr != nil {
							switch targetErr := targetErr.(type) { //nolint:errorlint // don't inspect error chain
							case mongo.WriteException:
								integration.AssertMatchesWriteError(t, compatErr, targetErr)
							case mongo.BulkWriteException:
								integration.AssertMatchesBulkException(t, compatErr, targetErr)
							default:
								assert.Equal(t, compatErr, targetErr)
							}

							return
						}

						require.NoError(t, compatErr, "compat error; target returned no error")
						require.Equal(t, compatInsertRes, targetInsertRes)

						targetFindRes := integration.FindAll(t, ctx, targetCollection)
						compatFindRes := integration.FindAll(t, ctx, compatCollection)

						require.Equal(t, len(compatFindRes), len(targetFindRes))

						for i := range compatFindRes {
							integration.AssertEqualDocuments(t, compatFindRes[i], targetFindRes[i])
						}
					})
				}

				var t testtb.TB = tt
				if tc.failsForSQLite != "" {
					t = setup.FailsForSQLite(tt, tc.failsForSQLite)
				}

				switch tc.resultType {
				case integration.NonEmptyResult:
					assert.True(t, NonEmptyResults, "expected non-empty results")
				case integration.EmptyResult:
					assert.False(t, NonEmptyResults, "expected empty results")
				default:
					t.Fatalf("unknown result type %v", tc.resultType)
				}
			})
		})
	}
}

func TestInsertCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]insertCompatTestCase{
		"Normal": {
			insert: []any{bson.D{{"_id", int32(42)}}},
		},

		"IDArray": {
			insert:         []any{bson.D{{"_id", bson.A{"foo", "bar"}}}},
			resultType:     integration.EmptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/2750",
		},
		"IDRegex": {
			insert:         []any{bson.D{{"_id", primitive.Regex{Pattern: "^regex$", Options: "i"}}}},
			resultType:     integration.EmptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/2750",
		},

		"OrderedAllErrors": {
			insert: []any{
				bson.D{{"_id", bson.A{"foo", "bar"}}},
				bson.D{{"_id", primitive.Regex{Pattern: "^regex$", Options: "i"}}},
			},
			ordered:        true,
			resultType:     integration.EmptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/2750",
		},
		"UnorderedAllErrors": {
			insert: []any{
				bson.D{{"_id", bson.A{"foo", "bar"}}},
				bson.D{{"_id", primitive.Regex{Pattern: "^regex$", Options: "i"}}},
			},
			ordered:        false,
			resultType:     integration.EmptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/2750",
		},

		"OrderedOneError": {
			insert: []any{
				bson.D{{"_id", "1"}},
				bson.D{{"_id", primitive.Regex{Pattern: "^regex$", Options: "i"}}},
				bson.D{{"_id", "2"}},
			},
			ordered:        true,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/2750",
		},
		"UnorderedOneError": {
			insert: []any{
				bson.D{{"_id", "1"}},
				bson.D{{"_id", "1"}}, // to test duplicate key error
				bson.D{{"_id", primitive.Regex{Pattern: "^regex$", Options: "i"}}},
				bson.D{{"_id", "2"}},
			},
			ordered:        false,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/2750",
		},
	}

	testInsertCompat(t, testCases)
}
