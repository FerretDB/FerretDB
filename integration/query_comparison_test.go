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
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestQueryComparisonEq(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
	}{
		"DoubleBigInt64PlusOne": {
			filter:      bson.D{{"v", bson.D{{"$eq", float64(2<<61 + 1)}}}},
			expectedIDs: []any{"int64-big"},
		},
		"Int64BigPlusOne": {
			filter:      bson.D{{"v", bson.D{{"$eq", int64(2<<60 + 1)}}}},
			expectedIDs: []any{"int64-big"},
		},
		"IDNull": {
			filter:      bson.D{{"_id", bson.D{{"$eq", nil}}}},
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
func TestQueryComparisonGte(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		// TODO document

		"ArrayEmpty": {
			value: bson.A{},
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three", "array-three-reverse", "array-two",
			},
		},
		"ArrayOne": {
			value: bson.A{int32(42)},
			expectedIDs: []any{
				"array", "array-three", "array-two",
			},
		},
		"Array": {
			value:       bson.A{int32(42), "foo", nil},
			expectedIDs: []any{"array-three", "array-two"},
		},
		"ArrayReverse": {
			value: bson.A{nil, "foo", int32(42)},
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse", "array-two",
			},
		},
		"ArrayNull": {
			value: bson.A{nil},
			expectedIDs: []any{
				"array", "array-null", "array-three", "array-three-reverse", "array-two",
			},
		},
		"ArraySlice": {
			value: bson.A{int32(42), "foo"},
			expectedIDs: []any{
				"array-three", "array-two",
			},
		},
		"ArrayShuffledValues": {
			value:       bson.A{"foo", nil, int32(42)},
			expectedIDs: []any{},
		},

		"Double": {
			value: 42.13,
			expectedIDs: []any{
				"array-two", "double", "double-big", "double-max", "int32-max", "int64-big", "int64-max",
			},
		},
		"DoubleNegativeZero": {
			value: math.Copysign(0, -1),
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse", "array-two",
				"double", "double-big", "double-max", "double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-zero",
				"int64", "int64-big", "int64-max", "int64-zero",
			},
		},
		"DoubleMax": {
			value:       math.MaxFloat64,
			expectedIDs: []any{"double-max"},
		},

		"DoubleNaN": {
			value:       math.NaN(),
			expectedIDs: []any{"array-two", "double-nan"},
		},

		"String": {
			value:       "foo",
			expectedIDs: []any{"array-three", "array-three-reverse", "string"},
		},
		"StringWhole": {
			value:       "42",
			expectedIDs: []any{"array-three", "array-three-reverse", "string", "string-double", "string-whole"},
		},
		"StringEmpty": {
			value:       "",
			expectedIDs: []any{"array-three", "array-three-reverse", "string", "string-double", "string-empty", "string-whole"},
		},

		"Binary": {
			value:       primitive.Binary{Subtype: 0x80, Data: []byte{42}},
			expectedIDs: []any{"binary"},
		},
		"BinaryNoSubtype": {
			value:       primitive.Binary{Data: []byte{42}},
			expectedIDs: []any{"binary"},
		},
		"BinaryEmpty": {
			value:       primitive.Binary{},
			expectedIDs: []any{"binary", "binary-empty"},
		},

		"ObjectID": {
			value:       must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011")),
			expectedIDs: []any{"objectid"},
		},
		"ObjectIDEmpty": {
			value:       primitive.NilObjectID,
			expectedIDs: []any{"objectid", "objectid-empty"},
		},

		"Bool": {
			value:       false,
			expectedIDs: []any{"bool-false", "bool-true"},
		},

		"Datetime": {
			value:       time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC),
			expectedIDs: []any{"datetime", "datetime-year-max"},
		},

		"Null": {
			value: nil,
			expectedIDs: []any{
				"array-null", "array-three",
				"array-three-reverse", "null",
			},
		},

		"Regex": {
			value: primitive.Regex{Pattern: "foo"},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Can't have RegEx as arg to predicate over field 'v'.",
			},
		},

		"Int32": {
			value: int32(42),
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse", "array-two",
				"double", "double-big", "double-max", "double-whole",
				"int32", "int32-max",
				"int64", "int64-big", "int64-max",
			},
		},
		"Int32Max": {
			value:       int32(math.MaxInt32),
			expectedIDs: []any{"double-big", "double-max", "int32-max", "int64-big", "int64-max"},
		},

		"Timestamp": {
			value:       primitive.Timestamp{T: 42, I: 12},
			expectedIDs: []any{"timestamp"},
		},
		"TimestampNoI": {
			value:       primitive.Timestamp{T: 42},
			expectedIDs: []any{"timestamp"},
		},
		"TimestampNoT": {
			value:       primitive.Timestamp{I: 13},
			expectedIDs: []any{"timestamp"},
		},

		"Int64": {
			value: int64(42),
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse", "array-two",
				"double", "double-big", "double-max", "double-whole",
				"int32", "int32-max",
				"int64", "int64-big", "int64-max",
			},
		},
		"Int64Max": {
			value:       int64(math.MaxInt64),
			expectedIDs: []any{"double-max", "int64-max"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$gte", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
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

func TestQueryComparisonLt(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		// TODO document

		"ArrayEmpty": {
			value:       bson.A{},
			expectedIDs: []any{},
		},
		"ArrayOne": {
			value: bson.A{int32(42)},
			expectedIDs: []any{
				"array-empty",
				"array-null", "array-three-reverse",
			},
		},
		"Array": {
			value: bson.A{int32(42), "foo", nil},
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three-reverse",
			},
		},
		"ArrayReverse": {
			value:       bson.A{nil, "foo", int32(42)},
			expectedIDs: []any{"array-empty", "array-null"},
		},
		"ArrayNull": {
			value:       bson.A{nil},
			expectedIDs: []any{"array-empty"},
		},
		"ArraySlice": {
			value: bson.A{int32(42), "foo"},
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three-reverse",
			},
		},
		"ArrayShuffledValues": {
			value: bson.A{"foo", nil, int32(42)},
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three", "array-three-reverse", "array-two",
			},
		},

		"Double": {
			value: 43.13,
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse", "array-two",
				"double", "double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-min", "int32-zero",
				"int64", "int64-min", "int64-zero",
			},
		},
		"DoubleNegativeZero": {
			value:       math.Copysign(0, -1),
			expectedIDs: []any{"int32-min", "int64-min"},
		},
		"DoubleSmallest": {
			value: math.SmallestNonzeroFloat64,
			expectedIDs: []any{
				"double-negative-zero", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"DoubleNaN": {
			value:       math.NaN(),
			expectedIDs: []any{},
		},

		"String": {
			value:       "goo",
			expectedIDs: []any{"array-three", "array-three-reverse", "string", "string-double", "string-empty", "string-whole"},
		},
		"StringWhole": {
			value:       "42",
			expectedIDs: []any{"string-empty"},
		},
		"StringEmpty": {
			value:       "",
			expectedIDs: []any{},
		},

		"Binary": {
			value:       primitive.Binary{Subtype: 0x80, Data: []byte{43}},
			expectedIDs: []any{"binary-empty"},
		},
		"BinaryNoSubtype": {
			value:       primitive.Binary{Data: []byte{43}},
			expectedIDs: []any{"binary-empty"},
		},
		"BinaryEmpty": {
			value:       primitive.Binary{},
			expectedIDs: []any{},
		},

		"ObjectID": {
			value:       must.NotFail(primitive.ObjectIDFromHex("000102030405060708091012")),
			expectedIDs: []any{"objectid", "objectid-empty"},
		},
		"ObjectIDEmpty": {
			value:       primitive.NilObjectID,
			expectedIDs: []any{},
		},

		"Bool": {
			value:       true,
			expectedIDs: []any{"bool-false"},
		},

		"Datetime": {
			value:       time.Date(2021, 11, 1, 10, 18, 43, 123000000, time.UTC),
			expectedIDs: []any{"datetime", "datetime-epoch", "datetime-year-min"},
		},

		"Null": {
			value:       nil,
			expectedIDs: []any{},
		},

		"Regex": {
			value: primitive.Regex{Pattern: "foo"},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Can't have RegEx as arg to predicate over field 'v'.",
			},
		},

		"Int32": {
			value: int32(42),
			expectedIDs: []any{
				"double-negative-zero", "double-smallest", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"Int32Min": {
			value:       int32(math.MinInt32),
			expectedIDs: []any{"int64-min"},
		},

		"Timestamp": {
			value:       primitive.Timestamp{T: 43, I: 14},
			expectedIDs: []any{"timestamp", "timestamp-i"},
		},
		"TimestampNoI": {
			value:       primitive.Timestamp{T: 43},
			expectedIDs: []any{"timestamp", "timestamp-i"},
		},
		"TimestampNoT": {
			value:       primitive.Timestamp{I: 14},
			expectedIDs: []any{"timestamp-i"},
		},

		"Int64": {
			value: int64(42),
			expectedIDs: []any{
				"double-negative-zero", "double-smallest", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"Int64Min": {
			value:       int64(math.MinInt64),
			expectedIDs: []any{},
		},
		"Int64Big": {
			value: int64(2<<60 + 1),
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse", "array-two",
				"double", "double-big", "double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-min", "int64-zero",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$lt", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
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

func TestQueryComparisonLte(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		// TODO document

		"ArrayEmpty": {
			value:       bson.A{},
			expectedIDs: []any{"array-empty"},
		},
		"ArrayOne": {
			value: bson.A{int32(42)},
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three-reverse",
			},
		},
		"Array": {
			value: bson.A{int32(42), "foo", nil},
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three", "array-three-reverse",
			},
		},
		"ArrayReverse": {
			value:       bson.A{nil, "foo", int32(42)},
			expectedIDs: []any{"array-empty", "array-null", "array-three-reverse"},
		},
		"ArrayNull": {
			value:       bson.A{nil},
			expectedIDs: []any{"array-empty", "array-null"},
		},
		"ArraySlice": {
			value: bson.A{int32(42), "foo"},
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three-reverse",
			},
		},
		"ArrayShuffledValues": {
			value: bson.A{"foo", nil, int32(42)},
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three", "array-three-reverse", "array-two",
			},
		},

		"Double": {
			value: 42.13,
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse", "array-two",
				"double", "double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-min", "int32-zero",
				"int64", "int64-min", "int64-zero",
			},
		},
		"DoubleNegativeZero": {
			value: math.Copysign(0, -1),
			expectedIDs: []any{
				"double-negative-zero", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"DoubleSmallest": {
			value: math.SmallestNonzeroFloat64,
			expectedIDs: []any{
				"double-negative-zero", "double-smallest", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"DoubleNaN": {
			value:       math.NaN(),
			expectedIDs: []any{"array-two", "double-nan"},
		},

		"String": {
			value:       "foo",
			expectedIDs: []any{"array-three", "array-three-reverse", "string", "string-double", "string-empty", "string-whole"},
		},
		"StringWhole": {
			value:       "42",
			expectedIDs: []any{"string-empty", "string-whole"},
		},
		"StringEmpty": {
			value:       "",
			expectedIDs: []any{"string-empty"},
		},

		"Binary": {
			value:       primitive.Binary{Subtype: 0x80, Data: []byte{42}},
			expectedIDs: []any{"binary-empty"},
		},
		"BinaryNoSubtype": {
			value:       primitive.Binary{Data: []byte{42}},
			expectedIDs: []any{"binary-empty"},
		},
		"BinaryEmpty": {
			value:       primitive.Binary{},
			expectedIDs: []any{"binary-empty"},
		},

		"ObjectID": {
			value:       must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011")),
			expectedIDs: []any{"objectid", "objectid-empty"},
		},
		"ObjectIDEmpty": {
			value:       primitive.NilObjectID,
			expectedIDs: []any{"objectid-empty"},
		},

		"Bool": {
			value:       true,
			expectedIDs: []any{"bool-false", "bool-true"},
		},

		"Datetime": {
			value:       time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC),
			expectedIDs: []any{"datetime", "datetime-epoch", "datetime-year-min"},
		},

		"Null": {
			value: nil,
			expectedIDs: []any{
				"array-null",
				"array-three", "array-three-reverse", "null",
			},
		},

		"Regex": {
			value: primitive.Regex{Pattern: "foo"},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Can't have RegEx as arg to predicate over field 'v'.",
			},
		},

		"Int32": {
			value: int32(42),
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse",
				"double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-min", "int32-zero",
				"int64", "int64-min", "int64-zero",
			},
		},
		"Int32Min": {
			value:       int32(math.MinInt32),
			expectedIDs: []any{"int32-min", "int64-min"},
		},

		"Timestamp": {
			value:       primitive.Timestamp{T: 42, I: 13},
			expectedIDs: []any{"timestamp", "timestamp-i"},
		},
		"TimestampNoI": {
			value:       primitive.Timestamp{T: 42},
			expectedIDs: []any{"timestamp-i"},
		},
		"TimestampNoT": {
			value:       primitive.Timestamp{I: 13},
			expectedIDs: []any{"timestamp-i"},
		},

		"Int64": {
			value: int64(42),
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse",
				"double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-min", "int32-zero",
				"int64", "int64-min", "int64-zero",
			},
		},
		"Int64Min": {
			value:       int64(math.MinInt64),
			expectedIDs: []any{"int64-min"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$lte", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
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

func TestQueryComparisonNin(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	var scalarDataTypesFilter bson.A
	for _, scalarDataType := range shareddata.Scalars.Docs() {
		scalarDataTypesFilter = append(scalarDataTypesFilter, scalarDataType.Map()["v"])
	}

	var compositeDataTypesFilter bson.A
	for _, compositeDataType := range shareddata.Composites.Docs() {
		compositeDataTypesFilter = append(compositeDataTypesFilter, compositeDataType.Map()["v"])
	}

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"ForScalarDataTypes": {
			value: scalarDataTypesFilter,
			expectedIDs: []any{
				"array-empty", "document", "document-composite",
				"document-composite-reverse", "document-empty", "document-null",
			},
		},
		"ForCompositeDataTypes": {
			value: compositeDataTypesFilter,
			expectedIDs: []any{
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"double", "double-big", "double-max", "double-nan", "double-negative-zero",
				"double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-big", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},

		"RegexString": {
			value: bson.A{bson.D{{"$regex", "/foo/"}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `cannot nest $ under $in`,
			},
		},
		"Regex": {
			value: bson.A{primitive.Regex{Pattern: "foo", Options: "i"}},
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-two",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-composite", "document-composite-reverse", "document-empty", "document-null",
				"double", "double-big", "double-max", "double-nan", "double-negative-zero",
				"double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-big", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex-empty",
				"string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},

		"NilInsteadOfArray": {
			value: nil,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$nin needs an array`,
			},
		},
		"StringInsteadOfArray": {
			value: "foo",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$nin needs an array`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$nin", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
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

func TestQueryComparisonIn(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	var scalarDataTypesFilter bson.A
	for _, scalarDataType := range shareddata.Scalars.Docs() {
		scalarDataTypesFilter = append(scalarDataTypesFilter, scalarDataType.Map()["v"])
	}

	var compositeDataTypesFilter bson.A
	for _, compositeDataType := range shareddata.Composites.Docs() {
		compositeDataTypesFilter = append(compositeDataTypesFilter, compositeDataType.Map()["v"])
	}

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"ForScalarDataTypes": {
			value: scalarDataTypesFilter,
			expectedIDs: []any{
				"array", "array-null",
				"array-three", "array-three-reverse", "array-two",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"double", "double-big", "double-max", "double-nan", "double-negative-zero",
				"double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-big", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},
		"ForCompositeDataTypes": {
			value: compositeDataTypesFilter,
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three", "array-three-reverse", "array-two",
				"document", "document-composite", "document-composite-reverse", "document-empty", "document-null",
			},
		},

		"RegexString": {
			value: bson.A{bson.D{{"$regex", "/foo/"}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `cannot nest $ under $in`,
			},
		},
		"Regex": {
			value:       bson.A{primitive.Regex{Pattern: "foo", Options: "i"}},
			expectedIDs: []any{"array-three", "array-three-reverse", "regex", "string"},
		},

		"NilInsteadOfArray": {
			value: nil,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$in needs an array`,
			},
		},
		"StringInsteadOfArray": {
			value: "foo",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$in needs an array`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$in", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
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

func TestQueryComparisonNe(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	for name, tc := range map[string]struct {
		value        any
		unexpectedID string
		err          *mongo.CommandError
	}{
		"Document": {
			value:        bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}},
			unexpectedID: "document-composite",
		},
		"DocumentShuffledKeys": {
			value:        bson.D{{"v", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}},
			unexpectedID: "",
		},

		"Array": {
			value:        bson.A{int32(42), "foo", nil},
			unexpectedID: "array-three",
		},
		"ArrayShuffledValues": {
			value:        bson.A{"foo", nil, int32(42)},
			unexpectedID: "",
		},

		"Double": {
			value:        42.13,
			unexpectedID: "double",
		},
		"DoubleNegativeZero": {
			value:        math.Copysign(0, -1),
			unexpectedID: "double-negative-zero",
		},
		"DoubleMax": {
			value:        math.MaxFloat64,
			unexpectedID: "double-max",
		},
		"DoubleSmallest": {
			value:        math.SmallestNonzeroFloat64,
			unexpectedID: "double-smallest",
		},
		"DoubleZero": {
			value:        0.0,
			unexpectedID: "double-zero",
		},
		"DoubleNaN": {
			value:        math.NaN(),
			unexpectedID: "double-nan",
		},
		"DoubleBig": {
			value:        float64(2 << 60),
			unexpectedID: "double-big",
		},

		"String": {
			value:        "foo",
			unexpectedID: "string",
		},
		"EmptyString": {
			value:        "",
			unexpectedID: "string-empty",
		},

		"Binary": {
			value:        primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}},
			unexpectedID: "binary",
		},
		"EmptyBinary": {
			value:        primitive.Binary{Data: []byte{}},
			unexpectedID: "binary-empty",
		},

		"BoolFalse": {
			value:        false,
			unexpectedID: "bool-false",
		},
		"BoolTrue": {
			value:        true,
			unexpectedID: "bool-true",
		},

		"Datetime": {
			value:        primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC)),
			unexpectedID: "datetime",
		},
		"DatetimeEpoch": {
			value:        primitive.NewDateTimeFromTime(time.Unix(0, 0)),
			unexpectedID: "datetime-epoch",
		},
		"DatetimeYearMax": {
			value:        primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)),
			unexpectedID: "datetime-year-min",
		},
		"DatetimeYearMin": {
			value:        primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC)),
			unexpectedID: "datetime-year-max",
		},

		"Timestamp": {
			value:        primitive.Timestamp{T: 42, I: 13},
			unexpectedID: "timestamp",
		},
		"TimestampI": {
			value:        primitive.Timestamp{I: 1},
			unexpectedID: "timestamp-i",
		},

		"Null": {
			value:        nil,
			unexpectedID: "null",
		},

		"Int32": {
			value:        int32(42),
			unexpectedID: "int32",
		},
		"Int32Zero": {
			value:        int32(0),
			unexpectedID: "int32-zero",
		},
		"Int32Max": {
			value:        int32(math.MaxInt32),
			unexpectedID: "int32-max",
		},
		"Int32Min": {
			value:        int32(math.MinInt32),
			unexpectedID: "int32-min",
		},

		"Int64": {
			value:        int64(42),
			unexpectedID: "int64",
		},
		"Int64Zero": {
			value:        int64(0),
			unexpectedID: "int64-zero",
		},
		"Int64Max": {
			value:        int64(math.MaxInt64),
			unexpectedID: "int64-max",
		},
		"Int64Min": {
			value:        int64(math.MinInt64),
			unexpectedID: "int64-min",
		},
		"Int64Big": {
			value:        int64(2 << 61),
			unexpectedID: "int64-big",
		},

		"Regex": {
			value: primitive.Regex{Pattern: "foo"},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Can't have regex as arg to $ne.`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$ne", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.NotContains(t, CollectIDs(t, actual), tc.unexpectedID)
		})
	}
}

func TestQueryComparisonMultipleOperators(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		filter      any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"InLteGte": {
			filter: bson.D{
				{"_id", bson.D{{"$in", bson.A{"int32"}}}},
				{"v", bson.D{{"$lte", int32(42)}, {"$gte", int32(0)}}},
			},
			expectedIDs: []any{"int32"},
		},
		"NinEqNe": {
			filter: bson.D{
				{"_id", bson.D{{"$nin", bson.A{"int64"}}, {"$ne", "int32"}}},
				{"v", bson.D{{"$eq", int32(42)}}},
			},
			expectedIDs: []any{"array", "array-three", "array-three-reverse", "double-whole"},
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
