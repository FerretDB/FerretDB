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

package cursors

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestCursorsGetMoreCommand(t *testing.T) {
	// do not run tests in parallel to avoid using too many backend connections

	// options are applied to create a client that uses single connection pool
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		ExtraOptions: url.Values{
			"minPoolSize":   []string{"1"},
			"maxPoolSize":   []string{"1"},
			"maxIdleTimeMS": []string{"0"},
		},
	})

	ctx, collection := s.Ctx, s.Collection

	// the number of documents is set above the default batchSize of 101
	// for testing unset batchSize returning default batchSize
	bsonArr, arr := integration.GenerateDocuments(0, 110)

	_, err := collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		firstBatchSize   any // optional, nil to leave firstBatchSize unset
		getMoreBatchSize any // optional, nil to leave getMoreBatchSize unset
		collection       any // optional, nil to leave collection unset
		cursorID         any // optional, defaults to cursorID from find()/aggregate()

		firstBatch []*types.Document   // required, expected find firstBatch
		nextBatch  []*types.Document   // optional, expected getMore nextBatch
		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
	}{
		"Int": {
			firstBatchSize:   1,
			getMoreBatchSize: int32(1),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			nextBatch:        integration.ConvertDocuments(t, arr[1:2]),
		},
		"IntNegative": {
			firstBatchSize:   1,
			getMoreBatchSize: int32(-1),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"IntZero": {
			firstBatchSize:   1,
			getMoreBatchSize: int32(0),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			nextBatch:        integration.ConvertDocuments(t, arr[1:]),
		},
		"Long": {
			firstBatchSize:   1,
			getMoreBatchSize: int64(1),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			nextBatch:        integration.ConvertDocuments(t, arr[1:2]),
		},
		"LongNegative": {
			firstBatchSize:   1,
			getMoreBatchSize: int64(-1),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"LongZero": {
			firstBatchSize:   1,
			getMoreBatchSize: int64(0),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			nextBatch:        integration.ConvertDocuments(t, arr[1:]),
		},
		"Double": {
			firstBatchSize:   1,
			getMoreBatchSize: float64(1),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			nextBatch:        integration.ConvertDocuments(t, arr[1:2]),
		},
		"DoubleNegative": {
			firstBatchSize:   1,
			getMoreBatchSize: float64(-1),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"DoubleZero": {
			firstBatchSize:   1,
			getMoreBatchSize: float64(0),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			nextBatch:        integration.ConvertDocuments(t, arr[1:]),
		},
		"DoubleFloor": {
			firstBatchSize:   1,
			getMoreBatchSize: 1.9,
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			nextBatch:        integration.ConvertDocuments(t, arr[1:2]),
		},
		"GetMoreCursorExhausted": {
			firstBatchSize:   200,
			getMoreBatchSize: int32(1),
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:110]),
			err: &mongo.CommandError{
				Code:    43,
				Name:    "CursorNotFound",
				Message: "cursor id 0 not found",
			},
		},
		"Bool": {
			firstBatchSize:   1,
			getMoreBatchSize: false,
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.batchSize' is the wrong type 'bool', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'getMore.batchSize' is the wrong type 'bool', expected types '[long, int, decimal, double]'",
		},
		"Unset": {
			firstBatchSize: 1,
			// unset getMore batchSize gets all remaining documents
			getMoreBatchSize: nil,
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			nextBatch:        integration.ConvertDocuments(t, arr[1:]),
		},
		"LargeBatchSize": {
			firstBatchSize:   1,
			getMoreBatchSize: 105,
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			nextBatch:        integration.ConvertDocuments(t, arr[1:106]),
		},
		"StringCursorID": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       collection.Name(),
			cursorID:         "invalid",
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.getMore' is the wrong type 'string', expected type 'long'",
			},
			altMessage: "BSON field 'getMore.getMore' is the wrong type, expected type 'long'",
		},
		"Int32CursorID": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       collection.Name(),
			cursorID:         int32(1111),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.getMore' is the wrong type 'int', expected type 'long'",
			},
			altMessage: "BSON field 'getMore.getMore' is the wrong type, expected type 'long'",
		},
		"NotFoundCursorID": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       collection.Name(),
			cursorID:         int64(1234),
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    43,
				Name:    "CursorNotFound",
				Message: "cursor id 1234 not found",
			},
		},
		"WrongTypeNamespace": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       bson.D{},
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.collection' is the wrong type 'object', expected type 'string'",
			},
		},
		"InvalidNamespace": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       "invalid",
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code: 13,
				Name: "Unauthorized",
				Message: "Requested getMore on namespace 'TestCursorsGetMoreCommand.invalid'," +
					" but cursor belongs to a different namespace TestCursorsGetMoreCommand.TestCursorsGetMoreCommand",
			},
		},
		"EmptyCollectionName": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       "",
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Collection names cannot be empty",
			},
		},
		"MissingCollectionName": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       nil,
			firstBatch:       integration.ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'getMore.collection' is missing but a required field",
			},
		},
		"UnsetAllBatchSize": {
			firstBatchSize:   nil,
			getMoreBatchSize: nil,
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:101]),
			nextBatch:        integration.ConvertDocuments(t, arr[101:]),
		},
		"UnsetFindBatchSize": {
			firstBatchSize:   nil,
			getMoreBatchSize: 5,
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:101]),
			nextBatch:        integration.ConvertDocuments(t, arr[101:106]),
		},
		"UnsetGetMoreBatchSize": {
			firstBatchSize:   5,
			getMoreBatchSize: nil,
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:5]),
			nextBatch:        integration.ConvertDocuments(t, arr[5:]),
		},
		"BatchSize": {
			firstBatchSize:   3,
			getMoreBatchSize: 5,
			collection:       collection.Name(),
			firstBatch:       integration.ConvertDocuments(t, arr[:3]),
			nextBatch:        integration.ConvertDocuments(t, arr[3:8]),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			// Do not run subtests in t.Parallel() to eliminate the occurrence
			// of session error.
			// Supporting session would help us understand fix it
			// https://github.com/FerretDB/FerretDB/issues/153.
			//
			// > Location50738
			// > Cannot run getMore on cursor 2053655655200551971,
			// > which was created in session 2926eea5-9775-41a3-a563-096969f1c7d5 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  - ,
			// > in session 774d9ac6-b24a-4fd8-9874-f92ab1c9c8f5 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  -

			require.NotNil(t, tc.firstBatch, "firstBatch must not be nil")

			var findRest bson.D
			aggregateCursor := bson.D{}

			if tc.firstBatchSize != nil {
				findRest = append(findRest, bson.E{Key: "batchSize", Value: tc.firstBatchSize})
				aggregateCursor = bson.D{{"batchSize", tc.firstBatchSize}}
			}

			aggregateCommand := bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", aggregateCursor},
			}

			findCommand := append(
				bson.D{{"find", collection.Name()}},
				findRest...,
			)

			for _, command := range []bson.D{findCommand, aggregateCommand} {
				var res bson.D
				err := collection.Database().RunCommand(ctx, command).Decode(&res)
				require.NoError(t, err)

				firstBatch, cursorID := getFirstBatch(t, res)

				require.Equal(t, len(tc.firstBatch), firstBatch.Len(), "expected: %v, got: %v", tc.firstBatch, firstBatch)
				for i, elem := range tc.firstBatch {
					require.Equal(t, elem, must.NotFail(firstBatch.Get(i)))
				}

				if tc.cursorID != nil {
					cursorID = tc.cursorID
				}

				var getMoreRest bson.D
				if tc.getMoreBatchSize != nil {
					getMoreRest = append(getMoreRest, bson.E{Key: "batchSize", Value: tc.getMoreBatchSize})
				}

				if tc.collection != nil {
					getMoreRest = append(getMoreRest, bson.E{Key: "collection", Value: tc.collection})
				}

				getMoreCommand := append(
					bson.D{
						{"getMore", cursorID},
					},
					getMoreRest...,
				)

				err = collection.Database().RunCommand(ctx, getMoreCommand).Decode(&res)
				if tc.err != nil {
					integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

					// upon error response contains firstBatch field.
					firstBatch, _ = getFirstBatch(t, res)

					require.Equal(t, len(tc.firstBatch), firstBatch.Len(), "expected: %v, got: %v", tc.firstBatch, firstBatch)
					for i, elem := range tc.firstBatch {
						require.Equal(t, elem, must.NotFail(firstBatch.Get(i)))
					}

					return
				}

				require.NoError(t, err)

				nextBatch, _ := getNextBatch(t, res)

				require.Equal(t, len(tc.nextBatch), nextBatch.Len(), "expected: %v, got: %v", tc.nextBatch, nextBatch)
				for i, elem := range tc.nextBatch {
					require.Equal(t, elem, must.NotFail(nextBatch.Get(i)))
				}
			}
		})
	}
}

