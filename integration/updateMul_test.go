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
)

func TestUpdateMul(t *testing.T) {
	//	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "zero_zero"}, {"value", 0}},
		bson.D{{"_id", "zero_int32"}, {"value", 0}},
		bson.D{{"_id", "zero_int64"}, {"value", 0}},
		bson.D{{"_id", "zero_float64"}, {"value", 0}},
		bson.D{{"_id", "int32_zero"}, {"value", int32(1000)}},
		bson.D{{"_id", "int64_zero"}, {"value", int64(1000)}},
		bson.D{{"_id", "float64_zero"}, {"value", float64(1000)}},

		bson.D{{"_id", "int32_int32"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_int64"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_float64"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_maxInt32"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_maxInt64"}, {"value", int32(2)}},

		bson.D{{"_id", "int64_int32"}, {"value", int64(10)}},
		bson.D{{"_id", "int64_int64"}, {"value", int64(10)}},
		bson.D{{"_id", "int64_float64"}, {"value", int64(10)}},
		bson.D{{"_id", "int64_maxInt32"}, {"value", int64(2)}},
		bson.D{{"_id", "int64_maxInt64"}, {"value", int64(2)}},

		bson.D{{"_id", "maxInt32_maxInt32"}, {"value", math.MaxInt32}},
		bson.D{{"_id", "maxInt64_maxInt64"}, {"value", math.MaxInt64}},

		bson.D{{"_id", "maxInt64_int32"}, {"value", math.MaxInt64}},
		bson.D{{"_id", "maxInt64_int64"}, {"value", math.MaxInt64}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter     bson.D
		update     bson.D
		expected   map[string]any
		err        *mongo.WriteError
		altMessage string
	}{
		"Zero_Zero": {
			filter:   bson.D{{"_id", "zero_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: map[string]any{"_id": "zero_zero", "value": int32(0)},
		},
		"Zero_Int32": {
			filter:   bson.D{{"_id", "zero_int32"}},
			update:   bson.D{{"$mul", bson.D{{"value", int32(10)}}}},
			expected: map[string]any{"_id": "zero_int32", "value": int32(0)},
		},
		"Zero_Int64": {
			filter:   bson.D{{"_id", "zero_int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(10)}}}},
			expected: map[string]any{"_id": "zero_int64", "value": int64(0)},
		},
		"Zero_Float64": {
			filter:   bson.D{{"_id", "zero_float64"}},
			update:   bson.D{{"$mul", bson.D{{"value", float64(10)}}}},
			expected: map[string]any{"_id": "zero_float64", "value": float64(0)},
		},
		"Int32_Zero": {
			filter:   bson.D{{"_id", "int32_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: map[string]any{"_id": "int32_zero", "value": int32(0)},
		},
		"Int64_Zero": {
			filter:   bson.D{{"_id", "int64_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: map[string]any{"_id": "int64_zero", "value": int64(0)},
		},
		"Float64_Zero": {
			filter:   bson.D{{"_id", "float64_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: map[string]any{"_id": "float64_zero", "value": float64(0)},
		},
		"Int32_Int32": {
			filter:   bson.D{{"_id", "int32_int32"}},
			update:   bson.D{{"$mul", bson.D{{"value", int32(10)}}}},
			expected: map[string]any{"_id": "int32_int32", "value": int32(100)},
		},
		"Int32_Int64": {
			filter:   bson.D{{"_id", "int32_int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(10)}}}},
			expected: map[string]any{"_id": "int32_int64", "value": int64(100)},
		},
		"Int32_Float64": {
			filter:   bson.D{{"_id", "int32_float64"}},
			update:   bson.D{{"$mul", bson.D{{"value", float64(10)}}}},
			expected: map[string]any{"_id": "int32_float64", "value": float64(100)},
		},
		"Int32_MaxInt32": {
			filter:   bson.D{{"_id", "int32_maxInt32"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
			expected: map[string]any{"_id": "int32_maxInt32", "value": int64(21474836470)},
		},
		"Int32_MaxInt64": {
			filter:   bson.D{{"_id", "int32_maxInt64"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
			expected: map[string]any{"_id": "int32_maxInt64", "value": int32(2)},
		},
		"Int64_Int32": {
			filter:   bson.D{{"_id", "int64_int32"}},
			update:   bson.D{{"$mul", bson.D{{"value", int32(10)}}}},
			expected: map[string]any{"_id": "int64_int32", "value": int64(100)},
		},
		"Int64_Int64": {
			filter:   bson.D{{"_id", "int64_int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(10)}}}},
			expected: map[string]any{"_id": "int64_int64", "value": int64(100)},
		},
		"Int64_Float64": {
			filter:   bson.D{{"_id", "int64_float64"}},
			update:   bson.D{{"$mul", bson.D{{"value", float64(10)}}}},
			expected: map[string]any{"_id": "int64_float64", "value": float64(100)},
		},
		"Int64_MaxInt32": {
			filter:   bson.D{{"_id", "int64_maxInt32"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
			expected: map[string]any{"_id": "int64_maxInt32", "value": int64(4294967294)},
		},
		"Int64_MaxInt64": {
			filter:   bson.D{{"_id", "int64_maxInt64"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
			expected: map[string]any{"_id": "int64_maxInt64", "value": int64(2)},
		},
		"MaxInt64_MaxInt64": {
			filter:   bson.D{{"_id", "maxInt64_maxInt64"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
			expected: map[string]any{"_id": "maxInt64_maxInt64", "value": int64(math.MaxInt64)},
		},
		"MaxInt64_Int32": {
			filter:   bson.D{{"_id", "maxInt64_int32"}},
			update:   bson.D{{"$mul", bson.D{{"value", int32(2)}}}},
			expected: map[string]any{"_id": "maxInt64_int32", "value": int64(math.MaxInt64)},
		},
		"MaxInt64_Int64": {
			filter:   bson.D{{"_id", "maxInt64_int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(2)}}}},
			expected: map[string]any{"_id": "maxInt64_int64", "value": int64(math.MaxInt64)},
		},

		// "Float64_Zero": {
		// 	filter:   bson.D{{"_id", "float64_zero"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
		// 	expected: map[string]any{"_id": "float64_zero", "value": float64(0)},
		// },

		// "FieldDoc": {
		// 	filter: bson.D{{"_id", "1"}},
		// 	update: bson.D{{"$rename", bson.D{{"name", primitive.D{}}}}},
		// 	err: &mongo.WriteError{
		// 		Code:    2,
		// 		Message: `The 'to' field for $rename must be a string: name: {}`,
		// 	},
		// 	altMessage: `The 'to' field for $rename must be a string: name: object`,
		// },

		// TODO issues #673
		/* "FieldDoc": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", bson.D{{}}}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `The 'to' field for $rename must be a string: name: { : null }`,
			},
			altMessage: `The 'to' field for $rename must be a string: name: object`,
		}, */

		// TODO issues #673
		/* "RenameDoc_1": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{}}}},
			err: &mongo.WriteError{
				Code:    56,
				Message: `An empty update path is not valid.`,
			},
			altMessage: `An empty update path is not valid.`,
		}, */
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			//	t.Parallel()

			_, err = collection.UpdateOne(ctx, tc.filter, tc.update)
			if tc.err != nil {
				require.NotNil(t, tc.err)
				AssertEqualAltWriteError(t, *tc.err, tc.altMessage, err)
				return
			}

			var actual bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()
			k := CollectKeys(t, actual)

			for key, item := range tc.expected {
				assert.Contains(t, k, key)
				assert.Equal(t, m[key], item)
			}
		})
	}
}
