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

func TestQueryBitwiseAllClear(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/508
	_, err := collection.DeleteOne(ctx, bson.D{{"_id", "binary"}})
	require.NoError(t, err)
	_, err = collection.DeleteOne(ctx, bson.D{{"_id", "binary-empty"}})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         *mongo.CommandError
		altMessage  string
	}{
		"Array": {
			value: primitive.A{1, 5},
			expectedIDs: []any{
				"double-1", "double-2", "double-4",
				"double-big", "double-min-overflow-verge", "double-zero",
				"int32-min", "int32-zero",
				"int64-3", "int64-big", "int64-min", "int64-zero",
			},
		},
		"ArrayNegativeBitPositionValue": {
			value: primitive.A{-1},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Failed to parse bit position. Expected a non-negative number in: 0: -1",
			},
		},
		"ArrayBadValue": {
			value: primitive.A{"123"},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Failed to parse bit position. Expected a number in: 0: "123"`,
			},
		},

		"Double": {
			value: 1.2,
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected an integer: $bitsAllClear: 1.2",
			},
		},
		"DoubleWhole": {
			value: 2.0,
			expectedIDs: []any{
				"double-1", "double-2", "double-4",
				"double-big", "double-min-overflow-verge", "double-zero",
				"int32-1", "int32-2", "int32-3", "int32-min", "int32-zero",
				"int64-1", "int64-2", "int64-3",
				"int64-big", "int64-min", "int64-zero",
			},
		},
		"DoubleNegativeValue": {
			value: float64(-1),
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAllClear: -1.0",
			},
		},

		"String": {
			value: "123",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `v takes an Array, a number, or a BinData but received: $bitsAllClear: "123"`,
			},
			altMessage: `value takes an Array, a number, or a BinData but received: $bitsAllClear: "123"`,
		},

		"Binary": {
			value: primitive.Binary{Data: []byte{2}},
			expectedIDs: []any{
				"double-1", "double-2", "double-4",
				"double-big", "double-min-overflow-verge", "double-zero",
				"int32-1", "int32-2", "int32-3", "int32-min", "int32-zero",
				"int64-1", "int64-2", "int64-3",
				"int64-big", "int64-min", "int64-zero",
			},
		},
		"BinaryWithZeroBytes": {
			value: primitive.Binary{Data: []byte{0, 0, 2}},
			expectedIDs: []any{
				"double-1", "double-2", "double-3", "double-big",
				"double-min-overflow-verge", "double-whole", "double-zero",
				"int32", "int32-1", "int32-min", "int32-zero",
				"int64", "int64-1", "int64-big", "int64-min", "int64-zero",
			},
		},
		"Binary9Bytes": {
			value: primitive.Binary{Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}},
			expectedIDs: []any{
				"double-big", "double-whole", "double-zero",
				"int32", "int32-zero",
				"int64", "int64-1", "int64-big", "int64-zero",
			},
		},

		"Int32": {
			value: int32(2),
			expectedIDs: []any{
				"double-1", "double-2", "double-4", "double-big",
				"double-min-overflow-verge", "double-zero",
				"int32-1", "int32-2", "int32-3",
				"int32-min", "int32-zero",
				"int64-1", "int64-2", "int64-3", "int64-big",
				"int64-min", "int64-zero",
			},
		},
		"Int32NegativeValue": {
			value: int32(-1),
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAllClear: -1",
			},
		},

		"Int64Max": {
			value: math.MaxInt64,
			expectedIDs: []any{
				"double-1", "double-2",
				"double-min-overflow-verge", "double-zero",
				"int32-zero",
				"int64-min",
				"int64-zero",
			},
		},
		"Int64NegativeValue": {
			value: int64(-1),
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAllClear: -1",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$bitsAllClear", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
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

func TestQueryBitwiseAllSet(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/508
	_, err := collection.DeleteOne(ctx, bson.D{{"_id", "binary"}})
	require.NoError(t, err)
	_, err = collection.DeleteOne(ctx, bson.D{{"_id", "binary-empty"}})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         *mongo.CommandError
		altMessage  string
	}{
		"Array": {
			value:       primitive.A{1, 5},
			expectedIDs: []any{"double-3", "double-whole", "int32", "int32-max", "int64", "int64-max"},
		},
		"ArrayNegativeBitPositionValue": {
			value: primitive.A{-1},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Failed to parse bit position. Expected a non-negative number in: 0: -1",
			},
		},
		"ArrayBadValue": {
			value: primitive.A{"123"},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Failed to parse bit position. Expected a number in: 0: "123"`,
			},
		},

		"Double": {
			value: 1.2,
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected an integer: $bitsAllSet: 1.2",
			},
		},
		"DoubleWhole": {
			value:       2.0,
			expectedIDs: []any{"double-3", "double-whole", "int32", "int32-max", "int64", "int64-max"},
		},
		"DoubleNegativeValue": {
			value: -1.0,
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAllSet: -1.0",
			},
		},

		"String": {
			value: "123",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `v takes an Array, a number, or a BinData but received: $bitsAllSet: "123"`,
			},
			altMessage: `value takes an Array, a number, or a BinData but received: $bitsAllSet: "123"`,
		},

		"Binary": {
			value:       primitive.Binary{Data: []byte{2}},
			expectedIDs: []any{"double-3", "double-whole", "int32", "int32-max", "int64", "int64-max"},
		},
		"BinaryWithZeroBytes": {
			value:       primitive.Binary{Data: []byte{0, 0, 2}},
			expectedIDs: []any{"double-4", "int32-2", "int32-3", "int32-max", "int64-2", "int64-3", "int64-max"},
		},

		"Int32": {
			value:       int32(2),
			expectedIDs: []any{"double-3", "double-whole", "int32", "int32-max", "int64", "int64-max"},
		},
		"Int32NegativeValue": {
			value: int32(-1),
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAllSet: -1",
			},
		},

		"Int64Max": {
			value:       math.MaxInt64,
			expectedIDs: []any{"int64-max"},
		},
		"Int64NegativeValue": {
			value: int64(-1),
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAllSet: -1",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$bitsAllSet", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
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

func TestQueryBitwiseAnyClear(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/508
	_, err := collection.DeleteOne(ctx, bson.D{{"_id", "binary"}})
	require.NoError(t, err)
	_, err = collection.DeleteOne(ctx, bson.D{{"_id", "binary-empty"}})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         *mongo.CommandError
		altMessage  string
	}{
		"Array": {
			value: primitive.A{1, 5},
			expectedIDs: []any{
				"double-1", "double-2", "double-4",
				"double-big", "double-min-overflow-verge", "double-zero",
				"int32-1", "int32-2", "int32-3",
				"int32-min", "int32-zero",
				"int64-1", "int64-2", "int64-3",
				"int64-big", "int64-min", "int64-zero",
			},
		},
		"ArrayNegativeBitPositionValue": {
			value: primitive.A{-1},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Failed to parse bit position. Expected a non-negative number in: 0: -1",
			},
		},
		"ArrayBadValue": {
			value: primitive.A{"123"},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Failed to parse bit position. Expected a number in: 0: "123"`,
			},
		},

		"Double": {
			value: 1.2,
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected an integer: $bitsAnyClear: 1.2",
			},
		},
		"DoubleWhole": {
			value: 2.0,
			expectedIDs: []any{
				"double-1", "double-2", "double-4",
				"double-big", "double-min-overflow-verge", "double-zero",
				"int32-1", "int32-2", "int32-3",
				"int32-min", "int32-zero",
				"int64-1", "int64-2", "int64-3",
				"int64-big", "int64-min", "int64-zero",
			},
		},
		"DoubleNegativeValue": {
			value: -1.0,
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAnyClear: -1.0",
			},
		},

		"String": {
			value: "123",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `v takes an Array, a number, or a BinData but received: $bitsAnyClear: "123"`,
			},
			altMessage: `value takes an Array, a number, or a BinData but received: $bitsAnyClear: "123"`,
		},

		"Binary": {
			value: primitive.Binary{Data: []byte{2}},
			expectedIDs: []any{
				"double-1", "double-2", "double-4",
				"double-big", "double-min-overflow-verge", "double-zero",
				"int32-1", "int32-2", "int32-3",
				"int32-min", "int32-zero",
				"int64-1", "int64-2", "int64-3",
				"int64-big", "int64-min", "int64-zero",
			},
		},
		"BinaryWithZeroBytes": {
			value: primitive.Binary{Data: []byte{0, 0, 2}},
			expectedIDs: []any{
				"double-1", "double-2", "double-3",
				"double-big", "double-min-overflow-verge", "double-whole", "double-zero",
				"int32", "int32-1", "int32-min", "int32-zero",
				"int64", "int64-1", "int64-big", "int64-min", "int64-zero",
			},
		},

		"Int32": {
			value: int32(2),
			expectedIDs: []any{
				"double-1", "double-2", "double-4",
				"double-big", "double-min-overflow-verge", "double-zero",
				"int32-1", "int32-2", "int32-3",
				"int32-min", "int32-zero",
				"int64-1", "int64-2", "int64-3",
				"int64-big", "int64-min", "int64-zero",
			},
		},
		"Int32NegativeValue": {
			value: int32(-1),
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAnyClear: -1",
			},
		},

		"Int64Max": {
			value: math.MaxInt64,
			expectedIDs: []any{
				"double-1", "double-2", "double-3", "double-4",
				"double-big", "double-min-overflow-verge", "double-whole", "double-zero",
				"int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-min", "int64-zero",
			},
		},
		"Int64NegativeValue": {
			value: int64(-1),
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAnyClear: -1",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$bitsAnyClear", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
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

func TestQueryBitwiseAnySet(t *testing.T) {
	setup.SkipForPostgresWithReason(t, "todo")
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/508
	_, err := collection.DeleteOne(ctx, bson.D{{"_id", "binary"}})
	require.NoError(t, err)
	_, err = collection.DeleteOne(ctx, bson.D{{"_id", "binary-empty"}})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         *mongo.CommandError
		altMessage  string
	}{
		"Array": {
			value: primitive.A{1, 5},
			expectedIDs: []any{
				"double-3", "double-whole",
				"int32", "int32-1", "int32-2", "int32-3", "int32-max",
				"int64", "int64-1", "int64-2", "int64-max",
			},
		},
		"ArrayNegativeBitPositionValue": {
			value: primitive.A{-1},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Failed to parse bit position. Expected a non-negative number in: 0: -1",
			},
		},
		"ArrayBadValue": {
			value: primitive.A{"123"},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Failed to parse bit position. Expected a number in: 0: "123"`,
			},
		},

		"Double": {
			value: 1.2,
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected an integer: $bitsAnySet: 1.2",
			},
		},
		"DoubleWhole": {
			value: 2.0,
			expectedIDs: []any{
				"double-3", "double-whole",
				"int32", "int32-max",
				"int64", "int64-max",
			},
		},
		"DoubleNegativeValue": {
			value: -1.0,
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAnySet: -1.0",
			},
		},

		"String": {
			value: "123",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `v takes an Array, a number, or a BinData but received: $bitsAnySet: "123"`,
			},
			altMessage: `value takes an Array, a number, or a BinData but received: $bitsAnySet: "123"`,
		},

		"Binary": {
			value:       primitive.Binary{Data: []byte{2}},
			expectedIDs: []any{"double-3", "double-whole", "int32", "int32-max", "int64", "int64-max"},
		},
		"BinaryWithZeroBytes": {
			value:       primitive.Binary{Data: []byte{0, 0, 2}},
			expectedIDs: []any{"double-4", "int32-2", "int32-3", "int32-max", "int64-2", "int64-3", "int64-max"},
		},

		"Int32": {
			value: int32(2),
			expectedIDs: []any{
				"double-3", "double-whole",
				"int32", "int32-max",
				"int64", "int64-max",
			},
		},
		"Int32NegativeValue": {
			value: int32(-1),
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAnySet: -1",
			},
		},

		"Int64Max": {
			value: math.MaxInt64,
			expectedIDs: []any{
				"double-3", "double-4",
				"double-big", "double-whole",
				"int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min",
				"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-max",
			},
		},
		"Int64NegativeValue": {
			value: int64(-1),
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Expected a non-negative number in: $bitsAnySet: -1",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$bitsAnySet", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
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