func TestCursorsGetMoreBatchSizeCursor(t *testing.T) {
	// do not run tests in parallel to avoid using too many backend connections

	ctx, collection := setup.Setup(t)

	// The test cases call `find`/`aggregate`, then may implicitly call `getMore` upon `cursor.Next()`.
	// The batchSize set by `find`/`aggregate` is used also by `getMore` unless
	// `find`/`aggregate` has default batchSize or 0 batchSize, then `getMore` has unlimited batchSize.
	// To test that, the number of documents is set to more than the double of default batchSize 101.
	arr, _ := integration.GenerateDocuments(0, 220)
	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	findFunc := func(batchSize *int32) (*mongo.Cursor, error) {
		opts := options.Find()
		if batchSize != nil {
			opts = opts.SetBatchSize(*batchSize)
		}

		return collection.Find(ctx, bson.D{}, opts)
	}

	aggregateFunc := func(batchSize *int32) (*mongo.Cursor, error) {
		opts := options.Aggregate()
		if batchSize != nil {
			opts = opts.SetBatchSize(*batchSize)
		}

		return collection.Aggregate(ctx, bson.D{}, opts)
	}

	cursorFuncs := []func(batchSize *int32) (*mongo.Cursor, error){findFunc, aggregateFunc}

	t.Run("SetBatchSize", func(t *testing.T) {
		t.Parallel()

		for _, f := range cursorFuncs {
			cursor, err := f(pointer.ToInt32(2))
			require.NoError(t, err)

			defer cursor.Close(ctx)

			require.Equal(t, 2, cursor.RemainingBatchLength(), "expected 2 documents in first batch")

			for i := 2; i > 0; i-- {
				ok := cursor.Next(ctx)
				require.True(t, ok, "expected to have next document in first batch")
				require.Equal(t, i-1, cursor.RemainingBatchLength())
			}

			// batchSize of 2 is applied to second batch which is obtained by implicit call to `getMore`
			for i := 2; i > 0; i-- {
				ok := cursor.Next(ctx)
				require.True(t, ok, "expected to have next document in second batch")
				require.Equal(t, i-1, cursor.RemainingBatchLength())
			}

			cursor.SetBatchSize(5)

			for i := 5; i > 0; i-- {
				ok := cursor.Next(ctx)
				require.True(t, ok, "expected to have next document in third batch")
				require.Equal(t, i-1, cursor.RemainingBatchLength())
			}

			// get rest of documents from the cursor to ensure cursor is exhausted
			var res bson.D
			err = cursor.All(ctx, &res)
			require.NoError(t, err)

			ok := cursor.Next(ctx)
			require.False(t, ok, "cursor exhausted, not expecting next document")
		}
	})

	t.Run("DefaultBatchSize", func(t *testing.T) {
		t.Parallel()

		for _, f := range cursorFuncs {
			// unset batchSize uses default batchSize 101 for the first batch
			cursor, err := f(nil)
			require.NoError(t, err)

			defer cursor.Close(ctx)

			require.Equal(t, 101, cursor.RemainingBatchLength())

			for i := 101; i > 0; i-- {
				ok := cursor.Next(ctx)
				require.True(t, ok, "expected to have next document")
				require.Equal(t, i-1, cursor.RemainingBatchLength())
			}

			// next batch obtain from implicit call to `getMore` has the rest of the documents, not default batchSize
			// 16MB batchSize limit
			// TODO https://github.com/FerretDB/FerretDB/issues/2824
			ok := cursor.Next(ctx)
			require.True(t, ok, "expected to have next document")
			require.Equal(t, 118, cursor.RemainingBatchLength())
		}
	})

	t.Run("ZeroBatchSize", func(t *testing.T) {
		t.Parallel()

		for _, f := range cursorFuncs {
			cursor, err := f(pointer.ToInt32(0))
			require.NoError(t, err)

			defer cursor.Close(ctx)

			require.Equal(t, 0, cursor.RemainingBatchLength())

			// next batch obtain from implicit call to `getMore` has the rest of the documents, not 0 batchSize
			// 16MB batchSize limit
			// TODO https://github.com/FerretDB/FerretDB/issues/2824
			ok := cursor.Next(ctx)
			require.True(t, ok, "expected to have next document")
			require.Equal(t, 219, cursor.RemainingBatchLength())
		}
	})

	t.Run("NegativeLimit", func(t *testing.T) {
		t.Parallel()

		// set limit to negative, it ignores batchSize and returns single document in the firstBatch.
		cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(10).SetLimit(-1))
		require.NoError(t, err)

		defer cursor.Close(ctx)

		require.Equal(t, 1, cursor.RemainingBatchLength(), "expected 1 document in first batch")

		ok := cursor.Next(ctx)
		require.True(t, ok, "expected to have next document")
		require.Equal(t, 0, cursor.RemainingBatchLength())

		// there is no remaining batch due to negative limit
		ok = cursor.Next(ctx)
		require.False(t, ok, "cursor exhausted, not expecting next document")
		require.Equal(t, 0, cursor.RemainingBatchLength())
	})
}

