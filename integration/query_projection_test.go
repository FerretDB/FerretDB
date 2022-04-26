package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryProjectionElemMatch(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Composites}
	ctx, collection := setup(t, providers...)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{
			{"_id", "document-composite-2"},
			{"value", bson.A{
				bson.D{{"field", int32(42)}},
				bson.D{{"field", int32(44)}},
			}},
		},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		projection  any
		expectedIDs []any
	}{
		"ElemMatch": {
			projection: bson.D{{
				"value",
				bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$eq", 42}}}}}},
			}},
			expectedIDs: []any{"document-composite-2"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}
