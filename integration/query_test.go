package integration

import (
	"testing"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestUnknownFilterOperator(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars)

	filter := bson.D{{"value", bson.D{{"$aboba", 42}}}}
	errExpected := mongo.CommandError{Code: 2, Name: "BadValue", Message: "unknown operator: $aboba"}
	_, err := collection.Find(ctx, filter)
	AssertEqualError(t, errExpected, err)
}
