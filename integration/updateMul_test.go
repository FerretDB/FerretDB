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
)

func TestUpdateMul(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		filter     bson.D
		update     bson.D
		expected   map[string]any
		err        *mongo.WriteError
		altMessage string
	}{
		"OneI": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$mul", bson.D{{"value", 10}}}},
			expected: map[string]any{
				"_id":   "1",
				"value": 100,
			},
		},

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
			t.Parallel()
			ctx, collection := setup(t)

			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", "1"}, {"value", 10}},
				bson.D{{"_id", "2"}, {"value", 1}},
			})
			require.NoError(t, err)

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