func TestCursorsGetMoreCommandConnection(t *testing.T) {
	// do not run tests in parallel to avoid using too many backend connections

	// options are applied to create a client that uses single connection pool
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		ExtraOptions: url.Values{
			"minPoolSize":   []string{"1"},
			"maxPoolSize":   []string{"1"},
			"maxIdleTimeMS": []string{"0"},
		},
	})

	ctx := s.Ctx
	collection1 := s.Collection
	databaseName := s.Collection.Database().Name()
	collectionName := s.Collection.Name()

	arr, _ := integration.GenerateDocuments(0, 5)
	_, err := collection1.InsertMany(ctx, arr)
	require.NoError(t, err)

	t.Run("SameClient", func(t *testing.T) {
		// Do not run subtests in t.Parallel() to eliminate the occurrence
		// of session error.
		// Supporting session would help us understand fix it
		// https://github.com/FerretDB/FerretDB/issues/153.
		//
		// > Location50738
		// > Cannot run getMore on cursor 2053655655200551971,
		// > which was created in session 2926eea5-9775-41a3-a563-096969f1c7d5 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  - ,
		// > in session 774d9ac6-b24a-4fd8-9874-f92ab1c9c8f5 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  -

		var res bson.D
		err = collection1.Database().RunCommand(
			ctx,
			bson.D{
				{"find", collection1.Name()},
				{"batchSize", 2},
			},
		).Decode(&res)
		require.NoError(t, err)

		doc := integration.ConvertDocument(t, res)

		v, _ := doc.Get("cursor")
		require.NotNil(t, v)

		cursor, ok := v.(*types.Document)
		require.True(t, ok)

		cursorID, _ := cursor.Get("id")
		assert.NotNil(t, cursorID)

		err = collection1.Database().RunCommand(
			ctx,
			bson.D{
				{"getMore", cursorID},
				{"collection", collection1.Name()},
			},
		).Decode(&res)
		require.NoError(t, err)
	})

	t.Run("DifferentClient", func(tt *testing.T) {
		// The error returned from MongoDB is a session error, FerretDB does not
		// return an error because db, collection and username are the same.
		t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/153")

		// do not run subtest in parallel to avoid breaking another parallel subtest
		var res bson.D
		err = collection1.Database().RunCommand(
			ctx,
			bson.D{
				{"find", collection1.Name()},
				{"batchSize", 2},
			},
		).Decode(&res)
		require.NoError(t, err)

		doc := integration.ConvertDocument(t, res)

		v, _ := doc.Get("cursor")
		require.NotNil(t, v)

		cursor, ok := v.(*types.Document)
		require.True(t, ok)

		cursorID, _ := cursor.Get("id")
		assert.NotNil(t, cursorID)

		client2, err := mongo.Connect(ctx, options.Client().ApplyURI(s.MongoDBURI))
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, client2.Disconnect(ctx))
		})

		err = client2.Database(databaseName).RunCommand(
			ctx,
			bson.D{
				{"getMore", cursorID},
				{"collection", client2.Database(databaseName).Collection(collectionName).Name()},
			},
		).Decode(&res)

		integration.AssertMatchesCommandError(t, mongo.CommandError{Code: 13, Name: "Unauthorized"}, err)
	})
}

