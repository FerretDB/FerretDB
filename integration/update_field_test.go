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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestUpdateFieldCurrentDate(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	t.Run("readBack", func(t *testing.T) {
		maxDifference := time.Duration(10 * time.Second)
		nowTimestamp := primitive.Timestamp{T: uint32(time.Now().Unix()), I: uint32(0)}
		id := "string-empty"

		stat := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 1,
			UpsertedCount: 0,
		}
		path := types.NewPathFromString("v")
		result := bson.D{{"_id", id}, {"v", nowTimestamp}}

		ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

		// store the current timestamp with $currentDate operator;
		update := bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "timestamp"}}}}}}
		res, err := collection.UpdateOne(ctx, bson.D{{"_id", id}}, update)
		require.NoError(t, err)
		require.Equal(t, stat, res)

		// read it, check that it is close to the current time;
		var actualBSON bson.D
		err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&actualBSON)
		require.NoError(t, err)

		expected := ConvertDocument(t, result)
		actualDocument := ConvertDocument(t, actualBSON)

		testutil.CompareAndSetByPathTime(t, expected, actualDocument, maxDifference, path)

		// write a new timestamp value with the same time;
		updateBSON := bson.D{{"$set", bson.D{{"v", nowTimestamp}}}}
		expectedBSON := bson.D{{"_id", id}, {"v", nowTimestamp}}
		res, err = collection.UpdateOne(ctx, bson.D{{"_id", id}}, updateBSON)
		require.NoError(t, err)
		require.Equal(t, stat, res)

		// read it back, and check that it is still close to the current time.
		err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&actualBSON)
		require.NoError(t, err)

		AssertEqualDocuments(t, expectedBSON, actualBSON)
		actualY := ConvertDocument(t, actualBSON)
		testutil.CompareAndSetByPathTime(t, actualY, actualDocument, maxDifference, path)
	})

	t.Run("currentDate", func(t *testing.T) {
		// maxDifference is a maximum amount of seconds can differ the value in placeholder from actual value
		maxDifference := time.Duration(3 * time.Minute)

		now := primitive.NewDateTimeFromTime(time.Now().UTC())
		nowTimestamp := primitive.Timestamp{T: uint32(time.Now().UTC().Unix()), I: uint32(0)}

		for name, tc := range map[string]struct {
			id       string
			update   bson.D
			expected bson.D
			stat     *mongo.UpdateResult
			paths    []types.Path
			err      *mongo.WriteError
			alt      string
		}{
			"DocumentEmpty": {
				id:       "double",
				update:   bson.D{{"$currentDate", bson.D{}}},
				expected: bson.D{{"_id", "double"}, {"v", float64(42.13)}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 0,
					UpsertedCount: 0,
				},
			},
			"ArrayEmpty": {
				id:     "double",
				update: bson.D{{"$currentDate", bson.A{}}},
				err: &mongo.WriteError{
					Code: 9,
					Message: "Modifiers operate on fields but we found type array instead. " +
						"For example: {$mod: {<field>: ...}} not {$currentDate: []}",
				},
				alt: "Modifiers operate on fields but we found another type instead",
			},
			"Int32Wrong": {
				id:     "double",
				update: bson.D{{"$currentDate", int32(1)}},
				err: &mongo.WriteError{
					Code: 9,
					Message: "Modifiers operate on fields but we found type int instead. " +
						"For example: {$mod: {<field>: ...}} not {$currentDate: 1}",
				},
				alt: "Modifiers operate on fields but we found another type instead",
			},
			"Nil": {
				id:     "double",
				update: bson.D{{"$currentDate", nil}},
				err: &mongo.WriteError{
					Code: 9,
					Message: "Modifiers operate on fields but we found type null instead. " +
						"For example: {$mod: {<field>: ...}} not {$currentDate: null}",
				},
				alt: "Modifiers operate on fields but we found another type instead",
			},
			"BoolTrue": {
				id:       "double",
				update:   bson.D{{"$currentDate", bson.D{{"v", true}}}},
				expected: bson.D{{"_id", "double"}, {"v", now}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
				paths: []types.Path{types.NewPathFromString("v")},
			},
			"BoolTwoTrue": {
				id:       "double",
				update:   bson.D{{"$currentDate", bson.D{{"v", true}, {"unexistent", true}}}},
				expected: bson.D{{"_id", "double"}, {"v", now}, {"unexistent", now}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
				paths: []types.Path{
					types.NewPathFromString("v"),
					types.NewPathFromString("unexistent"),
				},
			},
			"BoolFalse": {
				id:       "double",
				update:   bson.D{{"$currentDate", bson.D{{"v", false}}}},
				expected: bson.D{{"_id", "double"}, {"v", now}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
				paths: []types.Path{types.NewPathFromString("v")},
			},
			"Int32": {
				id:     "double",
				update: bson.D{{"$currentDate", bson.D{{"v", int32(1)}}}},
				err: &mongo.WriteError{
					Code:    2,
					Message: "int is not valid type for $currentDate. Please use a boolean ('true') or a $type expression ({$type: 'timestamp/date'}).",
				},
			},
			"Timestamp": {
				id:       "double",
				update:   bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "timestamp"}}}}}},
				expected: bson.D{{"_id", "double"}, {"v", nowTimestamp}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
				paths: []types.Path{types.NewPathFromString("v")},
			},
			"TimestampCapitalised": {
				id:     "double",
				update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "Timestamp"}}}}}},
				err: &mongo.WriteError{
					Code:    2,
					Message: "The '$type' string field is required to be 'date' or 'timestamp': {$currentDate: {field : {$type: 'date'}}}",
				},
				alt: "The '$type' string field is required to be 'date' or 'timestamp'",
			},
			"Date": {
				id:       "double",
				update:   bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "date"}}}}}},
				expected: bson.D{{"_id", "double"}, {"v", now}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
				paths: []types.Path{types.NewPathFromString("v")},
			},
			"WrongType": {
				id:     "double",
				update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", bson.D{{"abcd", int32(1)}}}}}}}},
				err: &mongo.WriteError{
					Code:    2,
					Message: "The '$type' string field is required to be 'date' or 'timestamp': {$currentDate: {field : {$type: 'date'}}}",
				},
				alt: "The '$type' string field is required to be 'date' or 'timestamp'",
			},
			"NoField": {
				id:       "double",
				update:   bson.D{{"$currentDate", bson.D{{"unexsistent", bson.D{{"$type", "date"}}}}}},
				expected: bson.D{{"_id", "double"}, {"v", 42.13}, {"unexsistent", now}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
				paths: []types.Path{types.NewPathFromString("unexsistent")},
			},
			"UnrecognizedOption": {
				id: "array",
				update: bson.D{{
					"$currentDate",
					bson.D{{"v", bson.D{{"array", bson.D{{"unexsistent", bson.D{}}}}}}},
				}},
				err: &mongo.WriteError{
					Code:    2,
					Message: "Unrecognized $currentDate option: array",
				},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

				res, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
				if tc.err != nil {
					require.Nil(t, tc.paths)
					require.Nil(t, tc.stat)
					AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.stat, res)

				var actualB bson.D
				err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actualB)
				require.NoError(t, err)

				expected := ConvertDocument(t, tc.expected)
				actual := ConvertDocument(t, actualB)

				for _, path := range tc.paths {
					testutil.CompareAndSetByPathTime(t, expected, actual, maxDifference, path)
				}
				assert.Equal(t, expected, actual)
			})
		}
	})
}

