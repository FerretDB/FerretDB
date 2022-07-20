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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryArraySize(t *testing.T) {
	t.Parallel()
	ctx, collection := Setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "array-empty"}, {"v", bson.A{}}},
		bson.D{{"_id", "array-one"}, {"v", bson.A{"1"}}},
		bson.D{{"_id", "array-two"}, {"v", bson.A{"1", nil}}},
		bson.D{{"_id", "array-three"}, {"v", bson.A{"1", "2", math.NaN()}}},
		bson.D{{"_id", "string"}, {"v", "12"}},
		bson.D{{"_id", "document"}, {"v", bson.D{{"v", bson.A{"1", "2"}}}}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"int32": {
			filter:      bson.D{{"v", bson.D{{"$size", int32(2)}}}},
			expectedIDs: []any{"array-two"},
		},
		"int64": {
			filter:      bson.D{{"v", bson.D{{"$size", int64(2)}}}},
			expectedIDs: []any{"array-two"},
		},
		"float64": {
			filter:      bson.D{{"v", bson.D{{"$size", float64(2)}}}},
			expectedIDs: []any{"array-two"},
		},
		"Zero": {
			filter:      bson.D{{"v", bson.D{{"$size", 0}}}},
			expectedIDs: []any{"array-empty"},
		},
		"NegativeZero": {
			filter:      bson.D{{"v", bson.D{{"$size", math.Copysign(0, -1)}}}},
			expectedIDs: []any{"array-empty"},
		},
		"NotFound": {
			filter:      bson.D{{"v", bson.D{{"$size", 4}}}},
			expectedIDs: []any{},
		},
		"InvalidType": {
			filter: bson.D{{"v", bson.D{{"$size", bson.D{{"$gt", 1}}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size needs a number`,
			},
		},
		"NotWhole": {
			filter: bson.D{{"v", bson.D{{"$size", 2.1}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size must be a whole number`,
			},
		},
		"NaN": {
			filter: bson.D{{"v", bson.D{{"$size", math.NaN()}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size must be a whole number`,
			},
		},
		"Infinity": {
			filter: bson.D{{"v", bson.D{{"$size", math.Inf(+1)}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size must be a whole number`,
			},
		},
		"Negative": {
			filter: bson.D{{"v", bson.D{{"$size", -1}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size may not be negative`,
			},
		},
		"InvalidUse": {
			filter: bson.D{{"$size", 2}},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `unknown top level operator: $size. ` +
					`If you have a field name that starts with a '$' symbol, consider using $getField or $setField.`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
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

func TestQueryArrayDotNotation(t *testing.T) {
	t.Parallel()
	ctx, collection := Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"PositionIndexGreaterThanArrayLength": {
			filter:      bson.D{{"v.5", bson.D{{"$type", "double"}}}},
			expectedIDs: []any{},
		},
		"PositionIndexAtTheEndOfArray": {
			filter:      bson.D{{"v.1", bson.D{{"$type", "double"}}}},
			expectedIDs: []any{"array-two"},
		},

		"PositionTypeNull": {
			filter:      bson.D{{"v.0", bson.D{{"$type", "null"}}}},
			expectedIDs: []any{"array-last-embedded", "array-middle-embedded", "array-null", "array-three-reverse"},
		},
		"PositionRegex": {
			filter:      bson.D{{"v.1", primitive.Regex{Pattern: "foo"}}},
			expectedIDs: []any{"array-three", "array-three-reverse"},
		},
		"PositionArray": {
			filter:      bson.D{{"v.0", bson.A{"42", "foo"}}},
			expectedIDs: []any{"array-embedded"},
		},
		"PositionArrayEmpty": {
			filter:      bson.D{{"v.0", bson.A{}}},
			expectedIDs: []any{"array-empty-nested"},
		},

		"NoSuchFieldPosition": {
			filter:      bson.D{{"v.some.0", bson.A{42}}},
			expectedIDs: []any{},
		},
		"Field": {
			filter:      bson.D{{"v.array", int32(42)}},
			expectedIDs: []any{"document-composite", "document-composite-reverse"},
		},
		"FieldPosition": {
			filter:      bson.D{{"v.array.0", int32(42)}},
			expectedIDs: []any{"document-composite", "document-composite-reverse"},
		},
		"FieldPositionQuery": {
			filter:      bson.D{{"v.array.0", bson.D{{"$gte", int32(42)}}}},
			expectedIDs: []any{"document-composite", "document-composite-reverse"},
		},
		"FieldPositionQueryNonArray": {
			filter:      bson.D{{"v.document.0", bson.D{{"$lt", int32(42)}}}},
			expectedIDs: []any{},
		},
		"FieldPositionField": {
			filter:      bson.D{{"v.array.2.foo", "bar"}},
			expectedIDs: []any{},
		},

		"FieldPositionQueryRegex": {
			filter: bson.D{{"v.array.0", bson.D{{"$lt", primitive.Regex{Pattern: "^$"}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Can't have RegEx as arg to predicate over field 'v.array.0'.",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
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

func TestQueryElemMatchOperator(t *testing.T) {
	t.Parallel()
	ctx, collection := Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"DoubleTarget": {
			filter: bson.D{
				{"_id", "double"},
				{"v", bson.D{{"$elemMatch", bson.D{{"$gt", int32(0)}}}}},
			},
			expectedIDs: []any{},
		},
		"GtZero": {
			filter:      bson.D{{"v", bson.D{{"$elemMatch", bson.D{{"$gt", int32(0)}}}}}},
			expectedIDs: []any{"array", "array-three", "array-three-reverse", "array-two"},
		},
		"GtZeroWithTypeArray": {
			filter: bson.D{
				{"v", bson.D{
					{"$elemMatch", bson.D{
						{"$gt", int32(0)},
					}},
					{"$type", "array"},
				}},
			},
			expectedIDs: []any{"array", "array-three", "array-three-reverse", "array-two"},
		},
		"GtZeroWithTypeString": {
			filter: bson.D{
				{"v", bson.D{
					{"$elemMatch", bson.D{
						{"$gt", int32(0)},
					}},
					{"$type", "string"},
				}},
			},
			expectedIDs: []any{"array-three", "array-three-reverse"},
		},
		"GtLt": {
			filter: bson.D{
				{"v", bson.D{
					{"$elemMatch", bson.D{
						{"$gt", int32(0)},
						{"$lt", int32(43)},
					}},
				}},
			},
			expectedIDs: []any{"array", "array-three", "array-three-reverse", "array-two"},
		},

		"UnexpectedFilterString": {
			filter: bson.D{{"v", bson.D{{"$elemMatch", "foo"}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$elemMatch needs an Object",
			},
		},
		"WhereInsideElemMatch": {
			filter: bson.D{{"v", bson.D{{"$elemMatch", bson.D{{"$where", "123"}}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$where can only be applied to the top-level document",
			},
		},
		"TextInsideElemMatch": {
			filter: bson.D{{"v", bson.D{{"$elemMatch", bson.D{{"$text", "123"}}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$text can only be applied to the top-level document",
			},
		},
		"GtField": {
			filter: bson.D{{"v", bson.D{
				{
					"$elemMatch",
					bson.D{
						{"$gt", int32(0)},
						{"foo", int32(42)},
					},
				},
			}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "unknown operator: foo",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
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

func TestArrayEquality(t *testing.T) {
	t.Parallel()
	ctx, collection := Setup(t, shareddata.Composites)

	for name, tc := range map[string]struct {
		array       bson.A
		expectedIDs []any
	}{
		"One": {
			array:       bson.A{int32(42)},
			expectedIDs: []any{"array"},
		},
		"Two": {
			array:       bson.A{42, "foo"},
			expectedIDs: []any{"array-first-embedded", "array-last-embedded", "array-middle-embedded"},
		},
		"Three": {
			array:       bson.A{int32(42), "foo", nil},
			expectedIDs: []any{"array-three"},
		},
		"Three-reverse": {
			array:       bson.A{nil, "foo", int32(42)},
			expectedIDs: []any{"array-three-reverse"},
		},
		"Empty": {
			array:       bson.A{},
			expectedIDs: []any{"array-empty", "array-empty-nested"},
		},
		"Null": {
			array:       bson.A{nil},
			expectedIDs: []any{"array-null"},
		},
		"EmptyNested": {
			array:       bson.A{bson.A{}},
			expectedIDs: []any{"array-empty-nested"},
		},
		"OneEmbedded": {
			array:       bson.A{bson.A{"42", "foo"}},
			expectedIDs: []any{"array-embedded"},
		},
		"FirstEmbedded": {
			array:       bson.A{bson.A{int32(42), "foo"}, nil},
			expectedIDs: []any{"array-first-embedded"},
		},
		"MiddleEmbedded": {
			array:       bson.A{nil, bson.A{int32(42), "foo"}, nil},
			expectedIDs: []any{"array-middle-embedded"},
		},
		"LastEmbedded": {
			array:       bson.A{nil, bson.A{int32(42), "foo"}},
			expectedIDs: []any{"array-last-embedded"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", tc.array}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}
