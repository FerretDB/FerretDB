package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestHello(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)
	db := collection.Database()

	var res bson.D

	require.NoError(t, db.RunCommand(ctx, bson.D{
		{"hello", "1"},
	}).Decode(&res))

	actual := ConvertDocument(t, res)

	assert.Equal(t, actual.Keys(), []string{
		"isWritablePrimary",
		"maxBsonObjectSize",
		"maxMessageSizeBytes",
		"maxWriteBatchSize",
		"localTime",
		"connectionId",
		"minWireVersion",
		"maxWireVersion",
		"readOnly",
		"ok",
	})
}

func TestHelloWithSupportedMechs(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)
	db := collection.Database()

	usersPayload := []bson.D{
		{
			{"createUser", "hello_user"},
			{"roles", bson.A{}},
			{"pwd", "hello_password"},
		},
		{
			{"createUser", "hello_user_plain"},
			{"roles", bson.A{}},
			{"pwd", "hello_password"},
			{"mechanisms", bson.A{"PLAIN"}},
		},
		{
			{"createUser", "hello_user_scram1"},
			{"roles", bson.A{}},
			{"pwd", "hello_password"},
			{"mechanisms", bson.A{"SCRAM-SHA-1"}},
		},
		{
			{"createUser", "hello_user_scram256"},
			{"roles", bson.A{}},
			{"pwd", "hello_password"},
			{"mechanisms", bson.A{"SCRAM-SHA-256"}},
		},
	}

	for _, u := range usersPayload {
		require.NoError(t, db.RunCommand(ctx, u).Err())
	}

	testCases := []struct {
		username string
		db       string
		mechs    *types.Array
		err      bool
	}{
		{
			username: "not_found",
			db:       db.Name(),
		},
		{
			username: "another_db",
			db:       db.Name() + "_not_found",
		},
		{
			username: "hello_user",
			db:       db.Name(),
			mechs:    must.NotFail(types.NewArray("SCRAM-SHA-1", "SCRAM-SHA-256")),
		},
		{
			username: "hello_user_plain",
			db:       db.Name(),
			mechs:    must.NotFail(types.NewArray("PLAIN")),
		},
		{
			username: "hello_user_scram1",
			db:       db.Name(),
			mechs:    must.NotFail(types.NewArray("SCRAM-SHA-1")),
		},
		{
			username: "hello_user_scram256",
			db:       db.Name(),
			mechs:    must.NotFail(types.NewArray("SCRAM-SHA-256")),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.username, func(t *testing.T) {
			t.Parallel()

			var res bson.D

			err := db.RunCommand(ctx, bson.D{
				{"hello", "1"},
				{"saslSupportedMechs", tc.db + "." + tc.username},
			}).Decode(&res)

			if tc.err {
				require.Error(t, err)
			}

			actual := ConvertDocument(t, res)

			keys := []string{
				"isWritablePrimary",
				"maxBsonObjectSize",
				"maxMessageSizeBytes",
				"maxWriteBatchSize",
				"localTime",
				"connectionId",
				"minWireVersion",
				"maxWireVersion",
				"readOnly",
			}

			if tc.mechs != nil {
				keys = append(keys, "saslSupportedMechs")
				mechanisms := must.NotFail(actual.Get("saslSupportedMechs"))
				assert.Equal(t, tc.mechs, mechanisms)
			} else {
				assert.False(t, actual.Has("saslSupportedMechs"))
			}

			keys = append(keys, "ok")
			assert.Equal(t, keys, actual.Keys())
		})
	}
}
