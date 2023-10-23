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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
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

func TestCreateStressSameCollection(t *testing.T) {
	// It should be rewritten to use teststress.Stress.

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

			err := db.CreateCollection(ctx, collName)
			if err == nil {
				created.Add(1)
			} else {
				AssertEqualCommandError(
					t,
					mongo.CommandError{
						Code:    48,
						Name:    "NamespaceExists",
						Message: `Collection TestCreateStressSameCollection.stress_same_collection already exists.`,
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
	var doc bson.D
	err = db.Collection(collName).FindOne(ctx, bson.D{{"_id", "foo_1"}}).Decode(&doc)
	require.NoError(t, err)
	require.Equal(t, bson.D{{"_id", "foo_1"}, {"v", "bar"}}, doc)

	require.Equal(t, int32(1), created.Load(), "Only one attempt to create a collection should succeed")
}

// TestCreateCappedCommandInvalidSpec checks that invalid create capped collection commands are handled correctly.
// For valid test cases see collStats for capped collections tests.
func TestCreateCappedCommandInvalidSpec(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		collectionName any
		capped         any
		size           any
		max            any

		err        *mongo.CommandError // required, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
	}{
		"ZeroSize": {
			collectionName: "zero_size",
			capped:         true,
			size:           0,
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'size' value must be >= 1, actual value '0'",
			},
		},
		"EmptySize": {
			collectionName: "no_size",
			capped:         true,
			err: &mongo.CommandError{
				Code:    72,
				Name:    "InvalidOptions",
				Message: "the 'size' field is required when 'capped' is true",
			},
		},
		"EmptySizeWithMax": {
			collectionName: "no_size_with_max",
			capped:         true,
			max:            500,
			err: &mongo.CommandError{
				Code:    72,
				Name:    "InvalidOptions",
				Message: "the 'size' field is required when 'capped' is true",
			},
		},
		"WrongSizeType": {
			collectionName: "wrong_size",
			capped:         true,
			size:           "foo",
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'create.size' is the wrong type 'string', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'create.size' is the wrong type 'string', expected types '[long, int, decimal, double]'",
		},
		"WrongMaxType": {
			collectionName: "wrong_max",
			capped:         true,
			size:           500,
			max:            "foo",
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'create.max' is the wrong type 'string', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'create.max' is the wrong type 'string', expected types '[long, int, decimal, double]'",
		},
		"WrongCappedType": {
			collectionName: "wrong_capped",
			capped:         "foo",
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'create.capped' is the wrong type 'string', expected types '[bool, long, int, decimal, double']",
			},
			altMessage: "BSON field 'capped' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
		},
		"NegativeSize": {
			collectionName: "negative_size",
			capped:         true,
			size:           -500,
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'size' value must be >= 1, actual value '-500'",
			},
		},
	} {
		tc, name := tc, name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// one provider is enough to check for errors
			ctx, collection := setup.Setup(t, shareddata.ArrayDocuments)

			command := bson.D{
				{"create", tc.collectionName},
				{"capped", tc.capped},
				{"size", tc.size},
				{"max", tc.max},
			}

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)

			require.Error(t, err)
			require.Nil(t, res)

			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

			// In addition, checking a case when size is not set at all (the error is the same as for empty size)
			if tc.size == nil {
				command := bson.D{
					{"create", tc.collectionName},
					{"capped", tc.capped},
					{"max", tc.max},
				}

				var res bson.D
				err := collection.Database().RunCommand(ctx, command).Decode(&res)

				require.Error(t, err)
				require.Nil(t, res)

				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
			}
		})
	}
}
