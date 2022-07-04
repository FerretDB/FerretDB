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

func TestMatch(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		match    bson.D
		expected []bson.D
	}{
		"Simple": {
			match: bson.D{{"a", 1}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"MultipleFields": {
			match: bson.D{{"a", 1}, {"b", 2}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
			},
		},
		"Nested": {
			match: bson.D{{"c", bson.D{{"name", "Felipe"}, {"age", 12}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)}, {"a", int32(2)}, {"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}}},
			},
		},
		"NestedAsFlat": {
			match: bson.D{{"c.name", "Felipe"}, {"c.age", 12}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)}, {"a", int32(2)}, {"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}}},
			},
		},
		"And": {
			match: bson.D{{"$and", bson.A{
				bson.D{{"a", 1}},
				bson.D{{"b", 2}},
			}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
			},
		},
		"Or": {
			match: bson.D{{"$or", bson.A{
				bson.D{{"a", 1}},
				bson.D{{"b", 8}},
			}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"GreaterThan": {
			match: bson.D{{"a", bson.D{{"$gt", 1}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)},
					{"a", int32(2)},
					{"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}},
				},
			},
		},
		"NotEqual": {
			match: bson.D{{"_id", bson.D{{"$ne", 3}}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"In": {
			match: bson.D{{"_id", bson.D{{"$in", bson.A{1, 2}}}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"NotIn": {
			match: bson.D{{"_id", bson.D{{"$nin", bson.A{1, 2}}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)}, {"a", int32(2)}, {"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}}},
			},
		},
		"Exists": {
			match: bson.D{{"c", bson.D{{"$exists", true}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)}, {"a", int32(2)}, {"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}}},
			},
		},
		"ExistsFalse": {
			match: bson.D{{"c", bson.D{{"$exists", false}}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"NestedExists": {
			match: bson.D{{"c.name", bson.D{{"$exists", true}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)}, {"a", int32(2)}, {"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}}},
			},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)

			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", 1}, {"a", 1}, {"b", 2}},
				bson.D{{"_id", 2}, {"a", 1}, {"b", 8}},
				bson.D{{"_id", 3}, {"a", 2}, {"b", 3}, {"c", bson.D{{"name", "Felipe"}, {"age", 12}}}},
			})
			require.NoError(t, err)

			match := bson.D{{"$match", tc.match}}
			cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match})
			require.NoError(t, err)

			var results []bson.D
			if err := cursor.All(ctx, &results); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expected, results)
		})
	}
}

func TestMatchAll(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"colors", bson.A{"white", "black"}}},
		bson.D{{"_id", 2}, {"colors", bson.A{"white", "blue"}}},
		bson.D{{"_id", 3}, {"colors", bson.A{"blue"}}},
		bson.D{{"_id", 4}, {"colors", bson.A{"white"}}},
	})
	require.NoError(t, err)

	match := bson.D{{"$match", bson.D{{"colors", bson.D{{"$all", bson.A{"white", "blue"}}}}}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []bson.D{
		bson.D{{"_id", int32(2)}, {"colors", bson.A{"white", "blue"}}},
	}, results)
}

func TestMatchRegex(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"color", "white"}},
		bson.D{{"_id", 2}, {"color", "black"}},
		bson.D{{"_id", 3}, {"color", "blue"}},
		bson.D{{"_id", 4}, {"color", "yellow"}},
	})
	require.NoError(t, err)

	match := bson.D{{"$match", bson.D{{"color", bson.D{{"$regex", "e$"}}}}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []bson.D{
		bson.D{{"_id", int32(1)}, {"color", "white"}},
		bson.D{{"_id", int32(3)}, {"color", "blue"}},
	}, results)
}

func TestCount(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"color", "white"}},
		bson.D{{"_id", 2}, {"color", "black"}},
		bson.D{{"_id", 3}, {"color", "blue"}},
		bson.D{{"_id", 4}, {"color", "yellow"}},
	})
	require.NoError(t, err)

	match := bson.D{{"$match", bson.D{{"color", bson.D{{"$regex", "e$"}}}}}}
	count := bson.D{{"$count", "cnt"}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match, count})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []bson.D{
		bson.D{{"cnt", int32(2)}},
	}, results)
}

