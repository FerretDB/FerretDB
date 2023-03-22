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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestQueryComparisonCompatImplicit(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Document": {
			filter:        bson.D{{"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"DocumentReverse": {
			filter:        bson.D{{"v", bson.D{{"array", bson.A{int32(42), "foo", nil}}, {"42", "foo"}, {"foo", int32(42)}}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"DocumentNull": {
			filter:        bson.D{{"v", bson.D{{"foo", nil}}}},
			skipForTigris: "Tigris does not support null values in objects",
		},
		"DocumentEmpty": {
			filter: bson.D{{"v", bson.D{}}},
		},
		"DocumentShuffledKeys": {
			filter:     bson.D{{"v", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}},
			resultType: emptyResult,
		},
		"DocumentDotNotation": {
			filter:        bson.D{{"v.foo", int32(42)}},
			skipForTigris: "No suitable Tigris-compatible provider to test this data",
		},
		"DocumentDotNotationNoSuchField": {
			filter:     bson.D{{"no-such-field.some", 42}},
			resultType: emptyResult,
		},
		"ArrayNoSuchField": {
			filter:     bson.D{{"no-such-field", bson.A{42}}},
			resultType: emptyResult,
		},
		"ArrayShuffledValues": {
			filter:     bson.D{{"v", bson.A{"foo", nil, int32(42)}}},
			resultType: emptyResult,
		},
		"ArrayDotNotationNoSuchField": {
			filter:     bson.D{{"v.some.0", bson.A{42}}},
			resultType: emptyResult,
		},
		"Int32": {
			filter:         bson.D{{"v", int32(42)}},
			resultPushdown: true,
		},
		"Int64": {
			filter:         bson.D{{"v", int64(42)}},
			resultPushdown: true,
		},
		"Double": {
			filter:         bson.D{{"v", 42.13}},
			resultPushdown: true,
		},
		"DoubleMax": {
			filter:         bson.D{{"v", math.MaxFloat64}},
			resultPushdown: true,
		},
		"DoubleSmallest": {
			filter:         bson.D{{"v", math.SmallestNonzeroFloat64}},
			resultPushdown: true,
		},
		"DoubleBig": {
			filter:         bson.D{{"v", float64(1 << 61)}},
			resultPushdown: true,
		},
		"DoubleBigPlus": {
			filter:         bson.D{{"v", float64((1 << 61) + 1)}},
			resultPushdown: true,
		},
		"DoubleBigMinus": {
			filter:         bson.D{{"v", float64((1 << 61) - 1)}},
			resultPushdown: true,
		},
		"DoubleNegBig": {
			filter:         bson.D{{"v", -float64(1 << 61)}},
			resultPushdown: true,
			skipForTigris:  "Values smaller than -(1<<53) are not pushdowned on Tigris",
		},
		"DoubleNegBigPlus": {
			filter:         bson.D{{"v", -float64(1<<61) + 1}},
			resultPushdown: true,
			skipForTigris:  "Values smaller than -(1<<53) are not pushdowned on Tigris",
		},
		"DoubleNegBigMinus": {
			filter:         bson.D{{"v", -float64(1<<61) - 1}},
			resultPushdown: true,
			skipForTigris:  "Values smaller than -(1<<53) are not pushdowned on Tigris",
		},
		"Int64Max": {
			filter:         bson.D{{"v", int64(math.MaxInt64)}},
			resultPushdown: true,
		},
		"Int64Min": {
			filter:         bson.D{{"v", int64(math.MinInt64)}},
			resultPushdown: true,
		},

		"Int64Big": {
			filter:         bson.D{{"v", int64(1 << 61)}},
			resultPushdown: true,
		},
		"Int64BigPlus": {
			filter:         bson.D{{"v", int64(1<<61) + 1}},
			resultPushdown: true,
		},
		"Int64BigMinus": {
			filter:         bson.D{{"v", int64(1<<61) - 1}},
			resultPushdown: true,
		},
		"Int64NegBig": {
			filter:         bson.D{{"v", -int64(1 << 61)}},
			resultPushdown: true,
		},
		"Int64NegBigPlus": {
			filter:         bson.D{{"v", -int64(1<<61) + 1}},
			resultPushdown: true,
		},
		"Int64NegBigMinus": {
			filter:         bson.D{{"v", -int64(1<<61) - 1}},
			resultPushdown: true,
		},

		"String": {
			filter:         bson.D{{"v", "foo"}},
			resultPushdown: true,
		},
		"StringInt": {
			filter:         bson.D{{"v", "42"}},
			resultPushdown: true,
		},
		"StringDouble": {
			filter:         bson.D{{"v", "42.13"}},
			resultPushdown: true,
		},
		"StringEmpty": {
			filter:         bson.D{{"v", ""}},
			resultPushdown: true,
		},
		"Binary": {
			filter: bson.D{{"v", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}},
		},
		"BinaryEmpty": {
			filter: bson.D{{"v", primitive.Binary{}}},
		},
		"BoolFalse": {
			filter:         bson.D{{"v", false}},
			resultPushdown: true,
		},
		"BoolTrue": {
			filter:         bson.D{{"v", true}},
			resultPushdown: true,
		},
		"Datetime": {
			filter:         bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))}},
			resultPushdown: true,
		},
		"DatetimeEpoch": {
			filter:         bson.D{{"v", primitive.NewDateTimeFromTime(time.Unix(0, 0))}},
			resultPushdown: true,
		},
		"DatetimeYearMin": {
			filter:         bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))}},
			resultPushdown: true,
		},
		"DatetimeYearMax": {
			filter:         bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC))}},
			resultPushdown: true,
		},
		"IDNull": {
			filter:     bson.D{{"_id", nil}},
			resultType: emptyResult,
		},
		"IDInt32": {
			filter:         bson.D{{"_id", int32(1)}},
			resultType:     emptyResult,
			resultPushdown: true,
		},
		"IDInt64": {
			filter:         bson.D{{"_id", int64(1)}},
			resultType:     emptyResult,
			resultPushdown: true,
		},
		"IDDouble": {
			filter:         bson.D{{"_id", 4.2}},
			resultType:     emptyResult,
			resultPushdown: true,
		},
		"IDString": {
			filter:         bson.D{{"_id", "string"}},
			resultPushdown: true,
		},
		"IDObjectID": {
			filter:         bson.D{{"_id", primitive.NilObjectID}},
			resultPushdown: true,
		},
		"ValueNull": {
			filter: bson.D{{"v", nil}},
		},
		"NoSuchFieldNull": {
			filter: bson.D{{"no-such-field", nil}},
		},
		"ValueNumber": {
			filter:         bson.D{{"v", 42}},
			resultPushdown: true,
		},
		"ValueRegex": {
			filter: bson.D{{"v", primitive.Regex{Pattern: "^fo"}}},
		},

		"EmptyKey": {
			filter:         bson.D{{"", "foo"}},
			resultType:     emptyResult,
			resultPushdown: true,
			skipForTigris:  "Tigris field name cannot be empty",
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryComparisonCompatEq(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Document": {
			filter: bson.D{{"v", bson.D{
				{"$eq", bson.D{
					{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}},
				}},
			}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"DocumentShuffledKeys": {
			filter: bson.D{{"v", bson.D{
				{"$eq", bson.D{
					{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)},
				}},
			}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
			resultType:    emptyResult,
		},
		"DocumentDotNotation": {
			filter: bson.D{{"v.foo", bson.D{{"$eq", int32(42)}}}},
		},
		"DocumentReverse": {
			filter: bson.D{{"v", bson.D{
				{"$eq", bson.D{
					{"array", bson.A{int32(42), "foo", nil}}, {"42", "foo"}, {"foo", int32(42)},
				}},
			}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"DocumentNull": {
			filter:        bson.D{{"v", bson.D{{"$eq", bson.D{{"foo", nil}}}}}},
			skipForTigris: "Tigri does not support null values in objects",
		},
		"DocumentEmpty": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.D{}}}}},
		},
		"Array": {
			filter:        bson.D{{"v", bson.D{{"$eq", bson.A{int32(42), "foo", nil}}}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"ArrayShuffledValues": {
			filter:     bson.D{{"v", bson.D{{"$eq", bson.A{"foo", nil, int32(42)}}}}},
			resultType: emptyResult,
		},
		"ArrayReverse": {
			filter:        bson.D{{"v", bson.D{{"$eq", bson.A{nil, "foo", int32(42)}}}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"ArrayNull": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.A{nil}}}}},
		},
		"ArrayEmpty": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.A{}}}}},
		},
		"Double": {
			filter:         bson.D{{"v", bson.D{{"$eq", 42.13}}}},
			resultPushdown: true,
		},
		"DoubleWhole": {
			filter:         bson.D{{"v", bson.D{{"$eq", 42.0}}}},
			resultPushdown: true,
		},
		"DoubleZero": {
			filter:         bson.D{{"v", bson.D{{"$eq", 0.0}}}},
			resultPushdown: true,
		},
		"DoubleMax": {
			filter:         bson.D{{"v", bson.D{{"$eq", math.MaxFloat64}}}},
			resultPushdown: true,
		},
		"DoubleSmallest": {
			filter:         bson.D{{"v", bson.D{{"$eq", math.SmallestNonzeroFloat64}}}},
			resultPushdown: true,
		},

		"DoubleBig": {
			filter:         bson.D{{"v", bson.D{{"$eq", float64(1 << 61)}}}},
			resultPushdown: true,
		},
		"DoubleBigPlus": {
			filter:         bson.D{{"v", bson.D{{"$eq", float64((1 << 61) + 1)}}}},
			resultPushdown: true,
		},
		"DoubleBigMinus": {
			filter:         bson.D{{"v", bson.D{{"$eq", float64((1 << 61) - 1)}}}},
			resultPushdown: true,
		},
		"DoubleNegBig": {
			filter:         bson.D{{"v", bson.D{{"$eq", -float64(1 << 61)}}}},
			resultPushdown: true,
			skipForTigris:  "Values smaller than -(1<<53) are not pushdowned on Tigris",
		},
		"DoubleNegBigPlus": {
			filter:         bson.D{{"v", bson.D{{"$eq", -float64((1 << 61) + 1)}}}},
			resultPushdown: true,
			skipForTigris:  "Values smaller than -(1<<53) are not pushdowned on Tigris",
		},
		"DoubleNegBigMinus": {
			filter:         bson.D{{"v", bson.D{{"$eq", -float64((1 << 61) - 1)}}}},
			resultPushdown: true,
			skipForTigris:  "Values smaller than -(1<<53) are not pushdowned on Tigris",
		},

		"DoublePrecMax": {
			filter:         bson.D{{"v", bson.D{{"$eq", float64(1 << 53)}}}},
			resultPushdown: true,
		},
		"DoublePrecMaxPlus": {
			filter:         bson.D{{"v", bson.D{{"$eq", float64(1<<53) + 1}}}},
			resultPushdown: true,
		},
		"DoublePrecMaxMinus": {
			filter:         bson.D{{"v", bson.D{{"$eq", float64(1<<53) - 1}}}},
			resultPushdown: true,
		},
		"DoublePrecMin": {
			filter:         bson.D{{"v", bson.D{{"$eq", -float64(1<<53 - 1)}}}},
			resultPushdown: true,
		},
		"DoublePrecMinPlus": {
			filter:         bson.D{{"v", bson.D{{"$eq", -float64(1<<53-1) + 1}}}},
			resultPushdown: true,
		},
		"DoublePrecMinMinus": {
			filter:         bson.D{{"v", bson.D{{"$eq", -float64(1<<53-1) - 1}}}},
			resultPushdown: true,
		},

		"String": {
			filter:         bson.D{{"v", bson.D{{"$eq", "foo"}}}},
			resultPushdown: true,
		},
		"StringDouble": {
			filter:         bson.D{{"v", bson.D{{"$eq", "42.13"}}}},
			resultPushdown: true,
		},
		"StringWhole": {
			filter:         bson.D{{"v", bson.D{{"$eq", "42"}}}},
			resultPushdown: true,
		},
		"StringEmpty": {
			filter:         bson.D{{"v", bson.D{{"$eq", ""}}}},
			resultPushdown: true,
		},
		"Binary": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
		},
		"BinaryEmpty": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Binary{Data: []byte{}}}}}},
		},
		"ObjectID": {
			filter:         bson.D{{"v", bson.D{{"$eq", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011"))}}}},
			resultPushdown: true,
		},
		"ObjectIDEmpty": {
			filter:         bson.D{{"v", bson.D{{"$eq", primitive.NilObjectID}}}},
			resultPushdown: true,
		},
		"BoolFalse": {
			filter:         bson.D{{"v", bson.D{{"$eq", false}}}},
			resultPushdown: true,
		},
		"BoolTrue": {
			filter:         bson.D{{"v", bson.D{{"$eq", true}}}},
			resultPushdown: true,
		},
		"Datetime": {
			filter:         bson.D{{"v", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))}}}},
			resultPushdown: true,
		},
		"DatetimeEpoch": {
			filter:         bson.D{{"v", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
			resultPushdown: true,
		},
		"DatetimeYearMin": {
			filter:         bson.D{{"v", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))}}}},
			resultPushdown: true,
		},
		"DatetimeYearMax": {
			filter:         bson.D{{"v", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC))}}}},
			resultPushdown: true,
		},
		"Null": {
			filter: bson.D{{"v", bson.D{{"$eq", nil}}}},
		},
		"RegexWithoutOption": {
			filter:     bson.D{{"v", bson.D{{"$eq", primitive.Regex{Pattern: "foo"}}}}},
			resultType: emptyResult,
		},
		"RegexWithOption": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Regex{Pattern: "foo", Options: "i"}}}}},
		},
		"RegexEmpty": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Regex{}}}}},
		},
		"Int32": {
			filter:         bson.D{{"v", bson.D{{"$eq", int32(42)}}}},
			resultPushdown: true,
		},
		"Int32Zero": {
			filter:         bson.D{{"v", bson.D{{"$eq", int32(0)}}}},
			resultPushdown: true,
		},
		"Int32Max": {
			filter:         bson.D{{"v", bson.D{{"$eq", int32(math.MaxInt32)}}}},
			resultPushdown: true,
		},
		"Int32Min": {
			filter:         bson.D{{"v", bson.D{{"$eq", int32(math.MinInt32)}}}},
			resultPushdown: true,
		},
		"Timestamp": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Timestamp{T: 42, I: 13}}}}},
		},
		"TimestampI": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Timestamp{I: 1}}}}},
		},
		"Int64": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(42)}}}},
			resultPushdown: true,
		},
		"Int64Zero": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(0)}}}},
			resultPushdown: true,
		},
		"Int64Max": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(math.MaxInt64)}}}},
			resultPushdown: true,
		},
		"Int64Min": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(math.MinInt64)}}}},
			resultPushdown: true,
		},

		"Int64Big": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(1 << 61)}}}},
			resultPushdown: true,
		},
		"Int64BigPlusOne": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(1<<61) + 1}}}},
			resultPushdown: true,
		},
		"Int64BigMinusOne": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(1<<61) - 1}}}},
			resultPushdown: true,
		},
		"Int64NegBig": {
			filter:         bson.D{{"v", bson.D{{"$eq", -int64(1 << 61)}}}},
			resultPushdown: true,
		},
		"Int64NegBigPlusOne": {
			filter:         bson.D{{"v", bson.D{{"$eq", -int64(1<<61) + 1}}}},
			resultPushdown: true,
		},
		"Int64NegBigMinusOne": {
			filter:         bson.D{{"v", bson.D{{"$eq", -int64(1<<61) - 1}}}},
			resultPushdown: true,
		},

		"Int64PrecMax": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(1 << 53)}}}},
			resultPushdown: true,
		},
		"Int64PrecMaxPlusOne": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(1<<53 + 1)}}}},
			resultPushdown: true,
		},
		"Int64PrecMaxMinusOne": {
			filter:         bson.D{{"v", bson.D{{"$eq", int64(1<<53 - 1)}}}},
			resultPushdown: true,
		},
		"Int64PrecMin": {
			filter:         bson.D{{"v", bson.D{{"$eq", -int64(1<<53 - 1)}}}},
			resultPushdown: true,
		},
		"Int64PrecMinPlus": {
			filter:         bson.D{{"v", bson.D{{"$eq", -int64(1<<53-1) + 1}}}},
			resultPushdown: true,
		},
		"Int64PrecMinMinus": {
			filter:         bson.D{{"v", bson.D{{"$eq", -int64(1<<53-1) - 1}}}},
			resultPushdown: true,
		},

		"IDNull": {
			filter:     bson.D{{"_id", bson.D{{"$eq", nil}}}},
			resultType: emptyResult,
		},
		"NoSuchFieldNull": {
			filter: bson.D{{"no-such-field", bson.D{{"$eq", nil}}}},
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryComparisonCompatGt(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Document": {
			filter: bson.D{{"v", bson.D{
				{"$gt", bson.D{
					{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}},
				}},
			}}},
		},
		"DocumentShuffledKeys": {
			filter: bson.D{{"v", bson.D{
				{"$gt", bson.D{
					{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)},
				}},
			}}},
		},
		"DocumentDotNotation": {
			filter: bson.D{{"v.foo", bson.D{{"$gt", int32(41)}}}},
		},
		"DocumentReverse": {
			filter: bson.D{
				{"v", bson.D{
					{"$gt", bson.D{
						{"array", bson.A{int32(42), "foo", nil}}, {"42", "foo"}, {"foo", int32(42)},
					}},
				}},
			},
			resultPushdown: false,
			skipForTigris:  "No suitable Tigris-compatible provider to test this data",
		},
		"DocumentNull": {
			filter: bson.D{{"v", bson.D{{"$gt", bson.D{{"foo", nil}}}}}},
		},
		"DocumentEmpty": {
			filter: bson.D{{"v", bson.D{{"$gt", bson.D{}}}}},
		},
		"ArrayEmpty": {
			filter: bson.D{{"v", bson.D{{"$gt", bson.A{}}}}},
		},
		"ArrayOne": {
			filter: bson.D{{"v", bson.D{{"$gt", bson.A{int32(42)}}}}},
		},
		"Array": {
			filter: bson.D{{"v", bson.D{{"$gt", bson.A{int32(42), "foo", nil}}}}},
		},
		"ArrayReverse": {
			filter: bson.D{{"v", bson.D{{"$gt", bson.A{nil, "foo", int32(42)}}}}},
		},
		"ArrayNull": {
			filter: bson.D{{"v", bson.D{{"$gt", bson.A{nil}}}}},
		},
		"ArraySlice": {
			filter: bson.D{{"v", bson.D{{"$gt", bson.A{int32(42), "foo"}}}}},
		},
		"ArrayShuffledValues": {
			filter:        bson.D{{"v", bson.D{{"$gt", bson.A{"foo", nil, int32(42)}}}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"Double": {
			filter: bson.D{{"v", bson.D{{"$gt", 41.13}}}},
		},
		"DoubleMax": {
			filter:     bson.D{{"v", bson.D{{"$gt", math.MaxFloat64}}}},
			resultType: emptyResult,
		},
		"String": {
			filter: bson.D{{"v", bson.D{{"$gt", "boo"}}}},
		},
		"StringWhole": {
			filter: bson.D{{"v", bson.D{{"$gt", "42"}}}},
		},
		"StringEmpty": {
			filter: bson.D{{"v", bson.D{{"$gt", ""}}}},
		},
		"Binary": {
			filter: bson.D{{"v", bson.D{{"$gt", primitive.Binary{Subtype: 0x80, Data: []byte{42}}}}}},
		},
		"BinaryNoSubtype": {
			filter: bson.D{{"v", bson.D{{"$gt", primitive.Binary{Data: []byte{42}}}}}},
		},
		"BinaryEmpty": {
			filter: bson.D{{"v", bson.D{{"$gt", primitive.Binary{}}}}},
		},
		"ObjectID": {
			filter: bson.D{{"v", bson.D{{"$gt", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091010"))}}}},
		},
		"ObjectIDEmpty": {
			filter: bson.D{{"v", bson.D{{"$gt", primitive.NilObjectID}}}},
		},
		"Bool": {
			filter: bson.D{{"v", bson.D{{"$gt", false}}}},
		},
		"Datetime": {
			filter: bson.D{{"v", bson.D{{"$gt", time.Date(2021, 11, 1, 10, 18, 41, 123000000, time.UTC)}}}},
		},
		"Null": {
			filter:     bson.D{{"v", bson.D{{"$gt", nil}}}},
			resultType: emptyResult,
		},
		"Regex": {
			filter:     bson.D{{"v", bson.D{{"$gt", primitive.Regex{Pattern: "foo"}}}}},
			resultType: emptyResult,
		},
		"Int32": {
			filter: bson.D{{"v", bson.D{{"$gt", int32(42)}}}},
		},
		"Int32Max": {
			filter: bson.D{{"v", bson.D{{"$gt", int32(math.MaxInt32)}}}},
		},
		"Timestamp": {
			filter: bson.D{{"v", bson.D{{"$gt", primitive.Timestamp{T: 41, I: 12}}}}},
		},
		"TimestampNoI": {
			filter: bson.D{{"v", bson.D{{"$gt", primitive.Timestamp{T: 41}}}}},
		},
		"TimestampNoT": {
			filter: bson.D{{"v", bson.D{{"$gt", primitive.Timestamp{I: 12}}}}},
		},
		"Int64": {
			filter: bson.D{{"v", bson.D{{"$gt", int64(42)}}}},
		},
		"Int64Max": {
			filter: bson.D{{"v", bson.D{{"$gt", int64(math.MaxInt64)}}}},
		},
		"Int64Big": {
			filter: bson.D{{"v", bson.D{{"$gt", int64(1 << 61)}}}},
		},
		"Int64BigPlusOne": {
			filter: bson.D{{"v", bson.D{{"$gt", int64(1<<61) + 1}}}},
		},
		"Int64BigMinusOne": {
			filter: bson.D{{"v", bson.D{{"$gt", int64(1<<61) - 1}}}},
		},
		"Int64NegBig": {
			filter: bson.D{{"v", bson.D{{"$gt", -int64(1 << 61)}}}},
		},
		"Int64NegBigPlusOne": {
			filter: bson.D{{"v", bson.D{{"$gt", -int64(1<<61) + 1}}}},
		},
		"Int64NegBigMinusOne": {
			filter: bson.D{{"v", bson.D{{"$gt", -int64(1<<61) - 1}}}},
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryComparisonCompatGte(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Document": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}}}},
		},
		"DocumentShuffledKeys": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}}}},
		},
		"DocumentDotNotation": {
			filter: bson.D{{"v.foo", bson.D{{"$gte", int32(42)}}}},
		},
		"DocumentReverse": {
			filter:        bson.D{{"v", bson.D{{"$gte", bson.D{{"array", bson.A{int32(42), "foo", nil}}, {"42", "foo"}, {"foo", int32(42)}}}}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"DocumentNull": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.D{{"foo", nil}}}}}},
		},
		"DocumentEmpty": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.D{}}}}},
		},
		"ArrayEmpty": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.A{}}}}},
		},
		"ArrayOne": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.A{int32(42)}}}}},
		},
		"Array": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.A{int32(42), "foo", nil}}}}},
		},
		"ArrayReverse": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.A{nil, "foo", int32(42)}}}}},
		},
		"ArrayNull": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.A{nil}}}}},
		},
		"ArraySlice": {
			filter: bson.D{{"v", bson.D{{"$gte", bson.A{int32(42), "foo"}}}}},
		},
		"ArrayShuffledValues": {
			filter:        bson.D{{"v", bson.D{{"$gte", bson.A{"foo", nil, int32(42)}}}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"Double": {
			filter: bson.D{{"v", bson.D{{"$gte", 41.13}}}},
		},
		"DoubleMax": {
			filter: bson.D{{"v", bson.D{{"$gte", math.MaxFloat64}}}},
		},
		"String": {
			filter: bson.D{{"v", bson.D{{"$gte", "foo"}}}},
		},
		"StringWhole": {
			filter: bson.D{{"v", bson.D{{"$gte", "42"}}}},
		},
		"StringEmpty": {
			filter: bson.D{{"v", bson.D{{"$gte", ""}}}},
		},
		"Binary": {
			filter: bson.D{{"v", bson.D{{"$gte", primitive.Binary{Subtype: 0x80, Data: []byte{42}}}}}},
		},
		"BinaryNoSubtype": {
			filter: bson.D{{"v", bson.D{{"$gte", primitive.Binary{Data: []byte{42}}}}}},
		},
		"BinaryEmpty": {
			filter: bson.D{{"v", bson.D{{"$gte", primitive.Binary{}}}}},
		},
		"ObjectID": {
			filter: bson.D{{"v", bson.D{{"$gte", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011"))}}}},
		},
		"ObjectIDEmpty": {
			filter: bson.D{{"v", bson.D{{"$gte", primitive.NilObjectID}}}},
		},
		"Bool": {
			filter: bson.D{{"v", bson.D{{"$gte", false}}}},
		},
		"Datetime": {
			filter: bson.D{{"v", bson.D{{"$gte", time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC)}}}},
		},
		"Null": {
			filter: bson.D{{"v", bson.D{{"$gte", nil}}}},
		},
		"Regex": {
			filter:     bson.D{{"v", bson.D{{"$gte", primitive.Regex{Pattern: "foo"}}}}},
			resultType: emptyResult,
		},
		"Int32": {
			filter: bson.D{{"v", bson.D{{"$gte", int32(42)}}}},
		},
		"Int32Max": {
			filter: bson.D{{"v", bson.D{{"$gte", int32(math.MaxInt32)}}}},
		},
		"Int32Desc": {
			filter: bson.D{{"v", bson.D{{"$gte", int32(45)}}}},
		},
		"Timestamp": {
			filter: bson.D{{"v", bson.D{{"$gte", primitive.Timestamp{T: 41, I: 12}}}}},
		},
		"TimestampNoI": {
			filter: bson.D{{"v", bson.D{{"$gte", primitive.Timestamp{T: 42}}}}},
		},
		"TimestampNoT": {
			filter: bson.D{{"v", bson.D{{"$gte", primitive.Timestamp{I: 13}}}}},
		},
		"Int64": {
			filter: bson.D{{"v", bson.D{{"$gte", int64(42)}}}},
		},
		"Int64Max": {
			filter: bson.D{{"v", bson.D{{"$gte", int64(math.MaxInt64)}}}},
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryComparisonCompatLt(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Document": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}}}},
		},
		"DocumentShuffledKeys": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}}}},
		},
		"DocumentDotNotation": {
			filter: bson.D{{"v.foo", bson.D{{"$lt", int32(43)}}}},
		},
		"DocumentReverse": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.D{{"array", bson.A{int32(42), "foo", nil}}, {"42", "foo"}, {"foo", int32(42)}}}}}},
		},
		"DocumentNull": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.D{{"foo", nil}}}}}},
		},
		"DocumentEmpty": {
			filter:     bson.D{{"v", bson.D{{"$lt", bson.D{}}}}},
			resultType: emptyResult,
		},
		"ArrayEmpty": {
			filter:     bson.D{{"v", bson.D{{"$lt", bson.A{}}}}},
			resultType: emptyResult,
		},
		"ArrayOne": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.A{int32(42)}}}}},
		},
		"Array": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.A{int32(42), "foo", nil}}}}},
		},
		"ArrayReverse": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.A{nil, "foo", int32(42)}}}}},
		},
		"ArrayNull": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.A{nil}}}}},
		},
		"ArraySlice": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.A{int32(42), "foo"}}}}},
		},
		"ArrayShuffledValues": {
			filter: bson.D{{"v", bson.D{{"$lt", bson.A{"foo", nil, int32(42)}}}}},
		},
		"Double": {
			filter: bson.D{{"v", bson.D{{"$lt", 43.13}}}},
		},
		"DoubleSmallest": {
			filter: bson.D{{"v", bson.D{{"$lt", math.SmallestNonzeroFloat64}}}},
		},
		"String": {
			filter: bson.D{{"v", bson.D{{"$lt", "goo"}}}},
		},
		"StringWhole": {
			filter: bson.D{{"v", bson.D{{"$lt", "42"}}}},
		},
		"StringEmpty": {
			filter:     bson.D{{"v", bson.D{{"$lt", ""}}}},
			resultType: emptyResult,
		},
		"StringAsc": {
			filter: bson.D{{"v", bson.D{{"$lt", "b"}}}},
		},
		"Binary": {
			filter: bson.D{{"v", bson.D{{"$lt", primitive.Binary{Subtype: 0x80, Data: []byte{43}}}}}},
		},
		"BinaryNoSubtype": {
			filter: bson.D{{"v", bson.D{{"$lt", primitive.Binary{Data: []byte{43}}}}}},
		},
		"BinaryEmpty": {
			filter:     bson.D{{"v", bson.D{{"$lt", primitive.Binary{}}}}},
			resultType: emptyResult,
		},
		"ObjectID": {
			filter: bson.D{{"v", bson.D{{"$lt", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091012"))}}}},
		},
		"ObjectIDEmpty": {
			filter:     bson.D{{"v", bson.D{{"$lt", primitive.NilObjectID}}}},
			resultType: emptyResult,
		},
		"Bool": {
			filter: bson.D{{"v", bson.D{{"$lt", true}}}},
		},
		"Datetime": {
			filter: bson.D{{"v", bson.D{{"$lt", time.Date(2021, 11, 1, 10, 18, 43, 123000000, time.UTC)}}}},
		},
		"Null": {
			filter:     bson.D{{"v", bson.D{{"$lt", nil}}}},
			resultType: emptyResult,
		},
		"Regex": {
			filter:     bson.D{{"v", bson.D{{"$lt", primitive.Regex{Pattern: "foo"}}}}},
			resultType: emptyResult,
		},
		"Int32": {
			filter: bson.D{{"v", bson.D{{"$lt", int32(42)}}}},
		},
		"Int32Min": {
			filter: bson.D{{"v", bson.D{{"$lt", int32(math.MinInt32)}}}},
		},
		"Timestamp": {
			filter: bson.D{{"v", bson.D{{"$lt", primitive.Timestamp{T: 43, I: 14}}}}},
		},
		"TimestampNoI": {
			filter: bson.D{{"v", bson.D{{"$lt", primitive.Timestamp{T: 43}}}}},
		},
		"TimestampNoT": {
			filter: bson.D{{"v", bson.D{{"$lt", primitive.Timestamp{I: 14}}}}},
		},
		"Int64": {
			filter: bson.D{{"v", bson.D{{"$lt", int64(42)}}}},
		},
		"Int64Min": {
			filter: bson.D{{"v", bson.D{{"$lt", int64(math.MinInt64)}}}},
		},
		"Int64Big": {
			filter: bson.D{{"v", bson.D{{"$lt", int64(1<<61 + 1)}}}},
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryComparisonCompatLte(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Document": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}}}},
		},
		"DocumentShuffledKeys": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}}}},
		},
		"DocumentDotNotation": {
			filter: bson.D{{"v.foo", bson.D{{"$lte", int32(42)}}}},
		},
		"DocumentReverse": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.D{{"array", bson.A{int32(42), "foo", nil}}, {"42", "foo"}, {"foo", int32(42)}}}}}},
		},
		"DocumentNull": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.D{{"foo", nil}}}}}},
		},
		"DocumentEmpty": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.D{}}}}},
		},
		"ArrayEmpty": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.A{}}}}},
		},
		"ArrayOne": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.A{int32(42)}}}}},
		},
		"Array": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.A{int32(42), "foo", nil}}}}},
		},
		"ArrayReverse": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.A{nil, "foo", int32(42)}}}}},
		},
		"ArrayNull": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.A{nil}}}}},
		},
		"ArraySlice": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.A{int32(42), "foo"}}}}},
		},
		"ArrayShuffledValues": {
			filter: bson.D{{"v", bson.D{{"$lte", bson.A{"foo", nil, int32(42)}}}}},
		},
		"Double": {
			filter: bson.D{{"v", bson.D{{"$lte", 42.13}}}},
		},
		"DoubleSmallest": {
			filter: bson.D{{"v", bson.D{{"$lte", math.SmallestNonzeroFloat64}}}},
		},
		"String": {
			filter: bson.D{{"v", bson.D{{"$lte", "foo"}}}},
		},
		"StringWhole": {
			filter: bson.D{{"v", bson.D{{"$lte", "42"}}}},
		},
		"StringEmpty": {
			filter: bson.D{{"v", bson.D{{"$lte", ""}}}},
		},
		"StringAsc": {
			filter: bson.D{{"v", bson.D{{"$lte", "a"}}}},
		},
		"Binary": {
			filter: bson.D{{"v", bson.D{{"$lte", primitive.Binary{Subtype: 0x80, Data: []byte{42}}}}}},
		},
		"BinaryNoSubtype": {
			filter: bson.D{{"v", bson.D{{"$lte", primitive.Binary{Data: []byte{42}}}}}},
		},
		"BinaryEmpty": {
			filter: bson.D{{"v", bson.D{{"$lte", primitive.Binary{}}}}},
		},
		"ObjectID": {
			filter: bson.D{{"v", bson.D{{"$lte", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011"))}}}},
		},
		"ObjectIDEmpty": {
			filter: bson.D{{"v", bson.D{{"$lte", primitive.NilObjectID}}}},
		},
		"Bool": {
			filter: bson.D{{"v", bson.D{{"$lte", true}}}},
		},
		"Datetime": {
			filter: bson.D{{"v", bson.D{{"$lte", time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC)}}}},
		},
		"Null": {
			filter: bson.D{{"v", bson.D{{"$lte", nil}}}},
		},
		"Regex": {
			filter:     bson.D{{"v", bson.D{{"$lte", primitive.Regex{Pattern: "foo"}}}}},
			resultType: emptyResult,
		},
		"Int32": {
			filter: bson.D{{"v", bson.D{{"$lte", int32(42)}}}},
		},
		"Int32Min": {
			filter: bson.D{{"v", bson.D{{"$lte", int32(math.MinInt32)}}}},
		},
		"Timestamp": {
			filter: bson.D{{"v", bson.D{{"$lte", primitive.Timestamp{T: 42, I: 13}}}}},
		},
		"TimestampNoI": {
			filter: bson.D{{"v", bson.D{{"$lte", primitive.Timestamp{T: 42}}}}},
		},
		"TimestampNoT": {
			filter: bson.D{{"v", bson.D{{"$lte", primitive.Timestamp{I: 13}}}}},
		},
		"Int64": {
			filter: bson.D{{"v", bson.D{{"$lte", int64(42)}}}},
		},
		"Int64Min": {
			filter: bson.D{{"v", bson.D{{"$lte", int64(math.MinInt64)}}}},
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryComparisonCompatNin(t *testing.T) {
	t.Parallel()

	var scalarDataTypesFilter bson.A
	for _, scalarDataType := range shareddata.Scalars.Docs() {
		scalarDataTypesFilter = append(scalarDataTypesFilter, scalarDataType.Map()["v"])
	}

	var compositeDataTypesFilter bson.A
	for _, compositeDataType := range shareddata.Composites.Docs() {
		compositeDataTypesFilter = append(compositeDataTypesFilter, compositeDataType.Map()["v"])
	}

	testCases := map[string]queryCompatTestCase{
		"ForScalarDataTypes": {
			filter: bson.D{{"v", bson.D{{"$nin", scalarDataTypesFilter}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1781",
		},
		"ForCompositeDataTypes": {
			filter: bson.D{{"v", bson.D{{"$nin", compositeDataTypesFilter}}}},
		},
		"RegexString": {
			filter:     bson.D{{"v", bson.D{{"$nin", bson.A{bson.D{{"$regex", "/foo/"}}}}}}},
			resultType: emptyResult,
		},
		"Regex": {
			filter: bson.D{{"v", bson.D{{"$nin", bson.A{primitive.Regex{Pattern: "foo", Options: "i"}}}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1781",
		},
		"NilInsteadOfArray": {
			filter:     bson.D{{"v", bson.D{{"$nin", nil}}}},
			resultType: emptyResult,
		},
		"StringInsteadOfArray": {
			filter:     bson.D{{"v", bson.D{{"$nin", "foo"}}}},
			resultType: emptyResult,
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryComparisonCompatIn(t *testing.T) {
	t.Parallel()

	var scalarDataTypesFilter bson.A
	for _, scalarDataType := range shareddata.Scalars.Docs() {
		scalarDataTypesFilter = append(scalarDataTypesFilter, scalarDataType.Map()["v"])
	}

	var compositeDataTypesFilter bson.A
	for _, compositeDataType := range shareddata.Composites.Docs() {
		compositeDataTypesFilter = append(compositeDataTypesFilter, compositeDataType.Map()["v"])
	}

	testCases := map[string]queryCompatTestCase{
		"ForScalarDataTypes": {
			filter: bson.D{{"v", bson.D{{"$in", scalarDataTypesFilter}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1781",
		},
		"ForCompositeDataTypes": {
			filter: bson.D{{"v", bson.D{{"$in", compositeDataTypesFilter}}}},
		},
		"RegexString": {
			filter:     bson.D{{"v", bson.D{{"$in", bson.A{bson.D{{"$regex", "/foo/"}}}}}}},
			resultType: emptyResult,
		},
		"Regex": {
			filter: bson.D{{"v", bson.D{{"$in", bson.A{primitive.Regex{Pattern: "foo", Options: "i"}}}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1781",
		},
		"NilInsteadOfArray": {
			filter:     bson.D{{"v", bson.D{{"$in", nil}}}},
			resultType: emptyResult,
		},
		"StringInsteadOfArray": {
			filter:     bson.D{{"v", bson.D{{"$in", "foo"}}}},
			resultType: emptyResult,
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryComparisonCompatNe(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Array": {
			filter: bson.D{{"v", bson.D{{"$ne", bson.A{int32(42), "foo", nil}}}}},
		},
		"ArrayShuffledValues": {
			filter: bson.D{{"v", bson.D{{"$ne", bson.A{"foo", nil, int32(42)}}}}},
		},
		"Double": {
			filter:         bson.D{{"v", bson.D{{"$ne", 41.13}}}},
			resultPushdown: true,
		},
		"DoubleMax": {
			filter:         bson.D{{"v", bson.D{{"$ne", math.MaxFloat64}}}},
			resultPushdown: true,
		},
		"DoubleSmallest": {
			filter:         bson.D{{"v", bson.D{{"$ne", math.SmallestNonzeroFloat64}}}},
			resultPushdown: true,
		},
		"DoubleZero": {
			filter:         bson.D{{"v", bson.D{{"$ne", 0.0}}}},
			resultPushdown: true,
		},
		"DoubleBig": {
			filter:         bson.D{{"v", bson.D{{"$ne", float64(1 << 61)}}}},
			resultPushdown: true,
		},
		"DoubleBigPlus": {
			filter:         bson.D{{"v", bson.D{{"$ne", float64(1<<61) + 1}}}},
			resultPushdown: true,
		},
		"DoubleBigMinus": {
			filter:         bson.D{{"v", bson.D{{"$ne", float64(1<<61) - 1}}}},
			resultPushdown: true,
		},
		"DoubleNegBig": {
			filter:         bson.D{{"v", bson.D{{"$ne", -float64(1 << 61)}}}},
			resultPushdown: true,
		},
		"DoubleNegBigPlus": {
			filter:         bson.D{{"v", bson.D{{"$ne", -float64(1<<61) + 1}}}},
			resultPushdown: true,
		},
		"DoubleNegBigMinus": {
			filter:         bson.D{{"v", bson.D{{"$ne", -float64(1<<61) - 1}}}},
			resultPushdown: true,
		},
		"DoublePrecMax": {
			filter:         bson.D{{"v", bson.D{{"$ne", float64(1 << 53)}}}},
			resultPushdown: true,
		},
		"DoublePrecMaxPlus": {
			filter:         bson.D{{"v", bson.D{{"$ne", float64(1<<53) + 1}}}},
			resultPushdown: true,
		},
		"DoublePrecMaxMinus": {
			filter:         bson.D{{"v", bson.D{{"$ne", float64(1<<53) - 1}}}},
			resultPushdown: true,
		},
		"DoublePrecMin": {
			filter:         bson.D{{"v", bson.D{{"$ne", -float64(1<<53 - 1)}}}},
			resultPushdown: true,
		},
		"DoublePrecMinPlus": {
			filter:         bson.D{{"v", bson.D{{"$ne", -float64(1<<53-1) + 1}}}},
			resultPushdown: true,
		},
		"DoublePrecMinMinus": {
			filter:         bson.D{{"v", bson.D{{"$ne", -float64(1<<53-1) - 1}}}},
			resultPushdown: true,
		},
		"String": {
			filter:         bson.D{{"v", bson.D{{"$ne", "foo"}}}},
			resultPushdown: true,
		},
		"StringEmpty": {
			filter:         bson.D{{"v", bson.D{{"$ne", ""}}}},
			resultPushdown: true,
		},
		"Binary": {
			filter: bson.D{{"v", bson.D{{"$ne", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
		},
		"BinaryEmpty": {
			filter: bson.D{{"v", bson.D{{"$ne", primitive.Binary{Data: []byte{}}}}}},
		},
		"BoolFalse": {
			filter:         bson.D{{"v", bson.D{{"$ne", false}}}},
			resultPushdown: true,
		},
		"BoolTrue": {
			filter:         bson.D{{"v", bson.D{{"$ne", true}}}},
			resultPushdown: true,
		},
		"Datetime": {
			filter:         bson.D{{"v", bson.D{{"$ne", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))}}}},
			resultPushdown: true,
		},
		"DatetimeEpoch": {
			filter:         bson.D{{"v", bson.D{{"$ne", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
			resultPushdown: true,
		},
		"DatetimeYearMin": {
			filter:         bson.D{{"v", bson.D{{"$ne", primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))}}}},
			resultPushdown: true,
		},
		"DatetimeYearMax": {
			filter:         bson.D{{"v", bson.D{{"$ne", primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC))}}}},
			resultPushdown: true,
		},
		"Timestamp": {
			filter: bson.D{{"v", bson.D{{"$ne", primitive.Timestamp{T: 42, I: 13}}}}},
		},
		"TimestampNoI": {
			filter: bson.D{{"v", bson.D{{"$ne", primitive.Timestamp{T: 1}}}}},
		},
		"Null": {
			filter: bson.D{{"v", bson.D{{"$ne", nil}}}},
		},
		"Int32": {
			filter:         bson.D{{"v", bson.D{{"$ne", int32(42)}}}},
			resultPushdown: true,
		},
		"Int32Zero": {
			filter:         bson.D{{"v", bson.D{{"$ne", int32(0)}}}},
			resultPushdown: true,
		},
		"Int32Max": {
			filter:         bson.D{{"v", bson.D{{"$ne", int32(math.MaxInt32)}}}},
			resultPushdown: true,
		},
		"Int32Min": {
			filter:         bson.D{{"v", bson.D{{"$ne", int32(math.MinInt32)}}}},
			resultPushdown: true,
		},
		"Int64": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64(42)}}}},
			resultPushdown: true,
		},
		"Int64Zero": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64(0)}}}},
			resultPushdown: true,
		},
		"Int64Max": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64(math.MaxInt64)}}}},
			resultPushdown: true,
		},
		"Int64Min": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64(math.MinInt64)}}}},
			resultPushdown: true,
		},
		"Int64Big": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64(1 << 61)}}}},
			resultPushdown: true,
		},
		"Int64BigPlusOne": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64((1 << 61) + 1)}}}},
			resultPushdown: true,
		},
		"Int64BigMinusOne": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64((1 << 61) - 1)}}}},
			resultPushdown: true,
		},
		"Int64NegBig": {
			filter:         bson.D{{"v", bson.D{{"$ne", -int64(1 << 61)}}}},
			resultPushdown: true,
		},
		"Int64NegBigPlusOne": {
			filter:         bson.D{{"v", bson.D{{"$ne", -int64(1<<61) + 1}}}},
			resultPushdown: true,
		},
		"Int64NegBigMinusOne": {
			filter:         bson.D{{"v", bson.D{{"$ne", -int64(1<<61) - 1}}}},
			resultPushdown: true,
		},

		"Int64PrecMax": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64(1 << 53)}}}},
			resultPushdown: true,
		},
		"Int64PrecMaxPlusOne": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64((1 << 53) + 1)}}}},
			resultPushdown: true,
		},
		"Int64PrecMaxMinusOne": {
			filter:         bson.D{{"v", bson.D{{"$ne", int64((1 << 53) - 1)}}}},
			resultPushdown: true,
		},
		"Int64PrecMin": {
			filter:         bson.D{{"v", bson.D{{"$ne", -int64(1<<53 - 1)}}}},
			resultPushdown: true,
		},
		"Int64PrecMinPlus": {
			filter:         bson.D{{"v", bson.D{{"$ne", -int64(1<<53-1) + 1}}}},
			resultPushdown: true,
		},
		"Int64PrecMinMinus": {
			filter:         bson.D{{"v", bson.D{{"$ne", -int64(1<<53-1) - 1}}}},
			resultPushdown: true,
		},
		"Regex": {
			filter:     bson.D{{"v", bson.D{{"$ne", primitive.Regex{Pattern: "foo"}}}}},
			resultType: emptyResult,
		},
		"Document": {
			filter: bson.D{{"v", bson.D{{"$ne", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}}}},
		},
		"DocumentShuffledKeys": {
			filter: bson.D{{"v", bson.D{{"$ne", bson.D{{"v", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}}}}}},
		},
	}

	for k, tc := range testCases {
		tc.skipForTigris = "https://github.com/FerretDB/FerretDB/issues/2052"
		testCases[k] = tc
	}

	testQueryCompat(t, testCases)
}

