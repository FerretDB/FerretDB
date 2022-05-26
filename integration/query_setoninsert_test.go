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
)

func TestSetOnInsert(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		filter      bson.D
		setOnInsert bson.D
		res         bson.D
	}{
		"double": {
			filter:      bson.D{{"_id", "double"}},
			setOnInsert: bson.D{{"value", 43.13}},
			res:         bson.D{{"_id", "double"}, {"value", 43.13}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var err error
			ctx, collection := setup(t)
			expectedRes := &mongo.UpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
			}
			var res *mongo.UpdateResult

			opts := options.Update().SetUpsert(true)
			res, err = collection.UpdateOne(ctx, tc.filter, bson.D{{"$setOnInsert", tc.setOnInsert}}, opts)
			require.NoError(t, err)
			id := res.UpsertedID
			assert.NotEmpty(t, id)
			res.UpsertedID = nil
			assert.Equal(t, expectedRes, res)

			var actual bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&actual)
			require.NoError(t, err)
			if !AssertEqualDocuments(t, tc.res, actual) {
				t.FailNow()
			}
		})
	}
}
