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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestBatchSize(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	// the number of documents is set above the default batchSize of 101
	// for testing unset batchSize returning default batchSize
	arr, _ := integration.GenerateDocuments(0, 110)
	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		filter    any // optional, nil to leave filter unset
		batchSize any // optional, nil to leave batchSize unset

		firstBatch primitive.A         // optional, expected firstBatch
		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"Int": {
			batchSize:  1,
			firstBatch: arr[:1],
		},
		"Long": {
			batchSize:  int64(2),
			firstBatch: arr[:2],
		},
		"LongZero": {
			batchSize:  int64(0),
			firstBatch: bson.A{},
		},
		"LongNegative": {
			batchSize: int64(-1),
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
			altMessage: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
		},
		"DoubleZero": {
			batchSize:  float64(0),
			firstBatch: bson.A{},
		},
		"DoubleNegative": {
			batchSize: -1.1,
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"DoubleFloor": {
			batchSize:  1.9,
			firstBatch: arr[:1],
		},
		"Bool": {
			batchSize:  true,
			firstBatch: arr[:1],
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'FindCommandRequest.batchSize' is the wrong type 'bool', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'find.batchSize' is the wrong type 'bool', expected types '[long, int, decimal, double]'",
		},
		"Unset": {
			// default batchSize is 101 when unset
			batchSize:  nil,
			firstBatch: arr[:101],
		},
		"LargeBatchSize": {
			batchSize:  102,
			firstBatch: arr[:102],
		},
		"LargeBatchSizeFilter": {
			filter:     bson.D{{"_id", bson.D{{"$in", bson.A{0, 1, 2, 3, 4, 5}}}}},
			batchSize:  102,
			firstBatch: arr[:6],
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			var rest bson.D
			if tc.filter != nil {
				rest = append(rest, bson.E{Key: "filter", Value: tc.filter})
			}

			if tc.batchSize != nil {
				rest = append(rest, bson.E{Key: "batchSize", Value: tc.batchSize})
			}

			command := append(
				bson.D{{"find", collection.Name()}},
				rest...,
			)

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)
			if tc.err != nil {
				assert.Nil(t, res)
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

				return
			}

			require.NoError(t, err)

			v, ok := res.Map()["cursor"]
			require.True(t, ok)

			cursor, ok := v.(bson.D)
			require.True(t, ok)

			// Do not check the value of cursor id, FerretDB has a different id.
			cursorID := cursor.Map()["id"]
			assert.NotNil(t, cursorID)

			firstBatch, ok := cursor.Map()["firstBatch"]
			require.True(t, ok)
			require.Equal(t, tc.firstBatch, firstBatch)
		})
	}
}

func TestSingleBatch(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	arr, _ := integration.GenerateDocuments(0, 5)
	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		batchSize   any // optional, nil to leave batchSize unset
		singleBatch any // optional, nil to leave singleBatch unset

		cursorClosed bool                // optional, set true for expecting cursor to be closed
		err          *mongo.CommandError // optional, expected error from MongoDB
		altMessage   string              // optional, alternative error message for FerretDB, ignored if empty
		skip         string              // optional, skip test with a specified reason
	}{
		"True": {
			singleBatch:  true,
			batchSize:    3,
			cursorClosed: true,
		},
		"False": {
			singleBatch:  false,
			batchSize:    3,
			cursorClosed: false,
		},
		"Int": {
			singleBatch: int32(1),
			batchSize:   3,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Field 'singleBatch' should be a boolean value, but found: int",
			},
			altMessage: "BSON field 'find.singleBatch' is the wrong type 'int', expected type 'bool'",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			var rest bson.D
			if tc.batchSize != nil {
				rest = append(rest, bson.E{Key: "batchSize", Value: tc.batchSize})
			}

			if tc.singleBatch != nil {
				rest = append(rest, bson.E{Key: "singleBatch", Value: tc.singleBatch})
			}

			command := append(
				bson.D{{"find", collection.Name()}},
				rest...,
			)

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)
			if tc.err != nil {
				assert.Nil(t, res)
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

				return
			}

			require.NoError(t, err)

			v, ok := res.Map()["cursor"]
			require.True(t, ok)

			cursor, ok := v.(bson.D)
			require.True(t, ok)

			cursorID := cursor.Map()["id"]
			assert.NotNil(t, cursorID)

			if !tc.cursorClosed {
				assert.NotZero(t, cursorID)
				return
			}

			assert.Equal(t, int64(0), cursorID)
		})
	}
}
