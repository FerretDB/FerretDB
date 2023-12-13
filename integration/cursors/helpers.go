package cursors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/internal/types"
)

// getFirstBatch takes the response from the query that generates the cursors,
// validates if it contains cursor.firstBatch, and cursor ID, and returns those.
func getFirstBatch(t testing.TB, res bson.D) (*types.Array, any) {
	t.Helper()

	doc := integration.ConvertDocument(t, res)

	v, _ := doc.Get("cursor")
	require.NotNil(t, v)

	cursor, ok := v.(*types.Document)
	require.True(t, ok)

	cursorID, _ := cursor.Get("id")
	assert.NotNil(t, cursorID)

	v, _ = cursor.Get("firstBatch")
	require.NotNil(t, v)

	firstBatch, ok := v.(*types.Array)
	require.True(t, ok)

	return firstBatch, cursorID
}

// getNextBatch takes the response from the getMore query,
// validates if it contains cursor.nextBatch, and cursor ID, and returns those.
func getNextBatch(t testing.TB, res bson.D) (*types.Array, any) {
	t.Helper()

	doc := integration.ConvertDocument(t, res)

	v, _ := doc.Get("cursor")
	require.NotNil(t, v)

	cursor, ok := v.(*types.Document)
	require.True(t, ok)

	cursorID, _ := cursor.Get("id")
	assert.NotNil(t, cursorID)

	v, _ = cursor.Get("nextBatch")
	require.NotNil(t, v)

	firstBatch, ok := v.(*types.Array)
	require.True(t, ok)

	return firstBatch, cursorID
}
