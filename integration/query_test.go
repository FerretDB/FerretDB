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

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryUnknownFilterOperator(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	filter := bson.D{{"v", bson.D{{"$someUnknownOperator", 42}}}}
	errExpected := mongo.CommandError{Code: 2, Name: "BadValue", Message: "unknown operator: $someUnknownOperator"}
	_, err := collection.Find(ctx, filter)
	AssertEqualError(t, errExpected, err)
}

func TestQuerySort(t *testing.T) {
	t.Skip("TODO https://github.com/FerretDB/FerretDB/issues/457")

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		sort        bson.D
		expectedIDs []any
	}{
		"Asc": {
			sort: bson.D{{"v", 1}, {"_id", 1}},
			expectedIDs: []any{
				"array-empty",
				"array-null",
				"array-three",
				"array-three-reverse",
				"null",
				"double-nan",
				"int64-min",
				"int32-min",
				"double-negative-zero",
				"double-zero",
				"int32-zero",
				"int64-zero",
				"double-smallest",
				"array",
				"double-whole",
				"int32",
				"int64",
				"double",
				"int32-max",
				"int64-max",
				"double-max",
				"string-empty",
				"string-whole",
				"string-double",
				"string",
				"document-empty",
				"document-null",
				"document",
				"document-composite",
				"document-composite-reverse",
				"binary-empty",
				"binary",
				"objectid-empty",
				"objectid",
				"bool-false",
				"bool-true",
				"datetime-year-min",
				"datetime-epoch",
				"datetime",
				"datetime-year-max",
				"timestamp-i",
				"timestamp",
				"regex-empty",
				"regex",
			},
		},
		"Desc": {
			sort: bson.D{{"v", -1}, {"_id", 1}},
			expectedIDs: []any{
				"regex",
				"regex-empty",
				"timestamp",
				"timestamp-i",
				"datetime-year-max",
				"datetime",
				"datetime-epoch",
				"datetime-year-min",
				"bool-true",
				"bool-false",
				"objectid",
				"objectid-empty",
				"binary",
				"binary-empty",
				"document-composite-reverse",
				"document-composite",
				"document",
				"document-null",
				"document-empty",
				"array-three",
				"array-three-reverse",
				"string",
				"string-double",
				"string-whole",
				"string-empty",
				"double-max",
				"int64-max",
				"int32-max",
				"double",
				"array",
				"double-whole",
				"int32",
				"int64",
				"double-smallest",
				"double-negative-zero",
				"double-zero",
				"int32-zero",
				"int64-zero",
				"int32-min",
				"int64-min",
				"double-nan",
				"array-null",
				"null",
				"array-empty",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(tc.sort))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

// TODO: https://github.com/FerretDB/FerretDB/issues/636
func TestQuerySortValue(t *testing.T) {
	setup.SkipForTigris(t)

	ctx, collection := setup.Setup(t, shareddata.Scalars)

	for name, tc := range map[string]struct {
		sort        bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"AscValueScalar": {
			sort: bson.D{{"v", 1}, {"_id", 1}},
			expectedIDs: []any{
				"null",
				"double-nan",
				"int64-min",
				"int32-min",
				"double-negative-zero",
				"double-zero",
				"int32-zero",
				"int64-zero",
				"double-smallest",
				"double-whole",
				"int32",
				"int64",
				"double",
				"int32-max",
				"double-big",
				"int64-big",
				"int64-max",
				"double-max",
				"string-empty",
				"string-whole",
				"string-double",
				"string",
				"binary-empty",
				"binary",
				"objectid-empty",
				"objectid",
				"bool-false",
				"bool-true",
				"datetime-year-min",
				"datetime-epoch",
				"datetime",
				"datetime-year-max",
				"timestamp-i",
				"timestamp",
				"regex-empty",
				"regex",
			},
		},
		"DescValueScalar": {
			sort: bson.D{{"v", -1}, {"_id", 1}},
			expectedIDs: []any{
				"regex",
				"regex-empty",
				"timestamp",
				"timestamp-i",
				"datetime-year-max",
				"datetime",
				"datetime-epoch",
				"datetime-year-min",
				"bool-true",
				"bool-false",
				"objectid",
				"objectid-empty",
				"binary",
				"binary-empty",
				"string",
				"string-double",
				"string-whole",
				"string-empty",
				"double-max",
				"int64-max",
				"int64-big",
				"double-big",
				"int32-max",
				"double",
				"double-whole",
				"int32",
				"int64",
				"double-smallest",
				"double-negative-zero",
				"double-zero",
				"int32-zero",
				"int64-zero",
				"int32-min",
				"int64-min",
				"double-nan",
				"null",
			},
		},
		"BadSortValue": {
			sort: bson.D{{"v", 11}},
			err: &mongo.CommandError{
				Code:    15975,
				Name:    "Location15975",
				Message: "$sort key ordering must be 1 (for ascending) or -1 (for descending)",
			},
		},
		"BadSortZeroValue": {
			sort: bson.D{{"v", 0}},
			err: &mongo.CommandError{
				Code:    15975,
				Name:    "Location15975",
				Message: "$sort key ordering must be 1 (for ascending) or -1 (for descending)",
			},
		},
		"BadSortNullValue": {
			sort: bson.D{{"v", nil}},
			err: &mongo.CommandError{
				Code:    15974,
				Name:    "Location15974",
				Message: "Illegal key in $sort specification: v: null",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(tc.sort))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

func TestQueryCount(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command  any
		response int32
	}{
		"CountAllDocuments": {
			command:  bson.D{{"count", collection.Name()}},
			response: 47,
		},
		"CountExactlyOneDocument": {
			command: bson.D{
				{"count", collection.Name()},
				{"query", bson.D{{"v", true}}},
			},
			response: 1,
		},
		"CountExactlyOneDocumentWithIdFilter": {
			command: bson.D{
				{"count", collection.Name()},
				{"query", bson.D{{"_id", "bool-true"}}},
			},
			response: 1,
		},
		"CountArrays": {
			command: bson.D{
				{"count", collection.Name()},
				{"query", bson.D{{"v", bson.D{{"$type", "array"}}}}},
			},
			response: 6,
		},
		"CountNonExistingCollection": {
			command: bson.D{
				{"count", "doesnotexist"},
				{"query", bson.D{{"v", true}}},
			},
			response: 0,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()

			assert.Equal(t, float64(1), m["ok"])

			keys := CollectKeys(t, actual)
			assert.Contains(t, keys, "n")
			assert.Equal(t, tc.response, m["n"])
		})
	}
}

func TestQueryBadFindType(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command bson.D
		err     *mongo.CommandError
	}{
		"Document": {
			command: bson.D{
				{"find", bson.D{}},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type object",
			},
		},
		"Array": {
			command: bson.D{
				{"find", primitive.A{}},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type array",
			},
		},
		"Double": {
			command: bson.D{
				{"find", 3.14},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type double",
			},
		},
		"DoubleWhole": {
			command: bson.D{
				{"find", 42.0},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type double",
			},
		},
		"Binary": {
			command: bson.D{
				{"find", primitive.Binary{}},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type binData",
			},
		},
		"ObjectID": {
			command: bson.D{
				{"find", primitive.ObjectID{}},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type objectId",
			},
		},
		"Bool": {
			command: bson.D{
				{"find", true},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type bool",
			},
		},
		"Date": {
			command: bson.D{
				{"find", time.Now()},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type date",
			},
		},
		"Null": {
			command: bson.D{
				{"find", nil},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type null",
			},
		},
		"Regex": {
			command: bson.D{
				{"find", primitive.Regex{Pattern: "/foo/"}},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type regex",
			},
		},
		"Int": {
			command: bson.D{
				{"find", int32(42)},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type int",
			},
		},
		"Timestamp": {
			command: bson.D{
				{"find", primitive.Timestamp{}},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type timestamp",
			},
		},
		"Long": {
			command: bson.D{
				{"find", int64(42)},
				{"projection", bson.D{{"v", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type long",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.Error(t, err)
			AssertEqualError(t, *tc.err, err)
		})
	}
}

func TestQueryBadSortType(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command    bson.D
		err        *mongo.CommandError
		altMessage string
	}{
		"BadSortTypeDouble": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", "some"}}},
				{"sort", 42.13},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Expected field sortto be of type object",
			},
			altMessage: "Expected field sort to be of type object",
		},
		"BadSortType": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", "some"}}},
				{"sort", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Expected field sortto be of type object",
			},
			altMessage: "Expected field sort to be of type object",
		},
		"BadSortTypeValue": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", 42}}},
				{"sort", bson.D{{"asc", "123"}}},
			},
			err: &mongo.CommandError{
				Code:    15974,
				Name:    "Location15974",
				Message: `Illegal key in $sort specification: asc: "123"`,
			},
			altMessage: `Illegal key in $sort specification: asc: 123`,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.Error(t, err)
			AssertEqualAltError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestQueryBadMaxTimeMSType(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		command    bson.D
		err        *mongo.CommandError
		altMessage string
	}{
		"BadMaxTimeMSTypeDouble": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", 43.15},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS has non-integral value",
			},
		},
		"BadMaxTimeMSNegativeDouble": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", -14245345234123245.55},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-14245345234123246 value for maxTimeMS is out of range",
			},
		},
		"BadMaxTimeMSTypeString": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", "string"},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
		"BadMaxTimeMSMaxInt64": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", math.MaxInt64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "9223372036854775807 value for maxTimeMS is out of range",
			},
		},
		"BadMaxTimeMSMinInt64": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", math.MinInt64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-9223372036854775808 value for maxTimeMS is out of range",
			},
		},
		"BadMaxTimeMSNull": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", nil},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
		"BadMaxTimeMSArray": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", bson.A{int32(42), "foo", nil}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
		"BadMaxTimeMSDocument": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
		"BadMaxTimeMSTypeNegativeInt32": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", -1123123},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-1123123 value for maxTimeMS is out of range",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.Error(t, err)
			AssertEqualAltError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestQueryMaxTimeMSAvailableValues(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command any
	}{
		"Double": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", float64(10000)},
			},
		},
		"DoubleZero": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", float64(0)},
			},
		},
		"Int32": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", int32(10000)},
			},
		},
		"Int32Zero": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", int32(0)},
			},
		},
		"Int64": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", int64(10000)},
			},
		},
		"Int64Zero": {
			command: bson.D{
				{"find", collection.Name()},
				{"maxTimeMS", int64(0)},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.NoError(t, err)
		})
	}
}

