package cursors

import (
	"testing"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		Providers: []shareddata.Provider{shareddata.Composites},
		//	ExtraOptions: url.Values{
		//		"minPoolSize":   []string{"1"},
		//		"maxPoolSize":   []string{"1"},
		//		"maxIdleTimeMS": []string{"0"},
		//	},
	})

	coll, ctx := s.Collection, s.Ctx

	sort := bson.D{{"v", 1}}

	opts := options.Find().SetSort(sort)
	cur, err := coll.Find(ctx, bson.D{{}}, opts)
	require.NoError(t, err)

	var expectedRes []bson.D
	require.NoError(t, cur.All(ctx, &expectedRes))

	expectedDocs := integration.ConvertDocuments(t, expectedRes)

	var res bson.D
	err = coll.Database().RunCommand(ctx, bson.D{
		{"find", coll.Name()},
		{"sort", sort},
		{"batchSize", 1},
	}).Decode(&res)

	require.NoError(t, err)

	doc := integration.ConvertDocument(t, res)

	v, _ := doc.Get("cursor")
	require.NotNil(t, v)

	cursor, ok := v.(*types.Document)
	require.True(t, ok)

	cursorID, _ := cursor.Get("id")
	require.NotNil(t, cursorID)

	v, _ = cursor.Get("firstBatch")
	assert.Equal(t, 1, v.(*types.Array).Len())

	actualdoc, _ := v.(*types.Array).Get(0)

	require.Equal(t, expectedDocs[0], actualdoc)

	for i := 1; i < len(expectedDocs); i++ {
		var getMoreRes bson.D
		err = coll.Database().RunCommand(ctx, bson.D{
			{"getMore", cursorID},
			{"collection", coll.Name()},
			{"batchSize", 1},
		}).Decode(&getMoreRes)
		require.NoError(t, err)

		getMoreDoc := integration.ConvertDocument(t, res)

		path, err := types.NewPathFromString("cursor.firstBatch")
		require.NoError(t, err)

		getMore, err := getMoreDoc.GetByPath(path)
		require.NoError(t, err)

		resDocs := getMore.(*types.Array)
		require.Equal(t, 1, resDocs.Len())

		doc, _ := resDocs.Get(0)

		assert.Equal(t, expectedDocs[i], doc, res)
	}
}
