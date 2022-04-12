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
)

func TestExistsOperator(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "empty-array"}, {"empty-array", []any{}}},
		bson.D{{"_id", "nan"}, {"nan", math.NaN()}},
		bson.D{{"_id", "null"}, {"null", nil}},
		bson.D{{"_id", "string"}, {"value", "12"}},
		bson.D{{"_id", "two-fields"}, {"value", "12"}, {"field", 42}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		q           bson.D
		expectedIDs []any
	}{
		"Exists": {
			q:           bson.D{{"_id", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"empty-array", "nan", "null", "string", "two-fields"},
		},
		"ExistsSecondField": {
			q:           bson.D{{"field", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"two-fields"},
		},
		"NullField": {
			q:           bson.D{{"null", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"null"},
		},
		"NonExistentField": {
			q:           bson.D{{"non-existent", bson.D{{"$exists", true}}}},
			expectedIDs: []any{},
		},
		"EmptyArray": {
			q:           bson.D{{"empty-array", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"empty-array"},
		},
		"NanField": {
			q:           bson.D{{"nan", bson.D{{"$exists", true}}}},
			expectedIDs: []any{"nan"},
		},
		"ExistsFalse": {
			q:           bson.D{{"field", bson.D{{"$exists", false}}}},
			expectedIDs: []any{"empty-array", "nan", "null", "string"},
		},
		"NonBool": {
			q:           bson.D{{"_id", bson.D{{"$exists", -123}}}},
			expectedIDs: []any{"empty-array", "nan", "null", "string", "two-fields"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := collection.Database()
			cursor, err := db.RunCommandCursor(ctx, bson.D{
				{"find", collection.Name()},
				{"filter", tc.q},
			})
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}
