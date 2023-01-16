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

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
)

func TestGetMoreErrors(t *testing.T) {
	t.Skip("TODO: https://github.com/FerretDB/FerretDB/issues/1807")

	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Int32BigAmounts)

	for name, tc := range map[string]struct {
		id         any
		err        *mongo.CommandError
		altMessage string
		command    bson.D
	}{
		"InvalidCursorID": {
			id: int64(2),
			command: bson.D{
				{"collection", collection.Name()},
			},
			err: &mongo.CommandError{
				Code:    43,
				Name:    "CursorNotFound",
				Message: "cursor id 2 not found",
			},
		},
		"CursorIdInt32": {
			id: int32(1),
			command: bson.D{
				{"collection", collection.Name()},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.getMore' is the wrong type 'int', expected type 'long'",
			},
		},
		"CursorIDNegative": {
			id: int64(-1),
			command: bson.D{
				{"collection", collection.Name()},
			},
			err: &mongo.CommandError{
				Code:    43,
				Name:    "CursorNotFound",
				Message: "cursor id -1 not found",
			},
		},
		"BatchSizeNegative": {
			command: bson.D{
				{"batchSize", int64(-1)},
				{"collection", collection.Name()},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"BatchSizeDocument": {
			command: bson.D{
				{"batchSize", bson.D{}},
				{"collection", collection.Name()},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'getMore.batchSize' is the wrong type 'object', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'batchSize' is the wrong type 'object', expected type 'long'",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			id := tc.id
			if tc.id == nil {
				res := collection.Database().RunCommand(ctx, bson.D{{"find", collection.Name()}, {"filter", bson.D{}}})
				require.NoError(t, res.Err())

				var result bson.D
				err := res.Decode(&result)
				require.NoError(t, err)

				responseDoc := ConvertDocument(t, result)
				cursor, err := responseDoc.Get("cursor")
				require.NoError(t, err)

				id, err = cursor.(*types.Document).Get("id")
				require.NoError(t, err)
			}
			tc.command = append(bson.D{{"getMore", id}}, tc.command...)

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			if tc.err != nil {
				require.Error(t, err)
				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
