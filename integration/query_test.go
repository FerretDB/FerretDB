package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestUnknownFilterOperator(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars)

	filter := bson.D{{"value", bson.D{{"$someUnknownOperator", 42}}}}
	errExpected := mongo.CommandError{Code: 2, Name: "BadValue", Message: "unknown operator: $someUnknownOperator"}
	_, err := collection.Find(ctx, filter)
	AssertEqualError(t, errExpected, err)
}

func TestQueryCount(t *testing.T) {
	t.Parallel()
	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	ctx, collection := setup(t, providers...)

	for name, tc := range map[string]struct {
		command  any
		response int32
	}{
		"CountAllDocuments": {
			command:  bson.D{{"count", collection.Name()}},
			response: 43,
		},
		"CountExactlyOneDocument": {
			command: bson.D{
				{"count", collection.Name()},
				{"query", bson.D{{"value", true}}},
			},
			response: 1,
		},
		"CountArrays": {
			command: bson.D{
				{"count", collection.Name()},
				{"query", bson.D{{"value", bson.D{{"$type", "array"}}}}},
			},
			response: 4,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()

			assert.Equal(t, 1.0, m["ok"])

			keys := CollectKeys(t, actual)
			assert.Contains(t, keys, "n")
			assert.Equal(t, tc.response, m["n"])
		})
	}
}