func TestCursorsGetMoreCommandMaxTimeMSErrors(t *testing.T) {
	// do not run tests in parallel to avoid using too many backend connections

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		command bson.D // required, command to run

		err        *mongo.CommandError // required, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
	}{
		"NegativeLong": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", int64(-1)},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-1 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "-1 value for maxTimeMS is out of range",
		},
		"MaxLong": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", math.MaxInt64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "9223372036854775807 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "9223372036854775807 value for maxTimeMS is out of range",
		},
		"Double": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", 1000.5},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS has non-integral value",
			},
			altMessage: "BSON field 'getMore.maxTimeMS' is the wrong type 'double', expected types '[long, int, decimal, double]'",
		},
		"NegativeDouble": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", -14245345234123245.55},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-14245345234123246 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "-1.4245345234123246e+16 value for maxTimeMS is out of range",
		},
		"BigDouble": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", math.MaxFloat64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "9223372036854775807 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "1.797693134862316e+308 value for maxTimeMS is out of range",
		},
		"BigNegativeDouble": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", -math.MaxFloat64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-9223372036854775808 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "-1.797693134862316e+308 value for maxTimeMS is out of range",
		},
		"NegativeInt": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", -1123123},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "-1123123 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "-1123123 value for maxTimeMS is out of range",
		},
		"MaxInt": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", math.MaxInt32 + 1},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "2147483648 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage: "2147483648 value for maxTimeMS is out of range",
		},
		"Null": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", nil},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
		},
		"String": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", "string"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.maxTimeMS' is the wrong type 'string', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'getMore.maxTimeMS' is the wrong type 'string', expected types '[long, int, decimal, double]'",
		},
		"Array": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", bson.A{int32(42), "foo", nil}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.maxTimeMS' is the wrong type 'array', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'getMore.maxTimeMS' is the wrong type 'array', expected types '[long, int, decimal, double]'",
		},
		"Document": {
			command: bson.D{
				{"getMore", int64(112233)},
				{"collection", collection.Name()},
				{"maxTimeMS", bson.D{{"foo", int32(42)}}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.maxTimeMS' is the wrong type 'object', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'getMore.maxTimeMS' is the wrong type 'object', expected types '[long, int, decimal, double]'",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, tc.err, "err must not be nil")

			var res bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&res)
			integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
			require.Nil(t, res)
		})
	}
}

