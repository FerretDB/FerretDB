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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryLogicalOr(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	for name, tc := range map[string]struct {
		filter      any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"Or": {
			filter: bson.A{
				bson.D{{"v", bson.D{{"$lt", 0}}}},
				bson.D{{"v", bson.D{{"$lt", 42}}}},
			},
			expectedIDs: []any{
				"double-negative-infinity", "double-negative-zero",
				"double-smallest", "double-zero",
				"int32-min", "int32-zero", "int64-min", "int64-zero",
			},
		},
		"OrAnd": {
			filter: bson.A{
				bson.D{{"$and", bson.A{
					bson.D{},
					bson.D{},
				}}},
				bson.D{},
			},
			expectedIDs: []any{
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"double", "double-big", "double-max", "double-nan", "double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-big", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},
		"BadInput": {
			filter: nil,
			err: &mongo.CommandError{
				Code:    2,
				Message: "$or must be an array",
				Name:    "BadValue",
			},
		},
		"BadExpressionValue": {
			filter: bson.A{
				bson.D{{"v", bson.D{{"$gt", 0}}}},
				nil,
			},
			err: &mongo.CommandError{
				Code:    2,
				Message: "$or/$and/$nor entries need to be full objects",
				Name:    "BadValue",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"$or", tc.filter}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

func TestQueryLogicalNor(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	for name, tc := range map[string]struct {
		filter      any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"Nor": {
			filter: bson.A{
				bson.D{{"v", bson.D{{"$gt", 0}}}},
				bson.D{{"v", bson.D{{"$gt", 42}}}},
			},
			expectedIDs: []any{
				"binary", "binary-empty", "bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"double-nan", "double-negative-infinity", "double-negative-zero", "double-zero",
				"int32-min", "int32-zero", "int64-min", "int64-zero",
				"null", "objectid", "objectid-empty",
				"regex", "regex-empty", "string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},
		"BadInput": {
			filter: nil,
			err: &mongo.CommandError{
				Code:    2,
				Message: "$nor must be an array",
				Name:    "BadValue",
			},
		},
		"BadExpressionValue": {
			filter: bson.A{
				bson.D{{"v", bson.D{{"$gt", 0}}}},
				nil,
			},
			err: &mongo.CommandError{
				Code:    2,
				Message: "$or/$and/$nor entries need to be full objects",
				Name:    "BadValue",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"$nor", tc.filter}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

func TestQueryLogicalNot(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup.Setup(t, providers...)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"Not": {
			filter: bson.D{{"v", bson.D{{"$not", bson.D{{"$eq", 42}}}}}},
			expectedIDs: []any{
				"array-embedded", "array-empty", "array-empty-nested", "array-first-embedded", "array-last-embedded",
				"array-middle-embedded", "array-null", "array-two",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-composite", "document-composite-reverse", "document-empty", "document-null",
				"double", "double-big", "double-max", "double-nan",
				"double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-zero",
				"int32-max", "int32-min", "int32-zero",
				"int64-big", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},
		"IDNull": {
			filter: bson.D{{"_id", bson.D{{"$not", nil}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$not needs a regex or a document",
			},
		},
		"NotEqNull": {
			filter: bson.D{{"v", bson.D{{"$not", bson.D{{"$eq", nil}}}}}},
			expectedIDs: []any{
				"array", "array-embedded", "array-empty", "array-empty-nested", "array-two",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-composite", "document-composite-reverse", "document-empty", "document-null",
				"double", "double-big", "double-max", "double-nan", "double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-big", "int64-max", "int64-min", "int64-zero",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},
		"ValueRegex": {
			filter: bson.D{{"v", bson.D{{"$not", primitive.Regex{Pattern: "^fo"}}}}},
			expectedIDs: []any{
				"array", "array-embedded", "array-empty", "array-empty-nested", "array-first-embedded",
				"array-last-embedded", "array-middle-embedded", "array-null", "array-two",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-composite", "document-composite-reverse", "document-empty", "document-null",
				"double", "double-big", "double-max", "double-nan", "double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-big", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},
		"NoSuchFieldRegex": {
			filter: bson.D{{"no-such-field", bson.D{{"$not", primitive.Regex{Pattern: "/someregex/"}}}}},
			expectedIDs: []any{
				"array", "array-embedded", "array-empty", "array-empty-nested", "array-first-embedded", "array-last-embedded",
				"array-middle-embedded", "array-null", "array-three", "array-three-reverse", "array-two",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-composite", "document-composite-reverse", "document-empty", "document-null",
				"double", "double-big", "double-max", "double-nan",
				"double-negative-infinity", "double-negative-zero", "double-positive-infinity",
				"double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-big", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},
		"NestedNot": {
			filter:      bson.D{{"v", bson.D{{"$not", bson.D{{"$not", bson.D{{"$eq", 42}}}}}}}},
			expectedIDs: []any{"array", "array-three", "array-three-reverse", "double-whole", "int32", "int64"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}
