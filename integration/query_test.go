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

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryBadFindType(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, nil)

	ctx, collection := s.Ctx, s.Collection

	for name, tc := range map[string]struct {
		value any // optional, used for find value

		err        *mongo.CommandError // required, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"Document": {
			value: bson.D{},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type object",
			},
			altMessage: "collection name has invalid type object",
		},
		"Array": {
			value: primitive.A{},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type array",
			},
		},
		"Double": {
			value: 3.14,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type double",
			},
		},
		"Binary": {
			value: primitive.Binary{},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type binData",
			},
		},
		"ObjectID": {
			value: primitive.ObjectID{},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type objectId",
			},
		},
		"Bool": {
			value: true,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type bool",
			},
		},
		"Date": {
			value: time.Now(),
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type date",
			},
		},
		"Null": {
			value: nil,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type null",
			},
		},
		"Regex": {
			value: primitive.Regex{Pattern: "/foo/"},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type regex",
			},
		},
		"Int": {
			value: int32(42),
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type int",
			},
		},
		"Timestamp": {
			value: primitive.Timestamp{},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type timestamp",
			},
		},
		"Long": {
			value: int64(42),
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type long",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.err, "err must not be nil")

			cmd := bson.D{
				{"find", tc.value},
				{"projection", bson.D{{"v", "some"}}},
			}

			var res bson.D
			err := collection.Database().RunCommand(ctx, cmd).Decode(&res)

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestQuerySortErrors(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command bson.D // required, command to run

		err        *mongo.CommandError // required
		altMessage string              // optional, alternative error message
		skip       string              // optional, skip test with a specified reason
	}{
		"SortTypeDouble": {
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
			altMessage: "BSON field 'find.sort' is the wrong type 'double', expected type 'object'",
		},
		"SortTypeString": {
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
			altMessage: "BSON field 'find.sort' is the wrong type 'string', expected type 'object'",
		},
		"SortStringValue": {
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
		"DoubleValue": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", 42}}},
				{"sort", bson.D{{"asc", 42.12}}},
			},
			err: &mongo.CommandError{
				Code:    15975,
				Name:    "Location15975",
				Message: `$sort key ordering must be 1 (for ascending) or -1 (for descending)`,
			},
		},
		"IncorrectIntValue": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", 42}}},
				{"sort", bson.D{{"asc", int32(12)}}},
			},
			err: &mongo.CommandError{
				Code:    15975,
				Name:    "Location15975",
				Message: `$sort key ordering must be 1 (for ascending) or -1 (for descending)`,
			},
		},
		"ExceedIntValue": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"v", 42}}},
				{"sort", bson.D{{"asc", int64(math.MaxInt64)}}},
			},
			err: &mongo.CommandError{
				Code:    15975,
				Name:    "Location15975",
				Message: `$sort key ordering must be 1 (for ascending) or -1 (for descending)`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.command, "command must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			var res bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&res)

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestQueryMaxTimeMSErrors(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		command bson.D // required, command to run

		err        *mongo.CommandError // required, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
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
			altMessage: "maxTimeMS has non-integral value",
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
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.command, "command must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			var res bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&res)

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestQueryMaxTimeMSAvailableValues(t *testing.T) {
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
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

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
			bson.D{{"_id", bson.D{{"foo", "bar"}}}},
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
		"FieldID": {
			filter:      bson.D{{"_id.foo", "bar"}},
			expectedIDs: []any{bson.D{{"foo", "bar"}}},
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

func TestQueryCommandBatchSize(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	// the number of documents is set above the default batchSize of 101
	// for testing unset batchSize returning default batchSize
	arr, _ := generateDocuments(0, 110)
	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		filter    any // optional, nil to leave filter unset
		batchSize any // optional, nil to leave batchSize unset

		firstBatch primitive.A         // optional, expected firstBatch
		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"Int": {
			batchSize:  1,
			firstBatch: arr[:1],
		},
		"Long": {
			batchSize:  int64(2),
			firstBatch: arr[:2],
		},
		"LongZero": {
			batchSize:  int64(0),
			firstBatch: bson.A{},
		},
		"LongNegative": {
			batchSize: int64(-1),
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
			altMessage: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
		},
		"DoubleZero": {
			batchSize:  float64(0),
			firstBatch: bson.A{},
		},
		"DoubleNegative": {
			batchSize: -1.1,
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"DoubleFloor": {
			batchSize:  1.9,
			firstBatch: arr[:1],
		},
		"Bool": {
			batchSize:  true,
			firstBatch: arr[:1],
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'FindCommandRequest.batchSize' is the wrong type 'bool', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'find.batchSize' is the wrong type 'bool', expected types '[long, int, decimal, double]'",
		},
		"Unset": {
			// default batchSize is 101 when unset
			batchSize:  nil,
			firstBatch: arr[:101],
		},
		"LargeBatchSize": {
			batchSize:  102,
			firstBatch: arr[:102],
		},
		"LargeBatchSizeFilter": {
			filter:     bson.D{{"_id", bson.D{{"$in", bson.A{0, 1, 2, 3, 4, 5}}}}},
			batchSize:  102,
			firstBatch: arr[:6],
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			var rest bson.D
			if tc.filter != nil {
				rest = append(rest, bson.E{Key: "filter", Value: tc.filter})
			}

			if tc.batchSize != nil {
				rest = append(rest, bson.E{Key: "batchSize", Value: tc.batchSize})
			}

			command := append(
				bson.D{{"find", collection.Name()}},
				rest...,
			)

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)
			if tc.err != nil {
				assert.Nil(t, res)
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

				return
			}

			require.NoError(t, err)

			v, ok := res.Map()["cursor"]
			require.True(t, ok)

			cursor, ok := v.(bson.D)
			require.True(t, ok)

			// Do not check the value of cursor id, FerretDB has a different id.
			cursorID := cursor.Map()["id"]
			assert.NotNil(t, cursorID)

			firstBatch, ok := cursor.Map()["firstBatch"]
			require.True(t, ok)
			require.Equal(t, tc.firstBatch, firstBatch)
		})
	}
}

func TestQueryCommandSingleBatch(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	arr, _ := generateDocuments(0, 5)
	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		batchSize   any // optional, nil to leave batchSize unset
		singleBatch any // optional, nil to leave singleBatch unset

		cursorClosed bool                // optional, set true for expecting cursor to be closed
		err          *mongo.CommandError // optional, expected error from MongoDB
		altMessage   string              // optional, alternative error message for FerretDB, ignored if empty
		skip         string              // optional, skip test with a specified reason
	}{
		"True": {
			singleBatch:  true,
			batchSize:    3,
			cursorClosed: true,
		},
		"False": {
			singleBatch:  false,
			batchSize:    3,
			cursorClosed: false,
		},
		"Int": {
			singleBatch: int32(1),
			batchSize:   3,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Field 'singleBatch' should be a boolean value, but found: int",
			},
			altMessage: "BSON field 'find.singleBatch' is the wrong type 'int', expected type 'bool'",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			var rest bson.D
			if tc.batchSize != nil {
				rest = append(rest, bson.E{Key: "batchSize", Value: tc.batchSize})
			}

			if tc.singleBatch != nil {
				rest = append(rest, bson.E{Key: "singleBatch", Value: tc.singleBatch})
			}

			command := append(
				bson.D{{"find", collection.Name()}},
				rest...,
			)

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)
			if tc.err != nil {
				assert.Nil(t, res)
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

				return
			}

			require.NoError(t, err)

			v, ok := res.Map()["cursor"]
			require.True(t, ok)

			cursor, ok := v.(bson.D)
			require.True(t, ok)

			cursorID := cursor.Map()["id"]
			assert.NotNil(t, cursorID)

			if !tc.cursorClosed {
				assert.NotZero(t, cursorID)
				return
			}

			assert.Equal(t, int64(0), cursorID)
		})
	}
}

