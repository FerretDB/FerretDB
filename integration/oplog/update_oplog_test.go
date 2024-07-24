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

package oplog

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestOplogUpdate(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t, shareddata.Composites)
	local := coll.Database().Client().Database("local")
	ns := fmt.Sprintf("%s.%s", coll.Database().Name(), coll.Name())
	opts := options.FindOne().SetSort(bson.D{{"$natural", -1}})

	if err := local.CreateCollection(ctx, "oplog.rs", options.CreateCollection().SetCapped(true).SetSizeInBytes(536870912)); err != nil {
		require.Contains(t, err.Error(), "local.oplog.rs already exists")
	}

	for name, tc := range map[string]struct { //nolint:vet // for readability
		update         bson.D
		filter         bson.D
		expectedDiffV1 *types.Document
		expectedO2     *types.Document
		expectedDiffV2 *types.Document
	}{
		"set": {
			update: bson.D{{"$set", bson.D{{"a", int32(1)}}}},
			filter: bson.D{{"_id", "array"}},
			expectedDiffV1: must.NotFail(types.NewDocument(
				"_id", "array",
				"v", must.NotFail(types.NewArray(int32(42))),
				"a", int32(1),
			)),
			expectedO2:     must.NotFail(types.NewDocument("_id", "array")),
			expectedDiffV2: must.NotFail(types.NewDocument("i", must.NotFail(types.NewDocument("a", int32(1))))),
		},
		"unset": {
			update:         bson.D{{"$unset", bson.D{{"v", int32(1)}}}},
			filter:         bson.D{{"_id", "array-two"}},
			expectedDiffV1: must.NotFail(types.NewDocument("_id", "array-two")),
			expectedO2:     must.NotFail(types.NewDocument("_id", "array-two")),
			expectedDiffV2: must.NotFail(types.NewDocument("d", must.NotFail(types.NewDocument("v", false)))),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			// Subtests are not run in parallel because we need to preserve oplog entries.

			_, err := coll.UpdateOne(ctx, tc.filter, tc.update)
			require.NoError(t, err)

			var lastOplogEntry bson.D
			err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
			require.NoError(t, err)

			actual := integration.ConvertDocument(t, lastOplogEntry)

			o := must.NotFail(actual.Get("o")).(*types.Document)
			version := must.NotFail(o.Get("$v")).(int32)
			switch version {
			case 1:
				diff := must.NotFail(o.Get("$set")).(*types.Document)
				assert.Equal(t, tc.expectedDiffV1, diff)

				o2 := must.NotFail(actual.Get("o2")).(*types.Document)
				assert.Equal(t, tc.expectedO2, o2)
			case 2:
				diff := must.NotFail(o.Get("diff")).(*types.Document)
				assert.Equal(t, tc.expectedDiffV2, diff)
			default:
				t.Fatalf("unexpected version %d", version)
			}

			unsetUnusedOplogFields(actual)
			actual.Remove("o")
			actual.Remove("o2")
			expected, err := types.NewDocument(
				"op", "u",
				"ns", ns,
				"ts", must.NotFail(actual.Get("ts")).(types.Timestamp),
				"v", int64(2),
			)
			require.NoError(t, err)
			assert.EqualValues(t, expected, actual)
		})
	}
}
