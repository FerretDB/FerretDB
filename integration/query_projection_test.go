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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryProjection(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Composites}
	ctx, collection := setup(t, providers...)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{
			{"_id", "document-composite-2"},
			{"value", bson.A{
				bson.D{{"field", int32(42)}},
				bson.D{{"field", int32(44)}},
			}},
		},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		projection any
		filter     any
		expected   bson.D
	}{
		"FindProjectionInclusions": {
			filter: bson.D{{"_id", "document-composite"}},
			// TODO: https://github.com/FerretDB/FerretDB/issues/537
			projection: bson.D{{"foo", int32(1)}, {"42", true}},
			expected:   bson.D{{"_id", "document-composite"}},
		},
		"FindProjectionExclusions": {
			filter: bson.D{{"_id", "document-composite"}},
			// TODO: https://github.com/FerretDB/FerretDB/issues/537
			projection: bson.D{{"foo", int32(0)}, {"array", false}},
			expected:   bson.D{{"_id", "document-composite"}, {"value", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
		},
		"FindProjectionIDExclusion": {
			filter: bson.D{{"_id", "document-composite"}},
			// TODO: https://github.com/FerretDB/FerretDB/issues/537
			projection: bson.D{{"_id", false}, {"array", int32(1)}},
			expected:   bson.D{},
		},
		"ProjectionSliceNonArrayField": {
			filter:     bson.D{{"_id", "document"}},
			projection: bson.D{{"_id", bson.D{{"$slice", 1}}}},
			expected:   bson.D{{"_id", "document"}, {"value", bson.D{{"foo", int32(42)}}}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetProjection(tc.projection))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			require.Len(t, actual, 1)
			AssertEqualDocuments(t, tc.expected, actual[0])
		})
	}
}

