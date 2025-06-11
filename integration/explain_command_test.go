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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestExplainCommandQueryErrors(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		command bson.D // required, command to run

		err              *mongo.CommandError // required, expected error from MongoDB
		altMessage       string              // optional, alternative error message for FerretDB, ignored if empty
		failsForFerretDB string
	}{
		"LimitDocument": {
			command: bson.D{
				{"explain", bson.D{
					{"find", collection.Name()},
					{"limit", bson.D{}},
				}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'FindCommandRequest.limit' is the wrong type 'object', expected types '[long, int, decimal, double']",
			},
			altMessage:       "BSON field 'limit' is the wrong type 'object', expected type 'long'",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/960",
		},
		"LimitNegative": {
			command: bson.D{
				{"explain", bson.D{
					{"find", collection.Name()},
					{"limit", int64(-1)},
				}},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'limit' value must be >= 0, actual value '-1'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/959",
		},
		"SkipDocument": {
			command: bson.D{
				{"explain", bson.D{
					{"find", collection.Name()},
					{"skip", bson.D{}},
				}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'FindCommandRequest.skip' is the wrong type 'object', expected types '[long, int, decimal, double']",
			},
			altMessage:       "BSON field 'skip' is the wrong type 'object', expected type 'long'",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/960",
		},
		"SkipNegative": {
			command: bson.D{
				{"explain", bson.D{
					{"find", collection.Name()},
					{"skip", int64(-1)},
				}},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'skip' value must be >= 0, actual value '-1'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/960",
		},
		"CollectionName": {
			command: bson.D{
				{"explain", bson.D{
					{"find", int32(1)},
				}},
			},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Failed to parse namespace element",
			},
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.command, "command must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			var res bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&res)

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestExplainNonExistentCollection(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	var res bson.D
	err := collection.Database().RunCommand(ctx, bson.D{
		{"explain", bson.D{
			{"find", "non-existent"},
		}},
	}).Decode(&res)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestExplainLimitInt(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	var res bson.D
	err := collection.Database().RunCommand(ctx, bson.D{
		{"explain", bson.D{
			{"find", collection.Name()},
			{"limit", int32(1)},
		}},
	}).Decode(&res)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}
