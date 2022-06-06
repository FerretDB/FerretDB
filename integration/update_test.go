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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestUpdateUpsert(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Composites)

	// this upsert inserts document
	filter := bson.D{{"foo", "bar"}}
	update := bson.D{{"$set", bson.D{{"foo", "baz"}}}}
	res, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	require.NoError(t, err)

	id := res.UpsertedID
	assert.NotEmpty(t, id)
	res.UpsertedID = nil
	expected := &mongo.UpdateResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: 1,
	}
	require.Equal(t, expected, res)

	// check inserted document
	var doc bson.D
	err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&doc)
	require.NoError(t, err)
	if !AssertEqualDocuments(t, bson.D{{"_id", id}, {"foo", "baz"}}, doc) {
		t.FailNow()
	}

	// this upsert updates document
	filter = bson.D{{"foo", "baz"}}
	update = bson.D{{"$set", bson.D{{"foo", "qux"}}}}
	res, err = collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	require.NoError(t, err)

	expected = &mongo.UpdateResult{
		MatchedCount:  1,
		ModifiedCount: 1,
		UpsertedCount: 0,
	}
	require.Equal(t, expected, res)

	// check updated document
	err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&doc)
	require.NoError(t, err)
	AssertEqualDocuments(t, bson.D{{"_id", id}, {"foo", "qux"}}, doc)
}

