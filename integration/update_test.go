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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// This file is for all remaining update tests.

func TestUpdateUpsert(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Composites)

	// this upsert inserts document
	filter := bson.D{{"foo", "bar"}}
	update := bson.D{{"$set", bson.D{{"foo", "baz"}}}}
	res, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	require.NoError(t, err)

	id := res.UpsertedID
	assert.NotEmpty(t, id)
	res.UpsertedID = nil
	expected := &mongo.UpdateResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: 1,
	}
	require.Equal(t, expected, res)

	// check inserted document
	var doc bson.D
	err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&doc)
	require.NoError(t, err)
	if !AssertEqualDocuments(t, bson.D{{"_id", id}, {"foo", "baz"}}, doc) {
		t.FailNow()
	}

	// this upsert updates document
	filter = bson.D{{"foo", "baz"}}
	update = bson.D{{"$set", bson.D{{"foo", "qux"}}}}
	res, err = collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	require.NoError(t, err)

	expected = &mongo.UpdateResult{
		MatchedCount:  1,
		ModifiedCount: 1,
		UpsertedCount: 0,
	}
	require.Equal(t, expected, res)

	// check updated document
	err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&doc)
	require.NoError(t, err)
	AssertEqualDocuments(t, bson.D{{"_id", id}, {"foo", "qux"}}, doc)
}

func TestMultiFlag(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		for name, tc := range map[string]struct {
			filter         bson.D
			update         bson.D
			multi          bool
			stat           bson.D
			expectedToFind int
		}{
			"False": {
				filter:         bson.D{{"v", int32(42)}},
				update:         bson.D{{"$set", bson.D{{"v", int32(43)}}}},
				multi:          false,
				stat:           bson.D{{"n", int32(1)}, {"nModified", int32(1)}, {"ok", float64(1)}},
				expectedToFind: 5,
			},
			"True": {
				filter:         bson.D{{"v", int32(42)}},
				update:         bson.D{{"$set", bson.D{{"v", int32(43)}}}},
				multi:          true,
				stat:           bson.D{{"n", int32(6)}, {"nModified", int32(6)}, {"ok", float64(1)}},
				expectedToFind: 0,
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

				command := bson.D{
					{"update", collection.Name()},
					{"updates", bson.A{
						bson.D{{"q", tc.filter}, {"u", tc.update}, {"multi", tc.multi}},
					}},
				}

				var result bson.D
				err := collection.Database().RunCommand(ctx, command).Decode(&result)
				require.NoError(t, err)

				AssertEqualDocuments(t, tc.stat, result)

				var after []bson.D
				cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
				require.NoError(t, err)

				err = cursor.All(ctx, &after)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedToFind, len(after))
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		for name, tc := range map[string]struct {
			filter     bson.D
			update     bson.D
			multi      any
			err        *mongo.CommandError
			altMessage string
		}{
			"String": {
				filter: bson.D{{"v", int32(42)}},
				update: bson.D{{"$set", bson.D{{"v", int32(43)}}}},
				multi:  "false",
				err: &mongo.CommandError{
					Code:    14,
					Name:    "TypeMismatch",
					Message: "BSON field 'update.updates.multi' is the wrong type 'string', expected type 'bool'",
				},
				altMessage: "BSON field 'multi' is the wrong type 'string', expected type 'bool'",
			},
			"Int": {
				filter: bson.D{{"v", int32(42)}},
				update: bson.D{{"$set", bson.D{{"v", int32(43)}}}},
				multi:  int32(0),
				err: &mongo.CommandError{
					Code:    14,
					Name:    "TypeMismatch",
					Message: "BSON field 'update.updates.multi' is the wrong type 'int', expected type 'bool'",
				},
				altMessage: "BSON field 'multi' is the wrong type 'int', expected type 'bool'",
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

				command := bson.D{
					{"update", collection.Name()},
					{"updates", bson.A{
						bson.D{{"q", tc.filter}, {"u", tc.update}, {"multi", tc.multi}},
					}},
				}

				var result bson.D
				err := collection.Database().RunCommand(ctx, command).Decode(&result)
				require.Error(t, err)

				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
			})
		}
	})
}

func TestUpdateNonExistingCollection(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	res, err := collection.Database().Collection("doesnotexist").UpdateOne(ctx, bson.D{}, bson.D{{"$set", bson.E{"foo", "bar"}}})
	require.NoError(t, err)

	assert.Equal(t, int64(0), res.MatchedCount)
}

func TestUpdateReplaceDocuments(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	stat := bson.D{{"n", int32(1)}, {"nModified", int32(1)}, {"ok", float64(1)}}

	for name, tc := range map[string]struct {
		update         bson.D
		expectedFilter bson.D
		expected       bson.D
	}{
		"Replace": {
			update: bson.D{
				{"q", bson.D{{"v", bson.D{{"$eq", true}}}}},
				{"u", bson.D{{"replacement-value", int32(1)}}},
			},
			expectedFilter: bson.D{{"_id", "bool-true"}},
			expected:       bson.D{{"_id", "bool-true"}, {"replacement-value", int32(1)}},
		},
		"ReplaceDotNotation": {
			update: bson.D{
				{"q", bson.D{{"v.array.0", bson.D{{"$eq", int32(42)}}}, {"_id", "document-composite"}}},
				{"u", bson.D{{"replacement-value", int32(1)}}},
			},
			expectedFilter: bson.D{{"_id", "document-composite"}},
			expected:       bson.D{{"_id", "document-composite"}, {"replacement-value", int32(1)}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Composites, shareddata.Scalars)

			res := collection.Database().RunCommand(
				ctx,
				bson.D{{"update", collection.Name()}, {"updates", bson.A{tc.update}}},
			)
			require.NoError(t, res.Err())

			var actual bson.D
			err := res.Decode(&actual)
			require.NoError(t, err)

			AssertEqualDocuments(t, stat, actual)

			err = collection.FindOne(ctx, tc.expectedFilter).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.expected, actual)
		})
	}
}
