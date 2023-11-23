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

func TestDiffDBStatsScale(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	testCases := map[string]struct {
		scale               any
		expectedMongoDBErr  string
		expectedFerretDBErr mongo.CommandError
	}{
		"Zero": {
			scale:              int32(0),
			expectedMongoDBErr: "scale has to be > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "Location51024",
				Code:    51024,
				Message: "BSON field 'scale' value must be >= 1, actual value '0'",
			},
		},
		"Negative": {
			scale:              int32(-100),
			expectedMongoDBErr: "scale has to be > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "Location51024",
				Code:    51024,
				Message: "BSON field 'scale' value must be >= 1, actual value '-100'",
			},
		},
		"MinFloat": {
			scale:              -math.MaxFloat64,
			expectedMongoDBErr: "scale has to be > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "Location51024",
				Code:    51024,
				Message: "BSON field 'scale' value must be >= 1, actual value '-9223372036854775808'",
			},
		},
		"String": {
			scale:              "1",
			expectedMongoDBErr: "scale has to be a number > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "TypeMismatch",
				Code:    14,
				Message: "BSON field 'dbStats.scale' is the wrong type 'string', expected types '[long, int, decimal, double]'",
			},
		},
		"Object": {
			scale:              bson.D{{"a", 1}},
			expectedMongoDBErr: "scale has to be a number > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "TypeMismatch",
				Code:    14,
				Message: "BSON field 'dbStats.scale' is the wrong type 'object', expected types '[long, int, decimal, double]'",
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := collection.Database().RunCommand(ctx, bson.D{{"dbStats", int32(1)}, {"scale", tc.scale}}).Err()
			require.Error(t, err)

			if setup.IsMongoDB(t) {
				expected := mongo.CommandError{
					Name:    "",
					Code:    0,
					Message: tc.expectedMongoDBErr,
				}
				AssertEqualCommandError(t, expected, err)
				return
			}

			AssertEqualCommandError(t, tc.expectedFerretDBErr, err)
		})
	}
}
