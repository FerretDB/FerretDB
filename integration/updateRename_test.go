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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestUpdateRename(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		filter     bson.D
		update     bson.D
		expected   map[string]any
		err        *mongo.WriteError
		altMessage string
	}{
		"OneField": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", "nickname"}}}},
			expected: map[string]any{
				"_id":      "1",
				"nickname": "alex",
			},
		},
		"ManyField": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", "nickname"}, {"phone", "mobile"}}}},
			expected: map[string]any{
				"_id":      "1",
				"nickname": "alex",
				"mobile":   "9012345678",
			},
		},
		"FieldInt": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", int64(1)}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `The 'to' field for $rename must be a string: name: 1`,
			},
			altMessage: `The 'to' field for $rename must be a string: name: long`,
		},
		"FieldDocument": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", primitive.D{}}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `The 'to' field for $rename must be a string: name: {}`,
			},
			altMessage: `The 'to' field for $rename must be a string: name: object`,
		},

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

		"FieldEmptyArray": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", primitive.A{}}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `The 'to' field for $rename must be a string: name: []`,
			},
			altMessage: `The 'to' field for $rename must be a string: name: array`,
		},
		"FieldArray": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", primitive.A{"nickname", "alias"}}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `The 'to' field for $rename must be a string: name: [ "nickname", "alias" ]`,
			},
			altMessage: `The 'to' field for $rename must be a string: name: array`,
		},
		"FieldNaN": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", math.NaN()}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `The 'to' field for $rename must be a string: name: nan.0`,
			},
			altMessage: `The 'to' field for $rename must be a string: name: double`,
		},
		"FieldNil": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", nil}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `The 'to' field for $rename must be a string: name: null`,
			},
			altMessage: `The 'to' field for $rename must be a string: name: null`,
		},
		"RenameString": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", "string"}},
			err: &mongo.WriteError{
				Code: 9,
				Message: `Modifiers operate on fields but we found type string instead.` +
					` For example: {$mod: {<field>: ...}} not {$rename: "string"}`,
			},
			altMessage: `Modifiers operate on fields but we found another type instead`,
		},
		"RenameNil": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", nil}},
			err: &mongo.WriteError{
				Code: 9,
				Message: `Modifiers operate on fields but we found type null instead.` +
					` For example: {$mod: {<field>: ...}} not {$rename: null}`,
			},
			altMessage: `Modifiers operate on fields but we found another type instead`,
		},
		"RenameDoc": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", primitive.D{}}},
		},

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
				bson.D{{"_id", "1"}, {"name", "alex"}, {"phone", "9012345678"}},
				bson.D{{"_id", "2"}, {"name", "bob"}},
			})
			require.NoError(t, err)

			_, err = collection.UpdateOne(ctx, tc.filter, tc.update)
			if tc.err != nil {
				require.NotNil(t, tc.err)
				AssertEqualAltWriteError(t, *tc.err, tc.altMessage, err)
				return
			}
			require.NoError(t, err)

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
