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
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestQueryComparisonImplicit(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
	}{
		"Document": {
			filter:      bson.D{{"value", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
			expectedIDs: []any{"document-composite"},
		},
		"DocumentShuffledKeys": {
			filter:      bson.D{{"value", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}},
			expectedIDs: []any{},
		},

		"Array": {
			filter:      bson.D{{"value", bson.A{int32(42), "foo", nil}}},
			expectedIDs: []any{"array-three"},
		},
		"ArrayEmbedded": {
			filter:      bson.D{{"value", bson.A{bson.A{int32(42), "foo"}, nil}}},
			expectedIDs: []any{"array-embedded"},
		},
		"ArrayShuffledValues": {
			filter:      bson.D{{"value", bson.A{"foo", nil, int32(42)}}},
			expectedIDs: []any{},
		},

		"IDNull": {
			filter:      bson.D{{"_id", nil}},
			expectedIDs: []any{},
		},
		"ValueNull": {
			filter:      bson.D{{"value", nil}},
			expectedIDs: []any{"array-embedded", "array-three", "null"},
		},
		"NoSuchFieldNull": {
			filter: bson.D{{"no-such-field", nil}},
			expectedIDs: []any{
				"array", "array-embedded", "array-empty", "array-three",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-composite", "document-empty",
				"double", "double-max", "double-nan", "double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},

		"ValueNumber": {
			filter:      bson.D{{"value", 42}},
			expectedIDs: []any{"array", "array-three", "double-whole", "int32", "int64"},
		},

		"ValueRegex": {
			filter:      bson.D{{"value", primitive.Regex{Pattern: "^fo"}}},
			expectedIDs: []any{"array-three", "string"},
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
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}

func TestQueryComparisonEq(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
	}{
		"Document": {
			filter:      bson.D{{"value", bson.D{{"$eq", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}}}},
			expectedIDs: []any{"document-composite"},
		},
		"DocumentShuffledKeys": {
			filter:      bson.D{{"value", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}},
			expectedIDs: []any{},
		},

		"Array": {
			filter:      bson.D{{"value", bson.D{{"$eq", bson.A{int32(42), "foo", nil}}}}},
			expectedIDs: []any{"array-three"},
		},
		"ArrayEmbedded": {
			filter:      bson.D{{"value", bson.D{{"$eq", bson.A{bson.A{int32(42), "foo"}, nil}}}}},
			expectedIDs: []any{"array-embedded"},
		},
		"ArrayShuffledValues": {
			filter:      bson.D{{"value", bson.A{"foo", nil, int32(42)}}},
			expectedIDs: []any{},
		},

		"Double": {
			filter:      bson.D{{"value", bson.D{{"$eq", 42.13}}}},
			expectedIDs: []any{"double"},
		},
		"DoubleWhole": {
			filter:      bson.D{{"value", bson.D{{"$eq", 42.0}}}},
			expectedIDs: []any{"array", "array-three", "double-whole", "int32", "int64"},
		},
		"DoubleZero": {
			filter:      bson.D{{"value", bson.D{{"$eq", 0.0}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"DoubleNegativeZero": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.Copysign(0, -1)}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"DoubleMax": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.MaxFloat64}}}},
			expectedIDs: []any{"double-max"},
		},
		"DoubleSmallest": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.SmallestNonzeroFloat64}}}},
			expectedIDs: []any{"double-smallest"},
		},
		"DoublePositiveInfinity": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.Inf(+1)}}}},
			expectedIDs: []any{"double-positive-infinity"},
		},
		"DoubleNegativeInfinity": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.Inf(-1)}}}},
			expectedIDs: []any{"double-negative-infinity"},
		},
		"DoubleNaN": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.NaN()}}}},
			expectedIDs: []any{"double-nan"},
		},

		"String": {
			filter:      bson.D{{"value", bson.D{{"$eq", "foo"}}}},
			expectedIDs: []any{"array-three", "string"},
		},
		"StringDouble": {
			filter:      bson.D{{"value", bson.D{{"$eq", "42.13"}}}},
			expectedIDs: []any{"string-double"},
		},
		"StringWhole": {
			filter:      bson.D{{"value", bson.D{{"$eq", "42"}}}},
			expectedIDs: []any{"string-whole"},
		},
		"StringEmpty": {
			filter:      bson.D{{"value", bson.D{{"$eq", ""}}}},
			expectedIDs: []any{"string-empty"},
		},

		"Binary": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
			expectedIDs: []any{"binary"},
		},
		"BinaryEmpty": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Binary{Data: []byte{}}}}}},
			expectedIDs: []any{"binary-empty"},
		},

		"ObjectID": {
			filter:      bson.D{{"value", bson.D{{"$eq", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011"))}}}},
			expectedIDs: []any{"objectid"},
		},
		"ObjectIDEmpty": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NilObjectID}}}},
			expectedIDs: []any{"objectid-empty"},
		},

		"BoolFalse": {
			filter:      bson.D{{"value", bson.D{{"$eq", false}}}},
			expectedIDs: []any{"bool-false"},
		},
		"BoolTrue": {
			filter:      bson.D{{"value", bson.D{{"$eq", true}}}},
			expectedIDs: []any{"bool-true"},
		},

		"Datetime": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))}}}},
			expectedIDs: []any{"datetime"},
		},
		"DatetimeEpoch": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
			expectedIDs: []any{"datetime-epoch"},
		},
		"DatetimeYearMax": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))}}}},
			expectedIDs: []any{"datetime-year-min"},
		},
		"DatetimeYearMin": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC))}}}},
			expectedIDs: []any{"datetime-year-max"},
		},

		"Null": {
			filter:      bson.D{{"value", bson.D{{"$eq", nil}}}},
			expectedIDs: []any{"array-embedded", "array-three", "null"},
		},

		"RegexWithoutOption": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Regex{Pattern: "foo"}}}}},
			expectedIDs: []any{},
		},
		"RegexWithOption": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Regex{Pattern: "foo", Options: "i"}}}}},
			expectedIDs: []any{"regex"},
		},
		"RegexEmpty": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Regex{}}}}},
			expectedIDs: []any{"regex-empty"},
		},

		"Int32": {
			filter:      bson.D{{"value", bson.D{{"$eq", int32(42)}}}},
			expectedIDs: []any{"array", "array-three", "double-whole", "int32", "int64"},
		},
		"Int32Zero": {
			filter:      bson.D{{"value", bson.D{{"$eq", int32(0)}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"Int32Max": {
			filter:      bson.D{{"value", bson.D{{"$eq", int32(math.MaxInt32)}}}},
			expectedIDs: []any{"int32-max"},
		},
		"Int32Min": {
			filter:      bson.D{{"value", bson.D{{"$eq", int32(math.MinInt32)}}}},
			expectedIDs: []any{"int32-min"},
		},

		"Timestamp": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Timestamp{T: 42, I: 13}}}}},
			expectedIDs: []any{"timestamp"},
		},
		"TimestampI": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Timestamp{I: 1}}}}},
			expectedIDs: []any{"timestamp-i"},
		},

		"Int64": {
			filter:      bson.D{{"value", bson.D{{"$eq", int64(42)}}}},
			expectedIDs: []any{"array", "array-three", "double-whole", "int32", "int64"},
		},
		"Int64Zero": {
			filter:      bson.D{{"value", bson.D{{"$eq", int64(0)}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"Int64Max": {
			filter:      bson.D{{"value", bson.D{{"$eq", int64(math.MaxInt64)}}}},
			expectedIDs: []any{"int64-max"},
		},
		"Int64Min": {
			filter:      bson.D{{"value", bson.D{{"$eq", int64(math.MinInt64)}}}},
			expectedIDs: []any{"int64-min"},
		},

		"IDNull": {
			filter:      bson.D{{"_id", bson.D{{"$eq", nil}}}},
			expectedIDs: []any{},
		},
		"NoSuchFieldNull": {
			filter: bson.D{{"no-such-field", bson.D{{"$eq", nil}}}},
			expectedIDs: []any{
				"array", "array-embedded", "array-empty", "array-three",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-composite", "document-empty",
				"double", "double-max", "double-nan", "double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
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
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}

func TestQueryComparisonGt(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         mongo.CommandError
	}{
		// TODO document, array

		"Double": {
			value: 41.13,
			expectedIDs: []any{
				"array", "array-three",
				"double", "double-max", "double-positive-infinity", "double-whole",
				"int32", "int32-max",
				"int64", "int64-max",
			},
		},
		"DoubleNegativeZero": {
			value: math.Copysign(0, -1),
			expectedIDs: []any{
				"array", "array-three",
				"double", "double-max", "double-positive-infinity", "double-smallest", "double-whole",
				"int32", "int32-max",
				"int64", "int64-max",
			},
		},
		"DoubleMax": {
			value:       math.MaxFloat64,
			expectedIDs: []any{"double-positive-infinity"},
		},
		"DoublePositiveInfinity": {
			value:       math.Inf(+1),
			expectedIDs: []any{},
		},
		"DoubleNaN": {
			value:       math.NaN(),
			expectedIDs: []any{},
		},

		"String": {
			value:       "boo",
			expectedIDs: []any{"array-three", "string"},
		},
		"StringWhole": {
			value:       "42",
			expectedIDs: []any{"array-three", "string", "string-double"},
		},
		"StringEmpty": {
			value:       "",
			expectedIDs: []any{"array-three", "string", "string-double", "string-whole"},
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
			expectedIDs: []any{"binary"},
		},

		"ObjectID": {
			value:       must.NotFail(primitive.ObjectIDFromHex("000102030405060708091010")),
			expectedIDs: []any{"objectid"},
		},
		"ObjectIDEmpty": {
			value:       primitive.NilObjectID,
			expectedIDs: []any{"objectid"},
		},

		"Bool": {
			value:       false,
			expectedIDs: []any{"bool-true"},
		},

		"Datetime": {
			value:       time.Date(2021, 11, 1, 10, 18, 41, 123000000, time.UTC),
			expectedIDs: []any{"datetime", "datetime-year-max"},
		},

		"Null": {
			value:       nil,
			expectedIDs: []any{},
		},

		"Regex": {
			value: primitive.Regex{Pattern: "foo"},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Can't have RegEx as arg to predicate over field 'value'.",
			},
		},

		"Int32": {
			value:       int32(42),
			expectedIDs: []any{"double", "double-max", "double-positive-infinity", "int32-max", "int64-max"},
		},
		"Int32Max": {
			value:       int32(math.MaxInt32),
			expectedIDs: []any{"double-max", "double-positive-infinity", "int64-max"},
		},

		"Timestamp": {
			value:       primitive.Timestamp{T: 41, I: 12},
			expectedIDs: []any{"timestamp"},
		},
		"TimestampNoI": {
			value:       primitive.Timestamp{T: 41},
			expectedIDs: []any{"timestamp"},
		},
		"TimestampNoT": {
			value:       primitive.Timestamp{I: 12},
			expectedIDs: []any{"timestamp"},
		},

		"Int64": {
			value:       int64(42),
			expectedIDs: []any{"double", "double-max", "double-positive-infinity", "int32-max", "int64-max"},
		},
		"Int64Max": {
			value:       int64(math.MaxInt64),
			expectedIDs: []any{"double-max", "double-positive-infinity"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"value", bson.D{{"$gt", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err.Code != 0 {
				require.Nil(t, tc.expectedIDs)
				assertEqualError(t, tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}

func TestQueryComparisonGte(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         mongo.CommandError
	}{
		// TODO document, array

		"Double": {
			value: 42.13,
			expectedIDs: []any{
				"double", "double-max", "double-positive-infinity", "int32-max", "int64-max",
			},
		},
		"DoubleNegativeZero": {
			value: math.Copysign(0, -1),
			expectedIDs: []any{
				"array", "array-three",
				"double", "double-max", "double-negative-zero", "double-positive-infinity", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-zero",
				"int64", "int64-max", "int64-zero",
			},
		},
		"DoubleMax": {
			value:       math.MaxFloat64,
			expectedIDs: []any{"double-max", "double-positive-infinity"},
		},
		"DoublePositiveInfinity": {
			value:       math.Inf(+1),
			expectedIDs: []any{"double-positive-infinity"},
		},
		"DoubleNaN": {
			value:       math.NaN(),
			expectedIDs: []any{"double-nan"},
		},

		"String": {
			value:       "foo",
			expectedIDs: []any{"array-three", "string"},
		},
		"StringWhole": {
			value:       "42",
			expectedIDs: []any{"array-three", "string", "string-double", "string-whole"},
		},
		"StringEmpty": {
			value:       "",
			expectedIDs: []any{"array-three", "string", "string-double", "string-empty", "string-whole"},
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
			value:       nil,
			expectedIDs: []any{"array-embedded", "array-three", "null"},
		},

		"Regex": {
			value: primitive.Regex{Pattern: "foo"},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Can't have RegEx as arg to predicate over field 'value'.",
			},
		},

		"Int32": {
			value: int32(42),
			expectedIDs: []any{
				"array", "array-three",
				"double", "double-max", "double-positive-infinity", "double-whole",
				"int32", "int32-max",
				"int64", "int64-max",
			},
		},
		"Int32Max": {
			value:       int32(math.MaxInt32),
			expectedIDs: []any{"double-max", "double-positive-infinity", "int32-max", "int64-max"},
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
				"array", "array-three",
				"double", "double-max", "double-positive-infinity", "double-whole",
				"int32", "int32-max",
				"int64", "int64-max",
			},
		},
		"Int64Max": {
			value:       int64(math.MaxInt64),
			expectedIDs: []any{"double-max", "double-positive-infinity", "int64-max"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"value", bson.D{{"$gte", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err.Code != 0 {
				require.Nil(t, tc.expectedIDs)
				assertEqualError(t, tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}

func TestQueryComparisonLt(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         mongo.CommandError
	}{
		// TODO document, array

		"Double": {
			value: 43.13,
			expectedIDs: []any{
				"array", "array-three",
				"double", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-min", "int32-zero",
				"int64", "int64-min", "int64-zero",
			},
		},
		"DoubleNegativeZero": {
			value:       math.Copysign(0, -1),
			expectedIDs: []any{"double-negative-infinity", "int32-min", "int64-min"},
		},
		"DoubleSmallest": {
			value: math.SmallestNonzeroFloat64,
			expectedIDs: []any{
				"double-negative-infinity", "double-negative-zero", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"DoubleNegativeInfinity": {
			value:       math.Inf(-1),
			expectedIDs: []any{},
		},
		"DoubleNaN": {
			value:       math.NaN(),
			expectedIDs: []any{},
		},

		"String": {
			value:       "goo",
			expectedIDs: []any{"array-three", "string", "string-double", "string-empty", "string-whole"},
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
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Can't have RegEx as arg to predicate over field 'value'.",
			},
		},

		"Int32": {
			value: int32(42),
			expectedIDs: []any{
				"double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"Int32Min": {
			value:       int32(math.MinInt32),
			expectedIDs: []any{"double-negative-infinity", "int64-min"},
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
				"double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"Int64Min": {
			value:       int64(math.MinInt64),
			expectedIDs: []any{"double-negative-infinity"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"value", bson.D{{"$lt", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err.Code != 0 {
				require.Nil(t, tc.expectedIDs)
				assertEqualError(t, tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}

func TestQueryComparisonLte(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		value       any
		expectedIDs []any
		err         mongo.CommandError
	}{
		// TODO document, array

		"Double": {
			value: 42.13,
			expectedIDs: []any{
				"array", "array-three",
				"double", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-min", "int32-zero",
				"int64", "int64-min", "int64-zero",
			},
		},
		"DoubleNegativeZero": {
			value: math.Copysign(0, -1),
			expectedIDs: []any{
				"double-negative-infinity", "double-negative-zero", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"DoubleSmallest": {
			value: math.SmallestNonzeroFloat64,
			expectedIDs: []any{
				"double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero",
				"int32-min", "int32-zero",
				"int64-min", "int64-zero",
			},
		},
		"DoubleNegativeInfinity": {
			value:       math.Inf(-1),
			expectedIDs: []any{"double-negative-infinity"},
		},
		"DoubleNaN": {
			value:       math.NaN(),
			expectedIDs: []any{"double-nan"},
		},

		"String": {
			value:       "foo",
			expectedIDs: []any{"array-three", "string", "string-double", "string-empty", "string-whole"},
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
			value:       nil,
			expectedIDs: []any{"array-embedded", "array-three", "null"},
		},

		"Regex": {
			value: primitive.Regex{Pattern: "foo"},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Can't have RegEx as arg to predicate over field 'value'.",
			},
		},

		"Int32": {
			value: int32(42),
			expectedIDs: []any{
				"array", "array-three",
				"double-negative-infinity", "double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-min", "int32-zero",
				"int64", "int64-min", "int64-zero",
			},
		},
		"Int32Min": {
			value:       int32(math.MinInt32),
			expectedIDs: []any{"double-negative-infinity", "int32-min", "int64-min"},
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
				"array", "array-three",
				"double-negative-infinity", "double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-min", "int32-zero",
				"int64", "int64-min", "int64-zero",
			},
		},
		"Int64Min": {
			value:       int64(math.MinInt64),
			expectedIDs: []any{"double-negative-infinity", "int64-min"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"value", bson.D{{"$lte", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err.Code != 0 {
				require.Nil(t, tc.expectedIDs)
				assertEqualError(t, tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}

// $in

// $ne

// $nin
