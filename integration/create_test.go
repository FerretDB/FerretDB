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
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCreateStress(t *testing.T) {
	ctx, collection := setup.Setup(t) // no providers there, we will create collections concurrently
	db := collection.Database()

	collNum := runtime.GOMAXPROCS(-1) * 10

	ready := make(chan struct{}, collNum)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < collNum; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

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
			}`, collName, i,
			)

			// Set $tigrisSchemaString for tigris only.
			var opts options.CreateCollectionOptions
			if setup.IsTigris(t) {
				opts.Validator = bson.D{{"$tigrisSchemaString", schema}}
			}

			err := db.CreateCollection(ctx, collName, &opts)
			assert.NoError(t, err)

			_, err = db.Collection(collName).InsertOne(ctx, bson.D{{"_id", "foo"}, {"v", "bar"}})

			assert.NoError(t, err)
		}(i)
	}

	for i := 0; i < collNum; i++ {
		<-ready
	}

	close(start)

	wg.Wait()

	colls, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

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

func TestCreateOnInsertStressSameCollection(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1341")
	ctx, collection := setup.Setup(t)
	db := collection.Database().Client().Database(strings.ToLower(t.Name()))

	collNum := runtime.GOMAXPROCS(-1) * 10
	collPrefix := "stress_same_collection"

	ready := make(chan struct{}, collNum)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < collNum; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			ready <- struct{}{}

			<-start

			_, err := db.Collection(collPrefix).InsertOne(ctx, bson.D{
				{"foo", "bar"},
			})
			assert.NoError(t, err)
		}(i)
	}

	for i := 0; i < collNum; i++ {
		<-ready
	}

	close(start)

	wg.Wait()
}

func TestCreateOnInsertStressDiffCollection(t *testing.T) {
	ctx, collection := setup.Setup(t)
	db := collection.Database().Client().Database(strings.ToLower(t.Name()))

	collNum := runtime.GOMAXPROCS(-1) * 10
	collPrefix := "stress_diff_collection_"

	ready := make(chan struct{}, collNum)
	start := make(chan struct{})

	var wg sync.WaitGroup

	for i := 0; i < collNum; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			ready <- struct{}{}

			<-start

			_, err := db.Collection(collPrefix+fmt.Sprint(i)).InsertOne(ctx, bson.D{
				{"foo", "bar"},
			})
			assert.NoError(t, err)
		}(i)
	}

	for i := 0; i < collNum; i++ {
		<-ready
	}

	close(start)

	wg.Wait()
}

func TestCreateStressSameCollection(t *testing.T) {
	ctx, collection := setup.Setup(t) // no providers there, we will create collection from the test
	db := collection.Database()

	collNum := runtime.GOMAXPROCS(-1) * 10
	collName := "stress_same_collection"

	ready := make(chan struct{}, collNum)
	start := make(chan struct{})

	var created atomic.Int32 // number of successful attempts to create a collection

	var wg sync.WaitGroup
	for i := 0; i < collNum; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			ready <- struct{}{}

			<-start

			schema := fmt.Sprintf(`{
				"title": "%s",
				"description": "Create Collection Stress %d",
				"primary_key": ["_id"],
				"properties": {
					"_id": {"type": "string"},
					"v": {"type": "string"}
				}
			}`, collName, i,
			)

			// Set $tigrisSchemaString for tigris only.
			var opts options.CreateCollectionOptions
			if setup.IsTigris(t) {
				opts.Validator = bson.D{{"$tigrisSchemaString", schema}}
			}

			err := db.CreateCollection(ctx, collName, &opts)
			if err == nil {
				created.Add(1)
			} else {
				AssertEqualError(
					t,
					mongo.CommandError{
						Code:    48,
						Name:    "NamespaceExists",
						Message: `Collection testcreatestresssamecollection.stress_same_collection already exists.`,
					},
					err,
				)
			}

			id := fmt.Sprintf("foo_%d", i)
			_, err = db.Collection(collName).InsertOne(ctx, bson.D{{"_id", id}, {"v", "bar"}})

			assert.NoError(t, err)
		}(i)
	}

	for i := 0; i < collNum; i++ {
		<-ready
	}

	close(start)

	wg.Wait()

	colls, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	require.Len(t, colls, 1)

	// check that the collection was created, and we can query it
	t.Run("check_stress", func(t *testing.T) {
		t.Parallel()

		var doc bson.D
		err := db.Collection(collName).FindOne(ctx, bson.D{{"_id", "foo_1"}}).Decode(&doc)
		require.NoError(t, err)
		require.Equal(t, bson.D{{"_id", "foo_1"}, {"v", "bar"}}, doc)
	})

	setup.SkipForTigrisWithReason(t, "In case of Tigris, CreateOrUpdate is called, "+
		"and it's not possible to check the number of creation attempts as some of them might be updates.")
	require.Equal(t, int32(1), created.Load(), "Only one attempt to create a collection should succeed")
}

func TestCreateTigris(t *testing.T) {
	setup.TigrisOnlyWithReason(t, "Tigris-specific schema is used")

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
