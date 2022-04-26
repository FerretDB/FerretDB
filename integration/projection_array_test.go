package integration

import (
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"testing"
)

func TestProjectionQuerySlice(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)
	_, err := collection.InsertOne(ctx, []any{
		bson.D{{"_id", "array"}, {"value", bson.A{1, 2, 3, 4}}},
	})
	require.NoError(t, err)
	type testCase struct {
		projection bson.D
		expected   *types.Array
		err        mongo.CommandError
	}

	t.Run("SingleArg", func(t *testing.T) {
		t.Parallel()
		for name, tc := range map[string]testCase{
			"InvalidType": {
				projection: bson.D{{"value", bson.D{{"$slice", "string"}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"Positive<Len": {
				projection: bson.D{{"value", bson.D{{"$slice", 2}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"Positive>=Len": {
				projection: bson.D{{"value", bson.D{{"$slice", 10}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"NegativeAbs>Len": {
				projection: bson.D{{"value", bson.D{{"$slice", -10}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"NegativeAbs=Len": {
				projection: bson.D{{"value", bson.D{{"$slice", -4}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"NegativeAbs<Len": {
				projection: bson.D{{"value", bson.D{{"$slice", -3}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()
			})
		}
	})
	t.Run("MultipleArgs", func(t *testing.T) {
		t.Parallel()
		for name, tc := range map[string]testCase{
			// $slice: [ <number to skip>, <number to return> ]
			"InvalidNumberOfArgs": {
				projection: bson.D{{"value", bson.D{{"$slice", bson.A{1, 2, 3}}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkipInvalidType": {
				projection: bson.D{{"value", bson.D{{"$slice", bson.A{"string", 2}}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToReturnInvalidValue": { // can't be negative
				projection: bson.D{{"value", bson.D{{"$slice", bson.A{1, -2}}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToReturnInvalidType": {
				projection: bson.D{{"value", bson.D{{"$slice", bson.A{1, "string"}}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkip>=0_ToReturn>Len": {
				projection: bson.D{{"value", bson.D{{"$slice", bson.A{2, 5}}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkip>=0_ToReturn<=Len": {
				projection: bson.D{{"value", bson.D{{"$slice", bson.A{2, 3}}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkip<0_ToReturn<=Len": {
				projection: bson.D{{"value", bson.D{{"$slice", bson.A{-2, 4}}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkip<0_ToReturn>Len": {
				projection: bson.D{{"value", bson.D{{"$slice", bson.A{-3, 10}}}}},
				expected:   nil,
				err:        mongo.CommandError{},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()
			})
		}
	})

}
