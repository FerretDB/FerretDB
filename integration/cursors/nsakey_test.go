package cursors

import (
	"testing"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

	var expectedDocs []bson.D
	require.NoError(t, cur.All(ctx, &expectedDocs))

	var res bson.M
	err = coll.Database().RunCommand(ctx, bson.D{
		{"find", coll.Name()},
		{"sort", sort},
		{"batchSize", 1},
	}).Decode(&res)

	require.NoError(t, err)

	//doc := integration.ConvertDocument(t, res)

	cursor := res["cursor"].(bson.M)
	//v, _ := doc.Get("cursor")
	//require.NotNil(t, v)

	//cursor, ok := v.(*types.Document)
	//require.True(t, ok)

	cursorID := cursor["id"]
	require.NotNil(t, cursorID)

	firstBatch := cursor["firstBatch"].(bson.A)
	require.Equal(t, 1, len(firstBatch))

	actualDoc := firstBatch[0].(bson.D)

	integration.AssertEqualDocuments(t, expectedDocs[0], actualDoc)

	for i := 1; i < len(expectedDocs); i++ {
		var getMoreRes bson.M

		err = coll.Database().RunCommand(ctx, bson.D{
			{"getMore", cursorID},
			{"collection", coll.Name()},
			{"batchSize", 1},
		}).Decode(&getMoreRes)
		require.NoError(t, err)

		integration.AssertEqualDocuments(t, expectedDocs[0], firstBatch[0].(bson.D))

		cursor := getMoreRes["cursor"].(bson.M)
		firstBatch := cursor["firstBatch"].(bson.A)
		require.Equal(t, 1, len(firstBatch))

		integration.AssertEqualDocuments(t, expectedDocs[i], firstBatch[0].(bson.D))
	}
}
