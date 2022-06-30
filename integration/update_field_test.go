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

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// This file is for tests of:
// - $currentDate
// - $inc
// - $min
// - $max
// - $mul
// - $rename
// - $set
// - $setOnInsert
// - $unset

func TestUpdateFieldCurrentDateTimestamp(t *testing.T) {
	t.Parallel()

	// store the current timestamp with $currentDate operator;
	t.Run("currentDateReadBack", func(t *testing.T) {
		maxDifference := time.Duration(2 * time.Second)
		nowTimestamp := primitive.Timestamp{T: uint32(time.Now().Unix()), I: uint32(0)}
		id := "string-empty"

		stat := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 1,
			UpsertedCount: 0,
		}
		path := types.NewPathFromString("value")
		result := bson.D{{"_id", id}, {"value", nowTimestamp}}

		ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

		update := bson.D{{"$currentDate", bson.D{{"value", bson.D{{"$type", "timestamp"}}}}}}
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
		updateBSON := bson.D{{"$set", bson.D{{"value", nowTimestamp}}}}
		expectedBSON := bson.D{{"_id", id}, {"value", nowTimestamp}}
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
}

func TestUpdateFieldIncErrors(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		filter bson.D
		update bson.D
		err    *mongo.WriteError
		alt    string
	}{
		"IncOnDocument": {
			filter: bson.D{{"_id", "document"}},
			update: bson.D{{"$inc", bson.D{{"value", int32(1)}}}},
			err: &mongo.WriteError{
				Code: 14,
				Message: `Cannot apply $inc to a value of non-numeric type. ` +
					`{_id: "document"} has the field 'value' of non-numeric type object`,
			},
		},
		"IncOnArray": {
			filter: bson.D{{"_id", "array"}},
			update: bson.D{{"$inc", bson.D{{"value", int32(1)}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot apply $inc to a value of non-numeric type. {_id: "array"} has the field 'value' of non-numeric type array`,
			},
		},
		"IncOnString": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$inc", "string"}},
			err: &mongo.WriteError{
				Code: 9,
				Message: `Modifiers operate on fields but we found type string instead.` +
					` For example: {$mod: {<field>: ...}} not {$inc: "string"}`,
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"IncWithStringValue": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$inc", bson.D{{"value", "bad value"}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot increment with non-numeric argument: {value: "bad value"}`,
			},
		},
		"DoubleIncOnNullValue": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$inc", bson.D{{"value", float64(1)}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot apply $inc to a value of non-numeric type. {_id: "string"} has the field 'value' of non-numeric type string`,
			},
		},
		"IntIncOnNullValue": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$inc", bson.D{{"value", int32(1)}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot apply $inc to a value of non-numeric type. {_id: "string"} has the field 'value' of non-numeric type string`,
			},
		},
		"LongIncOnNullValue": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$inc", bson.D{{"value", int64(1)}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot apply $inc to a value of non-numeric type. {_id: "string"} has the field 'value' of non-numeric type string`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)

			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", "document"}, {"value", bson.D{{"foo", "bar"}}}},
				bson.D{{"_id", "array"}, {"value", bson.A{"foo"}}},
				bson.D{{"_id", "string"}, {"value", "foo"}},
			})
			require.NoError(t, err)

			_, err = collection.UpdateOne(ctx, tc.filter, tc.update)
			require.NotNil(t, tc.err)
			AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
		})
	}
}

