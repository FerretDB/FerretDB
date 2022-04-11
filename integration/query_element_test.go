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

func TestTypeOperator(t *testing.T) {
	t.Parallel()
	// TODO: add this data types to collection "objectId", "decimal", "minKey", "maxKey"
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		v           any
		expectedIDs []any
		err         error
	}{
		"array": {
			v: "array",
			expectedIDs: []any{
				"array", "array-empty", "array-three",
			},
		},
		"bad-type-code": {
			v: 42,
			err: mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: 42",
				Name:    "BadValue",
			},
		},
		"bad-input-string": {
			v: "123",
			err: mongo.CommandError{
				Code:    2,
				Message: "Unknown type name alias: 123",
				Name:    "BadValue",
			},
		},
		"bad-type-name": {
			v: "float",
			err: mongo.CommandError{
				Code:    2,
				Message: "Unknown type name alias: float",
				Name:    "BadValue",
			},
		},
		"not-matched-type": {
			v:           "decimal",
			expectedIDs: []any{},
		},
		"integer": {
			v:           "int",
			expectedIDs: []any{"array", "array-three", "int32", "int32-max", "int32-min", "int32-zero"},
		},
		"integer-numerical-input": {
			v:           16,
			expectedIDs: []any{"array", "array-three", "int32", "int32-max", "int32-min", "int32-zero"},
		},
		"long": {
			v:           "long",
			expectedIDs: []any{"int64", "int64-max", "int64-min", "int64-zero"},
		},
		"regex": {
			v:           "regex",
			expectedIDs: []any{"regex", "regex-empty"},
		},
		"null": {
			v:           "null",
			expectedIDs: []any{"array-three", "null"},
		},
		"timestamp": {
			v:           "timestamp",
			expectedIDs: []any{"timestamp", "timestamp-i"},
		},
		"object": {
			v:           "object",
			expectedIDs: []any{"document", "document-empty"},
		},
		"double": {
			v: "double",
			expectedIDs: []any{
				"double", "double-max", "double-nan", "double-negative-infinity",
				"double-negative-zero", "double-positive-infinity",
				"double-smallest", "double-zero",
			},
		},
		"string": {
			v:           "string",
			expectedIDs: []any{"array-three", "string", "string-empty"},
		},
		"binData": {
			v:           "binData",
			expectedIDs: []any{"binary", "binary-empty"},
		},
		"bool": {
			v:           "bool",
			expectedIDs: []any{"bool-false", "bool-true"},
		},
		"datetime": {
			v:           "date",
			expectedIDs: []any{"datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min"},
		},
		"type-array-aliases": {
			v:           []any{"bool", "binData"},
			expectedIDs: []any{"binary", "binary-empty", "bool-false", "bool-true"},
		},
		"type-array-codes": {
			v:           []any{5, 8},
			expectedIDs: []any{"binary", "binary-empty", "bool-false", "bool-true"},
		},
		"type-array-alias-and-code-mixed": {
			v:           []any{5, "binData"},
			expectedIDs: []any{"binary", "binary-empty"},
		},
		"type-array-bad-value": {
			v: []any{"binData", -123},
			err: mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: -123",
				Name:    "BadValue",
			},
		},
		"type-array-bad-value-nan": {
			v: []any{"binData", math.NaN()},
			err: mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: nan",
				Name:    "BadValue",
			},
		},
		"type-array-bad-value-plus-inf": {
			v: []any{"binData", math.Inf(1)},
			err: mongo.CommandError{
				Code:    2,
				Message: "Invalid numerical type code: inf",
				Name:    "BadValue",
			},
		},
		"type-array-float": {
			v:           []any{5, 8.0},
			expectedIDs: []any{"binary", "binary-empty", "bool-false", "bool-true"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			q := bson.D{{"value", bson.D{{"$type", tc.v}}}}
			cursor, err := collection.Find(ctx, q, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				require.Equal(t, tc.err, err)
				return
			}
			require.NoError(t, err)
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}
