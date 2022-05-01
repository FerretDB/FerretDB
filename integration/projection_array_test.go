package integration

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestProjectionQuerySlice(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)
	_, err := collection.InsertOne(ctx,
		bson.D{{"_id", "array"}, {"value", bson.A{1, 2, 3, 4}}},
	)
	require.NoError(t, err)

	type testCase struct {
		projection    bson.D
		expectedArray bson.A
		err           *mongo.CommandError
	}

	for name, tc := range map[string]testCase{
		//"SingleArgString": {
		//	projection:    bson.D{{"value", bson.D{{"$slice", "string"}}}},
		//	expectedArray: nil,
		//	err: &mongo.CommandError{
		//		Code: 28667,
		//		Name: "Location28667",
		//		Message: "Invalid $slice syntax. The given syntax { $slice: \"string\" } " +
		//			"did not match the find() syntax because :: Location31273: " +
		//			"$slice only supports numbers and [skip, limit] arrays " +
		//			":: The given syntax did not match the expression $slice syntax. " +
		//			":: caused by :: Expression $slice takes at least 2 arguments, and at most 3, " +
		//			"but 1 were passed in.",
		//	},
		//},
		//"SingleArgDocument": {
		//	projection:    bson.D{{"value", bson.D{{"$slice", bson.D{"a", 3}}}}},
		//	expectedArray: nil,
		//	err: &mongo.CommandError{
		//		Code: 28667,
		//		Name: "Location28667",
		//		Message: "Invalid $slice syntax. The given syntax { $slice: { a: 3 } } " +
		//			"did not match the find() syntax because :: Location31273: " +
		//			"$slice only supports numbers and [skip, limit] arrays " +
		//			":: The given syntax did not match the expression $slice syntax. " +
		//			":: caused by :: Expression $slice takes at least 2 arguments, and at most 3, " +
		//			"but 1 were passed in.",
		//	},
		//},
		"SkipIsString": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{"string", 5}}}}},
			expectedArray: nil,
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: string",
			},
		},
		"LimitIsString": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{int32(2), "string"}}}}},
			expectedArray: nil,
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: int",
			},
		},
		"ArgEmptyArr": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{}}}}},
			expectedArray: nil,
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. The given syntax { $slice: [] } " +
					"did not match the find() syntax because :: Location31272: " +
					"$slice array argument should be of form [skip, limit] :: " +
					"The given syntax did not match the expression " +
					"$slice syntax. :: caused by :: " +
					"Expression $slice takes at least 2 arguments, and at most 3, but 0 were passed in.",
			},
		},
		"TooManyArgs": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{1, 2, 3, 4}}}}},
			expectedArray: nil,
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. The given syntax { $slice: [ 1, 2, 3, 4 ] } " +
					"did not match the find() syntax because :: Location31272: " +
					"$slice array argument should be of form [skip, limit] :: " +
					"The given syntax did not match the expression " +
					"$slice syntax. :: caused by :: " +
					"Expression $slice takes at least 2 arguments, and at most 3, but 4 were passed in.",
			},
		},
		"PositiveSingleArg": {
			projection:    bson.D{{"value", bson.D{{"$slice", 2}}}},
			expectedArray: bson.A{int32(1), int32(2)},
		},
		"NegativeSingleArg": {
			projection:    bson.D{{"value", bson.D{{"$slice", -2}}}},
			expectedArray: bson.A{int32(3), int32(4)},
		},
		"SingleArgFloat": {
			projection:    bson.D{{"value", bson.D{{"$slice", 1.4}}}},
			expectedArray: bson.A{int32(1)},
		},
		"SkipFloat": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{-2.5, 2}}}}},
			expectedArray: bson.A{int32(3), int32(4)},
		},
		"LimitFloat": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{1, 2.8}}}}},
			expectedArray: bson.A{int32(2), int32(3)},
		},
		"PositiveSkip": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{1, 2}}}}},
			expectedArray: bson.A{int32(2), int32(3)},
		},
		"NegativeSkip": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{-3, 2}}}}},
			expectedArray: bson.A{int32(2), int32(3)},
		},
		"NegativeLimitSkipInt": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{3, -2}}}}},
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: int",
			},
		},
		"NegativeLimitSkipFloat": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{0.3, -2}}}}},
			err: &mongo.CommandError{
				Code:    28724,
				Name:    "Location28724",
				Message: "First argument to $slice must be an array, but is of type: double",
			},
		},
		"ArgNaN": {
			projection:    bson.D{{"value", bson.D{{"$slice", math.NaN()}}}},
			expectedArray: bson.A{},
		},
		"ArgInf": {
			projection:    bson.D{{"value", bson.D{{"$slice", math.Inf(1)}}}},
			expectedArray: bson.A{int32(1), int32(2), int32(3), int32(4)},
		},
		"SingleArgNull": {
			projection: bson.D{{"value", bson.D{{"$slice", nil}}}},
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. " +
					"The given syntax { $slice: null } did not match the find() syntax " +
					"because :: Location31273: $slice only supports numbers and [skip, limit] arrays :: " +
					"The given syntax did not match the expression $slice syntax. :: caused by :: " +
					"Expression $slice takes at least 2 arguments, and at most 3, but 1 were passed in.",
			},
		},
		"NullInArr": {
			projection: bson.D{{"value", bson.D{{"$slice", bson.A{nil}}}}},
			err: &mongo.CommandError{
				Code: 28667,
				Name: "Location28667",
				Message: "Invalid $slice syntax. The given syntax { $slice: [ null ] } " +
					"did not match the find() syntax because :: Location31272: " +
					"$slice array argument should be of form [skip, limit] " +
					":: The given syntax did not match the expression $slice syntax. " +
					":: caused by :: Expression $slice takes at least 2 arguments, " +
					"and at most 3, but 1 were passed in.",
			},
		},
		"NullInPair": {
			projection:    bson.D{{"value", bson.D{{"$slice", bson.A{2, nil}}}}},
			expectedArray: nil,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			cursor, err := collection.Find(
				ctx, bson.D{},
				options.Find().SetProjection(tc.projection),
			)

			if tc.err != nil {
				require.Nil(t, tc.expectedArray)
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)

			if tc.expectedArray == nil {
				assert.Nil(t, actual[0].Map()["value"])
			} else {
				assert.Equal(t, tc.expectedArray, actual[0].Map()["value"])
			}
		})
	}
}
