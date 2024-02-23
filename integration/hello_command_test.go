package integration

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/stretchr/testify/require"
)

func TestHelloFail(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)
	db := collection.Database()

	payload := ConvertDocument(t, bson.D{
		{"hello", "1"},
	})

	require.NoError(t, db.RunCommand(ctx, payload).Err())
}
