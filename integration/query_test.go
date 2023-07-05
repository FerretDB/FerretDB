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
	"net/url"
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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
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
	setup.SkipForTigris(t)

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
	setup.SkipForTigris(t)

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
	setup.SkipForTigris(t)

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
	setup.SkipForTigris(t)

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

func TestQueryBatchSize(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	// The test cases call `find`, then may implicitly call `getMore` upon `cursor.Next()`.
	// The batchSize set by `find` is used also by `getMore` unless
	// `find` has default batchSize or 0 batchSize, then `getMore` has unlimited batchSize.
	// To test that, the number of documents is set to more than the double of default batchSize 101.
	arr, _ := generateDocuments(0, 220)
	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	t.Run("SetBatchSize", func(t *testing.T) {
		t.Parallel()

		cursor, err := collection.Find(ctx, bson.D{}, &options.FindOptions{BatchSize: pointer.ToInt32(2)})
		require.NoError(t, err)

		defer cursor.Close(ctx)

		require.Equal(t, 2, cursor.RemainingBatchLength(), "expected 2 documents in first batch")

		for i := 2; i > 0; i-- {
			ok := cursor.Next(ctx)
			require.True(t, ok, "expected to have next document in first batch")
			require.Equal(t, i-1, cursor.RemainingBatchLength())
		}

		// batchSize of 2 is applied to second batch which is obtained by implicit call to `getMore`
		for i := 2; i > 0; i-- {
			ok := cursor.Next(ctx)
			require.True(t, ok, "expected to have next document in second batch")
			require.Equal(t, i-1, cursor.RemainingBatchLength())
		}

		cursor.SetBatchSize(5)

		for i := 5; i > 0; i-- {
			ok := cursor.Next(ctx)
			require.True(t, ok, "expected to have next document in third batch")
			require.Equal(t, i-1, cursor.RemainingBatchLength())
		}

		// get rest of documents from the cursor to ensure cursor is exhausted
		var res bson.D
		err = cursor.All(ctx, &res)
		require.NoError(t, err)

		ok := cursor.Next(ctx)
		require.False(t, ok, "cursor exhausted, not expecting next document")
	})

	t.Run("DefaultBatchSize", func(t *testing.T) {
		t.Parallel()

		// unset batchSize uses default batchSize 101 for the first batch
		cursor, err := collection.Find(ctx, bson.D{})
		require.NoError(t, err)

		defer cursor.Close(ctx)

		require.Equal(t, 101, cursor.RemainingBatchLength())

		for i := 101; i > 0; i-- {
			ok := cursor.Next(ctx)
			require.True(t, ok, "expected to have next document")
			require.Equal(t, i-1, cursor.RemainingBatchLength())
		}

		// next batch obtain from implicit call to `getMore` has the rest of the documents, not default batchSize
		// TODO: 16MB batchSize limit https://github.com/FerretDB/FerretDB/issues/2824
		ok := cursor.Next(ctx)
		require.True(t, ok, "expected to have next document")
		require.Equal(t, 118, cursor.RemainingBatchLength())
	})

	t.Run("ZeroBatchSize", func(t *testing.T) {
		t.Parallel()

		cursor, err := collection.Find(ctx, bson.D{}, &options.FindOptions{BatchSize: pointer.ToInt32(0)})
		require.NoError(t, err)

		defer cursor.Close(ctx)

		require.Equal(t, 0, cursor.RemainingBatchLength())

		// next batch obtain from implicit call to `getMore` has the rest of the documents, not 0 batchSize
		// TODO: 16MB batchSize limit https://github.com/FerretDB/FerretDB/issues/2824
		ok := cursor.Next(ctx)
		require.True(t, ok, "expected to have next document")
		require.Equal(t, 219, cursor.RemainingBatchLength())
	})

	t.Run("NegativeLimit", func(t *testing.T) {
		t.Parallel()

		// set limit to negative, it ignores batchSize and returns single document in the firstBatch.
		cursor, err := collection.Find(ctx, bson.D{}, &options.FindOptions{
			Limit:     pointer.ToInt64(-1),
			BatchSize: pointer.ToInt32(10),
		})
		require.NoError(t, err)

		defer cursor.Close(ctx)

		require.Equal(t, 1, cursor.RemainingBatchLength(), "expected 1 document in first batch")

		ok := cursor.Next(ctx)
		require.True(t, ok, "expected to have next document")
		require.Equal(t, 0, cursor.RemainingBatchLength())

		// there is no remaining batch due to negative limit
		ok = cursor.Next(ctx)
		require.False(t, ok, "cursor exhausted, not expecting next document")
		require.Equal(t, 0, cursor.RemainingBatchLength())
	})
}

