// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cursors

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"

	"github.com/FerretDB/FerretDB/integration"
)

// getFirstBatch takes the response from the query that generates the cursors,
// validates if it contains cursor.firstBatch, and cursor ID, and returns those.
func getFirstBatch(t testtb.TB, res bson.D) (*types.Array, any) {
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
func getNextBatch(t testtb.TB, res bson.D) (*types.Array, any) {
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