func TestCursorsGetMoreExhausted(t *testing.T) {
	s := setup.SetupWithOpts(t, nil)

	collection := s.Collection
	db, ctx := collection.Database(), s.Ctx

	arr, _ := integration.GenerateDocuments(0, 10)

	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	var res bson.D
	err = db.RunCommand(ctx, bson.D{
		{"find", collection.Name()},
		{"batchSize", 1},
	}).Decode(&res)

	require.NoError(t, err)

	firstBatch, cursorID := getFirstBatch(t, res)
	require.Equal(t, 1, firstBatch.Len())
	require.NotNil(t, cursorID)

	err = db.RunCommand(ctx, bson.D{
		{"getMore", cursorID},
		{"collection", collection.Name()},
		{"batchSize", 9},
	}).Decode(&res)

	require.NoError(t, err)

	nextBatch, nextID := getNextBatch(t, res)
	require.Equal(t, 9, nextBatch.Len())
	assert.Equal(t, cursorID, nextID)

	err = db.RunCommand(ctx, bson.D{
		{"getMore", cursorID},
		{"collection", collection.Name()},
		{"batchSize", 1},
	}).Decode(&res)

	require.NoError(t, err)

	nextBatch, nextID = getNextBatch(t, res)
	require.Equal(t, 0, nextBatch.Len())
	assert.Equal(t, int64(0), nextID)

	err = db.RunCommand(ctx, bson.D{
		{"getMore", cursorID},
		{"collection", collection.Name()},
		{"batchSize", 1},
	}).Err()

	expectedErr := mongo.CommandError{
		Code:    43,
		Name:    "CursorNotFound",
		Message: fmt.Sprintf("cursor id %d not found", cursorID),
	}

	integration.AssertEqualCommandError(t, expectedErr, err)
}

