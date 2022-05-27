// Copyright 2021 FerretDB Set.
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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type testCase struct {
	id     string
	update bson.D
	result bson.D
	err    *mongo.WriteError
	stat   *mongo.UpdateResult
	alt    string
}

func TestSetOperatorOnString(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]testCase{
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
		},
		"String": {
			id:     "string",
			update: bson.D{{"$set", "string"}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: \"string\"}",
			},
			alt: "Modifiers operate on fields but we found type string instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: string}",
		},
		"Double": {
			id:     "string",
			update: bson.D{{"$set", float64(42.12345)}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: 42.12345}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: double}",
		},
		"NaN": {
			id:     "string",
			update: bson.D{{"$set", math.NaN()}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: nan.0}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: double}",
		},
		"Inf": {
			id:     "string",
			update: bson.D{{"$set", math.Inf(+1)}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: inf.0}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: double}",
		},
		"MinusInf": {
			id:     "string",
			update: bson.D{{"$set", math.Inf(-1)}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: -inf.0}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: double}",
		},
		"Array": {
			id:     "string",
			update: bson.D{{"$set", bson.A{}}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: []}",
			},
			alt: "Modifiers operate on fields but we found type array instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: array}",
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
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)
			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", "string"}, {"value", "foo"}},
			})
			require.NoError(t, err)

			res, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
			if tc.err != nil {
				if !AssertEqualAltWriteError(t, tc.err, tc.alt, err) {
					t.Logf("%[1]T %[1]v", err)
					t.FailNow()
				}
				return
			}
			require.Equal(t, tc.stat, res)

			var actual bson.D
			err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
			require.NoError(t, err)
			if !AssertEqualDocuments(t, tc.result, actual) {
				t.FailNow()
			}
		})
	}
}

func TestSetOperatorDoubleVal(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]testCase{
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
				bson.D{{"_id", "double"}, {"value", float64(0.0)}},
			})
			require.NoError(t, err)

			res, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
			if tc.err != nil {
				t.Log(err)
				if !AssertEqualAltWriteError(t, tc.err, tc.alt, err) {
					t.Log(err)
					t.FailNow()
				}
				return
			}
			require.Equal(t, tc.stat, res)

			var actual bson.D
			err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
			require.NoError(t, err)
			if !AssertEqualDocuments(t, tc.result, actual) {
				t.FailNow()
			}
		})
	}
}
