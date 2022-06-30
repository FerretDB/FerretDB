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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestUpdateUpsert(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Composites)

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

func TestUpdateMany(t *testing.T) {
	t.Parallel()

	notModified := &mongo.UpdateResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: 1,
	}

	for name, tc := range map[string]struct {
		filter bson.D
		update bson.D
		stat   *mongo.UpdateResult
		res    bson.D
		err    *mongo.WriteError
		alt    string
	}{
		"SetSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$setOnInsert", bson.D{{"value", math.NaN()}}},
			},
			res: bson.D{{"_id", "test"}, {"foo", int32(12)}, {"value", math.NaN()}},
		},
		"SetTwoFields": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}, {"value", math.NaN()}}},
			},
			res: bson.D{{"_id", "test"}, {"foo", int32(12)}, {"value", math.NaN()}},
		},
		"IncTwoFields": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$inc", bson.D{{"foo", int32(12)}, {"value", int32(1)}}},
			},
			res: bson.D{{"_id", "test"}, {"foo", int32(12)}, {"value", int32(1)}},
		},
		"SetIncSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$inc", bson.D{{"foo", int32(1)}}},
				{"$setOnInsert", bson.D{{"value", math.NaN()}}},
			},
			err: &mongo.WriteError{
				Code:    40,
				Message: "Updating the path 'foo' would create a conflict at 'foo'",
			},
		},
		"UnsetString": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$unset", bson.D{{"value", int32(1)}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			res: bson.D{{"_id", "string"}},
		},
		"UnsetEmpty": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$unset", bson.D{}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
			res: bson.D{{"_id", "string"}, {"value", "foo"}},
		},
		"UnsetField": {
			filter: bson.D{{"_id", "document-composite"}},
			update: bson.D{{"$unset", bson.D{{"value", bson.D{{"array", int32(1)}}}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
			},
			res: bson.D{{"_id", "document-composite"}},
		},
		"UnsetEmptyArray": {
			filter: bson.D{{"_id", "document-composite"}},
			update: bson.D{{"$unset", bson.A{}}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$unset: []}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"UnknownOperator": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{{"$foo", bson.D{{"foo", int32(1)}}}},
			err: &mongo.WriteError{
				Code:    9,
				Message: "Unknown modifier: $foo. Expected a valid update modifier or pipeline-style update specified as an array",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)

			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", "string"}, {"value", "foo"}},
			})
			require.NoError(t, err)

			opts := options.Update().SetUpsert(true)
			var res *mongo.UpdateResult
			res, err = collection.UpdateOne(ctx, tc.filter, tc.update, opts)

			if tc.err != nil {
				require.Nil(t, tc.res)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
				return
			}

			require.NoError(t, err)
			res.UpsertedID = nil

			expectedRes := notModified
			if tc.stat != nil {
				expectedRes = tc.stat
			}
			assert.Equal(t, expectedRes, res)

			var actual bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.res, actual)
		})
	}
}