func TestCursorsGetMoreCommandMaxTimeMSCursor(t *testing.T) {
	// do not run tests in parallel to avoid using too many backend connections

	// options are applied to create a client that uses single connection pool
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		ExtraOptions: url.Values{
			"minPoolSize":   []string{"1"},
			"maxPoolSize":   []string{"1"},
			"maxIdleTimeMS": []string{"0"},
		},
		Providers: []shareddata.Provider{shareddata.Composites},
	})

	ctx, collection := s.Ctx, s.Collection

	// need large amount of documents for time out to trigger
	arr, _ := integration.GenerateDocuments(0, 5000)

	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	t.Run("FindExpire", func(t *testing.T) {
		opts := options.Find().
			// set batchSize big enough to hit maxTimeMS
			SetBatchSize(2000).
			// set maxTimeMS small enough for find to expire
			SetMaxTime(1).
			// set sort to slow down the query more than 1ms
			SetSort(bson.D{{"v", 1}})

		_, err := collection.Find(ctx, bson.D{}, opts)

		integration.AssertMatchesCommandError(t, mongo.CommandError{Code: 50, Name: "MaxTimeMSExpired"}, err)
	})

	t.Run("AggregateExpire", func(t *testing.T) {
		opts := options.Aggregate().
			// set batchSize big enough to hit maxTimeMS
			SetBatchSize(2000).
			// set maxTimeMS small enough for aggregate to expire
			SetMaxTime(1)

		// use $sort stage to slow down the query more than 1ms
		_, err := collection.Aggregate(ctx, bson.A{bson.D{{"$sort", bson.D{{"v", 1}}}}}, opts)

		integration.AssertMatchesCommandError(t, mongo.CommandError{Code: 50, Name: "MaxTimeMSExpired"}, err)
	})
}