func TestQueryComparisonCompatMultipleOperators(t *testing.T) {
	t.Parallel()

	var scalarDataTypesFilter bson.A
	for _, scalarDataType := range shareddata.Scalars.Docs() {
		scalarDataTypesFilter = append(scalarDataTypesFilter, scalarDataType.Map()["v"])
	}

	var compositeDataTypesFilter bson.A
	for _, compositeDataType := range shareddata.Composites.Docs() {
		compositeDataTypesFilter = append(compositeDataTypesFilter, compositeDataType.Map()["v"])
	}

	testCases := map[string]queryCompatTestCase{
		"InLteGte": {
			filter: bson.D{
				{"_id", bson.D{{"$in", bson.A{"int32"}}}},
				{"v", bson.D{{"$lte", int32(42)}, {"$gte", int32(0)}}},
			},
		},
		"NinEqNe": {
			filter: bson.D{
				{"_id", bson.D{{"$nin", bson.A{"int64"}}, {"$ne", "int32"}}},
				{"v", bson.D{{"$eq", int32(42)}}},
			},
			resultPushdown: true,
		},
		"EqNe": {
			filter: bson.D{
				{"v", bson.D{{"$eq", int32(42)}, {"$ne", int32(0)}}},
			},
			resultPushdown: true,
		},
	}

	testQueryCompat(t, testCases)
}
