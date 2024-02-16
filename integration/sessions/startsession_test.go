package sessions

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestStartSessionCommand(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/1554")

	ctx, collection := setup.Setup(t)
	// sessionsCollection := collection.Database().Client().Database("config").Collection("system.sessions")

	var res bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"startSession", 1}}).Decode(&res)
	require.NoError(t, err)

	doc := integration.ConvertDocument(t, res)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
	idDoc := must.NotFail(doc.Get("id")).(*types.Document)
	id := must.NotFail(idDoc.Get("id")).(types.Binary)
	assert.Len(t, id.B, 16)
	assert.Equal(t, types.BinaryUUID, id.Subtype)
	assert.Equal(t, int32(30), must.NotFail(doc.Get("timeoutMinutes")))

	cur, err := collection.Database().Aggregate(ctx, bson.A{bson.D{{"$listLocalSessions", bson.D{}}}})
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
	assert.True(t, found, "Started session not found in $listLocalSessions results")

	idForFilter := primitive.Binary{Subtype: byte(id.Subtype), Data: id.B}
	err = collection.Database().RunCommand(ctx, bson.D{{"refreshSessions", bson.A{bson.D{{"id", idForFilter}}}}}).Decode(&res)
	require.NoError(t, err)

	doc = integration.ConvertDocument(t, res)
	assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))

	// TODO: would be nice to check an entry from system.sessions collection, but it's not created immediately
	// TODO: same about $listSessions (it uses system.sessions collection from the config db)
	// filter := bson.D{{"_id.id", primitive.Binary{Subtype: byte(id.Subtype), Data: id.B}}}
	// err = sessionsCollection.FindOne(ctx, filter).Decode(&res)
	// require.NoError(t, err)
}

func TestStartSession(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)
	client := collection.Database().Client()

	_, err := client.StartSession(options.Session().SetCausalConsistency(true))
	require.NoError(t, err)

	//	id := session.ID()

	var res bson.D
	err = client.Database("config").Collection("system.sessions").FindOne(ctx, nil).Decode(&res)
	require.NoError(t, err)
}
