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

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestQueryElementExists(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "empty-array"}, {"empty-array", []any{}}},
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
			expectedIDs: []any{"empty-array", "null", "string", "two-fields"},
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
		"ExistsFalse": {
			filter:      bson.D{{"field", bson.D{{"$exists", false}}}},
			expectedIDs: []any{"empty-array", "null", "string"},
		},
		"NonBool": {
			filter:      bson.D{{"_id", bson.D{{"$exists", -123}}}},
			expectedIDs: []any{"empty-array", "null", "string", "two-fields"},
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
