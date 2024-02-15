package sessions

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestWithTransaction(tt *testing.T) {
	tt.Parallel()

	t := tt
	// t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/1554")

	ctx, collection := setup.Setup(t)
	client := collection.Database().Client()

	session, err := client.StartSession(options.Session().SetCausalConsistency(true))
	require.NoError(t, err)

	defer session.EndSession(ctx)

	res, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		_, err = collection.InsertOne(ctx, bson.D{{"foo", "bar"}})
		require.NoError(t, err)

		return nil, nil
	})
	require.NoError(t, err)

	fmt.Println(res)
}
