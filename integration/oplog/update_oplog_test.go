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
		update        bson.D
		expectedOplog *types.Document
	}{
		"set": {
			update: bson.D{{"$set", bson.D{{"a", 1}}}},
		},
		"unset": {
			update: bson.D{{"$unset", bson.D{{"a", 1}}}},
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
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Run("UpdateOne", func(t *testing.T) {
				_, err = coll.UpdateOne(ctx, bson.D{}, tc.update)
				require.NoError(t, err)

				var lastOplogEntry bson.D
				err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
				require.NoError(t, err)

				actual := integration.ConvertDocument(t, lastOplogEntry)

				if must.NotFail(actual.Get("v")).(int64) == 2 {
					_ = convertDiff(t, must.NotFail(actual.Get("o")).(*types.Document))
				}

				assert.Equal(t, tc.expectedOplog, actual)
			})

			t.Run("UpdateMany", func(t *testing.T) {
			})
		})
	}
}

// convertDiff converts V2 oplog diff document to V1 format.
func convertDiff(t *testing.T, v2 *types.Document) *types.Document {
	v1 := must.NotFail(types.NewDocument())

	switch must.NotFail(v2.Get("op")).(string) {
	case "i":
		v1.Set("$set", must.NotFail(v2.Get("o")).(*types.Document))
	case "u":
		v1.Set("$set", must.NotFail(v2.Get("o2")).(*types.Document))
		v1.Set("$set", must.NotFail(v2.Get("o")).(*types.Document))
	case "d":
		v1.Set("$unset", must.NotFail(v2.Get("o")).(*types.Document))

	}

	/*
		iter := v2.Iterator()

		for {
			k, v, err := iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			require.NoError(t, err)

			switch k {
			case "":
			}
		}
	*/

	return v1
}
