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
	_, err := collection.InsertMany(ctx, []any{})
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
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"PositiveInBounds": {
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"PositiveOutOfBounds": {
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"NegativeInBounds": {
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"NegativeOutOfBounds": {
				projection: nil,
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
			"InvalidNumberOfArgs": {
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkipInvalidType": {
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToReturnInvalidValue": { // can't be negative
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToReturnInvalidType": {
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkip>=0_ToReturn>Len": {
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkip>=0_ToReturn<=Len": {
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkip<0_ToReturn<=Len": {
				projection: nil,
				expected:   nil,
				err:        mongo.CommandError{},
			},
			"ToSkip<0_ToReturn>Len": {
				projection: nil,
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