func TestMatchBeforeGroup(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"color", "white"}, {"quantity", 1}},
		bson.D{{"_id", 2}, {"color", "black"}, {"quantity", 5}},
		bson.D{{"_id", 3}, {"color", "blue"}, {"quantity", 3}},
		bson.D{{"_id", 4}, {"color", "yellow"}, {"quantity", 2}},
	})
	require.NoError(t, err)

	match := bson.D{{"$match", bson.D{{"color", bson.D{{"$regex", "e$"}}}}}}
	group := bson.D{{"$group", bson.D{{"averageQuantity", bson.D{{"$avg", "$quantity"}}}}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match, group})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []bson.D{
		bson.D{{"averageQuantity", float64(2)}},
	}, results)
}

func TestSimpleSort(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"color", "white"}, {"quantity", 1}},
		bson.D{{"_id", 2}, {"color", "black"}, {"quantity", 5}},
		bson.D{{"_id", 3}, {"color", "blue"}, {"quantity", 3}},
		bson.D{{"_id", 4}, {"color", "yellow"}, {"quantity", 2}},
	})
	require.NoError(t, err)

	sort := bson.D{{"$sort", bson.D{{"quantity", -1}}}}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{sort})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []bson.D{
		bson.D{{"_id", int32(2)}, {"color", "black"}, {"quantity", int32(5)}},
		bson.D{{"_id", int32(3)}, {"color", "blue"}, {"quantity", int32(3)}},
		bson.D{{"_id", int32(4)}, {"color", "yellow"}, {"quantity", int32(2)}},
		bson.D{{"_id", int32(1)}, {"color", "white"}, {"quantity", int32(1)}},
	}, results)
}