func TestQueryExactMatches(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{
			{"_id", "document-two-fields"},
			{"foo", "bar"},
			{"baz", int32(42)},
		},
		bson.D{
			{"_id", "document-value-two-fields"},
			{"v", bson.D{{"foo", "bar"}, {"baz", int32(42)}}},
		},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
	}{
		"Document": {
			filter:      bson.D{{"foo", "bar"}, {"baz", int32(42)}},
			expectedIDs: []any{"document-two-fields"},
		},
		"DocumentChangedFieldsOrder": {
			filter:      bson.D{{"baz", int32(42)}, {"foo", "bar"}},
			expectedIDs: []any{"document-two-fields"},
		},
		"DocumentValueFields": {
			filter:      bson.D{{"v", bson.D{{"foo", "bar"}, {"baz", int32(42)}}}},
			expectedIDs: []any{"document-value-two-fields"},
		},

		"Array": {
			filter:      bson.D{{"v", bson.A{int32(42), "foo", nil}}},
			expectedIDs: []any{"array-three"},
		},
		"ArrayChangedOrder": {
			filter:      bson.D{{"v", bson.A{int32(42), nil, "foo"}}},
			expectedIDs: []any{},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

func TestDotNotation(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t)

	_, err := collection.InsertMany(
		ctx,
		[]any{
			bson.D{
				{"_id", "document-deeply-nested"},
				{
					"foo",
					bson.D{
						{
							"bar",
							bson.D{{
								"baz",
								bson.D{{"qux", bson.D{{"quz", int32(42)}}}},
							}},
						},
						{
							"qaz",
							bson.A{bson.D{{"baz", int32(1)}}},
						},
					},
				},
				{
					"wsx",
					bson.A{bson.D{{"edc", bson.A{bson.D{{"rfv", int32(1)}}}}}},
				},
			},
		},
	)
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
	}{
		"DeeplyNested": {
			filter:      bson.D{{"foo.bar.baz.qux.quz", int32(42)}},
			expectedIDs: []any{"document-deeply-nested"},
		},
		"DottedField": {
			filter:      bson.D{{"foo.bar.baz", bson.D{{"qux.quz", int32(42)}}}},
			expectedIDs: []any{},
		},
		"FieldArrayField": {
			filter:      bson.D{{"foo.qaz.0.baz", int32(1)}},
			expectedIDs: []any{"document-deeply-nested"},
		},
		"FieldArrayFieldArrayField": {
			filter:      bson.D{{"wsx.0.edc.0.rfv", int32(1)}},
			expectedIDs: []any{"document-deeply-nested"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

// TestQueryNonExistingCollection tests that a query to a non existing collection doesn't fail but returns an empty result.
func TestQueryNonExistingCollection(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	cursor, err := collection.Database().Collection("doesnotexist").Find(ctx, bson.D{})
	require.NoError(t, err)

	var actual []bson.D
	err = cursor.All(ctx, &actual)
	require.NoError(t, err)
	require.Len(t, actual, 0)
}
