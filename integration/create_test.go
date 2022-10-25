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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCreateStress(t *testing.T) {
	setup.SkipForPostgresWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1206")

	t.Parallel()

	ctx, collection := setup.Setup(t) // no providers there, we will create collections concurrently
	db := collection.Database()

	const collNum = 20 // runtime.GOMAXPROCS * 10

	ready := make(chan struct{}, collNum)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < collNum; i++ {

		wg.Add(1)

		go func(i int) {
			ready <- struct{}{}
			<-start

			collName := fmt.Sprintf("stress_%d", i)

			schema := fmt.Sprintf(`{
				"title": "%s",
				"description": "Create Collection Stress %d",
				"primary_key": ["_id"],
				"properties": {
					"_id": {"type": "string"},
					"v": {"type": "string"}
				}
			}`, collName, i)
			opts := options.CreateCollectionOptions{
				Validator: bson.D{{"$tigrisSchemaString", schema}},
			}

			// Attempt to create a collection for Tigris with a schema.
			// If we get an error, it's not Tigris, so we create collection without schema
			err := db.CreateCollection(ctx, collName, &opts)
			if err != nil {
				err := db.CreateCollection(ctx, collName)
				assert.NoError(t, err)
			}

			_, err = db.Collection(collName).InsertOne(ctx, bson.D{{"_id", "foo"}, {"v", "bar"}})

			assert.NoError(t, err)
			defer wg.Done()
		}(i)

		close(start)
	}

	wg.Wait()

	colls, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	// TODO https://github.com/FerretDB/FerretDB/issues/1206
	// Without SkipForPostgres this test would fail.
	// Even though all the collections are created as separate tables in the database,
	// the settings table doesn't store all of them because of concurrency issues.
	require.Len(t, colls, collNum)

	// check that all collections were created, and we can query them
	for i := 0; i < collNum; i++ {
		i := i

		t.Run(fmt.Sprintf("check_stress_%d", i), func(t *testing.T) {
			t.Parallel()

			collName := fmt.Sprintf("stress_%d", i)

			var doc bson.D
			err := db.Collection(collName).FindOne(ctx, bson.D{{"_id", "foo"}}).Decode(&doc)
			require.NoError(t, err)
			require.Equal(t, bson.D{{"_id", "foo"}, {"v", "bar"}}, doc)
		})
	}
}

func TestCreateTigris(t *testing.T) {
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
		doc         bson.D
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
		"EmptySchema": {
			validator:  "$tigrisSchemaString",
			schema:     "",
			collection: collection.Name() + "_empty",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "empty schema is not allowed",
			},
		},
		"BadSchema": {
			validator:  "$tigrisSchemaString",
			schema:     "bad",
			collection: collection.Name() + "_bad",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "invalid character 'b' looking for beginning of value",
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
					"_id": {"type": "string"},
					"obj": {"type": "object", "properties": {"foo": {"type": "string"}}}
				}
			}`, collection.Name(),
			),
			collection: collection.Name() + "_good",
			doc: bson.D{
				{"_id", "foo"},
				{"balance", 1.0},
				{"age", 1},
				{"obj", bson.D{{"foo", "bar"}}},
			},
		},
		"WrongPKey": {
			validator: "$tigrisSchemaString",
			schema: fmt.Sprintf(`{
				"title": "%s_pkey",
				"description": "Foo Bar",
				"primary_key": [1, 2, 3],
				"properties": {
					"balance": {"type": "number"},
					"age": {"type": "integer", "format": "int32"},
					"_id": {"type": "string", "format": "byte"},
					"arr": {"type": "array", "items": {"type": "string"}}
				}
			}`, collection.Name(),
			),
			collection: collection.Name() + "_pkey",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "json: cannot unmarshal number into Go struct field Schema.primary_key of type string",
			},
		},
		"WrongProperties": {
			validator: "$tigrisSchemaString",
			schema: fmt.Sprintf(`{
				"title": "%s_wp",
				"description": "Foo Bar",
				"primary_key": ["_id"],
				"properties": "hello"
			}`, collection.Name(),
			),
			collection: collection.Name() + "_wp",
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "json: cannot unmarshal string into Go struct field Schema.properties of type map[string]*tjson.Schema",
			},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			opts := options.CreateCollectionOptions{
				Validator: bson.D{{tc.validator, tc.schema}},
			}

			err := db.Client().Database(dbName).CreateCollection(ctx, tc.collection, &opts)
			if tc.expectedErr != nil {
				AssertEqualError(t, *tc.expectedErr, err)
			} else {
				require.NoError(t, err)
			}

			// to make sure that schema is correct, we try to insert a document
			if tc.doc != nil {
				_, err = db.Collection(tc.collection).InsertOne(ctx, tc.doc)
				require.NoError(t, err)
			}
		})
	}
}
