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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryLogical(t *testing.T) {
	t.Parallel()

	// Use shared setup because find queries can't modify data.
	// TODO Use read-only user. https://github.com/FerretDB/FerretDB/issues/1025
	ctx, collection := setup.Setup(t, shareddata.Doubles)

	testCases := map[string]struct {
		filter      bson.D
		skip        string
		expectedIDs []any
		expectedErr mongo.CommandError
	}{
		// $and
		"AndZero": {
			filter: bson.D{{
				"$and", bson.A{},
			}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$and/$or/$nor must be a nonempty array",
			},
		},
		"AndOne": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
				},
			}},
			expectedIDs: []any{
				"double-smallest", "double-whole", "double", "double-big", "double-max",
			},
		},
		"AndTwo": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
					bson.D{{"v", bson.D{{"$lt", int64(42)}}}},
				},
			}},
			expectedIDs: []any{
				"double-smallest",
			},
		},
		"AndOr": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
					bson.D{{"$or", bson.A{
						bson.D{{"v", bson.D{{"$lt", int64(42)}}}},
						bson.D{{"v", bson.D{{"$lte", 42.13}}}},
					}}},
				},
			}},
			expectedIDs: []any{
				"double-smallest", "double-whole", "double",
			},
		},
		"AndAnd": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"$and", bson.A{
						bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
						bson.D{{"v", bson.D{{"$lte", 42.13}}}},
					}}},
					bson.D{{"v", bson.D{{"$type", "double"}}}},
				},
			}},
			expectedIDs: []any{
				"double-smallest", "double-whole", "double",
			},
		},
		"AndBadInput": {
			filter: bson.D{{"$and", nil}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$and must be an array",
			},
		},
		"AndBadValue": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
					true,
				},
			}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$or/$and/$nor entries need to be full objects",
			},
		},

		// $or
		"OrZero": {
			filter: bson.D{{
				"$or", bson.A{},
			}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$and/$or/$nor must be a nonempty array",
			},
		},
		"OrOne": {
			filter: bson.D{{
				"$or", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
				},
			}},
			expectedIDs: []any{},
		},
		"OrTwo": {
			filter: bson.D{{
				"$or", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					bson.D{{"v", bson.D{{"$gt", int64(42)}}}},
				},
			}},
			expectedIDs: []any{
				"double", "double-big", "double-max",
			},
		},
		"OrAnd": {
			filter: bson.D{{
				"$or", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					bson.D{{"$and", bson.A{
						bson.D{{"v", bson.D{{"$gt", int64(42)}}}},
						bson.D{{"v", bson.D{{"$lte", 42.13}}}},
					}}},
				},
			}},
			expectedIDs: []any{
				"double",
			},
		},
		"OrBadInput": {
			filter: bson.D{{"$or", nil}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$or must be an array",
			},
		},
		"OrBadValue": {
			filter: bson.D{{
				"$or", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					true,
				},
			}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$or/$and/$nor entries need to be full objects",
			},
		},

		// $nor
		"NorZero": {
			filter: bson.D{{
				"$nor", bson.A{},
			}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$and/$or/$nor must be a nonempty array",
			},
		},
		"NorOne": {
			filter: bson.D{{
				"$nor", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
				},
			}},
			expectedIDs: []any{
				"double-null", "double-zero", "double-smallest", "double-whole", "double", "double-big", "double-max",
			},
		},
		"NorTwo": {
			filter: bson.D{{
				"$nor", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					bson.D{{"v", bson.D{{"$gt", int64(42)}}}},
				},
			}},
			expectedIDs: []any{
				"double-null", "double-zero", "double-smallest", "double-whole",
			},
		},
		"NorBadInput": {
			filter: bson.D{{"$nor", nil}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$nor must be an array",
			},
		},
		"NorBadValue": {
			filter: bson.D{{
				"$nor", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					true,
				},
			}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$or/$and/$nor entries need to be full objects",
			},
		},

		// $not
		"Not": {
			filter: bson.D{{
				"v", bson.D{{"$not", bson.D{{"$eq", int64(42)}}}},
			}},
			expectedIDs: []any{
				"double-null", "double-zero", "double-smallest", "double", "double-big", "double-max",
			},
		},
		"NotNull": {
			filter: bson.D{{
				"v", bson.D{{"$not", nil}},
			}},
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "$not needs a regex or a document",
			},
		},
		"NotEqNull": {
			filter: bson.D{{
				"v", bson.D{{"$not", bson.D{{"$eq", nil}}}},
			}},
			expectedIDs: []any{
				"double-zero", "double-smallest", "double-whole", "double", "double-big", "double-max",
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			filter := tc.filter
			require.NotNil(t, filter)

			sort := bson.D{{"v", 1}}
			opts := options.Find().SetSort(sort)

			cursor, err := collection.Find(ctx, filter, opts)
			if err != nil {
				assert.Nil(t, tc.expectedIDs)
				AssertEqualError(t, tc.expectedErr, err)
				return
			}

			require.Empty(t, tc.expectedErr)

			actualIDs := CollectIDs(t, FetchAll(t, ctx, cursor))
			assert.Equal(t, tc.expectedIDs, actualIDs)
		})
	}
}