func TestUpdateIncOperatorErrors(t *testing.T) {
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

func TestUpdateIncOperator(t *testing.T) {
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

func TestUpdateSet(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		id     string
		update bson.D
		result bson.D
		err    *mongo.WriteError
		stat   *mongo.UpdateResult
		alt    string
	}{
		"Many": {
			id:     "string",
			update: bson.D{{"$set", bson.D{{"foo", int32(1)}, {"bar", bson.A{}}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			result: bson.D{{"_id", "string"}, {"value", "foo"}, {"bar", bson.A{}}, {"foo", int32(1)}},
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
			id:     "string",
			update: bson.D{{"$set", bson.D{}}},
			result: bson.D{{"_id", "string"}, {"value", "foo"}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"OkSetString": {
			id:     "string",
			update: bson.D{{"$set", bson.D{{"value", "ok value"}}}},
			result: bson.D{{"_id", "string"}, {"value", "ok value"}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"ArrayNil": {
			id:     "string",
			update: bson.D{{"$set", bson.D{{"value", bson.A{nil}}}}},
			result: bson.D{{"_id", "string"}, {"value", bson.A{nil}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"FieldNotExist": {
			id:     "string",
			update: bson.D{{"$set", bson.D{{"foo", int32(1)}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			result: bson.D{{"_id", "string"}, {"value", "foo"}, {"foo", int32(1)}},
		},
		"Double": {
			id:     "double",
			update: bson.D{{"$set", bson.D{{"value", float64(1)}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			result: bson.D{{"_id", "double"}, {"value", float64(1)}},
		},
		"NaN": {
			id:     "double",
			update: bson.D{{"$set", bson.D{{"value", math.NaN()}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			result: bson.D{{"_id", "double"}, {"value", math.NaN()}},
		},
		"EmptyArray": {
			id:     "double",
			update: bson.D{{"$set", bson.D{{"value", bson.A{}}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			result: bson.D{{"_id", "double"}, {"value", bson.A{}}},
		},
		"Null": {
			id:     "double",
			update: bson.D{{"$set", bson.D{{"value", nil}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			result: bson.D{{"_id", "double"}, {"value", nil}},
		},
		"Int32": {
			id:     "double",
			update: bson.D{{"$set", bson.D{{"value", int32(1)}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			result: bson.D{{"_id", "double"}, {"value", int32(1)}},
		},
		"Inf": {
			id:     "double",
			update: bson.D{{"$set", bson.D{{"value", math.Inf(+1)}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			result: bson.D{{"_id", "double"}, {"value", math.Inf(+1)}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)
			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", "string"}, {"value", "foo"}},
				bson.D{{"_id", "double"}, {"value", float64(0.0)}},
			})
			require.NoError(t, err)

			res, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
			if tc.err != nil {
				require.Nil(t, tc.result)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.stat, res)

			var actual bson.D
			err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.result, actual)
		})
	}
}

func TestUpdateSetOnInsertOperator(t *testing.T) {
	t.Parallel()

	notModified := &mongo.UpdateResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: 1,
	}

	for name, tc := range map[string]struct {
		filter      bson.D
		setOnInsert any
		stat        *mongo.UpdateResult
		res         bson.D
		err         *mongo.WriteError
		alt         string
	}{
		"Array": {
			filter:      bson.D{{"_id", "array"}},
			setOnInsert: bson.D{{"value", bson.A{}}},
			res:         bson.D{{"_id", "array"}, {"value", bson.A{}}},
		},
		"Nil": {
			filter:      bson.D{{"_id", "nil"}},
			setOnInsert: bson.D{{"value", nil}},
			res:         bson.D{{"_id", "nil"}, {"value", nil}},
		},
		"EmptyDoc": {
			filter:      bson.D{{"_id", "doc"}},
			setOnInsert: bson.D{},
			res:         bson.D{{"_id", "doc"}},
		},
		"EmptyArray": {
			filter:      bson.D{{"_id", "array"}},
			setOnInsert: bson.A{},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: []}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"DoubleDouble": {
			filter:      bson.D{{"_id", "double"}},
			setOnInsert: 43.13,
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: 43.13}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"ErrNaN": {
			filter:      bson.D{{"_id", "double-nan"}},
			setOnInsert: math.NaN(),
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: nan.0}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"ErrString": {
			filter:      bson.D{{"_id", "string"}},
			setOnInsert: "any string",
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: \"any string\"}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
		"ErrNil": {
			filter:      bson.D{{"_id", "nil"}},
			setOnInsert: nil,
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type null instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: null}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)

			opts := options.Update().SetUpsert(true)
			var res *mongo.UpdateResult
			res, err := collection.UpdateOne(ctx, tc.filter, bson.D{{"$setOnInsert", tc.setOnInsert}}, opts)
			if tc.err != nil {
				require.Nil(t, tc.res)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
				return
			}

			require.NoError(t, err)
			id := res.UpsertedID
			assert.NotEmpty(t, id)
			res.UpsertedID = nil

			expectedRes := notModified
			if tc.stat != nil {
				expectedRes = tc.stat
			}
			assert.Equal(t, expectedRes, res)

			var actual bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.res, actual)
		})
	}
}

func TestUpdateMany(t *testing.T) {
	t.Parallel()

	notModified := &mongo.UpdateResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: 1,
	}

	for name, tc := range map[string]struct {
		filter bson.D
		update bson.D
		stat   *mongo.UpdateResult
		res    bson.D
		err    *mongo.WriteError
		alt    string
	}{
		"SetSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$setOnInsert", bson.D{{"value", math.NaN()}}},
			},
			res: bson.D{{"_id", "test"}, {"foo", int32(12)}, {"value", math.NaN()}},
		},
		"SetTwoFields": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}, {"value", math.NaN()}}},
			},
			res: bson.D{{"_id", "test"}, {"foo", int32(12)}, {"value", math.NaN()}},
		},
		"IncTwoFields": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$inc", bson.D{{"foo", int32(12)}, {"value", int32(1)}}},
			},
			res: bson.D{{"_id", "test"}, {"foo", int32(12)}, {"value", int32(1)}},
		},
		"SetIncSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$inc", bson.D{{"foo", int32(1)}}},
				{"$setOnInsert", bson.D{{"value", math.NaN()}}},
			},
			err: &mongo.WriteError{
				Code:    40,
				Message: "Updating the path 'foo' would create a conflict at 'foo'",
			},
		},
		"UnsetString": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$unset", bson.D{{"value", int32(1)}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
			res: bson.D{{"_id", "string"}},
		},
		"UnsetEmpty": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$unset", bson.D{}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
			res: bson.D{{"_id", "string"}, {"value", "foo"}},
		},
		"UnsetField": {
			filter: bson.D{{"_id", "document-composite"}},
			update: bson.D{{"$unset", bson.D{{"value", bson.D{{"array", int32(1)}}}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
			},
			res: bson.D{{"_id", "document-composite"}},
		},
		"UnsetEmptyArray": {
			filter: bson.D{{"_id", "document-composite"}},
			update: bson.D{{"$unset", bson.A{}}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$unset: []}",
			},
			alt: "Modifiers operate on fields but we found another type instead",
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
			ctx, collection := setup(t)

			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", "string"}, {"value", "foo"}},
			})
			require.NoError(t, err)

			opts := options.Update().SetUpsert(true)
			var res *mongo.UpdateResult
			res, err = collection.UpdateOne(ctx, tc.filter, tc.update, opts)

			if tc.err != nil {
				require.Nil(t, tc.res)
				AssertEqualAltWriteError(t, *tc.err, tc.alt, err)
				return
			}

			require.NoError(t, err)
			res.UpsertedID = nil

			expectedRes := notModified
			if tc.stat != nil {
				expectedRes = tc.stat
			}
			assert.Equal(t, expectedRes, res)

			var actual bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.res, actual)
		})
	}
}