// FIXME this is not working when we group then sort, json is true but the field is no longer json
func TestGroupAndSort(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", int32(1)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(2)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-01T08:00:00.000Z"))}},
		bson.D{{"_id", int32(2)}, {"item", "jkl"}, {"price", int32(20)}, {"quantity", int32(1)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-01T09:00:00.000Z"))}},
		bson.D{{"_id", int32(3)}, {"item", "xyz"}, {"price", int32(5)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-15T09:00:00.000Z"))}},
		bson.D{{"_id", int32(4)}, {"item", "xyz"}, {"price", int32(5)}, {"quantity", int32(20)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-04-04T11:21:39.736Z"))}},
		bson.D{{"_id", int32(5)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-04-04T21:23:13.331Z"))}},
		bson.D{{"_id", int32(6)}, {"item", "def"}, {"price", float64(7.5)}, {"quantity", int32(5)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2015-06-04T05:08:13.000Z"))}},
		bson.D{{"_id", int32(7)}, {"item", "def"}, {"price", float64(7.5)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2015-09-10T08:43:00.000Z"))}},
		bson.D{{"_id", int32(8)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(5)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2016-02-06T20:20:13.000Z"))}},
	})
	require.NoError(t, err)

	group := bson.D{{"$group", bson.D{{"_id", bson.D{{"$dateToString", bson.D{{"date", "$date"}, {"format", "%Y-%m-%d"}}}}}}}}
	sort := bson.D{{"$sort", bson.D{{"_id", int32(1)}}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{group, sort})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []bson.D{
		bson.D{{"_id", "2014-03-01"}},
		bson.D{{"_id", "2014-03-15"}},
		bson.D{{"_id", "2014-04-04"}},
		bson.D{{"_id", "2015-06-04"}},
		bson.D{{"_id", "2015-09-10"}},
		bson.D{{"_id", "2016-02-06"}},
	}, results)
}

func TestMatchAndSort(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", int32(1)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(2)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-01T08:00:00.000Z"))}},
		bson.D{{"_id", int32(2)}, {"item", "jkl"}, {"price", int32(20)}, {"quantity", int32(1)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-01T09:00:00.000Z"))}},
		bson.D{{"_id", int32(3)}, {"item", "xyz"}, {"price", int32(5)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-15T09:00:00.000Z"))}},
		bson.D{{"_id", int32(4)}, {"item", "xyz"}, {"price", int32(5)}, {"quantity", int32(20)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-04-04T11:21:39.736Z"))}},
		bson.D{{"_id", int32(5)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-04-04T21:23:13.331Z"))}},
		bson.D{{"_id", int32(6)}, {"item", "def"}, {"price", float64(7.5)}, {"quantity", int32(5)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2015-06-04T05:08:13.000Z"))}},
		bson.D{{"_id", int32(7)}, {"item", "def"}, {"price", float64(7.5)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2015-09-10T08:43:00.000Z"))}},
		bson.D{{"_id", int32(8)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(5)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2016-02-06T20:20:13.000Z"))}},
	})
	require.NoError(t, err)

	match := bson.D{{"$match", bson.D{{"price", bson.D{{"$gt", int32(8)}}}}}}
	sort := bson.D{{"$sort", bson.D{{"price", int32(1)}}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match, sort})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int32(1), results[0][0].Value)
	assert.Equal(t, int32(5), results[1][0].Value)
	assert.Equal(t, int32(8), results[2][0].Value)
	assert.Equal(t, int32(2), results[3][0].Value)
}

func TestMatchAndSortByDate(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", int32(1)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(2)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-01T08:00:00.000Z"))}},
		bson.D{{"_id", int32(2)}, {"item", "jkl"}, {"price", int32(20)}, {"quantity", int32(1)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-01T09:00:00.000Z"))}},
		bson.D{{"_id", int32(3)}, {"item", "xyz"}, {"price", int32(5)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-03-15T09:00:00.000Z"))}},
		bson.D{{"_id", int32(4)}, {"item", "xyz"}, {"price", int32(5)}, {"quantity", int32(20)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-04-04T11:21:39.736Z"))}},
		bson.D{{"_id", int32(5)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2014-04-04T21:23:13.331Z"))}},
		bson.D{{"_id", int32(6)}, {"item", "def"}, {"price", float64(7.5)}, {"quantity", int32(5)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2015-06-04T05:08:13.000Z"))}},
		bson.D{{"_id", int32(7)}, {"item", "def"}, {"price", float64(7.5)}, {"quantity", int32(10)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2015-09-10T08:43:00.000Z"))}},
		bson.D{{"_id", int32(8)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(5)}, {"date", must.NotFail(time.Parse(time.RFC3339, "2016-02-06T20:20:13.000Z"))}},
	})
	require.NoError(t, err)

	match := bson.D{{"$match", bson.D{{"price", bson.D{{"$gt", int32(8)}}}}}}
	sort := bson.D{{"$sort", bson.D{{"date", int32(1)}}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match, sort})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int32(1), results[0][0].Value)
	assert.Equal(t, int32(2), results[1][0].Value)
	assert.Equal(t, int32(5), results[2][0].Value)
	assert.Equal(t, int32(8), results[3][0].Value)
}

func TestMatchAndCount(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", int32(1)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(2)}},
		bson.D{{"_id", int32(2)}, {"item", "jkl"}, {"price", int32(20)}, {"quantity", int32(1)}},
		bson.D{{"_id", int32(3)}, {"item", "xyz"}, {"price", int32(5)}, {"quantity", int32(10)}},
		bson.D{{"_id", int32(4)}, {"item", "xyz"}, {"price", int32(5)}, {"quantity", int32(20)}},
		bson.D{{"_id", int32(5)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(10)}},
		bson.D{{"_id", int32(6)}, {"item", "def"}, {"price", float64(7.5)}, {"quantity", int32(5)}},
		bson.D{{"_id", int32(7)}, {"item", "def"}, {"price", float64(7.5)}, {"quantity", int32(10)}},
		bson.D{{"_id", int32(8)}, {"item", "abc"}, {"price", int32(10)}, {"quantity", int32(5)}},
	})
	require.NoError(t, err)

	match := bson.D{{"$match", bson.D{{"price", bson.D{{"$gt", int32(8)}}}}}}
	count := bson.D{{"$count", "count"}}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match, count})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []bson.D{bson.D{{"count", int32(4)}}}, results)
}

func TestInvalidSort(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	sort := bson.D{{"$sort", "count"}}
	_, err := collection.Aggregate(ctx, mongo.Pipeline{sort})
	require.Error(t, err)

	assert.Equal(t, "(BadValue) the $sort key specification must be an object", err.Error())
}
