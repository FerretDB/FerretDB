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
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestGetMore(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		err        *mongo.CommandError
		altMessage string
		command    bson.D
	}{
		"BatchSizeNegative": {
			command: bson.D{
				{"getMore", collection.Name()},
				{"batchSize", int32(-1)},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
		},
		"BatchSizeZero": {
			command: bson.D{
				{"getMore", int64(0)},
				{"batchSize", int32(0)},
				{"collection", collection.Name()},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '0'",
			},
		},
		"BatchSizeDocument": {
			command: bson.D{
				{"getMore", collection.Name()},
				{"batchSize", bson.D{}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'FindCommandRequest.batchSize' is the wrong type 'object', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'batchSize' is the wrong type 'object', expected type 'int'",
		},
		"BatchSizeMaxInt32": {
			command: bson.D{
				{"getMore", collection.Name()},
				{"batchSize", math.MaxInt32},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

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