func TestQueryCommandGetMore(t *testing.T) {
	t.Parallel()

	// options are applied to create a client that uses single connection pool
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		ExtraOptions: url.Values{
			"minPoolSize":   []string{"1"},
			"maxPoolSize":   []string{"1"},
			"maxIdleTimeMS": []string{"0"},
		},
	})

	ctx, collection := s.Ctx, s.Collection

	// the number of documents is set above the default batchSize of 101
	// for testing unset batchSize returning default batchSize
	bsonArr, arr := generateDocuments(0, 110)

	_, err := collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		firstBatchSize   any // optional, nil to leave firstBatchSize unset
		getMoreBatchSize any // optional, nil to leave getMoreBatchSize unset
		collection       any // optional, nil to leave collection unset
		cursorID         any // optional, defaults to cursorID from find()

		firstBatch []*types.Document   // required, expected find firstBatch
		nextBatch  []*types.Document   // optional, expected getMore nextBatch
		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"Int": {
			firstBatchSize:   1,
			getMoreBatchSize: int32(1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:2]),
		},
		"IntNegative": {
			firstBatchSize:   1,
			getMoreBatchSize: int32(-1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"IntZero": {
			firstBatchSize:   1,
			getMoreBatchSize: int32(0),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:]),
		},
		"Long": {
			firstBatchSize:   1,
			getMoreBatchSize: int64(1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:2]),
		},
		"LongNegative": {
			firstBatchSize:   1,
			getMoreBatchSize: int64(-1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"LongZero": {
			firstBatchSize:   1,
			getMoreBatchSize: int64(0),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:]),
		},
		"Double": {
			firstBatchSize:   1,
			getMoreBatchSize: float64(1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:2]),
		},
		"DoubleNegative": {
			firstBatchSize:   1,
			getMoreBatchSize: float64(-1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"DoubleZero": {
			firstBatchSize:   1,
			getMoreBatchSize: float64(0),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:]),
		},
		"DoubleFloor": {
			firstBatchSize:   1,
			getMoreBatchSize: 1.9,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:2]),
		},
		"GetMoreCursorExhausted": {
			firstBatchSize:   200,
			getMoreBatchSize: int32(1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:110]),
			err: &mongo.CommandError{
				Code:    43,
				Name:    "CursorNotFound",
				Message: "cursor id 0 not found",
			},
		},
		"Bool": {
			firstBatchSize:   1,
			getMoreBatchSize: false,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.batchSize' is the wrong type 'bool', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'getMore.batchSize' is the wrong type 'bool', expected types '[long, int, decimal, double]'",
		},
		"Unset": {
			firstBatchSize: 1,
			// unset getMore batchSize gets all remaining documents
			getMoreBatchSize: nil,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:]),
		},
		"LargeBatchSize": {
			firstBatchSize:   1,
			getMoreBatchSize: 105,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:106]),
		},
		"StringCursorID": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       collection.Name(),
			cursorID:         "invalid",
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.getMore' is the wrong type 'string', expected type 'long'",
			},
			altMessage: "BSON field 'getMore.getMore' is the wrong type, expected type 'long'",
		},
		"Int32CursorID": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       collection.Name(),
			cursorID:         int32(1111),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.getMore' is the wrong type 'int', expected type 'long'",
			},
			altMessage: "BSON field 'getMore.getMore' is the wrong type, expected type 'long'",
		},
		"NotFoundCursorID": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       collection.Name(),
			cursorID:         int64(1234),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    43,
				Name:    "CursorNotFound",
				Message: "cursor id 1234 not found",
			},
		},
		"WrongTypeNamespace": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       bson.D{},
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.collection' is the wrong type 'object', expected type 'string'",
			},
		},
		"InvalidNamespace": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       "invalid",
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code: 13,
				Name: "Unauthorized",
				Message: "Requested getMore on namespace 'TestQueryCommandGetMore.invalid'," +
					" but cursor belongs to a different namespace TestQueryCommandGetMore.TestQueryCommandGetMore",
			},
		},
		"EmptyCollectionName": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       "",
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Collection names cannot be empty",
			},
		},
		"MissingCollectionName": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       nil,
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'getMore.collection' is missing but a required field",
			},
		},
		"UnsetAllBatchSize": {
			firstBatchSize:   nil,
			getMoreBatchSize: nil,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:101]),
			nextBatch:        ConvertDocuments(t, arr[101:]),
		},
		"UnsetFindBatchSize": {
			firstBatchSize:   nil,
			getMoreBatchSize: 5,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:101]),
			nextBatch:        ConvertDocuments(t, arr[101:106]),
		},
		"UnsetGetMoreBatchSize": {
			firstBatchSize:   5,
			getMoreBatchSize: nil,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:5]),
			nextBatch:        ConvertDocuments(t, arr[5:]),
		},
		"BatchSize": {
			firstBatchSize:   3,
			getMoreBatchSize: 5,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:3]),
			nextBatch:        ConvertDocuments(t, arr[3:8]),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			// Do not run subtests in t.Parallel() to eliminate the occurrence
			// of session error.
			// Supporting session would help us understand fix it
			// https://github.com/FerretDB/FerretDB/issues/153.
			//
			// > Location50738
			// > Cannot run getMore on cursor 2053655655200551971,
			// > which was created in session 2926eea5-9775-41a3-a563-096969f1c7d5 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  - ,
			// > in session 774d9ac6-b24a-4fd8-9874-f92ab1c9c8f5 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  -

			require.NotNil(t, tc.firstBatch, "firstBatch must not be nil")

			var findRest bson.D
			aggregateCursor := bson.D{}

			if tc.firstBatchSize != nil {
				findRest = append(findRest, bson.E{Key: "batchSize", Value: tc.firstBatchSize})
				aggregateCursor = bson.D{{"batchSize", tc.firstBatchSize}}
			}

			aggregateCommand := bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", aggregateCursor},
			}

			findCommand := append(
				bson.D{{"find", collection.Name()}},
				findRest...,
			)

			for _, command := range []bson.D{findCommand, aggregateCommand} {
				var res bson.D
				err := collection.Database().RunCommand(ctx, command).Decode(&res)
				require.NoError(t, err)

				doc := ConvertDocument(t, res)

				v, _ := doc.Get("cursor")
				require.NotNil(t, v)

				cursor, ok := v.(*types.Document)
				require.True(t, ok)

				cursorID, _ := cursor.Get("id")
				assert.NotNil(t, cursorID)

				v, _ = cursor.Get("firstBatch")
				require.NotNil(t, v)

				firstBatch, ok := v.(*types.Array)
				require.True(t, ok)

				require.Equal(t, len(tc.firstBatch), firstBatch.Len(), "expected: %v, got: %v", tc.firstBatch, firstBatch)
				for i, elem := range tc.firstBatch {
					require.Equal(t, elem, must.NotFail(firstBatch.Get(i)))
				}

				if tc.cursorID != nil {
					cursorID = tc.cursorID
				}

				var getMoreRest bson.D
				if tc.getMoreBatchSize != nil {
					getMoreRest = append(getMoreRest, bson.E{Key: "batchSize", Value: tc.getMoreBatchSize})
				}

				if tc.collection != nil {
					getMoreRest = append(getMoreRest, bson.E{Key: "collection", Value: tc.collection})
				}

				getMoreCommand := append(
					bson.D{
						{"getMore", cursorID},
					},
					getMoreRest...,
				)

				err = collection.Database().RunCommand(ctx, getMoreCommand).Decode(&res)
				if tc.err != nil {
					AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

					// upon error response contains firstBatch field.
					doc = ConvertDocument(t, res)

					v, _ = doc.Get("cursor")
					require.NotNil(t, v)

					cursor, ok = v.(*types.Document)
					require.True(t, ok)

					cursorID, _ = cursor.Get("id")
					assert.NotNil(t, cursorID)

					v, _ = cursor.Get("firstBatch")
					require.NotNil(t, v)

					firstBatch, ok = v.(*types.Array)
					require.True(t, ok)

					require.Equal(t, len(tc.firstBatch), firstBatch.Len(), "expected: %v, got: %v", tc.firstBatch, firstBatch)
					for i, elem := range tc.firstBatch {
						require.Equal(t, elem, must.NotFail(firstBatch.Get(i)))
					}

					return
				}

				require.NoError(t, err)

				doc = ConvertDocument(t, res)

				v, _ = doc.Get("cursor")
				require.NotNil(t, v)

				cursor, ok = v.(*types.Document)
				require.True(t, ok)

				cursorID, _ = cursor.Get("id")
				assert.NotNil(t, cursorID)

				v, _ = cursor.Get("nextBatch")
				require.NotNil(t, v)

				nextBatch, ok := v.(*types.Array)
				require.True(t, ok)

				require.Equal(t, len(tc.nextBatch), nextBatch.Len(), "expected: %v, got: %v", tc.nextBatch, nextBatch)
				for i, elem := range tc.nextBatch {
					require.Equal(t, elem, must.NotFail(nextBatch.Get(i)))
				}
			}
		})
	}
}

