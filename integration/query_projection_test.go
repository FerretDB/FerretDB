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

func TestQueryProjection(t *testing.T) {
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
		projection any
		filter     any
		expected   bson.D
	}{
		"FindProjectionInclusions": {
			filter: bson.D{{"_id", "document-composite"}},
			// TODO: https://github.com/FerretDB/FerretDB/issues/537
			projection: bson.D{{"foo", int32(1)}, {"42", true}},
			expected:   bson.D{{"_id", "document-composite"}},
		},
		"FindProjectionExclusions": {
			filter: bson.D{{"_id", "document-composite"}},
			// TODO: https://github.com/FerretDB/FerretDB/issues/537
			projection: bson.D{{"foo", int32(0)}, {"array", false}},
			expected:   bson.D{{"_id", "document-composite"}, {"value", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}},
		},
		"FindProjectionIDExclusion": {
			filter: bson.D{{"_id", "document-composite"}},
			// TODO: https://github.com/FerretDB/FerretDB/issues/537
			projection: bson.D{{"_id", false}, {"array", int32(1)}},
			expected:   bson.D{},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetProjection(tc.projection))
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			require.Len(t, actual, 1)
			AssertEqualDocuments(t, tc.expected, actual[0])
		})
	}
}

func TestQueryProjectionElemMatch(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Composites}
	ctx, collection := setup(t, providers...)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{
			{"_id", "document-composite-2"},
			{"value", bson.A{
				bson.D{{"field", int32(40)}},
				bson.D{{"field", int32(42)}},
				bson.D{{"field", int32(44)}},
				bson.D{{"field", bson.D{{"field", int32(42)}}}},
			}},
		},
		bson.D{
			{"_id", "document-composite-3"},
			{"value", bson.A{
				bson.D{{"field", int32(40)}},
				bson.D{{"field", bson.D{{"field", int32(41)}}}},
				bson.D{{"field", bson.D{{"field", int32(42)}}}},
			}},
		},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filterID    string
		projection  any
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"ElemMatch": {
			filterID: "document-composite-2",
			projection: bson.D{{
				"value",
				bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$eq", 42}}}}}},
			}},
			expectedIDs: []any{
				"document-composite-2",
			},
		},
		"ElemMatchNested1": {
			filterID: "document-composite-3",
			projection: bson.D{{
				"value",
				bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"field", bson.D{{"$eq", 42}}}}}}}},
			}},
			expectedIDs: []any{
				"document-composite-3",
			},
		},
		"ElemMatchNested2": {
			filterID: "document-composite-3",
			projection: bson.D{{
				"value",
				bson.D{{"field", bson.D{{"field", bson.D{{"$elemMatch", bson.D{{"$eq", 42}}}}}}}},
			}},
			err: &mongo.CommandError{
				Code:    31275,
				Name:    "Location31275",
				Message: "Cannot use $elemMatch projection on a nested field.",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(
				ctx,
				bson.D{{"_id", tc.filterID}},
				options.Find().SetProjection(tc.projection),
				options.Find().SetSort(bson.D{{"_id", 1}}),
			)

			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualError(t, *tc.err, err)
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
