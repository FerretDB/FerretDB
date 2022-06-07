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

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestText(t *testing.T) {
	t.Parallel()

	providers := []shareddata.Provider{shareddata.Composites}
	ctx, collection := setup(t, providers...)

	model := mongo.IndexModel{
		Keys:    bson.D{{"title", 1}},
		Options: options.Index(),
	}
	idxName, err := collection.Indexes().CreateOne(ctx, model)
	require.NoError(t, err)
	require.NotEmpty(t, idxName)
	cursor, err := collection.Indexes().List(ctx)
	var actual []bson.D
	err = cursor.All(ctx, &actual)
	for _, k := range actual {
		t.Logf("%v", k)
	}

	_, err = collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"title", "Ferret"}},
		bson.D{{"_id", 2}, {"title", "Ferret club"}},
		bson.D{{"_id", 3}, {"title", "Ferreters"}},
		bson.D{{"_id", 4}, {"title", "Ferret community"}},
		bson.D{{"_id", 5}, {"title", "ferreters and gophers"}},
		bson.D{{"_id", 6}, {"title", "ферретерс ^-^"}},
		bson.D{{"_id", 7}, {"title", "ferreters are cool"}},
		bson.D{{"_id", 8}, {"title", "FerretDB"}},
		bson.D{{"_id", 9}, {"title", "Ferret Database"}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
		alt         string
	}{
		"Bare": {
			filter:      bson.D{{"$text", bson.D{{"$search", "ferret"}}}},
			expectedIDs: []any{1, 2, 3, 4, 5, 7, 8, 9},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
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
