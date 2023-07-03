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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestQueryCommandGetMore(t *testing.T) {
	t.Parallel()

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
	bsonArr, arr := generateDocuments(0, 110)

	_, err := collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		firstBatchSize   any // optional, nil to leave firstBatchSize unset
		getMoreBatchSize any // optional, nil to leave getMoreBatchSize unset
		collection       any // optional, nil to leave collection unset
		cursorID         any // optional, defaults to cursorID from find()

		firstBatch []*types.Document   // required, expected find firstBatch
		nextBatch  []*types.Document   // optional, expected getMore nextBatch
		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"Int": {
			firstBatchSize:   1,
			getMoreBatchSize: int32(1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:2]),
		},
		"IntNegative": {
			firstBatchSize:   1,
			getMoreBatchSize: int32(-1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:]),
		},
		"Long": {
			firstBatchSize:   1,
			getMoreBatchSize: int64(1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:2]),
		},
		"LongNegative": {
			firstBatchSize:   1,
			getMoreBatchSize: int64(-1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:]),
		},
		"Double": {
			firstBatchSize:   1,
			getMoreBatchSize: float64(1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:2]),
		},
		"DoubleNegative": {
			firstBatchSize:   1,
			getMoreBatchSize: float64(-1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:]),
		},
		"DoubleFloor": {
			firstBatchSize:   1,
			getMoreBatchSize: 1.9,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:2]),
		},
		"GetMoreCursorExhausted": {
			firstBatchSize:   200,
			getMoreBatchSize: int32(1),
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:110]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:]),
		},
		"LargeBatchSize": {
			firstBatchSize:   1,
			getMoreBatchSize: 105,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:1]),
			nextBatch:        ConvertDocuments(t, arr[1:106]),
		},
		"StringCursorID": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       collection.Name(),
			cursorID:         "invalid",
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
			err: &mongo.CommandError{
				Code: 13,
				Name: "Unauthorized",
				Message: "Requested getMore on namespace 'TestQueryCommandGetMore.invalid'," +
					" but cursor belongs to a different namespace TestQueryCommandGetMore.TestQueryCommandGetMore",
			},
		},
		"EmptyCollectionName": {
			firstBatchSize:   1,
			getMoreBatchSize: 1,
			collection:       "",
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:1]),
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
			firstBatch:       ConvertDocuments(t, arr[:101]),
			nextBatch:        ConvertDocuments(t, arr[101:]),
		},
		"UnsetFindBatchSize": {
			firstBatchSize:   nil,
			getMoreBatchSize: 5,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:101]),
			nextBatch:        ConvertDocuments(t, arr[101:106]),
		},
		"UnsetGetMoreBatchSize": {
			firstBatchSize:   5,
			getMoreBatchSize: nil,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:5]),
			nextBatch:        ConvertDocuments(t, arr[5:]),
		},
		"BatchSize": {
			firstBatchSize:   3,
			getMoreBatchSize: 5,
			collection:       collection.Name(),
			firstBatch:       ConvertDocuments(t, arr[:3]),
			nextBatch:        ConvertDocuments(t, arr[3:8]),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

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

				doc := ConvertDocument(t, res)

				v, _ := doc.Get("cursor")
				require.NotNil(t, v)

				cursor, ok := v.(*types.Document)
				require.True(t, ok)

				cursorID, _ := cursor.Get("id")
				assert.NotNil(t, cursorID)

				v, _ = cursor.Get("firstBatch")
				require.NotNil(t, v)

				firstBatch, ok := v.(*types.Array)
				require.True(t, ok)

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
					AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

					// upon error response contains firstBatch field.
					doc = ConvertDocument(t, res)

					v, _ = doc.Get("cursor")
					require.NotNil(t, v)

					cursor, ok = v.(*types.Document)
					require.True(t, ok)

					cursorID, _ = cursor.Get("id")
					assert.NotNil(t, cursorID)

					v, _ = cursor.Get("firstBatch")
					require.NotNil(t, v)

					firstBatch, ok = v.(*types.Array)
					require.True(t, ok)

					require.Equal(t, len(tc.firstBatch), firstBatch.Len(), "expected: %v, got: %v", tc.firstBatch, firstBatch)
					for i, elem := range tc.firstBatch {
						require.Equal(t, elem, must.NotFail(firstBatch.Get(i)))
					}

					return
				}

				require.NoError(t, err)

				doc = ConvertDocument(t, res)

				v, _ = doc.Get("cursor")
				require.NotNil(t, v)

				cursor, ok = v.(*types.Document)
				require.True(t, ok)

				cursorID, _ = cursor.Get("id")
				assert.NotNil(t, cursorID)

				v, _ = cursor.Get("nextBatch")
				require.NotNil(t, v)

				nextBatch, ok := v.(*types.Array)
				require.True(t, ok)

				require.Equal(t, len(tc.nextBatch), nextBatch.Len(), "expected: %v, got: %v", tc.nextBatch, nextBatch)
				for i, elem := range tc.nextBatch {
					require.Equal(t, elem, must.NotFail(nextBatch.Get(i)))
				}
			}
		})
	}
}

func TestQueryCommandGetMoreConnection(t *testing.T) {
	t.Parallel()

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

	arr, _ := generateDocuments(0, 5)
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

		doc := ConvertDocument(t, res)

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

	t.Run("DifferentClient", func(t *testing.T) {
		// The error returned from MongoDB is a session error, FerretDB does not
		// return an error because db, collection and username are the same.
		setup.SkipExceptMongoDB(t, "https://github.com/FerretDB/FerretDB/issues/153")

		// do not run subtest in parallel to avoid breaking another parallel subtest

		u, err := url.Parse(s.MongoDBURI)
		require.NoError(t, err)

		client2, err := mongo.Connect(ctx, options.Client().ApplyURI(u.String()))
		require.NoError(t, err)

		defer client2.Disconnect(ctx)

		collection2 := client2.Database(databaseName).Collection(collectionName)

		var res bson.D
		err = collection1.Database().RunCommand(
			ctx,
			bson.D{
				{"find", collection1.Name()},
				{"batchSize", 2},
			},
		).Decode(&res)
		require.NoError(t, err)

		doc := ConvertDocument(t, res)

		v, _ := doc.Get("cursor")
		require.NotNil(t, v)

		cursor, ok := v.(*types.Document)
		require.True(t, ok)

		cursorID, _ := cursor.Get("id")
		assert.NotNil(t, cursorID)

		err = collection2.Database().RunCommand(
			ctx,
			bson.D{
				{"getMore", cursorID},
				{"collection", collection2.Name()},
			},
		).Decode(&res)

		// use AssertMatchesCommandError because message cannot be compared as it contains session ID
		AssertMatchesCommandError(
			t,
			mongo.CommandError{
				Code: 50738,
				Name: "Location50738",
				Message: "Cannot run getMore on cursor 5720627396082469624, which was created in session " +
					"95326129-ff9c-48a4-9060-464b4ea3ee06 - 47DEQpj8HBSa+/TImW+5JC\neuQeRkm5NMpJWZG3hSuFU= -  - , " +
					"in session 9e8902e9-338c-4156-9fd8-50e5d62ac992 - 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU= -  - ",
			},
			err,
		)
	})
}