func TestUpdateFieldInc(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	t.Run("Ok", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			id       string
			update   bson.D
			expected bson.D
			stat     *mongo.UpdateResult
		}{
			"DoubleIncrement": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", float64(42.13)}}}},
				expected: bson.D{{"_id", "double"}, {"v", float64(84.26)}},
			},
			"DoubleIncrementNaN": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", math.NaN()}}}},
				expected: bson.D{{"_id", "double"}, {"v", math.NaN()}},
			},
			"DoubleIncrementPlusInfinity": {
				id:       "double-nan",
				update:   bson.D{{"$inc", bson.D{{"v", math.Inf(+1)}}}},
				expected: bson.D{{"_id", "double-nan"}, {"v", math.NaN()}},
			},
			"DoubleNegativeIncrement": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", float64(-42.13)}}}},
				expected: bson.D{{"_id", "double"}, {"v", float64(0)}},
			},
			"DoubleIncrementIntField": {
				id:       "int32",
				update:   bson.D{{"$inc", bson.D{{"v", float64(1.13)}}}},
				expected: bson.D{{"_id", "int32"}, {"v", float64(43.13)}},
			},
			"DoubleIncrementLongField": {
				id:       "int64",
				update:   bson.D{{"$inc", bson.D{{"v", float64(1.13)}}}},
				expected: bson.D{{"_id", "int64"}, {"v", float64(43.13)}},
			},
			"DoubleIntIncrement": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", int32(1)}}}},
				expected: bson.D{{"_id", "double"}, {"v", float64(43.13)}},
			},
			"DoubleLongIncrement": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", int64(1)}}}},
				expected: bson.D{{"_id", "double"}, {"v", float64(43.13)}},
			},
			"DoubleDoubleMaxIncrement": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", math.MaxFloat64}}}},
				expected: bson.D{{"_id", "double"}, {"v", math.MaxFloat64}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
			},
			"DoubleDoubleNaNIncrement": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", math.NaN()}}}},
				expected: bson.D{{"_id", "double"}, {"v", math.NaN()}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
			},
			"DoubleDoubleBigIncrement": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", float64(2 << 60)}}}},
				expected: bson.D{{"_id", "double"}, {"v", float64(2 << 60)}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
			},
			"DoubleBigDoubleIncrement": {
				id:       "double-big",
				update:   bson.D{{"$inc", bson.D{{"v", 42.13}}}},
				expected: bson.D{{"_id", "double-big"}, {"v", float64(2 << 60)}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 0,
					UpsertedCount: 0,
				},
			},
			"DoubleMaxDoublePositiveIncrement": {
				id:       "double-max",
				update:   bson.D{{"$inc", bson.D{{"v", 42.13}}}},
				expected: bson.D{{"_id", "double-max"}, {"v", math.MaxFloat64}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 0,
					UpsertedCount: 0,
				},
			},
			"DoubleMaxDoubleNegativeIncrement": {
				id:       "double-max",
				update:   bson.D{{"$inc", bson.D{{"v", -42.13}}}},
				expected: bson.D{{"_id", "double-max"}, {"v", math.MaxFloat64}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 0,
					UpsertedCount: 0,
				},
			},
			"DoubleNaNDoublePositiveIncrement": {
				id:       "double-nan",
				update:   bson.D{{"$inc", bson.D{{"v", 42.13}}}},
				expected: bson.D{{"_id", "double-nan"}, {"v", math.NaN()}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
			},
			"DoubleNaNDoubleNegativeIncrement": {
				id:       "double-nan",
				update:   bson.D{{"$inc", bson.D{{"v", -42.13}}}},
				expected: bson.D{{"_id", "double-nan"}, {"v", math.NaN()}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
			},
			"IntIncrement": {
				id:       "int32",
				update:   bson.D{{"$inc", bson.D{{"v", int32(1)}}}},
				expected: bson.D{{"_id", "int32"}, {"v", int32(43)}},
			},
			"IntNegativeIncrement": {
				id:       "int32",
				update:   bson.D{{"$inc", bson.D{{"v", int32(-1)}}}},
				expected: bson.D{{"_id", "int32"}, {"v", int32(41)}},
			},
			"IntIncrementDoubleField": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", int32(1)}}}},
				expected: bson.D{{"_id", "double"}, {"v", float64(43.13)}},
			},
			"IntIncrementLongField": {
				id:       "int64",
				update:   bson.D{{"$inc", bson.D{{"v", int32(1)}}}},
				expected: bson.D{{"_id", "int64"}, {"v", int64(43)}},
			},
			"LongIncrement": {
				id:       "int64",
				update:   bson.D{{"$inc", bson.D{{"v", int64(1)}}}},
				expected: bson.D{{"_id", "int64"}, {"v", int64(43)}},
			},
			"LongNegativeIncrement": {
				id:       "int64",
				update:   bson.D{{"$inc", bson.D{{"v", int64(-1)}}}},
				expected: bson.D{{"_id", "int64"}, {"v", int64(41)}},
			},
			"LongIncrementDoubleField": {
				id:       "double",
				update:   bson.D{{"$inc", bson.D{{"v", int64(1)}}}},
				expected: bson.D{{"_id", "double"}, {"v", float64(43.13)}},
			},
			"LongIncrementIntField": {
				id:       "int32",
				update:   bson.D{{"$inc", bson.D{{"v", int64(1)}}}},
				expected: bson.D{{"_id", "int32"}, {"v", int64(43)}},
			},

			"FieldNotExist": {
				id:       "int32",
				update:   bson.D{{"$inc", bson.D{{"foo", int32(1)}}}},
				expected: bson.D{{"_id", "int32"}, {"v", int32(42)}, {"foo", int32(1)}},
			},
			"IncTwoFields": {
				id:       "int32",
				update:   bson.D{{"$inc", bson.D{{"foo", int32(12)}, {"v", int32(1)}}}},
				expected: bson.D{{"_id", "int32"}, {"v", int32(43)}, {"foo", int32(12)}},
			},
			"DotNotationDocumentFieldExist": {
				id:       "document-composite",
				update:   bson.D{{"$inc", bson.D{{"v.foo", int32(1)}}}},
				expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(43)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
			},
			"DotNotationDocumentFieldNotExist": {
				id:       "int32",
				update:   bson.D{{"$inc", bson.D{{"foo.bar", int32(1)}}}},
				expected: bson.D{{"_id", "int32"}, {"v", int32(42)}, {"foo", bson.D{{"bar", int32(1)}}}},
			},
			"DotNotationArrayFieldExist": {
				id:       "document-composite",
				update:   bson.D{{"$inc", bson.D{{"v.array.0", int32(1)}}}},
				expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(43), "foo", nil}}}}},
			},
			"DotNotationArrayFieldNotExist": {
				id:     "int32",
				update: bson.D{{"$inc", bson.D{{"foo.0.baz", int32(1)}}}},
				expected: bson.D{
					{"_id", "int32"},
					{"v", int32(42)},
					{"foo", bson.D{{"0", bson.D{{"baz", int32(1)}}}}},
				},
			},
			"DocumentDotNotationArrayFieldNotExist": {
				id:     "document",
				update: bson.D{{"$inc", bson.D{{"v.0.foo", int32(1)}}}},
				expected: bson.D{
					{"_id", "document"},
					{"v", bson.D{{"foo", int32(42)}, {"0", bson.D{{"foo", int32(1)}}}}},
				},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

				result, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
				require.NoError(t, err)

				if tc.stat != nil {
					require.Equal(t, tc.stat, result)
				}

				var actual bson.D
				err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
				require.NoError(t, err)

				AssertEqualDocuments(t, tc.expected, actual)
			})
		}
	})

	t.Run("Err", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			id     string
			update bson.D
			err    *mongo.WriteError
			alt    string
		}{
			"IncOnDocument": {
				id:     "document",
				update: bson.D{{"$inc", bson.D{{"v", int32(1)}}}},
				err: &mongo.WriteError{
					Code: 14,
					Message: `Cannot apply $inc to a value of non-numeric type. ` +
						`{_id: "document"} has the field 'v' of non-numeric type object`,
				},
			},
			"IncOnArray": {
				id:     "array",
				update: bson.D{{"$inc", bson.D{{"v", int32(1)}}}},
				err: &mongo.WriteError{
					Code: 14,
					Message: `Cannot apply $inc to a value of non-numeric type. ` +
						`{_id: "array"} has the field 'v' of non-numeric type array`,
				},
			},
			"IncOnString": {
				id:     "string",
				update: bson.D{{"$inc", "string"}},
				err: &mongo.WriteError{
					Code: 9,
					Message: `Modifiers operate on fields but we found type string instead.` +
						` For example: {$mod: {<field>: ...}} not {$inc: "string"}`,
				},
				alt: "Modifiers operate on fields but we found another type instead",
			},
			"IncWithStringValue": {
				id:     "string",
				update: bson.D{{"$inc", bson.D{{"v", "bad value"}}}},
				err: &mongo.WriteError{
					Code:    14,
					Message: `Cannot increment with non-numeric argument: {v: "bad value"}`,
				},
			},
			"DoubleIncOnNullValue": {
				id:     "string",
				update: bson.D{{"$inc", bson.D{{"v", float64(1)}}}},
				err: &mongo.WriteError{
					Code: 14,
					Message: `Cannot apply $inc to a value of non-numeric type. ` +
						`{_id: "string"} has the field 'v' of non-numeric type string`,
				},
			},
			"IntIncOnNullValue": {
				id:     "string",
				update: bson.D{{"$inc", bson.D{{"v", int32(1)}}}},
				err: &mongo.WriteError{
					Code: 14,
					Message: `Cannot apply $inc to a value of non-numeric type. ` +
						`{_id: "string"} has the field 'v' of non-numeric type string`,
				},
			},
			"LongIncOnNullValue": {
				id:     "string",
				update: bson.D{{"$inc", bson.D{{"v", int64(1)}}}},
				err: &mongo.WriteError{
					Code: 14,
					Message: `Cannot apply $inc to a value of non-numeric type. ` +
						`{_id: "string"} has the field 'v' of non-numeric type string`,
				},
			},
			"ArrayDotNotationFieldNotExist": {
				id:     "document-composite",
				update: bson.D{{"$inc", bson.D{{"v.array.foo", int32(1)}}}},
				err: &mongo.WriteError{
					Code:    28,
					Message: `Cannot create field 'foo' in element {array: [ 42, "foo", null ]}`,
				},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

				_, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
				require.NotNil(t, tc.err)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
			})
		}
	})
}

