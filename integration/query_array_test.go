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

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryArraySize(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "array-empty"}, {"v", bson.A{}}},
		bson.D{{"_id", "array-one"}, {"v", bson.A{"1"}}},
		bson.D{{"_id", "array-two"}, {"v", bson.A{"1", nil}}},
		bson.D{{"_id", "array-three"}, {"v", bson.A{"1", "2", nil}}},
		bson.D{{"_id", "string"}, {"v", "12"}},
		bson.D{{"_id", "document"}, {"v", bson.D{{"v", bson.A{"1", "2"}}}}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"NegativeZero": {
			filter:      bson.D{{"v", bson.D{{"$size", math.Copysign(0, -1)}}}},
			expectedIDs: []any{"array-empty"},
		},
		"InvalidType": {
			// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/1539
			filter: bson.D{{"v", bson.D{{"$size", bson.D{{"$gt", 1}}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Failed to parse $size. Expected a number in: $size: { $gt: 1 }`,
			},
		},
		"NotWhole": {
			// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/1539
			filter: bson.D{{"v", bson.D{{"$size", 2.1}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Failed to parse $size. Expected an integer: $size: 2.1",
			},
		},
		"Infinity": {
			// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/1539
			filter: bson.D{{"v", bson.D{{"$size", math.Inf(+1)}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Failed to parse $size. Cannot represent as a 64-bit integer: $size: inf.0`,
			},
		},
		"Negative": {
			// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/1539
			filter: bson.D{{"v", bson.D{{"$size", -1}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Failed to parse $size. Expected a non-negative number in: $size: -1`,
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
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"FieldPositionQueryRegex": {
			// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/1540
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
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"UnexpectedFilterString": {
			// TODO move to compat https://github.com/FerretDB/FerretDB/issues/1541
			filter: bson.D{{"v", bson.D{{"$elemMatch", "foo"}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$elemMatch needs an Object",
			},
		},
		"WhereInsideElemMatch": {
			// TODO move to compat https://github.com/FerretDB/FerretDB/issues/1542
			filter: bson.D{{"v", bson.D{{"$elemMatch", bson.D{{"$where", "123"}}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$where can only be applied to the top-level document",
			},
		},
		"TextInsideElemMatch": {
			// TODO move to compat https://github.com/FerretDB/FerretDB/issues/1542
			filter: bson.D{{"v", bson.D{{"$elemMatch", bson.D{{"$text", "123"}}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$text can only be applied to the top-level document",
			},
		},
		"GtField": {
			// TODO move to compat https://github.com/FerretDB/FerretDB/issues/1541
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

// TestQueryArrayAll covers the case where the $all operator is used on an array or scalar.
func TestQueryArrayAll(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Composites, shareddata.Scalars)

	// Insert additional data to check more complicated cases:
	// - a longer array of ints;
	// - a field is called differently and needs to be found with the {$all: [null]} case.
	// TODO Add "many-integers" to shareddata.Composites once more
	// query tests are moved to compat, then move remaining tests to compat.
	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "many-integers"}, {"customField", bson.A{42, 43, 45}}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		expectedErr *mongo.CommandError
	}{
		"WholeInTheMiddle": {
			filter:      bson.D{{"customField", bson.D{{"$all", bson.A{int32(43)}}}}},
			expectedIDs: []any{"many-integers"},
		},
		"WholeTwoRepeated": {
			filter:      bson.D{{"customField", bson.D{{"$all", bson.A{int32(42), int32(43), int32(43), int32(42)}}}}},
			expectedIDs: []any{"many-integers"},
		},
		"Nil": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{nil}}}}},
			expectedIDs: []any{
				"array-null", "array-three", "array-three-reverse", "many-integers", "null",
			},
			expectedErr: nil,
		},
		"NilRepeated": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{nil, nil, nil}}}}},
			expectedIDs: []any{
				"array-null", "array-three", "array-three-reverse", "many-integers", "null",
			},
			expectedErr: nil,
		},
		"NaN": {
			filter:      bson.D{{"v", bson.D{{"$all", bson.A{math.NaN()}}}}},
			expectedIDs: []any{"array-two", "double-nan"},
			expectedErr: nil,
		},

		"$allNeedsAnArrayNan": {
			filter:      bson.D{{"v", bson.D{{"$all", math.NaN()}}}},
			expectedIDs: nil,
			expectedErr: &mongo.CommandError{
				Code:    2,
				Message: "$all needs an array",
				Name:    "BadValue",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.expectedErr != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualError(t, *tc.expectedErr, err)
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