func TestQueryCommandGetMoreConnection(t *testing.T) {
	t.Parallel()

	// options are applied to create a client that uses single connection pool
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		ExtraOptions: url.Values{
			"minPoolSize":   []string{"1"},
			"maxPoolSize":   []string{"1"},
			"maxIdleTimeMS": []string{"0"},
		},
	})

	ctx := s.Ctx
	collection1 := s.Collection
	databaseName := s.Collection.Database().Name()
	collectionName := s.Collection.Name()

	arr, _ := generateDocuments(0, 5)
	_, err := collection1.InsertMany(ctx, arr)
	require.NoError(t, err)

	t.Run("SameClient", func(t *testing.T) {
		// Do not run subtests in t.Parallel() to eliminate the occurrence
		// of session error.
		// Supporting session would help us understand fix it
		// https://github.com/FerretDB/FerretDB/issues/153.
		//
		// > Location50738
		// > Cannot run getMore on cursor 2053655655200551971,
		// > which was created in session 2926eea5-9775-41a3-a563-096969f1c7d5 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  - ,
		// > in session 774d9ac6-b24a-4fd8-9874-f92ab1c9c8f5 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  -

		var res bson.D
		err = collection1.Database().RunCommand(ctx, bson.D{
			{"find", collection1.Name()},
			{"batchSize", 2},
		}).Decode(&res)
		require.NoError(t, err)

		doc := ConvertDocument(t, res)

		v, _ := doc.Get("cursor")
		require.NotNil(t, v)

		cursor, ok := v.(*types.Document)
		require.True(t, ok)

		cursorID, _ := cursor.Get("id")
		assert.NotNil(t, cursorID)

		err = collection1.Database().RunCommand(ctx, bson.D{
			{"getMore", cursorID},
			{"collection", collection1.Name()},
		}).Decode(&res)
		require.NoError(t, err)
	})

	t.Run("DifferentClient", func(tt *testing.T) {
		// The error returned from MongoDB is a session error, FerretDB does not
		// return an error because db, collection and username are the same.
		t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/153")

		// do not run subtest in parallel to avoid breaking another parallel subtest

		u, err := url.Parse(s.MongoDBURI)
		require.NoError(t, err)

		client2, err := mongo.Connect(ctx, options.Client().ApplyURI(u.String()))
		require.NoError(t, err)

		defer client2.Disconnect(ctx)

		collection2 := client2.Database(databaseName).Collection(collectionName)

		var res bson.D
		err = collection1.Database().RunCommand(ctx, bson.D{
			{"find", collection1.Name()},
			{"batchSize", 2},
		}).Decode(&res)
		require.NoError(t, err)

		doc := ConvertDocument(t, res)

		v, _ := doc.Get("cursor")
		require.NotNil(t, v)

		cursor, ok := v.(*types.Document)
		require.True(t, ok)

		cursorID, _ := cursor.Get("id")
		assert.NotNil(t, cursorID)

		err = collection2.Database().RunCommand(ctx, bson.D{
			{"getMore", cursorID},
			{"collection", collection2.Name()},
		}).Decode(&res)

		// use AssertMatchesCommandError because message cannot be compared as it contains session ID
		AssertMatchesCommandError(
			t,
			mongo.CommandError{
				Code: 50738,
				Name: "Location50738",
				Message: "Cannot run getMore on cursor 5720627396082469624, which was created in session " +
					"95326129-ff9c-48a4-9060-464b4ea3ee06 - 47DEQpj8HBSa+/TImW+5JC\neuQeRkm5NMpJWZG3hSuFU= -  - , " +
					"in session 9e8902e9-338c-4156-9fd8-50e5d62ac992 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  - ",
			},
			err,
		)
	})
}
