package oplog

import (
	"fmt"
	"testing"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestOplogUpdate(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t, shareddata.Composites)
	local := coll.Database().Client().Database("local")
	ns := fmt.Sprintf("%s.%s", coll.Database().Name(), coll.Name())
	opts := options.FindOne().SetSort(bson.D{{"$natural", -1}})

	err := local.CreateCollection(ctx, "oplog.rs", options.CreateCollection().SetCapped(true).SetSizeInBytes(536870912))
	if err != nil {
		require.Contains(t, err.Error(), "local.oplog.rs already exists")
		err = nil
	}

	require.NoError(t, err)

	for name, tc := range map[string]struct {
		update         bson.D
		filter         bson.D
		expectedDiffV1 *types.Document
		expectedDiffV2 *types.Document
	}{
		"set": {
			update:         bson.D{{"$set", bson.D{{"a", int32(1)}}}},
			filter:         bson.D{{"_id", "array"}},
			expectedDiffV1: must.NotFail(types.NewDocument("a", int32(1))),
			expectedDiffV2: must.NotFail(types.NewDocument("i", must.NotFail(types.NewDocument("a", int32(1))))),
		},
		/*"unset": {
			update:         bson.D{{"$unset", bson.D{{"a", 1}}}},
			expectedDiffV1: must.NotFail(types.NewDocument("$unset", must.NotFail(types.NewDocument("a", int64(1))))),
			expectedDiffV2: must.NotFail(types.NewDocument("d", must.NotFail(types.NewDocument("a", int64(1))))),
		},
		"inc": {
			update: bson.D{{"$inc", bson.D{{"a", 1}}}},
		},
		"mul": {
			update: bson.D{{"$mul", bson.D{{"a", 2}}}},
		},
		"rename": {
			update: bson.D{{"$rename", bson.D{{"a", "b"}}}},
		},
		"pull": {
			update: bson.D{{"$pull", bson.D{{"a", 1}}}},
		},
		"push": {
			update: bson.D{{"$push", bson.D{{"a", 1}}}},
		},*/
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Run("UpdateOne", func(t *testing.T) {
				_, err = coll.UpdateOne(ctx, tc.filter, tc.update)
				require.NoError(t, err)

				var lastOplogEntry bson.D
				err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
				require.NoError(t, err)

				actual := integration.ConvertDocument(t, lastOplogEntry)

				o := must.NotFail(actual.Get("o")).(*types.Document)
				version := must.NotFail(o.Get("$v")).(int32)
				if version == 2 {
					diff := must.NotFail(o.Get("diff")).(*types.Document)
					assert.Equal(t, tc.expectedDiffV2, diff)
				} else {
					diff := must.NotFail(o.Get("$set")).(*types.Document)
					assert.Equal(t, tc.expectedDiffV1, diff)

					o2 := must.NotFail(o.Get("o2")).(*types.Document)
					expectedO2 := must.NotFail(types.NewDocument("_id", must.NotFail(o2.Get("_id"))))
					assert.Equal(t, expectedO2, o2)
				}

				/*	o := must.NotFail(actual.Get("o")).(*types.Document)
					if o.Has("diff") {
						diff := must.NotFail(o.Get("diff")).(*types.Document)
						assert.Equal(t, tc.expectedDiffV2, diff)
					} else {
						diff := must.NotFail(o.Get("$set")).(*types.Document)
						assert.Equal(t, tc.expectedDiffV1, diff)

						o2 := must.NotFail(o.Get("o2")).(*types.Document)
						expectedO2 := must.NotFail(types.NewDocument("_id", must.NotFail(o2.Get("_id"))))
						assert.Equal(t, expectedO2, o2)
					}*/

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
				assert.Equal(t, expected, actual)
			})
		})
	}
}
