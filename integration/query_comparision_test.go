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
		"IDNull": {
			filter:      bson.D{{"_id", nil}},
			expectedIDs: []any{},
		},
		"ValueNull": {
			filter:      bson.D{{"value", nil}},
			expectedIDs: []any{"array-three", "null"},
		},
		"NoSuchFieldNull": {
			filter: bson.D{{"no-such-field", nil}},
			expectedIDs: []any{
				"array", "array-empty", "array-three",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-empty",
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

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)
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
		// TODO document, array

		"EqDouble": {
			filter:      bson.D{{"value", bson.D{{"$eq", 42.13}}}},
			expectedIDs: []any{"double"},
		},
		"EqDoubleWhole": {
			filter:      bson.D{{"value", bson.D{{"$eq", 42.0}}}},
			expectedIDs: []any{"array", "array-three", "double-whole", "int32", "int64"},
		},
		"EqDoubleZero": {
			filter:      bson.D{{"value", bson.D{{"$eq", 0.0}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"EqDoubleNegativeZero": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.Copysign(0, -1)}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"EqDoubleMax": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.MaxFloat64}}}},
			expectedIDs: []any{"double-max"},
		},
		"EqDoubleSmallest": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.SmallestNonzeroFloat64}}}},
			expectedIDs: []any{"double-smallest"},
		},
		"EqDoublePositiveInfinity": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.Inf(+1)}}}},
			expectedIDs: []any{"double-positive-infinity"},
		},
		"EqDoubleNegativeInfinity": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.Inf(-1)}}}},
			expectedIDs: []any{"double-negative-infinity"},
		},
		"EqDoubleNaN": {
			filter:      bson.D{{"value", bson.D{{"$eq", math.NaN()}}}},
			expectedIDs: []any{"double-nan"},
		},

		"EqString": {
			filter:      bson.D{{"value", bson.D{{"$eq", "foo"}}}},
			expectedIDs: []any{"array-three", "string"},
		},
		"EqStringDouble": {
			filter:      bson.D{{"value", bson.D{{"$eq", "42.13"}}}},
			expectedIDs: []any{"string-double"},
		},
		"EqStringWhole": {
			filter:      bson.D{{"value", bson.D{{"$eq", "42"}}}},
			expectedIDs: []any{"string-whole"},
		},
		"EqStringEmpty": {
			filter:      bson.D{{"value", bson.D{{"$eq", ""}}}},
			expectedIDs: []any{"string-empty"},
		},

		"EqBinary": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
			expectedIDs: []any{"binary"},
		},
		"EqBinaryEmpty": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Binary{Data: []byte{}}}}}},
			expectedIDs: []any{"binary-empty"},
		},

		"EqObjectID": {
			filter:      bson.D{{"value", bson.D{{"$eq", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011"))}}}},
			expectedIDs: []any{"objectid"},
		},
		"EqObjectIDEmpty": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NilObjectID}}}},
			expectedIDs: []any{"objectid-empty"},
		},

		"EqBoolFalse": {
			filter:      bson.D{{"value", bson.D{{"$eq", false}}}},
			expectedIDs: []any{"bool-false"},
		},
		"EqBoolTrue": {
			filter:      bson.D{{"value", bson.D{{"$eq", true}}}},
			expectedIDs: []any{"bool-true"},
		},

		"EqDatetime": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))}}}},
			expectedIDs: []any{"datetime"},
		},
		"EqDatetimeEpoch": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
			expectedIDs: []any{"datetime-epoch"},
		},
		"EqDatetimeYearMax": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))}}}},
			expectedIDs: []any{"datetime-year-min"},
		},
		"EqDatetimeYearMin": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC))}}}},
			expectedIDs: []any{"datetime-year-max"},
		},

		"EqNull": {
			filter:      bson.D{{"value", bson.D{{"$eq", nil}}}},
			expectedIDs: []any{"array-three", "null"},
		},

		"EqRegexWithoutOption": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Regex{Pattern: "foo"}}}}},
			expectedIDs: []any{},
		},
		"EqRegexWithOption": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Regex{Pattern: "foo", Options: "i"}}}}},
			expectedIDs: []any{"regex"},
		},
		"EqRegexEmpty": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Regex{}}}}},
			expectedIDs: []any{"regex-empty"},
		},

		"EqInt32": {
			filter:      bson.D{{"value", bson.D{{"$eq", int32(42)}}}},
			expectedIDs: []any{"array", "array-three", "double-whole", "int32", "int64"},
		},
		"EqInt32Zero": {
			filter:      bson.D{{"value", bson.D{{"$eq", int32(0)}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"EqInt32Max": {
			filter:      bson.D{{"value", bson.D{{"$eq", int32(math.MaxInt32)}}}},
			expectedIDs: []any{"int32-max"},
		},
		"EqInt32Min": {
			filter:      bson.D{{"value", bson.D{{"$eq", int32(math.MinInt32)}}}},
			expectedIDs: []any{"int32-min"},
		},

		"EqTimestamp": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Timestamp{T: 42, I: 13}}}}},
			expectedIDs: []any{"timestamp"},
		},
		"EqTimestampI": {
			filter:      bson.D{{"value", bson.D{{"$eq", primitive.Timestamp{I: 1}}}}},
			expectedIDs: []any{"timestamp-i"},
		},

		"EqInt64": {
			filter:      bson.D{{"value", bson.D{{"$eq", int64(42)}}}},
			expectedIDs: []any{"array", "array-three", "double-whole", "int32", "int64"},
		},
		"EqInt64Zero": {
			filter:      bson.D{{"value", bson.D{{"$eq", int64(0)}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"EqInt64Max": {
			filter:      bson.D{{"value", bson.D{{"$eq", int64(math.MaxInt64)}}}},
			expectedIDs: []any{"int64-max"},
		},
		"EqInt64Min": {
			filter:      bson.D{{"value", bson.D{{"$eq", int64(math.MinInt64)}}}},
			expectedIDs: []any{"int64-min"},
		},

		"EqIDNull": {
			filter:      bson.D{{"_id", bson.D{{"$eq", nil}}}},
			expectedIDs: []any{},
		},
		"EqNoSuchFieldNull": {
			filter: bson.D{{"no-such-field", bson.D{{"$eq", nil}}}},
			expectedIDs: []any{
				"array", "array-empty", "array-three",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-empty",
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

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)
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
	}{
		"GtDouble": {
			value: 41.13,
			expectedIDs: []any{
				"array", "array-three",
				"double", "double-max", "double-positive-infinity", "double-whole",
				"int32", "int32-max",
				"int64", "int64-max",
			},
		},
		"GtDoubleNegativeZero": {
			value: math.Copysign(0, -1),
			expectedIDs: []any{
				"array", "array-three",
				"double", "double-max", "double-positive-infinity", "double-smallest", "double-whole",
				"int32", "int32-max",
				"int64", "int64-max",
			},
		},
		"GtDoubleMax": {
			value:       math.MaxFloat64,
			expectedIDs: []any{"double-positive-infinity"},
		},
		"GtDoublePositiveInfinity": {
			value:       math.Inf(+1),
			expectedIDs: []any{},
		},
		"GtDoubleNaN": {
			value:       math.NaN(),
			expectedIDs: []any{},
		},

		"GtString": {
			value:       "boo",
			expectedIDs: []any{"array-three", "string"},
		},
		"GtStringWhole": {
			value:       "42",
			expectedIDs: []any{"array-three", "string", "string-double"},
		},
		"GtStringEmpty": {
			value:       "",
			expectedIDs: []any{"array-three", "string", "string-double", "string-whole"},
		},

		"GtBinary": {
			value:       primitive.Binary{Subtype: 0x80, Data: []byte{42}},
			expectedIDs: []any{"binary"},
		},
		"GtBinaryNoSubtype": {
			value:       primitive.Binary{Data: []byte{42}},
			expectedIDs: []any{"binary"},
		},
		"GtBinaryEmpty": {
			value:       primitive.Binary{},
			expectedIDs: []any{"binary"},
		},

		"GtObjectID": {
			value:       must.NotFail(primitive.ObjectIDFromHex("000102030405060708091010")),
			expectedIDs: []any{"objectid"},
		},
		"GtObjectIDEmpty": {
			value:       primitive.NilObjectID,
			expectedIDs: []any{"objectid"},
		},

		"GtBool": {
			value:       false,
			expectedIDs: []any{"bool-true"},
		},

		"GtDatetime": {
			value:       time.Date(2021, 11, 1, 10, 18, 41, 123000000, time.UTC),
			expectedIDs: []any{"datetime", "datetime-year-max"},
		},

		"GtNull": {
			value:       nil,
			expectedIDs: []any{},
		},

		// TODO
		// "GtRegex": {
		// 	value:       primitive.Regex{Pattern: "foo"},
		// 	expectedIDs: []any{},
		// },

		"GtInt32": {
			value:       int32(42),
			expectedIDs: []any{"double", "double-max", "double-positive-infinity", "int32-max", "int64-max"},
		},
		"GtInt32Max": {
			value:       int32(math.MaxInt32),
			expectedIDs: []any{"double-max", "double-positive-infinity", "int64-max"},
		},

		"GtTimestamp": {
			value:       primitive.Timestamp{T: 41, I: 12},
			expectedIDs: []any{"timestamp"},
		},
		"GtTimestampNoI": {
			value:       primitive.Timestamp{T: 41},
			expectedIDs: []any{"timestamp"},
		},
		"GtTimestampNoT": {
			value:       primitive.Timestamp{I: 12},
			expectedIDs: []any{"timestamp"},
		},

		"GtInt64": {
			value:       int64(42),
			expectedIDs: []any{"double", "double-max", "double-positive-infinity", "int32-max", "int64-max"},
		},
		"GtInt64Max": {
			value:       int64(math.MaxInt64),
			expectedIDs: []any{"double-max", "double-positive-infinity"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			filter := bson.D{{"value", bson.D{{"$gt", tc.value}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)
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
		q           bson.D
		expectedIDs []any
	}{
		"GteDouble": {
			q:           bson.D{{"value", bson.D{{"$gte", 42.13}}}},
			expectedIDs: []any{"double", "double-max", "double-positive-infinity", "int32-max", "int64-max"},
		},
		"GteDoublePositiveInfinity": {
			q:           bson.D{{"value", bson.D{{"$gte", math.Inf(+1)}}}},
			expectedIDs: []any{"double-positive-infinity"},
		},
		"GteDoubleMax": {
			q:           bson.D{{"value", bson.D{{"$gte", math.MaxFloat64}}}},
			expectedIDs: []any{"double-max", "double-positive-infinity"},
		},

		"GteString": {
			q:           bson.D{{"value", bson.D{{"$gte", "foo"}}}},
			expectedIDs: []any{"array-three", "string"},
		},

		"GteEmptyString": {
			q:           bson.D{{"value", bson.D{{"$gte", ""}}}},
			expectedIDs: []any{"array-three", "string", "string-empty"},
		},

		"GteInt32": {
			q:           bson.D{{"value", bson.D{{"$gte", int32(42)}}}},
			expectedIDs: []any{"array", "array-three", "double", "double-max", "double-positive-infinity", "int32", "int32-max", "int64", "int64-max"},
		},

		"GteInt32Max": {
			q:           bson.D{{"value", bson.D{{"$gte", int32(math.MaxInt32)}}}},
			expectedIDs: []any{"double-max", "double-positive-infinity", "int32-max", "int64-max"},
		},

		"GteInt64": {
			q:           bson.D{{"value", bson.D{{"$gte", int64(42)}}}},
			expectedIDs: []any{"array", "array-three", "double", "double-max", "double-positive-infinity", "int32", "int32-max", "int64", "int64-max"},
		},

		"GteInt64Max": {
			q:           bson.D{{"value", bson.D{{"$gte", int64(math.MaxInt64)}}}},
			expectedIDs: []any{"double-max", "double-positive-infinity", "int64-max"},
		},

		"GteDatetime": {
			q:           bson.D{{"value", bson.D{{"$gte", time.Date(2021, 11, 1, 10, 18, 41, 123000000, time.UTC)}}}},
			expectedIDs: []any{"datetime", "datetime-year-max"},
		},

		"GteTimeStamp": {
			q:           bson.D{{"value", bson.D{{"$gte", primitive.Timestamp{T: 41, I: 13}}}}},
			expectedIDs: []any{"timestamp"},
		},
		"GteNull": {
			q:           bson.D{{"value", bson.D{{"$gte", nil}}}},
			expectedIDs: []any{"array-three", "null"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.q, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)
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
		filter      bson.D
		expectedIDs []any
	}{
		"LtDouble": {
			filter:      bson.D{{"value", bson.D{{"$lt", 42.13}}}},
			expectedIDs: []any{"array", "array-three", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32", "int32-min", "int32-zero", "int64", "int64-min", "int64-zero"},
		},
		"LtDoubleNegativeInfinity": {
			filter:      bson.D{{"value", bson.D{{"$lt", math.Inf(-1)}}}},
			expectedIDs: []any{},
		},
		"LtDoubleSmallest": {
			filter:      bson.D{{"value", bson.D{{"$lt", math.SmallestNonzeroFloat64}}}},
			expectedIDs: []any{"double-negative-infinity", "double-negative-zero", "double-zero", "int32-min", "int32-zero", "int64-min", "int64-zero"},
		},

		"LtString": {
			filter:      bson.D{{"value", bson.D{{"$lt", "goo"}}}},
			expectedIDs: []any{"array-three", "string", "string-empty"},
		},

		"LtEmptyString": {
			filter:      bson.D{{"value", bson.D{{"$lt", ""}}}},
			expectedIDs: []any{},
		},

		"LtInt32": {
			filter:      bson.D{{"value", bson.D{{"$lt", int32(42)}}}},
			expectedIDs: []any{"double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32-min", "int32-zero", "int64-min", "int64-zero"},
		},

		"LtInt32Min": {
			filter:      bson.D{{"value", bson.D{{"$lt", int32(math.MinInt32)}}}},
			expectedIDs: []any{"double-negative-infinity", "int64-min"},
		},

		"LtInt64": {
			filter:      bson.D{{"value", bson.D{{"$lt", int64(42)}}}},
			expectedIDs: []any{"double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32-min", "int32-zero", "int64-min", "int64-zero"},
		},

		"LtInt64Min": {
			filter:      bson.D{{"value", bson.D{{"$lt", int64(math.MinInt64)}}}},
			expectedIDs: []any{"double-negative-infinity"},
		},

		"LtDatetime": {
			filter:      bson.D{{"value", bson.D{{"$lt", time.Date(2021, 11, 1, 10, 18, 41, 123000000, time.UTC)}}}},
			expectedIDs: []any{"datetime-epoch", "datetime-year-min"},
		},

		"LtTimeStamp": {
			filter:      bson.D{{"value", bson.D{{"$lt", primitive.Timestamp{T: 41, I: 13}}}}},
			expectedIDs: []any{"timestamp-i"},
		},
		"LtNull": {
			filter:      bson.D{{"value", bson.D{{"$lt", nil}}}},
			expectedIDs: []any{},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)
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
		filter      bson.D
		expectedIDs []any
	}{
		"LteDouble": {
			filter:      bson.D{{"value", bson.D{{"$lte", 42.13}}}},
			expectedIDs: []any{"array", "array-three", "double", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32", "int32-min", "int32-zero", "int64", "int64-min", "int64-zero"},
		},
		"LteDoubleNegativeInfinity": {
			filter:      bson.D{{"value", bson.D{{"$lte", math.Inf(-1)}}}},
			expectedIDs: []any{"double-negative-infinity"},
		},
		"LteDoubleSmallest": {
			filter:      bson.D{{"value", bson.D{{"$lte", math.SmallestNonzeroFloat64}}}},
			expectedIDs: []any{"double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32-min", "int32-zero", "int64-min", "int64-zero"},
		},

		"LteString": {
			filter:      bson.D{{"value", bson.D{{"$lte", "foo"}}}},
			expectedIDs: []any{"array-three", "string", "string-empty"},
		},

		"LteEmptyString": {
			filter:      bson.D{{"value", bson.D{{"$lte", ""}}}},
			expectedIDs: []any{"string-empty"},
		},

		"LteInt32": {
			filter:      bson.D{{"value", bson.D{{"$lte", int32(42)}}}},
			expectedIDs: []any{"array", "array-three", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32", "int32-min", "int32-zero", "int64", "int64-min", "int64-zero"},
		},

		"LteInt32Min": {
			filter:      bson.D{{"value", bson.D{{"$lte", int32(math.MinInt32)}}}},
			expectedIDs: []any{"double-negative-infinity", "int32-min", "int64-min"},
		},

		"LteInt64": {
			filter:      bson.D{{"value", bson.D{{"$lte", int64(42)}}}},
			expectedIDs: []any{"array", "array-three", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32", "int32-min", "int32-zero", "int64", "int64-min", "int64-zero"},
		},

		"LteInt64Min": {
			filter:      bson.D{{"value", bson.D{{"$lte", int64(math.MinInt64)}}}},
			expectedIDs: []any{"double-negative-infinity", "int64-min"},
		},

		"LteDatetime": {
			filter:      bson.D{{"value", bson.D{{"$lte", time.Date(2021, 11, 1, 10, 18, 41, 123000000, time.UTC)}}}},
			expectedIDs: []any{"datetime-epoch", "datetime-year-min"},
		},

		"LteTimeStamp": {
			filter:      bson.D{{"value", bson.D{{"$lte", primitive.Timestamp{T: 41, I: 13}}}}},
			expectedIDs: []any{"timestamp-i"},
		},
		"LteNull": {
			filter:      bson.D{{"value", bson.D{{"$lte", nil}}}},
			expectedIDs: []any{"array-three", "null"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}

// $in

// $ne

// $nin