func TestCursors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		ExtraOptions: url.Values{
			"minPoolSize": []string{"1"},
			"maxPoolSize": []string{"1"},
		},
	})

	collection, ctx := s.Collection, s.Ctx

	arr, _ := integration.GenerateDocuments(1, 5)
	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	var res bson.D

	t.Run("RemoveLastDocument", func(tt *testing.T) {
		t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/3818")

		err = collection.Database().RunCommand(ctx, bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
		}).Decode(&res)
		require.NoError(t, err)

		firstBatch, cursorID := getFirstBatch(t, res)
		require.NotNil(t, firstBatch)
		require.Equal(t, 1, firstBatch.Len())

		_, err = collection.DeleteOne(ctx, bson.D{{"_id", 4}})
		require.NoError(t, err)

		getMoreCmd := bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 1},
		}

		for i := 1; i < 3; i++ {
			err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
			nextBatch, nextID := getNextBatch(t, res)
			require.Equal(t, cursorID, nextID)
			require.Equal(t, 1, nextBatch.Len())
		}

		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, int64(0), nextID)
		require.Equal(t, 0, nextBatch.Len())
	})

	t.Run("QueryPlanKilledByDrop", func(tt *testing.T) {
		t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/3818")

		err = collection.Database().RunCommand(ctx, bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
		}).Decode(&res)
		require.NoError(t, err)

		firstBatch, cursorID := getFirstBatch(t, res)
		require.NotNil(t, firstBatch)
		require.Equal(t, 1, firstBatch.Len())

		err = collection.Database().Drop(ctx)
		require.NoError(t, err)

		getMoreCmd := bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 1},
		}

		err = collection.Database().RunCommand(ctx, getMoreCmd).Err()
		require.Error(t, err)

		var ce mongo.CommandError
		require.True(t, errors.As(err, &ce))
		require.Equal(t, int32(175), ce.Code, "invalid error: %v", ce)
	})
}

func TestCursorsFirstBatchMaxTimeMS(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, t.Name(), opts)
	require.NoError(t, err)

	collection := db.Collection(t.Name())

	bsonArr, arr := integration.GenerateDocuments(0, 3)

	_, err = collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	var cursorID any

	t.Run("FirstBatch", func(t *testing.T) {
		cmd := bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
			{"maxTimeMS", 200},
		}

		var res bson.D
		err = collection.Database().RunCommand(ctx, cmd).Decode(&res)
		require.NoError(t, err)

		var firstBatch *types.Array
		firstBatch, cursorID = getFirstBatch(t, res)

		expectedFirstBatch := integration.ConvertDocuments(t, arr[:1])
		require.Equal(t, len(expectedFirstBatch), firstBatch.Len())
		require.Equal(t, expectedFirstBatch[0], must.NotFail(firstBatch.Get(0)))
	})

	getMoreCmd := bson.D{
		{"getMore", cursorID},
		{"collection", collection.Name()},
		{"batchSize", 1},
	}

	t.Run("GetMore", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			time.Sleep(100 * time.Millisecond)
			var res bson.D
			err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
			require.NoError(t, err)

			nextBatch, nextID := getNextBatch(t, res)
			expectedNextBatch := integration.ConvertDocuments(t, arr[i+1:i+2])

			assert.Equal(t, cursorID, nextID)

			require.Equal(t, len(expectedNextBatch), nextBatch.Len())
			require.Equal(t, expectedNextBatch[0], must.NotFail(nextBatch.Get(0)))
		}
	})

	t.Run("GetMoreEmpty", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, int64(0), nextID)
	})
}

