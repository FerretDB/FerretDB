package sessions

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestStartSessionCommand(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/1554")

	s := setup.SetupWithOpts(t, nil)
	ctx := s.Ctx
	db := s.Collection.Database()

	optsNoAuth := options.Client().ApplyURI(s.MongoDBURI)
	clientNoAuth, err := mongo.Connect(ctx, optsNoAuth)
	require.NoError(t, err)
	require.NoError(t, clientNoAuth.Ping(ctx, nil))

	username1 := "sessionUser"
	username2 := "sessionUser2"

	err = db.RunCommand(ctx, bson.D{{"dropAllUsersFromDatabase", 1}}).Err()
	require.NoError(t, err)

	err = db.RunCommand(ctx, bson.D{
		{"createUser", username1},
		{"roles", bson.A{}},
		{"pwd", "password"},
		{"mechanisms", bson.A{"SCRAM-SHA-256"}},
	}).Err()
	require.NoError(t, err)

	err = db.RunCommand(ctx, bson.D{
		{"createUser", username2},
		{"roles", bson.A{}},
		{"pwd", "password2"},
		{"mechanisms", bson.A{"SCRAM-SHA-256"}},
	}).Err()
	require.NoError(t, err)

	credential := options.Credential{ // as created in setup.setupUser
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    db.Name(),
		Username:      username1,
		Password:      "password",
	}

	optsAuth := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)
	clientAuth, err := mongo.Connect(ctx, optsAuth)
	require.NoError(t, err)
	require.NoError(t, clientAuth.Ping(ctx, nil))

	credential = options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    db.Name(),
		Username:      username2,
		Password:      "password2",
	}
	optsAuth2 := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)
	clientAuth2, err := mongo.Connect(ctx, optsAuth2)
	require.NoError(t, err)
	require.NoError(t, clientAuth2.Ping(ctx, nil))

	for name, tc := range map[string]struct {
		clientToStartSession *mongo.Client
		clientToEndSession   *mongo.Client
	}{
		"noAuth": {
			clientToStartSession: clientNoAuth,
			clientToEndSession:   clientNoAuth,
		},
		"auth": {
			clientToStartSession: clientAuth,
			clientToEndSession:   clientAuth,
		},
		"otherClientEndsAuth": {
			clientToStartSession: clientAuth,
			clientToEndSession:   clientAuth2,
		},
		"authEndsNoAuth": {
			clientToStartSession: clientNoAuth,
			clientToEndSession:   clientAuth,
		},
		"noAuthEndsAuth": {
			clientToStartSession: clientAuth,
			clientToEndSession:   clientNoAuth,
		},
	} {
		name, tc := name, tc

		tt.Run(name, func(t *testing.T) {
			// Subtests are not run in parallel because MongoDB doesn't guarantee sessions being safe for concurrent use.
			sessionID := startSession(t, ctx, tc.clientToStartSession.Database(db.Name()))

			var res bson.D
			endSessionsCommand := bson.D{{"endSessions", bson.A{bson.D{{"id", sessionID}}}}}
			err := tc.clientToEndSession.Database(db.Name()).RunCommand(ctx, endSessionsCommand).Decode(&res)
			require.NoError(t, err)
		})
	}
}

func startSession(t *testing.T, ctx context.Context, db *mongo.Database) *primitive.Binary {
	var res bson.D
	err := db.RunCommand(ctx, bson.D{{"startSession", 1}}).Decode(&res)
	require.NoError(t, err)

	doc := integration.ConvertDocument(t, res)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	idDoc := must.NotFail(doc.Get("id")).(*types.Document)
	id := must.NotFail(idDoc.Get("id")).(types.Binary)
	assert.Len(t, id.B, 16)
	assert.Equal(t, types.BinaryUUID, id.Subtype)
	assert.Equal(t, int32(30), must.NotFail(doc.Get("timeoutMinutes")))

	/*cur, err := collection.Database().Aggregate(ctx, bson.A{bson.D{{"$listLocalSessions", bson.D{}}}})
	require.NoError(t, err)

	sessions := integration.FetchAll(t, ctx, cur)
	var found bool
	for _, session := range sessions {
		s := integration.ConvertDocument(t, session)
		sDoc := must.NotFail(s.Get("_id")).(*types.Document)
		sId := must.NotFail(sDoc.Get("id")).(types.Binary)

		if id.Subtype == sId.Subtype && bytes.Equal(id.B, sId.B) {
			found = true
			break
		}
	}
	assert.True(t, found, "Started session not found in $listLocalSessions results")*/

	idForFilter := primitive.Binary{Subtype: byte(id.Subtype), Data: id.B}
	err = db.RunCommand(ctx, bson.D{{"refreshSessions", bson.A{bson.D{{"id", idForFilter}}}}}).Decode(&res)
	require.NoError(t, err)

	doc = integration.ConvertDocument(t, res)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

	return &idForFilter
}
