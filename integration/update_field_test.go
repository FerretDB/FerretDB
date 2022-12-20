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
		// TODO Make duration lower https://github.com/FerretDB/FerretDB/issues/1347
		maxDifference := time.Duration(4 * time.Minute)

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
		"ArrayNil": {
			// TODO remove https://github.com/FerretDB/FerretDB/issues/1662
			id:       "string",
			update:   bson.D{{"$set", bson.D{{"v", bson.A{nil}}}}},
			expected: bson.D{{"_id", "string"}, {"v", bson.A{nil}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"SetSameValueInt": {
			// TODO remove https://github.com/FerretDB/FerretDB/issues/1662
			id:       "int32",
			update:   bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			expected: bson.D{{"_id", "int32"}, {"v", int32(42)}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"DotNotationDocumentFieldExist": {
			// TODO remove https://github.com/FerretDB/FerretDB/issues/1661
			id:       "document-composite",
			update:   bson.D{{"$set", bson.D{{"v.foo", int32(1)}}}},
			expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(1)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"DotNotationArrayFieldExist": {
			// TODO remove https://github.com/FerretDB/FerretDB/issues/1661
			id:       "document-composite",
			update:   bson.D{{"$set", bson.D{{"v.array.0", int32(1)}}}},
			expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(1), "foo", nil}}}}},
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"DocumentDotNotationArrayFieldNotExist": {
			// TODO remove https://github.com/FerretDB/FerretDB/issues/1661
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
				{"$setOnInsert", bson.D{{"v", nil}}},
			},
			expected: bson.D{{"_id", "test"}, {"foo", int32(12)}, {"v", nil}},
		},
		"SetIncSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$inc", bson.D{{"foo", int32(1)}}},
				{"$setOnInsert", bson.D{{"v", nil}}},
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
			"PopDotNotation": {
				// TODO remove https://github.com/FerretDB/FerretDB/issues/1663
				id:       "document-composite",
				update:   bson.D{{"$pop", bson.D{{"v.array", 1}}}},
				expected: bson.D{{"_id", "document-composite"}, {"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo"}}}}},
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
			"PopDotNotationNonArray": {
				// TODO remove https://github.com/FerretDB/FerretDB/issues/1663
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
