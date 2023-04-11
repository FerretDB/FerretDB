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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryProjection(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Composites)

	for name, tc := range map[string]struct {
		projection any
		filter     any
		expected   bson.D
		err        bool // TODO check error type
	}{
		"FindProjectionIDExclusion": {
			filter: bson.D{{"_id", "document-composite"}},
			// TODO: https://github.com/FerretDB/FerretDB/issues/537
			projection: bson.D{{"_id", false}, {"array", int32(1)}},
			expected:   bson.D{},
		},
		"Invalid": {
			filter:     bson.D{},
			projection: bson.D{{"a", "$"}},
			err:        true,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetProjection(tc.projection))
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			require.Len(t, actual, 1)
			AssertEqualDocuments(t, tc.expected, actual[0])
		})
	}
}
