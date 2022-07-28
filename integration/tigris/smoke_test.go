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

package tigris

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestSmoke(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.FixedScalars)

	var doc bson.D
	err := collection.FindOne(ctx, bson.D{{"_id", "fixed_double"}}).Decode(&doc)
	require.NoError(t, err)
	integration.AssertEqualDocuments(t, bson.D{{"_id", "fixed_double"}, {"double_value", 42.13}}, doc)

	del, err := collection.DeleteOne(ctx, bson.D{{"_id", "fixed_double"}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), del.DeletedCount)

	ins, err := collection.InsertOne(ctx, bson.D{{"double_value", 123}})
	require.NoError(t, err)
	del, err = collection.DeleteOne(ctx, bson.D{{"_id", ins.InsertedID}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), del.DeletedCount)
}

// TestSmokeMsgCount implements simple smoke tests for MsgCount.
// TODO Implement proper testing: https://github.com/FerretDB/FerretDB/issues/931.
func TestSmokeMsgCount(t *testing.T) {
	t.Parallel()

	// As Tigris require different fields for different types,
	// for this smoke test we only use the fixed scalars.
	ctx, collection := setup.Setup(t, shareddata.FixedScalars)

	for name, tc := range map[string]struct {
		command  any
		response int32
	}{
		"CountAllFixedScalars": {
			command:  bson.D{{"count", collection.Name()}},
			response: 6,
		},
		"CountExactlyOneDocument": {
			command: bson.D{
				{"count", collection.Name()},
				{"query", bson.D{{"double_value", math.MaxFloat64}}},
			},
			response: 1,
		},
		"CountNonExistingCollection": {
			command: bson.D{
				{"count", "doesnotexist"},
				{"query", bson.D{{"v", true}}},
			},
			response: 0,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()

			assert.Equal(t, float64(1), m["ok"])

			keys := integration.CollectKeys(t, actual)
			assert.Contains(t, keys, "n")
			assert.Equal(t, tc.response, m["n"])
		})
	}
}
