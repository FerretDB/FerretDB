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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// TestDeleteSimple checks simple cases of doc deletion.
func TestDeleteSimple(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		collection    string
		expectedCount int64
	}{
		"DeleteOne": {
			collection:    collection.Name(),
			expectedCount: 1,
		},
		"DeleteFromNonExistingCollection": {
			collection:    "doesnotexist",
			expectedCount: 0,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Database().Collection(tc.collection).DeleteOne(ctx, bson.D{})
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCount, cursor.DeletedCount)
		})
	}
}

// TestDeleteOrdered checks if oredered parameter works properly.
func TestDeleteOrdered(t *testing.T) {
	setup.SkipForTigris(t)
	t.Parallel()

	for name, tc := range map[string]struct {
		deletes            bson.A
		ordered            bool
		expectedRemovedIDs []string
	}{
		"test 1": {
			deletes: bson.A{
				bson.D{
					{"q", bson.D{{"_id", "string"}}},
				},
				bson.D{
					{"q", "fail"},
				},
			},
			ordered:            true,
			expectedRemovedIDs: []string{"string"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, collection := setup.Setup(t, shareddata.Scalars)

			cursor, err := collection.Find(ctx, bson.D{})
			require.NoError(t, err)

			var resBefore []bson.D
			err = cursor.All(ctx, &resBefore)
			require.NoError(t, err)

			collection.Database().RunCommand(
				ctx,
				bson.D{
					{"delete", collection.Name()},
					{"deletes", tc.deletes},
					{"ordered", tc.ordered},
				},
			)

			cursor, err = collection.Find(ctx, bson.D{})
			require.NoError(t, err)

			var resAfter []bson.D
			err = cursor.All(ctx, &resAfter)
			require.NoError(t, err)

			beforeIDs := make(map[string]struct{})
			for _, r := range resBefore {
				id, ok := r.Map()["_id"].(string)
				require.True(t, ok)

				beforeIDs[id] = struct{}{}
			}

			var created []string
			afterIDs := make(map[string]struct{})
			for _, r := range resAfter {
				id, ok := r.Map()["_id"].(string)
				require.True(t, ok)

				if _, ok := beforeIDs[id]; !ok {
					created = append(created, id)
				}
				afterIDs[id] = struct{}{}
			}

			var removed []string
			for id := range beforeIDs {
				if _, ok := afterIDs[id]; !ok {
					removed = append(removed, id)
				}
			}

			sort.Strings(tc.expectedRemovedIDs)
			sort.Strings(removed)

			assert.Empty(t, created)
			assert.Equal(t, tc.expectedRemovedIDs, removed)
		})
	}
}
