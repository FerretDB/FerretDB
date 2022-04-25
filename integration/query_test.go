package integration

import (
	"testing"

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