func TestGetMoreNonAwaitDataError(t *testing.T) {
	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, t.Name(), opts)
	require.NoError(t, err)

	collection := db.Collection(t.Name())
	bsonDocs, _ := integration.GenerateDocuments(0, 1)

	_, err = collection.InsertMany(ctx, bsonDocs)
	require.NoError(t, err)

	t.Run("Cursor", func(tt *testing.T) {
		var res bson.D
		err := collection.Database().RunCommand(ctx, bson.D{
			{"find", collection.Name()},
			{"tailable", true},
			{"batchSize", 0},
		}).Decode(&res)
		require.NoError(t, err)

		doc := integration.ConvertDocument(t, res)

		v, _ := doc.Get("cursor")
		require.NotNil(t, v)

		cursor, ok := v.(*types.Document)
		require.True(t, ok)

		cursorID, _ := cursor.Get("id")
		require.NotZero(t, cursorID)

		err = collection.Database().RunCommand(ctx, bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 2000},
			{"maxTimeMS", 1},
		}).Err()

		integration.AssertEqualCommandError(
			t,
			mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "cannot set maxTimeMS on getMore command for a non-awaitData cursor",
			},
			err,
		)
	})
	t.Run("TailableCursor", func(tt *testing.T) {
		var res bson.D
		err := collection.Database().RunCommand(ctx, bson.D{
			{"find", collection.Name()},
			{"tailable", true},
			{"batchSize", 0},
		}).Decode(&res)
		require.NoError(t, err)

		doc := integration.ConvertDocument(t, res)

		v, _ := doc.Get("cursor")
		require.NotNil(t, v)

		cursor, ok := v.(*types.Document)
		require.True(t, ok)

		cursorID, _ := cursor.Get("id")
		require.NotZero(t, cursorID)

		err = collection.Database().RunCommand(ctx, bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 2000},
			{"maxTimeMS", 1},
		}).Err()

		integration.AssertEqualCommandError(
			t,
			mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "cannot set maxTimeMS on getMore command for a non-awaitData cursor",
			},
			err,
		)
	})
	t.Run("AwaitDataCursor", func(tt *testing.T) {
		var res bson.D
		err := collection.Database().RunCommand(ctx, bson.D{
			{"find", collection.Name()},
			{"tailable", true},
			{"awaitData", true},
			{"batchSize", 0},
		}).Decode(&res)
		require.NoError(t, err)

		doc := integration.ConvertDocument(t, res)

		v, _ := doc.Get("cursor")
		require.NotNil(t, v)

		cursor, ok := v.(*types.Document)
		require.True(t, ok)

		cursorID, _ := cursor.Get("id")
		require.NotZero(t, cursorID)

		err = collection.Database().RunCommand(ctx, bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 2000},
			{"maxTimeMS", 1},
		}).Err()

		require.NoError(t, err)
	})
}

func TestCursorsGetMoreAfterInsertion(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, t.Name(), opts)
	require.NoError(t, err)

	collection := db.Collection(t.Name())

	bsonArr, arr := integration.GenerateDocuments(0, 3)

	_, err = collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	var cursorID any

	t.Run("FirstBatch", func(t *testing.T) {
		cmd := bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
		}

		var res bson.D
		err = collection.Database().RunCommand(ctx, cmd).Decode(&res)
		require.NoError(t, err)

		var firstBatch *types.Array
		firstBatch, cursorID = getFirstBatch(t, res)

		expectedFirstBatch := integration.ConvertDocuments(t, arr[:1])
		require.Equal(t, len(expectedFirstBatch), firstBatch.Len())
		require.Equal(t, expectedFirstBatch[0], must.NotFail(firstBatch.Get(0)))
	})

	getMoreCmd := bson.D{
		{"getMore", cursorID},
		{"collection", collection.Name()},
		{"batchSize", 1},
	}

	t.Run("GetMore", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			var res bson.D
			err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
			require.NoError(t, err)

			nextBatch, nextID := getNextBatch(t, res)
			expectedNextBatch := integration.ConvertDocuments(t, arr[i+1:i+2])

			assert.Equal(t, cursorID, nextID)

			require.Equal(t, len(expectedNextBatch), nextBatch.Len())
			require.Equal(t, expectedNextBatch[0], must.NotFail(nextBatch.Get(0)))
		}
	})

	t.Run("GetMoreEmpty", func(tt *testing.T) {
		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, int64(0), nextID)
	})

	t.Run("GetMoreNewDoc", func(tt *testing.T) {
		newDoc := bson.D{{"_id", "new"}}
		_, err = collection.InsertOne(ctx, newDoc)
		require.NoError(t, err)

		var res bson.D

		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		integration.AssertEqualCommandError(
			t,
			mongo.CommandError{
				Code:    43,
				Name:    "CursorNotFound",
				Message: fmt.Sprintf("cursor id %d not found", cursorID),
			},
			err,
		)
	})
}
