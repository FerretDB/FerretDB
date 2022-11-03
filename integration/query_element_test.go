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

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryElementExists(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "empty-array"}, {"empty-array", []any{}}},
		bson.D{{"_id", "nan"}, {"nan", math.NaN()}},
		bson.D{{"_id", "null"}, {"null", nil}},
		bson.D{{"_id", "string"}, {"v", "12"}},
		bson.D{{"_id", "two-fields"}, {"v", "12"}, {"field", 42}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
	}{
		"Exists": {
			filter:      bson.D{{"_id", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"empty-array", "nan", "null", "string", "two-fields"},
		},
		"ExistsSecondField": {
			filter:      bson.D{{"field", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"two-fields"},
		},
		"NullField": {
			filter:      bson.D{{"null", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"null"},
		},
		"NonExistentField": {
			filter:      bson.D{{"non-existent", bson.D{{"$exists", true}}}},
			expectedIDs: []any{},
		},
		"EmptyArray": {
			filter:      bson.D{{"empty-array", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"empty-array"},
		},
		"NanField": {
			filter:      bson.D{{"nan", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"nan"},
		},
		"ExistsFalse": {
			filter:      bson.D{{"field", bson.D{{"$exists", false}}}},
			expectedIDs: []any{"empty-array", "nan", "null", "string"},
		},
		"NonBool": {
			filter:      bson.D{{"_id", bson.D{{"$exists", -123}}}},
			expectedIDs: []any{"empty-array", "nan", "null", "string", "two-fields"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := collection.Database()
			cursor, err := db.RunCommandCursor(ctx, bson.D{
				{"find", collection.Name()},
				{"filter", tc.filter},
			})
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

func TestQueryElementType(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	// TODO: add cases for "decimal" when it would be added.
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		v           any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"Document": {
			v:           "object",
			expectedIDs: []any{"document", "document-composite", "document-composite-reverse", "document-empty", "document-null"},
		},
		"Array": {
			v: "array",
			expectedIDs: []any{
				"array", "array-empty",
				"array-null", "array-three", "array-three-reverse", "array-two",
			},
		},
		"Double": {
			v: "double",
			expectedIDs: []any{
				"array-two", "double", "double-big", "double-max", "double-nan",
				"double-negative-zero", "double-smallest", "double-whole", "double-zero",
			},
		},
		"String": {
			v:           "string",
			expectedIDs: []any{"array-three", "array-three-reverse", "string", "string-double", "string-empty", "string-whole"},
		},
		"Binary": {
			v:           "binData",
			expectedIDs: []any{"binary", "binary-empty"},
		},
		"ObjectID": {
			v:           "objectId",
			expectedIDs: []any{"objectid", "objectid-empty"},
		},
		"Bool": {
			v:           "bool",
			expectedIDs: []any{"bool-false", "bool-true"},
		},
		"Datetime": {
			v:           "date",
			expectedIDs: []any{"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min"},
		},
		"Null": {
			v: "null",
			expectedIDs: []any{
				"array-null", "array-three",
				"array-three-reverse", "null",
			},
		},
		"Regex": {
			v:           "regex",
			expectedIDs: []any{"regex", "regex-empty"},
		},
		"Integer": {
			v:           "int",
			expectedIDs: []any{"array", "array-three", "array-three-reverse", "int32", "int32-max", "int32-min", "int32-zero"},
		},
		"Timestamp": {
			v:           "timestamp",
			expectedIDs: []any{"timestamp", "timestamp-i"},
		},
		"Long": {
			v:           "long",
			expectedIDs: []any{"int64", "int64-big", "int64-max", "int64-min", "int64-zero"},
		},

		"Number": {
			v: "number",
			expectedIDs: []any{
				"array", "array-three", "array-three-reverse", "array-two",
				"double", "double-big", "double-max", "double-nan",
				"double-negative-zero", "double-smallest", "double-whole", "double-zero",
				"int32", "int32-max", "int32-min", "int32-zero",
				"int64", "int64-big", "int64-max", "int64-min", "int64-zero",
			},
		},

		"BadTypeCode": {
			v: 42,
			err: &mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: 42",
				Name:    "BadValue",
			},
		},
		"BadTypeName": {
			v: "float",
			err: &mongo.CommandError{
				Code:    2,
				Message: "Unknown type name alias: float",
				Name:    "BadValue",
			},
		},
		"IntegerNumericalInput": {
			v:           16,
			expectedIDs: []any{"array", "array-three", "array-three-reverse", "int32", "int32-max", "int32-min", "int32-zero"},
		},
		"FloatTypeCode": {
			v:           16.0,
			expectedIDs: []any{"array", "array-three", "array-three-reverse", "int32", "int32-max", "int32-min", "int32-zero"},
		},
		"TypeArrayAliases": {
			v:           []any{"bool", "binData"},
			expectedIDs: []any{"binary", "binary-empty", "bool-false", "bool-true"},
		},
		"TypeArrayCodes": {
			v:           []any{5, 8},
			expectedIDs: []any{"binary", "binary-empty", "bool-false", "bool-true"},
		},
		"TypeArrayAliasAndCodeMixed": {
			v:           []any{5, "binData"},
			expectedIDs: []any{"binary", "binary-empty"},
		},
		"TypeArrayBadValue": {
			v: []any{"binData", -123},
			err: &mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: -123",
				Name:    "BadValue",
			},
		},
		"TypeArrayBadValueNan": {
			v: []any{"binData", math.NaN()},
			err: &mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: nan",
				Name:    "BadValue",
			},
		},
		"TypeArrayBadValuePlusInf": {
			v: []any{"binData", math.Inf(+1)},
			err: &mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: inf",
				Name:    "BadValue",
			},
		},
		"TypeArrayBadValueMinusInf": {
			v: []any{"binData", math.Inf(-1)},
			err: &mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: -inf",
				Name:    "BadValue",
			},
		},
		"TypeArrayBadValueNegativeFloat": {
			v: []any{"binData", -1.123},
			err: &mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: -1.123",
				Name:    "BadValue",
			},
		},
		"TypeArrayFloat": {
			v:           []any{5, 8.0},
			expectedIDs: []any{"binary", "binary-empty", "bool-false", "bool-true"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"v", bson.D{{"$type", tc.v}}}}
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
