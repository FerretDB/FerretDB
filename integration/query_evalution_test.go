package integration

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestEvalMod(t *testing.T) {
	//	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "float64_1"}, {"value", float64(113.01)}},
		bson.D{{"_id", "float64_2"}, {"value", float64(114.99)}},
		bson.D{{"_id", "float64_3"}, {"value", float64(115.5)}},
		bson.D{{"_id", "int64_1"}, {"value", int64(141)}},
		bson.D{{"_id", "int64_2"}, {"value", int64(151)}},
		bson.D{{"_id", "int64_3"}, {"value", int64(161)}},
		bson.D{{"_id", "int32_1"}, {"value", int32(177)}},
		bson.D{{"_id", "int32_2"}, {"value", int32(178)}},
		bson.D{{"_id", "int32_3"}, {"value", int32(179)}},
		bson.D{{"_id", "nil"}, {"value", nil}},
		bson.D{{"_id", "string"}, {"value", "12"}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		q           bson.D
		expectedIDs []any
		err         error
	}{
		"Float64_1": {
			q:           bson.D{{"value", bson.D{{"$mod", bson.A{10, 3}}}}},
			expectedIDs: []any{"float64_1"},
		},
		"Float64_2": {
			q:           bson.D{{"value", bson.D{{"$mod", bson.A{10, float64(4.5)}}}}},
			expectedIDs: []any{"float64_2"},
		},
		"Float64_3": {
			q:           bson.D{{"value", bson.D{{"$mod", bson.A{float64(10.5), 5}}}}},
			expectedIDs: []any{"float64_3"},
		},
		"Int64_1": {
			q:           bson.D{{"value", bson.D{{"$mod", bson.A{70, 1}}}}},
			expectedIDs: []any{"int64_1"},
		},
		"Int64_2": {
			q:           bson.D{{"value", bson.D{{"$mod", bson.A{float64(70.5), 11}}}}},
			expectedIDs: []any{"int64_2"},
		},
		"Int64_3": {
			q:           bson.D{{"value", bson.D{{"$mod", bson.A{70, float64(21.99)}}}}},
			expectedIDs: []any{"int64_3"},
		},
		"Int32_1": {
			q:           bson.D{{"value", bson.D{{"$mod", bson.A{80, 17}}}}},
			expectedIDs: []any{"int32_1"},
		},
		"Int32_2": {
			q:           bson.D{{"value", bson.D{{"$mod", bson.A{float64(80.5), 18}}}}},
			expectedIDs: []any{"int32_2"},
		},
		"Int32_3": {
			q:           bson.D{{"value", bson.D{{"$mod", bson.A{80, float64(19.09)}}}}},
			expectedIDs: []any{"int32_3"},
		},
		"EmptyArray": {
			q: bson.D{{"value", bson.D{{"$mod", bson.A{}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, not enough elements`,
			},
		},
		"NotEnoughElements": {
			q: bson.D{{"value", bson.D{{"$mod", bson.A{1}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, not enough elements`,
			},
		},
		"TooManyElements": {
			q: bson.D{{"value", bson.D{{"$mod", bson.A{1, 2, 3}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, too many elements`,
			},
		},
		"DivisorNotNumber": {
			q: bson.D{{"value", bson.D{{"$mod", bson.A{"1", 2}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor not a number`,
			},
		},
		"RemainderNotNumber": {
			q: bson.D{{"value", bson.D{{"$mod", bson.A{1, "2"}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, remainder not a number`,
			},
		},
		"Nil": {
			q: bson.D{{"value", bson.D{{"$mod", bson.A{nil, 3}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor not a number`,
			},
		},
		"NaN": {
			q: bson.D{{"value", bson.D{{"$mod", bson.A{math.NaN(), 1}}}}},
			err: mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"Infinity": {
			q: bson.D{{"value", bson.D{{"$mod", bson.A{1, math.Inf(1)}}}}},
			err: mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, remainder value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"InvalidUse": {
			q: bson.D{{"$mod", bson.A{1, 1}}},
			err: mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `unknown top level operator: $mod. ` +
					`If you have a field name that starts with a '$' symbol, consider using $getField or $setField.`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			//			t.Parallel()

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.q)
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				require.Equal(t, tc.err, err)
				return
			}
			require.NoError(t, err)
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, collectIDs(t, actual))
		})
	}
}