func TestQueryCommandLimitPushDown(t *testing.T) {
	t.Parallel()

	// must use a collection of documents which does not support query pushdown to test limit pushdown
	s := setup.SetupWithOpts(t, &setup.SetupOpts{Providers: []shareddata.Provider{shareddata.Composites}})
	ctx, collection := s.Ctx, s.Collection

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		filter  bson.D // optional, defaults to bson.D{}
		limit   int64  // optional, defaults to zero which is unlimited
		sort    bson.D // optional, nil to leave sort unset
		optSkip *int64 // optional, nil to leave optSkip unset

		len           int                 // expected length of results
		queryPushdown bool                // optional, set true for expected pushdown for query
		limitPushdown bool                // optional, set true for expected pushdown for limit
		err           *mongo.CommandError // optional, expected error from MongoDB
		altMessage    string              // optional, alternative error message for FerretDB, ignored if empty
		skip          string              // optional, skip test with a specified reason
	}{
		"Simple": {
			limit:         1,
			len:           1,
			limitPushdown: true,
		},
		"AlmostAll": {
			limit:         int64(len(shareddata.Composites.Docs()) - 1),
			len:           len(shareddata.Composites.Docs()) - 1,
			limitPushdown: true,
		},
		"All": {
			limit:         int64(len(shareddata.Composites.Docs())),
			len:           len(shareddata.Composites.Docs()),
			limitPushdown: true,
		},
		"More": {
			limit:         int64(len(shareddata.Composites.Docs()) + 1),
			len:           len(shareddata.Composites.Docs()),
			limitPushdown: true,
		},
		"Big": {
			limit:         1000,
			len:           len(shareddata.Composites.Docs()),
			limitPushdown: true,
		},
		"Zero": {
			limit:         0,
			len:           len(shareddata.Composites.Docs()),
			limitPushdown: false,
		},
		"IDFilter": {
			filter:        bson.D{{"_id", "array"}},
			limit:         3,
			len:           1,
			queryPushdown: true,
			limitPushdown: false,
		},
		"ValueFilter": {
			filter:        bson.D{{"v", 42}},
			sort:          bson.D{{"_id", 1}},
			limit:         3,
			len:           3,
			queryPushdown: true,
			limitPushdown: false,
		},
		"DotNotationFilter": {
			filter:        bson.D{{"v.foo", 42}},
			limit:         3,
			len:           3,
			queryPushdown: false,
			limitPushdown: false,
		},
		"ObjectFilter": {
			filter:        bson.D{{"v", bson.D{{"foo", nil}}}},
			limit:         3,
			len:           1,
			queryPushdown: false,
			limitPushdown: false,
		},
		"Sort": {
			sort:          bson.D{{"_id", 1}},
			limit:         2,
			len:           2,
			queryPushdown: false,
			limitPushdown: true,
		},
		"ValueFilterSort": {
			filter:        bson.D{{"v", 42}},
			sort:          bson.D{{"_id", 1}},
			limit:         3,
			len:           3,
			queryPushdown: true,
			limitPushdown: false,
		},
		"DotNotationFilterSort": {
			filter:        bson.D{{"v.foo", 42}},
			sort:          bson.D{{"_id", 1}},
			limit:         3,
			len:           3,
			queryPushdown: false,
			limitPushdown: false,
		},
		"ObjectFilterSort": {
			filter:        bson.D{{"v", bson.D{{"foo", nil}}}},
			sort:          bson.D{{"_id", 1}},
			limit:         3,
			len:           1,
			queryPushdown: false,
			limitPushdown: false,
		},
		"Skip": {
			optSkip:       pointer.ToInt64(1),
			limit:         2,
			len:           2,
			limitPushdown: false,
		},
	} {
		tc, name := tc, name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var rest bson.D
			if tc.sort != nil {
				rest = append(rest, bson.E{Key: "sort", Value: tc.sort})
			}

			if tc.optSkip != nil {
				rest = append(rest, bson.E{Key: "skip", Value: tc.optSkip})
			}

			filter := tc.filter
			if filter == nil {
				filter = bson.D{}
			}

			query := append(
				bson.D{
					{"find", collection.Name()},
					{"filter", filter},
					{"limit", tc.limit},
				},
				rest...,
			)

			t.Run("Explain", func(t *testing.T) {
				setup.SkipForMongoDB(t, "pushdown is FerretDB specific feature")

				var res bson.D
				err := collection.Database().RunCommand(ctx, bson.D{{"explain", query}}).Decode(&res)
				if tc.err != nil {
					assert.Nil(t, res)
					AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

					return
				}

				assert.NoError(t, err)

				var msg string
				if !setup.IsSortPushdownEnabled() && tc.sort != nil {
					tc.limitPushdown = false
					msg = "Sort pushdown is disabled, but target resulted with limitPushdown"
				}

				if setup.IsPushdownDisabled() {
					tc.queryPushdown = false
					msg = "Query pushdown is disabled, but target resulted with pushdown"
				}

				doc := ConvertDocument(t, res)
				limitPushdown, _ := doc.Get("limitPushdown")
				assert.Equal(t, tc.limitPushdown, limitPushdown, msg)

				queryPushdown, _ := ConvertDocument(t, res).Get("pushdown")
				assert.Equal(t, tc.queryPushdown, queryPushdown, msg)
			})

			t.Run("Find", func(t *testing.T) {
				cursor, err := collection.Database().RunCommandCursor(ctx, query)
				if tc.err != nil {
					AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

					return
				}

				defer cursor.Close(ctx)

				require.NoError(t, err)

				docs := FetchAll(t, ctx, cursor)

				// do not check the content, limit without sort returns randomly ordered documents
				require.Len(t, docs, tc.len)
			})
		})
	}
}
