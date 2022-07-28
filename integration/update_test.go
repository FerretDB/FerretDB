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
	t.Parallel()

	for name, tc := range map[string]struct {
		filter bson.D
		update bson.D
		multi  string
		stat   bson.D
	}{
		"MultiFalse": {
			filter: bson.D{{"foo", "x"}},
			update: bson.D{{"$set", bson.D{{"foo", "y"}}}},
			multi:  "false",
			stat:   bson.D{{"n", int32(1)}, {"nModified", int32(1)}, {"ok", float64(1)}},
		},
		"MultiTrue": {
			filter: bson.D{{"foo", "x"}},
			update: bson.D{{"$set", bson.D{{"foo", "y"}}}},
			multi:  "true",
			stat:   bson.D{{"n", int32(2)}, {"nModified", int32(2)}, {"ok", float64(1)}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t)

			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", "first"}, {"foo", "x"}},
				bson.D{{"_id", "second"}, {"foo", "x"}},
			})
			require.NoError(t, err)

			var actual bson.D

			if tc.multi == "true" {
				collection.UpdateMany(ctx, tc.filter, tc.update)
			} else {
				collection.UpdateOne(ctx, tc.filter, tc.update)
			}

			err = collection.FindOne(ctx, bson.D{{"_id", "first"}}).Decode(&actual)
			require.NoError(t, err)

			require.Equal(t, bson.D{{"_id", "first"}, {"foo", "y"}}, actual)

			err = collection.FindOne(ctx, bson.D{{"_id", "second"}}).Decode(&actual)
			require.NoError(t, err)

			var expected bson.D
			if tc.multi == "true" {
				expected = bson.D{{"_id", "second"}, {"foo", "y"}}
			} else {
				expected = bson.D{{"_id", "second"}, {"foo", "x"}}
			}
			require.Equal(t, expected, actual)
		})
	}
}
