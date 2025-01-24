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
	"sync"
	"sync/atomic"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestCreateStress(t *testing.T) {
	// It should be rewritten to use teststress.Stress.

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

			err := db.CreateCollection(ctx, collName)
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
	// It should be rewritten to use teststress.Stress.

	ctx, collection := setup.Setup(t)
	// do not toLower() db name as it may contain uppercase letters
	db := collection.Database().Client().Database(t.Name())

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
	// It should be rewritten to use teststress.Stress.

	ctx, collection := setup.Setup(t)
	// do not toLower() db name as it may contain uppercase letters
	db := collection.Database().Client().Database(t.Name())

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

func TestCreateStressSameCollection(tt *testing.T) {
	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/3853")

	// It should be rewritten to use teststress.Stress.

	ctx, collection := setup.Setup(tt) // no providers there, we will create collection from the test
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

			err := db.CreateCollection(ctx, collName)
			require.NoError(t, err)
			created.Add(1)

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
	var doc bson.D
	err = db.Collection(collName).FindOne(ctx, bson.D{{"_id", "foo_1"}}).Decode(&doc)
	require.NoError(t, err)
	require.Equal(t, bson.D{{"_id", "foo_1"}, {"v", "bar"}}, doc)

	// Until Mongo 7.0, attempts to create a collection that existed would return a NamespaceExists error.
	require.Equal(t, int32(collNum), created.Load(), "All attempts to create a collection should succeed")

	assert.Error(t, db.CreateCollection(ctx, collName, &options.CreateCollectionOptions{
		SizeInBytes: pointer.ToInt64(int64(1024)),
	}))
}
