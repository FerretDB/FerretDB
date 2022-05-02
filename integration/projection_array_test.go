package integration

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectionQueryDollar(t *testing.T) {
	ctx, collection := setup(t)

	must.NotFail(collection.InsertOne(ctx,
		bson.D{
			{"_id", "array-embedded"},
			{"value", bson.A{bson.D{
				{"document", "abc"},
				{"score", 42.13},
			}, bson.D{
				{"document", "def"},
				{"score", 40},
			}, bson.D{
				{"document", "jkl"},
				{"score", 24},
			}}},
		}))
	for name, tc := range map[string]struct {
		projection    bson.D
		expectedArray bson.A
		err           *mongo.CommandError
	}{
		"Null": {
			projection:    bson.D{{"value.$", nil}},
			expectedArray: nil,
			err: &mongo.CommandError{
				Code:    31308,
				Name:    "Location31308",
				Message: "positional projection cannot be used with a literal",
			},
		},
		"String": {
			projection:    bson.D{{"value.$", "a"}},
			expectedArray: nil,
			err: &mongo.CommandError{
				Code:    31308,
				Name:    "Location31308",
				Message: "positional projection cannot be used with a literal",
			},
		},
		"Exclusion": {
			projection:    bson.D{{"value.$", 0}},
			expectedArray: nil,
			err: &mongo.CommandError{
				Code:    31395,
				Name:    "Location31395",
				Message: "positional projection cannot be used with exclusion",
			},
		},
		"Document": {
			projection:    bson.D{{"value.$", bson.D{{"a", 19}}}},
			expectedArray: nil,
			err: &mongo.CommandError{
				Code:    31271,
				Name:    "Location31271",
				Message: "positional projection cannot be used with an expression or sub object",
			},
		},
		"Array": {
			projection:    bson.D{{"value.$", bson.A{1, 2, 3, 4}}},
			expectedArray: nil,
			err: &mongo.CommandError{
				Code:    31308,
				Name:    "Location31308",
				Message: "positional projection cannot be used with a literal",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			var doc bson.D
			err := collection.FindOne(
				ctx, bson.D{},
				options.FindOne().SetProjection(tc.projection),
			).Decode(&doc)

			if tc.err != nil {
				require.Nil(t, tc.expectedArray)
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedArray, doc.Map()["value"])
		})
	}
}
