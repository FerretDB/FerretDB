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
	"github.com/FerretDB/FerretDB/internal/util/must"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestQueryComparisonCompatImplicit(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Document": {
			filter: bson.D{{"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
		},
		"DocumentReverse": {
			filter: bson.D{{"v", bson.D{{"array", bson.A{int32(42), "foo", nil}}, {"42", "foo"}, {"foo", int32(42)}}}},
		},
		"DocumentNull": {
			filter: bson.D{{"v", bson.D{{"foo", nil}}}},
		},
		"DocumentEmpty": {
			filter: bson.D{{"v", bson.D{}}},
		},
		"DocumentShuffledKeys": {
			filter:     bson.D{{"v", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}},
			resultType: emptyResult,
		},
		"DocumentDotNotation": {
			filter: bson.D{{"v.foo", int32(42)}},
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
		"Double": {
			filter: bson.D{{"v", 42.13}},
		},
		"DoubleMax": {
			filter: bson.D{{"v", math.MaxFloat64}},
		},
		"DoubleSmallest": {
			filter: bson.D{{"v", math.SmallestNonzeroFloat64}},
		},
		"DoubleBig": {
			filter: bson.D{{"v", float64(2 << 60)}},
		},
		"String": {
			filter: bson.D{{"v", "foo"}},
		},
		"StringEmpty": {
			filter: bson.D{{"v", ""}},
		},
		"Binary": {
			filter: bson.D{{"v", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}},
		},
		"BinaryEmpty": {
			filter: bson.D{{"v", primitive.Binary{}}},
		},
		"BoolFalse": {
			filter: bson.D{{"v", false}},
		},
		"BoolTrue": {
			filter: bson.D{{"v", true}},
		},
		"IDNull": {
			filter:     bson.D{{"_id", nil}},
			resultType: emptyResult,
		},
		"ValueNull": {
			filter: bson.D{{"v", nil}},
		},
		"NoSuchFieldNull": {
			filter: bson.D{{"no-such-field", nil}},
		},
		"ValueNumber": {
			filter: bson.D{{"v", 42}},
		},
		"ValueRegex": {
			filter: bson.D{{"v", primitive.Regex{Pattern: "^fo"}}},
		},
	}

	skipForTigris := "https://github.com/FerretDB/FerretDB/issues/908"
	testQueryCompat(t, skipForTigris, testCases)
}

func TestQueryComparisonCompatEq(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Document": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}}}},
		},
		"DocumentShuffledKeys": {
			filter:     bson.D{{"v", bson.D{{"$eq", bson.D{{"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}, {"foo", int32(42)}}}}}},
			resultType: emptyResult,
		},
		"DocumentDotNotation": {
			filter: bson.D{{"v.foo", bson.D{{"$eq", int32(42)}}}},
		},
		"DocumentReverse": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.D{{"array", bson.A{int32(42), "foo", nil}}, {"42", "foo"}, {"foo", int32(42)}}}}}},
		},
		"DocumentNull": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.D{{"foo", nil}}}}}},
		},
		"DocumentEmpty": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.D{}}}}},
		},
		"Array": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.A{int32(42), "foo", nil}}}}},
		},
		"ArrayShuffledValues": {
			filter:     bson.D{{"v", bson.D{{"$eq", bson.A{"foo", nil, int32(42)}}}}},
			resultType: emptyResult,
		},
		"ArrayReverse": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.A{nil, "foo", int32(42)}}}}},
		},
		"ArrayNull": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.A{nil}}}}},
		},
		"ArrayEmpty": {
			filter: bson.D{{"v", bson.D{{"$eq", bson.A{}}}}},
		},
		"Double": {
			filter: bson.D{{"v", bson.D{{"$eq", 42.13}}}},
		},
		"DoubleWhole": {
			filter: bson.D{{"v", bson.D{{"$eq", 42.0}}}},
		},
		"DoubleZero": {
			filter: bson.D{{"v", bson.D{{"$eq", 0.0}}}},
		},
		"DoubleMax": {
			filter: bson.D{{"v", bson.D{{"$eq", math.MaxFloat64}}}},
		},
		"DoubleSmallest": {
			filter: bson.D{{"v", bson.D{{"$eq", math.SmallestNonzeroFloat64}}}},
		},
		"DoubleNaN": {
			filter: bson.D{{"v", bson.D{{"$eq", math.NaN()}}}},
		},
		"DoubleBigInt64": {
			filter: bson.D{{"v", bson.D{{"$eq", float64(2 << 61)}}}},
		},
		"String": {
			filter: bson.D{{"v", bson.D{{"$eq", "foo"}}}},
		},
		"StringDouble": {
			filter: bson.D{{"v", bson.D{{"$eq", "42.13"}}}},
		},
		"StringWhole": {
			filter: bson.D{{"v", bson.D{{"$eq", "42"}}}},
		},
		"StringEmpty": {
			filter: bson.D{{"v", bson.D{{"$eq", ""}}}},
		},
		"Binary": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
		},
		"BinaryEmpty": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Binary{Data: []byte{}}}}}},
		},
		"ObjectID": {
			filter: bson.D{{"v", bson.D{{"$eq", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011"))}}}},
		},
		"ObjectIDEmpty": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.NilObjectID}}}},
		},
		"BoolFalse": {
			filter: bson.D{{"v", bson.D{{"$eq", false}}}},
		},
		"BoolTrue": {
			filter: bson.D{{"v", bson.D{{"$eq", true}}}},
		},
		"Datetime": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))}}}},
		},
		"DatetimeEpoch": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
		},
		"DatetimeYearMax": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))}}}},
		},
		"DatetimeYearMin": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC))}}}},
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
			filter: bson.D{{"v", bson.D{{"$eq", int32(42)}}}},
		},
		"Int32Zero": {
			filter: bson.D{{"v", bson.D{{"$eq", int32(0)}}}},
		},
		"Int32Max": {
			filter: bson.D{{"v", bson.D{{"$eq", int32(math.MaxInt32)}}}},
		},
		"Int32Min": {
			filter: bson.D{{"v", bson.D{{"$eq", int32(math.MinInt32)}}}},
		},
		"Timestamp": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Timestamp{T: 42, I: 13}}}}},
		},
		"TimestampI": {
			filter: bson.D{{"v", bson.D{{"$eq", primitive.Timestamp{I: 1}}}}},
		},
		"Int64": {
			filter: bson.D{{"v", bson.D{{"$eq", int64(42)}}}},
		},
		"Int64Zero": {
			filter: bson.D{{"v", bson.D{{"$eq", int64(0)}}}},
		},
		"Int64Max": {
			filter: bson.D{{"v", bson.D{{"$eq", int64(math.MaxInt64)}}}},
		},
		"Int64Min": {
			filter: bson.D{{"v", bson.D{{"$eq", int64(math.MinInt64)}}}},
		},
		"Int64DoubleBig": {
			filter: bson.D{{"v", bson.D{{"$eq", int64(2 << 60)}}}},
		},
		"NoSuchFieldNull": {
			filter: bson.D{{"no-such-field", bson.D{{"$eq", nil}}}},
		},
	}

	skipForTigris := "https://github.com/FerretDB/FerretDB/issues/908"
	testQueryCompat(t, skipForTigris, testCases)
}

func TestQueryComparisonCompatGt(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
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
			filter:     bson.D{{"v", bson.D{{"$gt", bson.A{"foo", nil, int32(42)}}}}},
			resultType: emptyResult,
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
			filter: bson.D{{"v", bson.D{{"$gt", int64(2<<60 - 1)}}}},
		},
	}

	skipForTigris := "https://github.com/FerretDB/FerretDB/issues/908"
	testQueryCompat(t, skipForTigris, testCases)
}