func TestUpdateFieldInc(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		filter bson.D
		update bson.D
		result bson.D
	}{
		"DoubleIncrement": {
			filter: bson.D{{"_id", "double"}},
			update: bson.D{{"$inc", bson.D{{"value", float64(42.13)}}}},
			result: bson.D{{"_id", "double"}, {"value", float64(84.26)}},
		},
		"DoubleIncrementNaN": {
			filter: bson.D{{"_id", "double"}},
			update: bson.D{{"$inc", bson.D{{"value", math.NaN()}}}},
			result: bson.D{{"_id", "double"}, {"value", math.NaN()}},
		},
		"DoubleIncrementPlusInfinity": {
			filter: bson.D{{"_id", "double-nan"}},
			update: bson.D{{"$inc", bson.D{{"value", math.Inf(+1)}}}},
			result: bson.D{{"_id", "double-nan"}, {"value", math.NaN()}},
		},
		"DoubleNegativeIncrement": {
			filter: bson.D{{"_id", "double"}},
			update: bson.D{{"$inc", bson.D{{"value", float64(-42.13)}}}},
			result: bson.D{{"_id", "double"}, {"value", float64(0)}},
		},
		"DoubleIncrementIntField": {
			filter: bson.D{{"_id", "int32"}},
			update: bson.D{{"$inc", bson.D{{"value", float64(1.13)}}}},
			result: bson.D{{"_id", "int32"}, {"value", float64(43.13)}},
		},
		"DoubleIncrementLongField": {
			filter: bson.D{{"_id", "int64"}},
			update: bson.D{{"$inc", bson.D{{"value", float64(1.13)}}}},
			result: bson.D{{"_id", "int64"}, {"value", float64(43.13)}},
		},
		"DoubleIntIncrement": {
			filter: bson.D{{"_id", "double"}},
			update: bson.D{{"$inc", bson.D{{"value", int32(1)}}}},
			result: bson.D{{"_id", "double"}, {"value", float64(43.13)}},
		},
		"DoubleLongIncrement": {
			filter: bson.D{{"_id", "double"}},
			update: bson.D{{"$inc", bson.D{{"value", int64(1)}}}},
			result: bson.D{{"_id", "double"}, {"value", float64(43.13)}},
		},
		"IntIncrement": {
			filter: bson.D{{"_id", "int32"}},
			update: bson.D{{"$inc", bson.D{{"value", int32(1)}}}},
			result: bson.D{{"_id", "int32"}, {"value", int32(43)}},
		},
		"IntNegativeIncrement": {
			filter: bson.D{{"_id", "int32"}},
			update: bson.D{{"$inc", bson.D{{"value", int32(-1)}}}},
			result: bson.D{{"_id", "int32"}, {"value", int32(41)}},
		},
		"IntIncrementDoubleField": {
			filter: bson.D{{"_id", "double"}},
			update: bson.D{{"$inc", bson.D{{"value", int32(1)}}}},
			result: bson.D{{"_id", "double"}, {"value", float64(43.13)}},
		},
		"IntIncrementLongField": {
			filter: bson.D{{"_id", "int64"}},
			update: bson.D{{"$inc", bson.D{{"value", int32(1)}}}},
			result: bson.D{{"_id", "int64"}, {"value", int64(43)}},
		},
		"LongIncrement": {
			filter: bson.D{{"_id", "int64"}},
			update: bson.D{{"$inc", bson.D{{"value", int64(1)}}}},
			result: bson.D{{"_id", "int64"}, {"value", int64(43)}},
		},
		"LongNegativeIncrement": {
			filter: bson.D{{"_id", "int64"}},
			update: bson.D{{"$inc", bson.D{{"value", int64(-1)}}}},
			result: bson.D{{"_id", "int64"}, {"value", int64(41)}},
		},
		"LongIncrementDoubleField": {
			filter: bson.D{{"_id", "double"}},
			update: bson.D{{"$inc", bson.D{{"value", int64(1)}}}},
			result: bson.D{{"_id", "double"}, {"value", float64(43.13)}},
		},
		"LongIncrementIntField": {
			filter: bson.D{{"_id", "int32"}},
			update: bson.D{{"$inc", bson.D{{"value", int64(1)}}}},
			result: bson.D{{"_id", "int32"}, {"value", int64(43)}},
		},

		"FieldNotExist": {
			filter: bson.D{{"_id", "int32"}},
			update: bson.D{{"$inc", bson.D{{"foo", int32(1)}}}},
			result: bson.D{{"_id", "int32"}, {"value", int32(42)}, {"foo", int32(1)}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)

			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", "double"}, {"value", 42.13}},
				bson.D{{"_id", "double-nan"}, {"value", math.NaN()}},
				bson.D{{"_id", "int32"}, {"value", int32(42)}},
				bson.D{{"_id", "int64"}, {"value", int64(42)}},
			})
			require.NoError(t, err)

			_, err = collection.UpdateOne(ctx, tc.filter, tc.update)
			require.NoError(t, err)

			var actual bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&actual)
			require.NoError(t, err)

			AssertEqualDocuments(t, tc.result, actual)
		})
	}
}
