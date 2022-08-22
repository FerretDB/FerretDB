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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCreateTigris(t *testing.T) {
	setup.SkipForMongoWithReason(t, "Tigris-specific schema is used")
	setup.SkipForPostgresWithReason(t, "Tigris-specific schema is used")

	t.Parallel()

	ctx, collection := setup.Setup(t) // no providers there
	db := collection.Database()
	dbName := db.Name()

	for name, tc := range map[string]struct {
		validator   bson.D
		collection  string
		expectedErr *mongo.CommandError
	}{
		"EmptyValidator": {
			validator:  bson.D{},
			collection: collection.Name() + "_empty",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name is not same as schema name 'TestCreateTigris' ''", // Tigris returns this
			},
		},
		"BadValidator": {
			validator:  bson.D{{"bsonType", "object"}},
			collection: collection.Name() + "_bad",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "invalid schema, the follwing keys are not supported: [bsonType]",
			},
		},
		"GoodValidator": {
			validator: bson.D{
				{"title", collection.Name() + "_good"},
				{"description", "Foo Bar"},
				{"primary_key", bson.A{"_id"}},
				{"properties", bson.D{
					{"balance", bson.D{{"type", "number"}}},
					{"age", bson.D{{"type", "integer"}, {"format", "int32"}}},
					{"_id", bson.D{{"type", "string"}, {"format", "byte"}}},
					{"arr", bson.D{{"type", "array"}, {"items", bson.D{{"type", "string"}}}}},
				}},
			},
			collection: collection.Name() + "_good",
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			opts := new(options.CreateCollectionOptions).SetValidator(bson.D{{"$jsonSchema", tc.validator}})

			err := db.Client().Database(dbName).CreateCollection(ctx, tc.collection, opts)
			if tc.expectedErr != nil {
				AssertEqualError(t, *tc.expectedErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
