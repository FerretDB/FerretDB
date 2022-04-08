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
)

func TestQueryComparisonEq(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		q           bson.D
		expectedIDs []any
	}{
		"EqDouble": {
			q:           bson.D{{"value", bson.D{{"$eq", 42.13}}}},
			expectedIDs: []any{"double"},
		},
		"EqDoubleNegativeInfinity": {
			q:           bson.D{{"value", bson.D{{"$eq", math.Inf(-1)}}}},
			expectedIDs: []any{"double-negative-infinity"},
		},
		"EqDoubleNegativeZero": {
			q:           bson.D{{"value", bson.D{{"$eq", math.Copysign(0, -1)}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"EqDoublePositiveInfinity": {
			q:           bson.D{{"value", bson.D{{"$eq", math.Inf(+1)}}}},
			expectedIDs: []any{"double-positive-infinity"},
		},
		"EqDoubleMax": {
			q:           bson.D{{"value", bson.D{{"$eq", math.MaxFloat64}}}},
			expectedIDs: []any{"double-max"},
		},
		"EqDoubleSmallest": {
			q:           bson.D{{"value", bson.D{{"$eq", math.SmallestNonzeroFloat64}}}},
			expectedIDs: []any{"double-smallest"},
		},
		"EqDoubleZero": {
			q:           bson.D{{"value", bson.D{{"$eq", 0.0}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"EqDoubleNaN": {
			q:           bson.D{{"value", bson.D{{"$eq", math.NaN()}}}},
			expectedIDs: []any{"double-nan"},
		},

		"EqString": {
			q:           bson.D{{"value", bson.D{{"$eq", "foo"}}}},
			expectedIDs: []any{"array-three", "string"},
		},
		"EqEmptyString": {
			q:           bson.D{{"value", bson.D{{"$eq", ""}}}},
			expectedIDs: []any{"string-empty"},
		},

		"EqBinary": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
			expectedIDs: []any{"binary"},
		},
		"EqEmptyBinary": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.Binary{Data: []byte{}}}}}},
			expectedIDs: []any{"binary-empty"},
		},

		"EqBoolFalse": {
			q:           bson.D{{"value", bson.D{{"$eq", false}}}},
			expectedIDs: []any{"bool-false"},
		},
		"EqBoolTrue": {
			q:           bson.D{{"value", bson.D{{"$eq", true}}}},
			expectedIDs: []any{"bool-true"},
		},

		"EqDatetime": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))}}}},
			expectedIDs: []any{"datetime"},
		},
		"EqDatetimeEpoch": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
			expectedIDs: []any{"datetime-epoch"},
		},
		"EqDatetimeYearMax": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))}}}},
			expectedIDs: []any{"datetime-year-min"},
		},
		"EqDatetimeYearMin": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC))}}}},
			expectedIDs: []any{"datetime-year-max"},
		},

		"EqTimestamp": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.Timestamp{T: 42, I: 13}}}}},
			expectedIDs: []any{"timestamp"},
		},
		"EqTimestampI": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.Timestamp{I: 1}}}}},
			expectedIDs: []any{"timestamp-i"},
		},

		"EqNull": {
			q:           bson.D{{"value", bson.D{{"$eq", nil}}}},
			expectedIDs: []any{"array-three", "null"},
		},

		"EqFindRegexWithoutOption": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.Regex{Pattern: "foo"}}}}},
			expectedIDs: []any{},
		},
		"EqFindRegexWithOption": {
			q:           bson.D{{"value", bson.D{{"$eq", primitive.Regex{Pattern: "foo", Options: "i"}}}}},
			expectedIDs: []any{"regex"},
		},

		"EqInt32": {
			q:           bson.D{{"value", bson.D{{"$eq", int32(42)}}}},
			expectedIDs: []any{"array", "array-three", "int32", "int64"},
		},
		"EqInt32Zero": {
			q:           bson.D{{"value", bson.D{{"$eq", int32(0)}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"EqInt32Max": {
			q:           bson.D{{"value", bson.D{{"$eq", int32(math.MaxInt32)}}}},
			expectedIDs: []any{"int32-max"},
		},
		"EqInt32Min": {
			q:           bson.D{{"value", bson.D{{"$eq", int32(math.MinInt32)}}}},
			expectedIDs: []any{"int32-min"},
		},

		"EqInt64": {
			q:           bson.D{{"value", bson.D{{"$eq", int64(42)}}}},
			expectedIDs: []any{"array", "array-three", "int32", "int64"},
		},
		"EqInt64Zero": {
			q:           bson.D{{"value", bson.D{{"$eq", int64(0)}}}},
			expectedIDs: []any{"double-negative-zero", "double-zero", "int32-zero", "int64-zero"},
		},
		"EqInt64Max": {
			q:           bson.D{{"value", bson.D{{"$eq", int64(math.MaxInt64)}}}},
			expectedIDs: []any{"int64-max"},
		},
		"EqInt64Min": {
			q:           bson.D{{"value", bson.D{{"$eq", int64(math.MinInt64)}}}},
			expectedIDs: []any{"int64-min"},
		},

		"EqNoSuchFieldNull": {
			q: bson.D{{"no-such-field", bson.D{{"$eq", nil}}}},
			expectedIDs: []any{
				"array", "array-empty", "array-three",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-empty",
				"double", "double-max", "double-nan",
				"double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero", "int64", "int64-max", "int64-min", "int64-zero",
				"null", "regex", "regex-empty", "string", "string-empty", "timestamp", "timestamp-i",
			},
		},
		"EqStringNull": {
			q:           bson.D{{"_id", bson.D{{"$eq", nil}}}},
			expectedIDs: []any{},
		},

		"EqCompareNoSuchField": {
			q: bson.D{{"no-such-field", nil}},
			expectedIDs: []any{
				"array", "array-empty", "array-three",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-empty",
				"double", "double-max", "double-nan",
				"double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero", "int64", "int64-max", "int64-min", "int64-zero",
				"null", "regex", "regex-empty", "string", "string-empty", "timestamp", "timestamp-i",
			},
		},
		"EqCompareWithNull": {
			q:           bson.D{{"_id", nil}},
			expectedIDs: []any{},
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

func TestQueryComparisonGt(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		q           bson.D
		expectedIDs []any
	}{
		"GtDouble": {
			q:           bson.D{{"value", bson.D{{"$gt", 40.123}}}},
			expectedIDs: []any{"array", "array-three", "double", "double-max", "double-positive-infinity", "int32", "int32-max", "int64", "int64-max"},
		},
		"GtDoublePositiveInfinity": {
			q:           bson.D{{"value", bson.D{{"$gt", math.Inf(+1)}}}},
			expectedIDs: []any{},
		},
		"GtDoubleMax": {
			q:           bson.D{{"value", bson.D{{"$gt", math.MaxFloat64}}}},
			expectedIDs: []any{"double-positive-infinity"},
		},

		"GtString": {
			q:           bson.D{{"value", bson.D{{"$gt", "boo"}}}},
			expectedIDs: []any{"array-three", "string"},
		},

		"GtEmptyString": {
			q:           bson.D{{"value", bson.D{{"$gt", ""}}}},
			expectedIDs: []any{"array-three", "string"},
		},

		"GtInt32": {
			q:           bson.D{{"value", bson.D{{"$gt", int32(42)}}}},
			expectedIDs: []any{"double", "double-max", "double-positive-infinity", "int32-max", "int64-max"},
		},

		"GtInt32Max": {
			q:           bson.D{{"value", bson.D{{"$gt", int32(math.MaxInt32)}}}},
			expectedIDs: []any{"double-max", "double-positive-infinity", "int64-max"},
		},

		"GtInt64": {
			q:           bson.D{{"value", bson.D{{"$gt", int64(42)}}}},
			expectedIDs: []any{"double", "double-max", "double-positive-infinity", "int32-max", "int64-max"},
		},

		"GtInt64Max": {
			q:           bson.D{{"value", bson.D{{"$gt", int64(math.MaxInt64)}}}},
			expectedIDs: []any{"double-max", "double-positive-infinity"},
		},

		"GtDatetime": {
			q:           bson.D{{"value", bson.D{{"$gt", time.Date(2021, 11, 1, 10, 18, 41, 123000000, time.UTC)}}}},
			expectedIDs: []any{"datetime", "datetime-year-max"},
		},

		"GtTimeStamp": {
			q:           bson.D{{"value", bson.D{{"$gt", primitive.Timestamp{T: 41, I: 13}}}}},
			expectedIDs: []any{"timestamp"},
		},
		"GtNull": {
			q:           bson.D{{"value", bson.D{{"$gt", nil}}}},
			expectedIDs: []any{},
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
		q           bson.D
		expectedIDs []any
	}{
		"LtDouble": {
			q:           bson.D{{"value", bson.D{{"$lt", 42.13}}}},
			expectedIDs: []any{"array", "array-three", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32", "int32-min", "int32-zero", "int64", "int64-min", "int64-zero"},
		},
		"LtDoubleNegativeInfinity": {
			q:           bson.D{{"value", bson.D{{"$lt", math.Inf(-1)}}}},
			expectedIDs: []any{},
		},
		"LtDoubleSmallest": {
			q:           bson.D{{"value", bson.D{{"$lt", math.SmallestNonzeroFloat64}}}},
			expectedIDs: []any{"double-negative-infinity", "double-negative-zero", "double-zero", "int32-min", "int32-zero", "int64-min", "int64-zero"},
		},

		"LtString": {
			q:           bson.D{{"value", bson.D{{"$lt", "goo"}}}},
			expectedIDs: []any{"array-three", "string", "string-empty"},
		},

		"LtEmptyString": {
			q:           bson.D{{"value", bson.D{{"$lt", ""}}}},
			expectedIDs: []any{},
		},

		"LtInt32": {
			q:           bson.D{{"value", bson.D{{"$lt", int32(42)}}}},
			expectedIDs: []any{"double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32-min", "int32-zero", "int64-min", "int64-zero"},
		},

		"LtInt32Min": {
			q:           bson.D{{"value", bson.D{{"$lt", int32(math.MinInt32)}}}},
			expectedIDs: []any{"double-negative-infinity", "int64-min"},
		},

		"LtInt64": {
			q:           bson.D{{"value", bson.D{{"$lt", int64(42)}}}},
			expectedIDs: []any{"double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32-min", "int32-zero", "int64-min", "int64-zero"},
		},

		"LtInt64Min": {
			q:           bson.D{{"value", bson.D{{"$lt", int64(math.MinInt64)}}}},
			expectedIDs: []any{"double-negative-infinity"},
		},

		"LtDatetime": {
			q:           bson.D{{"value", bson.D{{"$lt", time.Date(2021, 11, 1, 10, 18, 41, 123000000, time.UTC)}}}},
			expectedIDs: []any{"datetime-epoch", "datetime-year-min"},
		},

		"LtTimeStamp": {
			q:           bson.D{{"value", bson.D{{"$lt", primitive.Timestamp{T: 41, I: 13}}}}},
			expectedIDs: []any{"timestamp-i"},
		},
		"LtNull": {
			q:           bson.D{{"value", bson.D{{"$lt", nil}}}},
			expectedIDs: []any{},
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

func TestQueryComparisonLte(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		q           bson.D
		expectedIDs []any
	}{
		"LteDouble": {
			q:           bson.D{{"value", bson.D{{"$lte", 42.13}}}},
			expectedIDs: []any{"array", "array-three", "double", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32", "int32-min", "int32-zero", "int64", "int64-min", "int64-zero"},
		},
		"LteDoubleNegativeInfinity": {
			q:           bson.D{{"value", bson.D{{"$lte", math.Inf(-1)}}}},
			expectedIDs: []any{"double-negative-infinity"},
		},
		"LteDoubleSmallest": {
			q:           bson.D{{"value", bson.D{{"$lte", math.SmallestNonzeroFloat64}}}},
			expectedIDs: []any{"double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32-min", "int32-zero", "int64-min", "int64-zero"},
		},

		"LteString": {
			q:           bson.D{{"value", bson.D{{"$lte", "foo"}}}},
			expectedIDs: []any{"array-three", "string", "string-empty"},
		},

		"LteEmptyString": {
			q:           bson.D{{"value", bson.D{{"$lte", ""}}}},
			expectedIDs: []any{"string-empty"},
		},

		"LteInt32": {
			q:           bson.D{{"value", bson.D{{"$lte", int32(42)}}}},
			expectedIDs: []any{"array", "array-three", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32", "int32-min", "int32-zero", "int64", "int64-min", "int64-zero"},
		},

		"LteInt32Min": {
			q:           bson.D{{"value", bson.D{{"$lte", int32(math.MinInt32)}}}},
			expectedIDs: []any{"double-negative-infinity", "int32-min", "int64-min"},
		},

		"LteInt64": {
			q:           bson.D{{"value", bson.D{{"$lte", int64(42)}}}},
			expectedIDs: []any{"array", "array-three", "double-negative-infinity", "double-negative-zero", "double-smallest", "double-zero", "int32", "int32-min", "int32-zero", "int64", "int64-min", "int64-zero"},
		},

		"LteInt64Min": {
			q:           bson.D{{"value", bson.D{{"$lte", int64(math.MinInt64)}}}},
			expectedIDs: []any{"double-negative-infinity", "int64-min"},
		},

		"LteDatetime": {
			q:           bson.D{{"value", bson.D{{"$lte", time.Date(2021, 11, 1, 10, 18, 41, 123000000, time.UTC)}}}},
			expectedIDs: []any{"datetime-epoch", "datetime-year-min"},
		},

		"LteTimeStamp": {
			q:           bson.D{{"value", bson.D{{"$lte", primitive.Timestamp{T: 41, I: 13}}}}},
			expectedIDs: []any{"timestamp-i"},
		},
		"LteNull": {
			q:           bson.D{{"value", bson.D{{"$lte", nil}}}},
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

// $in

// $ne

// $nin
