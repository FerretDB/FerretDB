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

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestGroupDistinct(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"item", "abc"}},
		bson.D{{"_id", 2}, {"item", "jkl"}},
		bson.D{{"_id", 3}, {"item", "xyz"}},
		bson.D{{"_id", 4}, {"item", "xyz"}},
		bson.D{{"_id", 5}, {"item", "abc"}},
		bson.D{{"_id", 6}, {"item", "def"}},
		bson.D{{"_id", 7}, {"item", "def"}},
		bson.D{{"_id", 8}, {"item", "abc"}},
	})
	require.NoError(t, err)

	group := bson.D{{"$group", bson.D{{"_id", "$item"}}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{group})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []bson.D{
		bson.D{{"_id", "abc"}},
		bson.D{{"_id", "def"}},
		bson.D{{"_id", "jkl"}},
		bson.D{{"_id", "xyz"}},
	}, results)
}

func TestCountSumAndAverage(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"item", "abc"}, {"price", float32(10)}, {"quantity", int32(2)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-01T08:00:00Z"))}},
		bson.D{{"_id", 2}, {"item", "jkl"}, {"price", float32(20)}, {"quantity", int32(1)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-01T09:00:00Z"))}},
		bson.D{{"_id", 3}, {"item", "xyz"}, {"price", float32(5)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-15T09:00:00Z"))}},
		bson.D{{"_id", 4}, {"item", "xyz"}, {"price", float32(5)}, {"quantity", int32(20)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-04-04T11:21:39.736Z"))}},
		bson.D{{"_id", 5}, {"item", "abc"}, {"price", float32(10)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-04-04T21:23:13.331Z"))}},
		bson.D{{"_id", 6}, {"item", "def"}, {"price", float32(7.5)}, {"quantity", int32(5)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2015-06-04T05:08:13Z"))}},
		bson.D{{"_id", 7}, {"item", "def"}, {"price", float32(7.5)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2015-09-10T08:43:00Z"))}},
		bson.D{{"_id", 8}, {"item", "abc"}, {"price", float32(10)}, {"quantity", int32(5)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2016-02-06T20:20:13Z"))}},
	})
	require.NoError(t, err)

	// match := bson.D{{
	// 	"$match", bson.D{{
	// 		"date", bson.D{
	// 			{"$gte", time.Date(2014, time.March, 1, 0, 0, 0, 0, time.UTC)},
	// 			{"$lt", time.Date(2015, time.March, 1, 0, 0, 0, 0, time.UTC)},
	// 		},
	// 	}},
	// }}
	group := bson.D{{"$group", bson.D{
		{"_id", bson.D{{"$dateToString", bson.D{{"format", "%Y-%m-%d"}, {"date", "$date"}}}}},
		// {"totalSaleAmount", bson.D{{"$sum", bson.D{{"$multiply", bson.D{{"$price", "$quantity"}}}}}}},
		{"totalSaleAmount", bson.D{{"$sum", "$price"}}},
		{"averageQuantity", bson.D{{"$avg", "$quantity"}}},
		{"count", bson.D{{"$sum", 1}}},
	}}}
	// cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match, group})
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{group})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []bson.D{
		bson.D{{"_id", "2014-03-15"}, {"totalSaleAmount", int32(5)}, {"averageQuantity", int32(10)}, {"count", int32(1)}},
		bson.D{{"_id", "2014-03-01"}, {"totalSaleAmount", int32(30)}, {"averageQuantity", int32(1)}, {"count", int32(2)}},
		bson.D{{"_id", "2015-09-10"}, {"totalSaleAmount", int32(7)}, {"averageQuantity", int32(10)}, {"count", int32(1)}},
		bson.D{{"_id", "2016-02-06"}, {"totalSaleAmount", int32(10)}, {"averageQuantity", int32(5)}, {"count", int32(1)}},
		bson.D{{"_id", "2014-04-04"}, {"totalSaleAmount", int32(15)}, {"averageQuantity", int32(15)}, {"count", int32(2)}},
		bson.D{{"_id", "2015-06-04"}, {"totalSaleAmount", int32(7)}, {"averageQuantity", int32(5)}, {"count", int32(1)}},
	}, results)
}