func TestQueryProjectionElemMatch(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	testData := map[string]bson.D{
		"document-composite-2": {
			{"_id", "document-composite-2"},
			{"value", bson.A{
				bson.D{{"field", int32(40)}},
				bson.D{{"field", int32(41)}},
				bson.D{{"field", int32(42)}},
				bson.D{{"field", bson.D{{"field", int32(43)}}}},
				bson.D{{"field", bson.A{int32(44)}}},
			}},
		},
		"document-composite-3": {
			{"_id", "document-composite-3"},
			{"value", bson.A{
				bson.D{{"field", int32(10)}},
				bson.D{{"field", bson.D{{"field", int32(11)}}}},
				bson.D{{"field", bson.D{{"field", int32(12)}}}},
			}},
		},
	}

	_, err := collection.InsertMany(ctx, []any{
		testData["document-composite-2"], testData["document-composite-3"],
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filterIDs  bson.A
		projection bson.D
		expected   []bson.D
		err        *mongo.CommandError
	}{
		"ElemMatchMalformedArray": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.A{"field", "$eq", 42}}}}},
			err: &mongo.CommandError{
				Code:    31274,
				Name:    "Location31274",
				Message: "elemMatch: Invalid argument, object required, but got array",
			},
		},
		"ElemMatchMalformedString": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"field", bson.D{{"$elemMatch", ""}}}}}},
			err: &mongo.CommandError{
				Code:    31274,
				Name:    "Location31274",
				Message: "elemMatch: Invalid argument, object required, but got string",
			},
		},
		"ElemMatchNestedErr": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"field", bson.D{{"field", bson.D{{"$elemMatch", bson.D{{"$eq", 42}}}}}}}}}},
			err: &mongo.CommandError{
				Code:    31275,
				Name:    "Location31275",
				Message: "Cannot use $elemMatch projection on a nested field.",
			},
		},
		"ElemMatchNestedFieldEmpty": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"field", bson.D{{"$eq", 42}}}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}},
				{{"_id", "document-composite-3"}},
			},
		},
		"ElemMatchEq": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$eq", 42}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}, {"value", bson.A{bson.D{{"field", int32(42)}}}}},
				{{"_id", "document-composite-3"}},
			},
		},
		"ElemMatchEqArrayNotFound": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$eq", bson.A{42}}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}},
				{{"_id", "document-composite-3"}},
			},
		},
		"ElemMatchEqArrayFound": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$eq", bson.A{44}}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}, {"value", bson.A{bson.D{{"field", bson.A{int32(44)}}}}}},
				{{"_id", "document-composite-3"}},
			},
		},
		"ElemMatchNe": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$ne", 42}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}, {"value", bson.A{bson.D{{"field", int32(40)}}}}},
				{{"_id", "document-composite-3"}, {"value", bson.A{bson.D{{"field", int32(10)}}}}},
			},
		},
		"ElemMatchGt": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$gt", 13}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}, {"value", bson.A{bson.D{{"field", int32(40)}}}}},
				{{"_id", "document-composite-3"}},
			},
		},
		"ElemMatchGte": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$gte", 13}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}, {"value", bson.A{bson.D{{"field", int32(40)}}}}},
				{{"_id", "document-composite-3"}},
			},
		},
		"ElemMatchLt": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$lt", 10}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}},
				{{"_id", "document-composite-3"}},
			},
		},
		"ElemMatchLte": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$lte", 10}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}},
				{{"_id", "document-composite-3"}, {"value", bson.A{bson.D{{"field", int32(10)}}}}},
			},
		},
		"ElemMatchInErr": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$in", 41}}}}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$in needs an array",
			},
		},
		"ElemMatchIn": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$in", bson.A{41}}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}, {"value", bson.A{bson.D{{"field", int32(41)}}}}},
				{{"_id", "document-composite-3"}},
			},
		},
		"ElemMatchNin": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$nin", bson.A{40}}}}}}}}},
			expected: []bson.D{
				{{"_id", "document-composite-2"}, {"value", bson.A{bson.D{{"field", int32(41)}}}}},
				{{"_id", "document-composite-3"}, {"value", bson.A{bson.D{{"field", int32(10)}}}}},
			},
		},
		"ElemMatchNinErr": {
			filterIDs:  bson.A{"document-composite-3", "document-composite-2"},
			projection: bson.D{{"value", bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$nin", int32(40)}}}}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$nin needs an array",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(
				ctx,
				bson.D{{"_id", bson.D{{"$in", tc.filterIDs}}}},
				options.Find().SetProjection(tc.projection),
				options.Find().SetSort(bson.D{{"_id", 1}}),
			)

			if tc.err != nil {
				require.Nil(t, tc.expected)
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestQueryProjectionSlice(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)
	_, err := collection.InsertOne(ctx,
		bson.D{{"_id", "array"}, {"value", bson.A{1, 2, 3, 4}}},
	)
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		projection    bson.D
		expectedArray bson.A
		err           *mongo.CommandError
		altMessage    string
	}{
		"SingleArgDocument": {
			projection: bson.D{{"value", bson.D{
				{"$slice", bson.D{{"a", bson.D{{"b", 3}}}, {"b", "abcd"}}},
			}}},
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. The given syntax { $slice: { a: { b: 3 }, b: \"abcd\" } } " +
					"did not match the find() syntax because :: Location31273: " +
					"$slice only supports numbers and [skip, limit] arrays " +
					":: The given syntax did not match the expression $slice syntax. " +
					":: caused by :: Expression $slice takes at least 2 arguments, and at most 3, " +
					"but 1 were passed in.",
			},
			altMessage: "Invalid $slice syntax. The given syntax " +
				"did not match the find() syntax because :: Location31273: " +
				"$slice only supports numbers and [skip, limit] arrays " +
				":: The given syntax did not match the expression $slice syntax. " +
				":: caused by :: Expression $slice takes at least 2 arguments, and at most 3, " +
				"but 1 were passed in.",
		},
		"SingleArgString": {
			projection: bson.D{{"value", bson.D{{"$slice", "string"}}}},
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. The given syntax { $slice: \"string\" } " +
					"did not match the find() syntax because :: Location31273: " +
					"$slice only supports numbers and [skip, limit] arrays " +
					":: The given syntax did not match the expression $slice syntax. " +
					":: caused by :: Expression $slice takes at least 2 arguments, and at most 3, " +
					"but 1 were passed in.",
			},
			altMessage: "Invalid $slice syntax. The given syntax " +
				"did not match the find() syntax because :: Location31273: " +
				"$slice only supports numbers and [skip, limit] arrays " +
				":: The given syntax did not match the expression $slice syntax. " +
				":: caused by :: Expression $slice takes at least 2 arguments, and at most 3, " +
				"but 1 were passed in.",
		},
		"SkipIsString": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{"string", 5}}}}},
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: string",
			},
		},
		"LimitIsString": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{int32(2), "string"}}}}},
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: int",
			},
		},
		"ArgEmptyArr": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{}}}}},
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. The given syntax { $slice: [] } " +
					"did not match the find() syntax because :: Location31272: " +
					"$slice array argument should be of form [skip, limit] :: " +
					"The given syntax did not match the expression " +
					"$slice syntax. :: caused by :: " +
					"Expression $slice takes at least 2 arguments, and at most 3, but 0 were passed in.",
			},
			altMessage: "Invalid $slice syntax. The given syntax " +
				"did not match the find() syntax because :: Location31272: " +
				"$slice array argument should be of form [skip, limit] :: " +
				"The given syntax did not match the expression " +
				"$slice syntax. :: caused by :: " +
				"Expression $slice takes at least 2 arguments, and at most 3, but 0 were passed in.",
		},
		"ThreeArgs": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{"string", 2, 3}}}}},
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: string",
			},
		},
		"TooManyArgs": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{1, 2, 3, 4}}}}},
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. The given syntax { $slice: [ 1, 2, 3, 4 ] } " +
					"did not match the find() syntax because :: Location31272: " +
					"$slice array argument should be of form [skip, limit] :: " +
					"The given syntax did not match the expression " +
					"$slice syntax. :: caused by :: " +
					"Expression $slice takes at least 2 arguments, and at most 3, but 4 were passed in.",
			},
			altMessage: "Invalid $slice syntax. The given syntax " +
				"did not match the find() syntax because :: Location31272: " +
				"$slice array argument should be of form [skip, limit] :: " +
				"The given syntax did not match the expression " +
				"$slice syntax. :: caused by :: " +
				"Expression $slice takes at least 2 arguments, and at most 3, but 4 were passed in.",
		},
		"Int64SingleArg": {
			projection:    bson.D{{"value", bson.D{{"$slice", int64(2)}}}},
			expectedArray: bson.A{int32(1), int32(2)},
		},
		"PositiveSingleArg": {
			projection:    bson.D{{"value", bson.D{{"$slice", 2}}}},
			expectedArray: bson.A{int32(1), int32(2)},
		},
		"NegativeSingleArg": {
			projection:    bson.D{{"value", bson.D{{"$slice", -2}}}},
			expectedArray: bson.A{int32(3), int32(4)},
		},
		"SingleArgFloat": {
			projection:    bson.D{{"value", bson.D{{"$slice", 1.4}}}},
			expectedArray: bson.A{int32(1)},
		},
		"SkipFloat": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{-2.5, 2}}}}},
			expectedArray: bson.A{int32(3), int32(4)},
		},
		"LimitFloat": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{1, 2.8}}}}},
			expectedArray: bson.A{int32(2), int32(3)},
		},
		"PositiveSkip": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{1, 2}}}}},
			expectedArray: bson.A{int32(2), int32(3)},
		},
		"NegativeSkip": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{-3, 2}}}}},
			expectedArray: bson.A{int32(2), int32(3)},
		},
		"NegativeLimitSkipInt64": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{int64(3), -2}}}}},
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: long",
			},
		},
		"NegativeLimitSkipInt": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{3, -2}}}}},
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: int",
			},
		},
		"NegativeLimitSkipFloat": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{0.3, -2}}}}},
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: double",
			},
		},
		"ArgNaN": {
			projection:    bson.D{{"value", bson.D{{"$slice", math.NaN()}}}},
			expectedArray: bson.A{},
		},
		"ArgInf": {
			projection:    bson.D{{"value", bson.D{{"$slice", math.Inf(1)}}}},
			expectedArray: bson.A{int32(1), int32(2), int32(3), int32(4)},
		},
		"SingleArgNull": {
			projection: bson.D{{"value", bson.D{{"$slice", nil}}}},
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. " +
					"The given syntax { $slice: null } did not match the find() syntax " +
					"because :: Location31273: $slice only supports numbers and [skip, limit] arrays :: " +
					"The given syntax did not match the expression $slice syntax. :: caused by :: " +
					"Expression $slice takes at least 2 arguments, and at most 3, but 1 were passed in.",
			},
			altMessage: "Invalid $slice syntax. " +
				"The given syntax did not match the find() syntax " +
				"because :: Location31273: $slice only supports numbers and [skip, limit] arrays :: " +
				"The given syntax did not match the expression $slice syntax. :: caused by :: " +
				"Expression $slice takes at least 2 arguments, and at most 3, but 1 were passed in.",
		},
		"NullInArr": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{nil}}}}},
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. The given syntax { $slice: [ null ] } " +
					"did not match the find() syntax because :: Location31272: " +
					"$slice array argument should be of form [skip, limit] " +
					":: The given syntax did not match the expression $slice syntax. " +
					":: caused by :: Expression $slice takes at least 2 arguments, " +
					"and at most 3, but 1 were passed in.",
			},
			altMessage: "Invalid $slice syntax. The given syntax " +
				"did not match the find() syntax because :: Location31272: " +
				"$slice array argument should be of form [skip, limit] " +
				":: The given syntax did not match the expression $slice syntax. " +
				":: caused by :: Expression $slice takes at least 2 arguments, " +
				"and at most 3, but 1 were passed in.",
		},
		"NullInPair": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{2, nil}}}}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res := collection.FindOne(ctx, bson.D{}, options.FindOne().SetProjection(tc.projection))
			err = res.Err()
			if tc.err != nil {
				require.Nil(t, tc.expectedArray)
				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
				return
			}
			require.NoError(t, err)

			var actual bson.D
			err = res.Decode(&actual)
			require.NoError(t, err)

			if tc.expectedArray == nil {
				assert.Nil(t, actual.Map()["value"])
				return
			}
			assert.Equal(t, tc.expectedArray, actual.Map()["value"])
		})
	}
}
