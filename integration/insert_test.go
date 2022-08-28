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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/shareddata"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// TestInsertTigrisNull tests how the insert operation works with null values in Tigris.
func TestInsertTigrisNull(t *testing.T) {
	setup.SkipForPostgresWithReason(t, "TODO! Fix me!")

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Doubles)

	for name, tc := range map[string]struct {
		collection *mongo.Collection
		insert     bson.D
		err        *mongo.WriteError
	}{
		"ExistingCollectionNewField": {
			collection: collection,
			insert:     bson.D{{"_id", "foo-is-nil"}, {"v", 42.13}, {"foo", nil}},
			err:        nil, // valid even for Tigris, the data is inserted, but the field "foo" will not be present in the schema
		},
		"ExistingCollectionFieldNotSet": {
			collection: collection,
			insert:     bson.D{{"_id", "v-is-not-set"}},
			err:        nil,
		},
		"NewCollection": {
			collection: collection.Database().Collection(collection.Name() + "NewCollection"),
			insert:     bson.D{{"_id", "new-foo-is-nil"}, {"foo", nil}},
			err:        nil,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.collection.InsertOne(ctx, tc.insert)

			if tc.err != nil {
				require.Nil(t, res)
				AssertEqualWriteError(t, *tc.err, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)

			var actual bson.D
			err = tc.collection.FindOne(ctx, bson.D{{"_id", res.InsertedID}}).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.insert, actual)
		})
	}
}
