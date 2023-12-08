package cursors

import (
	"testing"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

//func TestSomethingElse(t *testing.T) {
//	t.Parallel()
//	ctx, collection := setup.Setup(t)
//
//	db1 := collection.Database()
//	t.Cleanup(func() { require.NoError(t, db1.Drop(ctx)) })
//
//	db2 := collection.Database().Client().Database(db1.Name() + "_2")
//	t.Cleanup(func() { require.NoError(t, db2.Drop(ctx)) })
//}

func TestSomething(t *testing.T) {
	t.Parallel()

	//	s := setup.SetupWithOpts(t, &setup.SetupOpts{
	//		Providers: []shareddata.Provider{shareddata.Composites},
	//		//	ExtraOptions: url.Values{
	//		//		"minPoolSize":   []string{"1"},
	//		//		"maxPoolSize":   []string{"1"},
	//		//		"maxIdleTimeMS": []string{"0"},
	//		//	},
	//	})
	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{shareddata.Composites},
		AddNonExistentCollection: true,
	})

	// TODO
	ctx, targetCollection, _ := s.Ctx, s.TargetCollections[0], s.CompatCollections[0]

	//targetCollection, ctx := s.Collection, s.Ctx

	var targetRes bson.D
	err := targetCollection.Database().RunCommand(ctx, bson.D{
		{"find", targetCollection.Name()},
		{"sort", bson.D{{"v", 1}}},
		{"batchSize", 1},
	}).Decode(&targetRes)

	require.NoError(t, err)

	doc := integration.ConvertDocument(t, targetRes)

	v, _ := doc.Get("cursor")
	require.NotNil(t, v)

	cursor, ok := v.(*types.Document)
	require.True(t, ok)

	cursorID, _ := cursor.Get("id")
	require.NotNil(t, cursorID)

	v, _ = cursor.Get("firstBatch")
	require.NotNil(t, v)

}