func TestUpdateFieldSet(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		id       string
		update   bson.D
		expected bson.D
		err      *mongo.WriteError
		stat     *mongo.UpdateResult
		alt      string
	}{
		"Many": {
			id:       "string",
			update:   bson.D{{"$set", bson.D{{"foo", int32(1)}, {"bar", bson.A{}}}}},
			expected: bson.D{{"_id", "string"}, {"v", "foo"}, {"bar", bson.A{}}, {"foo", int32(1)}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"NilOperand": {
			id:     "string",
			update: bson.D{{"$set", nil}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type null instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: null}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"String": {
			id:     "string",
			update: bson.D{{"$set", "string"}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: \"string\"}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"Array": {
			id:     "string",
			update: bson.D{{"$set", bson.A{}}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: []}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"EmptyDoc": {
			id:       "string",
			update:   bson.D{{"$set", bson.D{}}},
			expected: bson.D{{"_id", "string"}, {"v", "foo"}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"OkSetString": {
			id:       "string",
			update:   bson.D{{"$set", bson.D{{"v", "ok value"}}}},
			expected: bson.D{{"_id", "string"}, {"v", "ok value"}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"ArrayNil": {
			id:       "string",
			update:   bson.D{{"$set", bson.D{{"v", bson.A{nil}}}}},
			expected: bson.D{{"_id", "string"}, {"v", bson.A{nil}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"FieldNotExist": {
			id:       "string",
			update:   bson.D{{"$set", bson.D{{"foo", int32(1)}}}},
			expected: bson.D{{"_id", "string"}, {"v", "foo"}, {"foo", int32(1)}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"Double": {
			id:       "double",
			update:   bson.D{{"$set", bson.D{{"v", float64(1)}}}},
			expected: bson.D{{"_id", "double"}, {"v", float64(1)}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"NaN": {
			id:       "double",
			update:   bson.D{{"$set", bson.D{{"v", math.NaN()}}}},
			expected: bson.D{{"_id", "double"}, {"v", math.NaN()}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"EmptyArray": {
			id:       "double",
			update:   bson.D{{"$set", bson.D{{"v", bson.A{}}}}},
			expected: bson.D{{"_id", "double"}, {"v", bson.A{}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"Null": {
			id:       "double",
			update:   bson.D{{"$set", bson.D{{"v", nil}}}},
			expected: bson.D{{"_id", "double"}, {"v", nil}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"Int32": {
			id:       "double",
			update:   bson.D{{"$set", bson.D{{"v", int32(1)}}}},
			expected: bson.D{{"_id", "double"}, {"v", int32(1)}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"SetTwoFields": {
			id:       "int32-zero",
			update:   bson.D{{"$set", bson.D{{"foo", int32(12)}, {"v", math.NaN()}}}},
			expected: bson.D{{"_id", "int32-zero"}, {"v", math.NaN()}, {"foo", int32(12)}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"SetSameValueInt": {
			id:       "int32",
			update:   bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			expected: bson.D{{"_id", "int32"}, {"v", int32(42)}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"SetSameValueNan": {
			id:       "double-nan",
			update:   bson.D{{"$set", bson.D{{"v", math.NaN()}}}},
			expected: bson.D{{"_id", "double-nan"}, {"v", math.NaN()}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"DotNotationDocumentFieldExist": {
			id:       "document-composite",
			update:   bson.D{{"$set", bson.D{{"v.foo", int32(1)}}}},
			expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(1)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"DotNotationDocumentFieldNotExist": {
			id:       "int32",
			update:   bson.D{{"$set", bson.D{{"foo.bar", int32(1)}}}},
			expected: bson.D{{"_id", "int32"}, {"v", int32(42)}, {"foo", bson.D{{"bar", int32(1)}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"DotNotationArrayFieldExist": {
			id:       "document-composite",
			update:   bson.D{{"$set", bson.D{{"v.array.0", int32(1)}}}},
			expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(1), "foo", nil}}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"DotNotationArrayFieldNotExist": {
			id:     "int32",
			update: bson.D{{"$set", bson.D{{"foo.0.baz", int32(1)}}}},
			expected: bson.D{
				{"_id", "int32"},
				{"v", int32(42)},
				{"foo", bson.D{{"0", bson.D{{"baz", int32(1)}}}}},
			},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"DocumentDotNotationArrayFieldNotExist": {
			id:     "document",
			update: bson.D{{"$set", bson.D{{"v.0.foo", int32(1)}}}},
			expected: bson.D{
				{"_id", "document"},
				{"v", bson.D{{"foo", int32(42)}, {"0", bson.D{{"foo", int32(1)}}}}},
			},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			res, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
			if tc.err != nil {
				require.Nil(t, tc.expected)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.stat, res)

			var actual bson.D
			err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.expected, actual)
		})
	}
}

func TestUpdateFieldSetOnInsert(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		id           string
		update       bson.D
		expected     bson.D
		err          *mongo.WriteError
		alt          string
		expectedStat *mongo.UpdateResult
		upserted     bool
	}{
		"Array": {
			id:       "array-set-on-insert",
			update:   bson.D{{"$setOnInsert", bson.D{{"v", bson.A{}}}}},
			expected: bson.D{{"_id", "array-set-on-insert"}, {"v", bson.A{}}},
			expectedStat: &mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
			},
			upserted: true,
		},
		"Nil": {
			id:       "nil",
			update:   bson.D{{"$setOnInsert", bson.D{{"v", nil}}}},
			expected: bson.D{{"_id", "nil"}, {"v", nil}},
			expectedStat: &mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
			},
			upserted: true,
		},
		"EmptyDoc": {
			id:       "doc",
			update:   bson.D{{"$setOnInsert", bson.D{}}},
			expected: bson.D{{"_id", "doc"}},
			expectedStat: &mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
			},
			upserted: true,
		},
		"EmptyArray": {
			id:     "array",
			update: bson.D{{"$setOnInsert", bson.A{}}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: []}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"DoubleDouble": {
			id:     "double",
			update: bson.D{{"$setOnInsert", 43.13}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: 43.13}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"ErrNaN": {
			id:     "double-nan",
			update: bson.D{{"$setOnInsert", math.NaN()}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: nan.0}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"ErrString": {
			id:     "string",
			update: bson.D{{"$setOnInsert", "any string"}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: \"any string\"}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"ErrNil": {
			id:     "nil",
			update: bson.D{{"$setOnInsert", nil}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type null instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: null}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"DotNotationDocumentFieldExist": {
			id:       "document-composite",
			update:   bson.D{{"$set", bson.D{{"v.foo", int32(1)}}}},
			expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(1)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
			expectedStat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"DotNotationDocumentFieldNotExist": {
			id:       "int32",
			update:   bson.D{{"$setOnInsert", bson.D{{"foo.bar", int32(1)}}}},
			expected: bson.D{{"_id", "int32"}, {"v", int32(42)}},
			expectedStat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"DotNotationArrayFieldExist": {
			id:       "document-composite",
			update:   bson.D{{"$setOnInsert", bson.D{{"v.array.0", int32(1)}}}},
			expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
			expectedStat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"DotNotationArrFieldNotExist": {
			id:     "int32",
			update: bson.D{{"$setOnInsert", bson.D{{"foo.0.baz", int32(1)}}}},
			expected: bson.D{
				{"_id", "int32"},
				{"v", int32(42)},
			},
			expectedStat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"DocumentDotNotationArrFieldNotExist": {
			id:     "document",
			update: bson.D{{"$setOnInsert", bson.D{{"v.0.foo", int32(1)}}}},
			expected: bson.D{
				{"_id", "document"},
				{"v", bson.D{{"foo", int32(42)}}},
			},
			expectedStat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, collection := setup.Setup(t, shareddata.Composites, shareddata.Scalars)

			opts := options.Update().SetUpsert(true)
			actualUpdateStat, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update, opts)
			if tc.err != nil {
				require.Nil(t, tc.expected)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
				return
			}

			require.NoError(t, err)

			expectedStat := tc.expectedStat
			if tc.upserted {
				expectedStat.UpsertedID = tc.id
			}
			assert.Equal(t, expectedStat, actualUpdateStat)

			var actual bson.D
			err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.expected, actual)
		})
	}
}

func TestUpdateFieldUnset(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		id           string
		update       bson.D
		expected     bson.D
		expectedStat *mongo.UpdateResult
		err          *mongo.WriteError
		alt          string
	}{
		"Empty": {
			id:       "string",
			update:   bson.D{{"$unset", bson.D{}}},
			expected: bson.D{{"_id", "string"}, {"v", "foo"}},
			expectedStat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"EmptyArray": {
			id:     "document-composite",
			update: bson.D{{"$unset", bson.A{}}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$unset: []}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			opts := options.Update().SetUpsert(true)
			actualStat, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update, opts)

			if tc.err != nil {
				require.Nil(t, tc.expected)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
				return
			}

			require.NoError(t, err)
			actualStat.UpsertedID = nil
			assert.Equal(t, tc.expectedStat, actualStat)

			var actual bson.D
			err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.expected, actual)
		})
	}
}

func TestUpdateFieldMixed(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	for name, tc := range map[string]struct {
		filter   bson.D
		update   bson.D
		expected bson.D
		err      *mongo.WriteError
	}{
		"SetSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$setOnInsert", bson.D{{"v", math.NaN()}}},
			},
			expected: bson.D{{"_id", "test"}, {"foo", int32(12)}, {"v", math.NaN()}},
		},
		"SetIncSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$inc", bson.D{{"foo", int32(1)}}},
				{"$setOnInsert", bson.D{{"v", math.NaN()}}},
			},
			err: &mongo.WriteError{
				Code:    40,
				Message: "Updating the path 'foo' would create a conflict at 'foo'",
			},
		},
		"UnknownOperator": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{{"$foo", bson.D{{"foo", int32(1)}}}},
			err: &mongo.WriteError{
				Code:    9,
				Message: "Unknown modifier: $foo. Expected a valid update modifier or pipeline-style update specified as an array",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			opts := options.Update().SetUpsert(true)
			actualStat, err := collection.UpdateOne(ctx, tc.filter, tc.update, opts)

			if tc.err != nil {
				require.Nil(t, tc.expected)
				AssertEqualWriteError(t, *tc.err, err)
				return
			}

			require.NoError(t, err)
			actualStat.UpsertedID = nil

			expectedStat := &mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
			}
			assert.Equal(t, expectedStat, actualStat)

			var actual bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.expected, actual)
		})
	}
}

func TestUpdateFieldPopArrayOperator(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()

	t.Run("Ok", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			id       string
			update   bson.D
			expected bson.D
			stat     *mongo.UpdateResult
		}{
			"Pop": {
				id:       "array-three",
				update:   bson.D{{"$pop", bson.D{{"v", 1}}}},
				expected: bson.D{{"_id", "array-three"}, {"v", bson.A{int32(42), "foo"}}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
			},
			"PopFirst": {
				id:       "array-three",
				update:   bson.D{{"$pop", bson.D{{"v", -1}}}},
				expected: bson.D{{"_id", "array-three"}, {"v", bson.A{"foo", nil}}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
			},
			"PopDotNotation": {
				id:       "document-composite",
				update:   bson.D{{"$pop", bson.D{{"v.array", 1}}}},
				expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo"}}}}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 1,
					UpsertedCount: 0,
				},
			},
			"PopEmptyArray": {
				id:       "array-empty",
				update:   bson.D{{"$pop", bson.D{{"v", 1}}}},
				expected: bson.D{{"_id", "array-empty"}, {"v", bson.A{}}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 0,
					UpsertedCount: 0,
				},
			},
			"PopNoSuchKey": {
				id:       "array",
				update:   bson.D{{"$pop", bson.D{{"foo", 1}}}},
				expected: bson.D{{"_id", "array"}, {"v", bson.A{int32(42)}}},
			},
			"PopEmptyValue": {
				id:       "array",
				update:   bson.D{{"$pop", bson.D{}}},
				expected: bson.D{{"_id", "array"}, {"v", bson.A{int32(42)}}},
				stat: &mongo.UpdateResult{
					MatchedCount:  1,
					ModifiedCount: 0,
					UpsertedCount: 0,
				},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

				result, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
				require.NoError(t, err)

				if tc.stat != nil {
					require.Equal(t, tc.stat, result)
				}

				var actual bson.D
				err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
				require.NoError(t, err)

				AssertEqualDocuments(t, tc.expected, actual)
			})
		}
	})

	t.Run("Err", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			id     string
			update bson.D
			err    *mongo.WriteError
			alt    string
		}{
			"PopNotValidValueString": {
				id:     "array",
				update: bson.D{{"$pop", bson.D{{"v", "foo"}}}},
				err: &mongo.WriteError{
					Code:    9,
					Message: "Expected a number in: v: \"foo\"",
				},
			},
			"PopNotValidValueInt": {
				id:     "array",
				update: bson.D{{"$pop", bson.D{{"v", int32(42)}}}},
				err: &mongo.WriteError{
					Code:    9,
					Message: "$pop expects 1 or -1, found: 42",
				},
			},
			"PopOnNonArray": {
				id:     "int32",
				update: bson.D{{"$pop", bson.D{{"v", 1}}}},
				err: &mongo.WriteError{
					Code:    14,
					Message: "Path 'v' contains an element of non-array type 'int'",
				},
			},
			// TODO: https://github.com/FerretDB/FerretDB/issues/364
			//"PopLastAndFirst": {
			//	id:     "array-three",
			//	update: bson.D{{"$pop", bson.D{{"v", 1}, {"v", -1}}}},
			//	err: &mongo.WriteError{
			//		Code:    40,
			//		Message: "Updating the path 'v' would create a conflict at 'v'",
			//	},
			//},
			"PopDotNotationNonArray": {
				id:     "document-composite",
				update: bson.D{{"$pop", bson.D{{"v.foo", 1}}}},
				err: &mongo.WriteError{
					Code:    14,
					Message: "Path 'v.foo' contains an element of non-array type 'int'",
				},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

				_, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
				require.NotNil(t, tc.err)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
			})
		}
	})
}

// This test is to ensure that the order of fields in the document is preserved.
func TestUpdateDocumentFieldsOrder(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Tigris schema would fail this test")

	ctx, collection := setup.Setup(t, shareddata.Composites)

	_, err := collection.UpdateOne(
		ctx,
		bson.D{{"_id", "document"}},
		bson.D{{"$set", bson.D{{"foo", int32(42)}, {"bar", "baz"}}}},
	)
	require.NoError(t, err)

	var updated bson.D

	err = collection.FindOne(ctx, bson.D{{"_id", "document"}}).Decode(&updated)
	require.NoError(t, err)

	expected := bson.D{
		{"_id", "document"},
		{"v", bson.D{{"foo", int32(42)}}},
		{"bar", "baz"},
		{"foo", int32(42)},
	}

	AssertEqualDocuments(t, expected, updated)

	_, err = collection.UpdateOne(
		ctx,
		bson.D{{"_id", "document"}},
		bson.D{{"$unset", bson.D{{"foo", ""}}}},
	)
	require.NoError(t, err)

	err = collection.FindOne(ctx, bson.D{{"_id", "document"}}).Decode(&updated)
	require.NoError(t, err)

	expected = bson.D{
		{"_id", "document"},
		{"v", bson.D{{"foo", int32(42)}}},
		{"bar", "baz"},
	}

	AssertEqualDocuments(t, expected, updated)

	_, err = collection.UpdateOne(
		ctx,
		bson.D{{"_id", "document"}},
		bson.D{{"$set", bson.D{{"abc", int32(42)}}}},
	)
	require.NoError(t, err)

	err = collection.FindOne(ctx, bson.D{{"_id", "document"}}).Decode(&updated)
	require.NoError(t, err)

	expected = bson.D{
		{"_id", "document"},
		{"v", bson.D{{"foo", int32(42)}}},
		{"bar", "baz"},
		{"abc", int32(42)},
	}

	AssertEqualDocuments(t, expected, updated)
}

// This test is to ensure that the order of fields in the document is preserved.
func TestUpdateDocumentFieldsOrderSimplified(t *testing.T) {
	ctx, collection := setup.Setup(t)

	_, err := collection.InsertOne(ctx, bson.D{{"_id", "document"}, {"foo", int32(42)}, {"bar", "baz"}})
	require.NoError(t, err)

	var inserted bson.D
	err = collection.FindOne(ctx, bson.D{{"_id", "document"}}).Decode(&inserted)
	require.NoError(t, err)

	expected := bson.D{
		{"_id", "document"},
		{"foo", int32(42)},
		{"bar", "baz"},
	}
	AssertEqualDocuments(t, expected, inserted)

	_, err = collection.UpdateOne(
		ctx,
		bson.D{{"_id", "document"}},
		bson.D{{"$unset", bson.D{{"foo", ""}, {"bar", ""}}}},
	)
	require.NoError(t, err)

	var updated bson.D
	err = collection.FindOne(ctx, bson.D{{"_id", "document"}}).Decode(&updated)
	require.NoError(t, err)

	expected = bson.D{
		{"_id", "document"},
	}
	AssertEqualDocuments(t, expected, updated)

	_, err = collection.UpdateOne(
		ctx,
		bson.D{{"_id", "document"}},
		bson.D{{"$set", bson.D{{"foo", int32(42)}, {"bar", "baz"}}}},
	)
	require.NoError(t, err)

	err = collection.FindOne(ctx, bson.D{{"_id", "document"}}).Decode(&updated)
	require.NoError(t, err)

	expected = bson.D{
		{"_id", "document"},
		{"bar", "baz"},
		{"foo", int32(42)},
	}
	AssertEqualDocuments(t, expected, updated)
}
