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

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryLogicalAnd(t *testing.T) {
	t.Parallel()
	ctx, collection := setupWithOpts(t, &setupOpts{
		providers: []shareddata.Provider{shareddata.Scalars},
	})

	for name, tc := range map[string]struct {
		q           bson.D
		expectedIDs []any
		err         error
	}{
		"And": {
			q: bson.D{{
				"$and",
				bson.A{
					bson.D{{"value", bson.D{{"$gt", 0}}}},
					bson.D{{"value", bson.D{{"$lte", 42}}}},
				},
			}},
			expectedIDs: []any{"double-smallest", "double-whole", "int32", "int64"},
		},
		"BadInput": {
			q: bson.D{{"$and", nil}},
			err: mongo.CommandError{
				Code:    2,
				Message: "$and must be an array",
				Name:    "BadValue",
			},
		},
		"BadExpressionValue": {
			q: bson.D{{
				"$and",
				bson.A{
					bson.D{{"value", bson.D{{"$gt", 0}}}},
					nil,
				},
			}},
			err: mongo.CommandError{
				Code:    2,
				Message: "$or/$and/$nor entries need to be full objects",
				Name:    "BadValue",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.q)
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				assertEqualError(t, tc.err.(mongo.CommandError), err)
				return
			}
			require.NoError(t, err)
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}

func TestQueryLogicalOr(t *testing.T) {
	t.Parallel()
	ctx, collection := setupWithOpts(t, &setupOpts{
		providers: []shareddata.Provider{shareddata.Scalars},
	})

	for name, tc := range map[string]struct {
		q           bson.D
		expectedIDs []any
		err         error
	}{
		"Or": {
			q: bson.D{{
				"$or",
				bson.A{
					bson.D{{"value", bson.D{{"$lt", 0}}}},
					bson.D{{"value", bson.D{{"$lt", 42}}}},
				},
			}},
			expectedIDs: []any{
				"double-negative-infinity", "double-negative-zero",
				"double-smallest", "double-zero",
				"int32-min", "int32-zero", "int64-min", "int64-zero",
			},
		},
		"BadInput": {
			q: bson.D{{"$or", nil}},
			err: mongo.CommandError{
				Code:    2,
				Message: "$or must be an array",
				Name:    "BadValue",
			},
		},
		"BadExpressionValue": {
			q: bson.D{{
				"$or",
				bson.A{
					bson.D{{"value", bson.D{{"$gt", 0}}}},
					nil,
				},
			}},
			err: mongo.CommandError{
				Code:    2,
				Message: "$or/$and/$nor entries need to be full objects",
				Name:    "BadValue",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.q)
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				assertEqualError(t, tc.err.(mongo.CommandError), err)
				return
			}
			require.NoError(t, err)
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}

func TestQueryLogicalNor(t *testing.T) {
	t.Parallel()
	ctx, collection := setupWithOpts(t, &setupOpts{
		providers: []shareddata.Provider{shareddata.Scalars},
	})

	for name, tc := range map[string]struct {
		q           bson.D
		expectedIDs []any
		err         mongo.CommandError
	}{
		"Nor": {
			q: bson.D{{
				"$nor",
				bson.A{
					bson.D{{"value", bson.D{{"$gt", 0}}}},
					bson.D{{"value", bson.D{{"$gt", 42}}}},
				},
			}},
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
			q: bson.D{{"$nor", nil}},
			err: mongo.CommandError{
				Code:    2,
				Message: "$nor must be an array",
				Name:    "BadValue",
			},
		},
		"BadExpressionValue": {
			q: bson.D{{
				"$nor",
				bson.A{
					bson.D{{"value", bson.D{{"$gt", 0}}}},
					nil,
				},
			}},
			err: mongo.CommandError{
				Code:    2,
				Message: "$or/$and/$nor entries need to be full objects",
				Name:    "BadValue",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.q, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err.Code != 0 {
				require.Nil(t, tc.expectedIDs)
				assertEqualError(t, tc.err, err)
				return
			}
			require.NoError(t, err)
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}
func TestQueryLogicalNot(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         mongo.CommandError
	}{
		"IDNull": {
			filter: bson.D{{"_id", bson.D{{"$not", nil}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$not needs a regex or a document",
			},
		},
		"ValueNull": {
			filter: bson.D{{"value", bson.D{{"$not", bson.D{{"$eq", nil}}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$not needs a regex or a document",
			},
		},
		"NoSuchFieldNull": {
			filter: bson.D{{"no-such-field", bson.D{{"$not", 0}}}},
			expectedIDs: []any{
				"array", "array-empty", "array-three",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-empty",
				"double", "double-max", "double-nan", "double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string", "string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},

		"ValueNumber": {
			filter:      bson.D{{"value", bson.D{{"$not", 42}}}},
			expectedIDs: []any{"array", "array-three", "double-whole", "int32", "int64"},
		},

		"ValueRegex": {
			filter: bson.D{{"value", bson.D{{"$not", primitive.Regex{Pattern: "^fo"}}}}},
			expectedIDs: []any{
				"array", "array-empty",
				"binary", "binary-empty",
				"bool-false", "bool-true",
				"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min",
				"document", "document-empty",
				"double", "double-max", "double-nan", "double-negative-infinity", "double-negative-zero",
				"double-positive-infinity", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-max", "int64-min", "int64-zero",
				"null",
				"objectid", "objectid-empty",
				"regex", "regex-empty",
				"string-double", "string-empty", "string-whole",
				"timestamp", "timestamp-i",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err.Code != 0 {
				require.Nil(t, tc.expectedIDs)
				assertEqualError(t, tc.err, err)
				return
			}
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}
