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
)

func TestSetOnInsert(t *testing.T) {
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
		"doc": {
			filter:      bson.D{{"_id", "doc"}},
			setOnInsert: bson.D{{"value", bson.D{}}},
			res:         bson.D{{"_id", "doc"}, {"value", bson.D{}}},
		},
		"array": {
			filter:      bson.D{{"_id", "array"}},
			setOnInsert: bson.D{{"value", bson.A{}}},
			res:         bson.D{{"_id", "array"}, {"value", bson.A{}}},
		},
		"double": {
			filter:      bson.D{{"_id", "double"}},
			setOnInsert: bson.D{{"value", 43.13}},
			res:         bson.D{{"_id", "double"}, {"value", 43.13}},
		},
		"NaN": {
			filter:      bson.D{{"_id", "double-nan"}},
			setOnInsert: bson.D{{"value", math.NaN()}},
			res:         bson.D{{"_id", "double-nan"}, {"value", math.NaN()}},
		},
		"string": {
			filter:      bson.D{{"_id", "string"}},
			setOnInsert: bson.D{{"value", "abcd"}},
			res:         bson.D{{"_id", "string"}, {"value", "abcd"}},
		},
		"nil": {
			filter:      bson.D{{"_id", "nil"}},
			setOnInsert: bson.D{{"value", nil}},
			res:         bson.D{{"_id", "nil"}, {"value", nil}},
		},
		"empty-doc": {
			filter:      bson.D{{"_id", "doc"}},
			setOnInsert: bson.D{},
			res:         bson.D{{"_id", "doc"}},
		},
		"empty-array": {
			filter:      bson.D{{"_id", "array"}},
			setOnInsert: bson.A{},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: []}",
			},
			alt: "Modifiers operate on fields but we found type array instead. " +
				"For example: {$mod: {<field>: ...}} not {$setOnInsert: array}",
		},
		"double-double": {
			filter:      bson.D{{"_id", "double"}},
			setOnInsert: 43.13,
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: 43.13}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$setOnInsert: double}",
		},
		"err-NaN": {
			filter:      bson.D{{"_id", "double-nan"}},
			setOnInsert: math.NaN(),
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: nan.0}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$setOnInsert: double}",
		},
		"err-string": {
			filter:      bson.D{{"_id", "string"}},
			setOnInsert: "any string",
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: \"any string\"}",
			},
			alt: "Modifiers operate on fields but we found type string instead. " +
				"For example: {$mod: {<field>: ...}} not {$setOnInsert: string}",
		},
		"err-nil": {
			filter:      bson.D{{"_id", "nil"}},
			setOnInsert: nil,
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type null instead. " +
					"For example: {$mod: {<field>: ...}} not {$setOnInsert: null}",
			},
			alt: "Modifiers operate on fields but we found type null instead. " +
				"For example: {$mod: {<field>: ...}} not {$setOnInsert: null}",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var err error
			ctx, collection := setup(t)

			opts := options.Update().SetUpsert(true)
			var res *mongo.UpdateResult
			res, err = collection.UpdateOne(ctx, tc.filter, bson.D{{"$setOnInsert", tc.setOnInsert}}, opts)
			if tc.err != nil {
				if !AssertEqualWriteError(t, tc.err, tc.alt, err) {
					t.Logf("%[1]T %[1]v", err)
					t.FailNow()
				}
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
			if !AssertEqualDocuments(t, tc.res, actual) {
				t.FailNow()
			}
		})
	}
}

func TestSetOnInsertMore(t *testing.T) {
	t.Parallel()

	notModified := &mongo.UpdateResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: 1,
	}

	for name, tc := range map[string]struct {
		filter bson.D
		query  bson.D
		stat   *mongo.UpdateResult
		res    bson.D
		err    *mongo.WriteError
		alt    string
	}{
		"tandem-set-setoninsert": {
			filter: bson.D{{"_id", "test"}},
			query: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$setOnInsert", bson.D{{"value", math.NaN()}}},
			},
			res: bson.D{{"_id", "test"}, {"foo", int32(12)}, {"value", math.NaN()}},
		},
		"trio": {
			filter: bson.D{{"_id", "test"}},
			query: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$inc", bson.D{{"foo", int32(1)}}},
				{"$setOnInsert", bson.D{{"value", math.NaN()}}},
			},
			err: &mongo.WriteError{
				Code:    40,
				Message: "Updating the path 'foo' would create a conflict at 'foo'",
			},
		},
		"unknown-operator": {
			filter: bson.D{{"_id", "test"}},
			query: bson.D{
				{"$foo", bson.D{{"foo", int32(1)}}},
			},
			err: &mongo.WriteError{
				Code:    9,
				Message: "Unknown modifier: $foo. Expected a valid update modifier or pipeline-style update specified as an array",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var err error
			ctx, collection := setup(t)

			opts := options.Update().SetUpsert(true)
			var res *mongo.UpdateResult
			res, err = collection.UpdateOne(ctx, tc.filter, tc.query, opts)
			if tc.err != nil {
				if !AssertEqualWriteError(t, tc.err, tc.alt, err) {
					t.Logf("%[1]T %[1]v", err)
					t.FailNow()
				}
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
			if !AssertEqualDocuments(t, tc.res, actual) {
				t.FailNow()
			}
		})
	}
}
