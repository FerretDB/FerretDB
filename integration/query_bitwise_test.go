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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryBitwiseAllClear(t *testing.T) {
	t.Skip("TODO https://github.com/FerretDB/FerretDB/issues/442")

	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars)

	for name, tc := range map[string]struct {
		v           any
		expectedIDs []any
		err         mongo.CommandError
	}{
		"int32": {
			v: int32(2),
			expectedIDs: []any{
				"binary-empty", "double-negative-zero", "double-zero",
				"int32-min", "int32-zero", "int64-min", "int64-zero",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			filter := bson.D{{"value", bson.D{{"$bitsAllClear", tc.v}}}}
			cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err.Code != 0 {
				require.Nil(t, tc.expectedIDs)
				AssertEqualError(t, tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}
