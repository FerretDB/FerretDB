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
	"fmt"
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
		validator   string
		schema      string
		collection  string
		expectedErr *mongo.CommandError
	}{
		"BadValidator": {
			validator:  "$bad",
			schema:     "{}",
			collection: collection.Name() + "wrong",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `required parameter "$tigrisSchemaString" is missing`,
			},
		},
		"EmptyValidator": {
			validator:  "",
			schema:     "{}",
			collection: collection.Name() + "_empty",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name is not same as schema name 'TestCreateTigris_empty' ''", // Tigris returns this
			},
		},
		"EmptySchema": {
			validator:  "$tigrisSchemaString",
			schema:     "",
			collection: collection.Name() + "_empty",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name is not same as schema name 'TestCreateTigris_empty' ''", // Tigris returns this
			},
		},
		"BadSchema": {
			validator:  "$tigrisSchemaString",
			schema:     "bad",
			collection: collection.Name() + "_bad",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "invalid schema, the following keys are not supported: [bsonType]",
			},
		},
		"Valid": {
			validator: "$tigrisSchemaString",
			schema: fmt.Sprintf(`{
				"title": "%s_good",
				"description": "Foo Bar",
				"primary_key": ["_id"],
				"properties": {
					"balance": {"type": "number"},
					"age": {"type": "integer", "format": "int32"},
					"_id": {"type": "string", "format": "byte"},
					"arr": {"type": "array", "items": {"type": "string"}}
				}
			}`, collection.Name()),
			collection: collection.Name() + "_good",
		},
		"WrongPKey": {
			validator: "$tigrisSchemaString",
			schema: fmt.Sprintf(`{
				"title": "%s_good",
				"description": "Foo Bar",
				"primary_key": [1, 2, 3],
				"properties": {
					"balance": {"type": "number"},
					"age": {"type": "integer", "format": "int32"},
					"_id": {"type": "string", "format": "byte"},
					"arr": {"type": "array", "items": {"type": "string"}}
				}
			}`, collection.Name()),
			collection: collection.Name() + "_good",
		},
		"WrongProperties": {
			validator: "$tigrisSchemaString",
			schema: fmt.Sprintf(`{
				"title": "%s_good",
				"description": "Foo Bar",
				"primary_key": [1, 2, 3],
				"properties": "hello"
			}`, collection.Name()),
			collection: collection.Name() + "_good",
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := new(options.CreateCollectionOptions).SetValidator(bson.D{{tc.validator, tc.schema}})

			err := db.Client().Database(dbName).CreateCollection(ctx, tc.collection, opts)
			if tc.expectedErr != nil {
				AssertEqualError(t, *tc.expectedErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
