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
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

type countCommandCompatTestCase struct {
	collectionName   any
	command          bson.D
	failsForFerretDB string
}

// testCountCommandCompat tests query compatibility test cases.
func testCountCommandCompat(t *testing.T, testCases map[string]countCommandCompatTestCase) {
	t.Helper()

	// Use shared setup because count queries can't modify data.
	//
	// Use read-only user.
	// TODO https://github.com/FerretDB/FerretDB/issues/1025
	ctx, targetCollections, compatCollections := setup.SetupCompat(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(tt *testing.T) {
					tt.Helper()

					var t testing.TB = tt

					if tc.failsForFerretDB != "" {
						t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
					}

					targetCollectionName := tc.collectionName
					compatCollectionName := tc.collectionName
					if tc.collectionName == nil {
						targetCollectionName = targetCollection.Name()
						compatCollectionName = compatCollection.Name()
					}

					targetCommand := append(
						bson.D{
							{"count", targetCollectionName},
						},
						tc.command...,
					)
					compatCommand := append(
						bson.D{
							{"count", compatCollectionName},
						},
						tc.command...,
					)

					targetResult := targetCollection.Database().RunCommand(ctx, targetCommand)
					compatResult := compatCollection.Database().RunCommand(ctx, compatCommand)

					targetErr := targetResult.Err()
					compatErr := compatResult.Err()

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					var targetRes, compatRes bson.D
					require.NoError(t, targetResult.Decode(&targetRes))
					require.NoError(t, compatResult.Decode(&compatRes))

					AssertEqualDocuments(t, compatRes, targetRes)

					targetCount := targetRes.Map()["n"].(int32)
					compatCount := compatRes.Map()["n"].(int32)

					require.Equal(t, compatCount, targetCount)
				})
			}
		})
	}
}

func TestCountCommandCompatErrors(t *testing.T) {
	t.Parallel()

	testCases := map[string]countCommandCompatTestCase{
		"Pass": {
			command: bson.D{
				{"query", bson.D{}},
			},
		},
		"CollectionDocument": {
			collectionName: bson.D{},
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionArray": {
			collectionName: primitive.A{},
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionDouble": {
			collectionName: 3.14,
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionBinary": {
			collectionName: primitive.Binary{},
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionObjectID": {
			collectionName: primitive.ObjectID{},
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionBool": {
			collectionName: true,
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionDate": {
			collectionName: time.Now(),
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionNull": {
			collectionName: nil,
			command: bson.D{
				{"query", bson.D{}},
			},
		},
		"CollectionRegex": {
			collectionName: primitive.Regex{Pattern: "/foo/"},
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionInt": {
			collectionName: int32(42),
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionTimestamp": {
			collectionName: primitive.Timestamp{},
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"CollectionLong": {
			collectionName: int64(42),
			command: bson.D{
				{"query", bson.D{}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/401",
		},
		"QueryArray": {
			command: bson.D{
				{"query", bson.A{}},
			},
		},
		"QueryInt": {
			command: bson.D{
				{"query", int32(42)},
			},
		},
	}

	testCountCommandCompat(t, testCases)
}
