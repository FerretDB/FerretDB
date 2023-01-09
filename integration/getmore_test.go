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

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestGetMore(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t)

	insertDoc := bson.D{{"a", int32(1)}}

	var docs []any
	for i := 0; i < 200; i++ {
		docs = append(docs, insertDoc)
	}

	_, err := coll.InsertMany(ctx, docs)
	require.NoError(t, err)

	cursor, err := coll.Find(ctx, bson.D{})
	require.NoError(t, err)

	resDoc := bson.D{}
	for cursor.Next(ctx) {
		t.Log(cursor.RemainingBatchLength())

		err = cursor.Decode(&resDoc)
		require.NoError(t, err)
	}

	//
	//more := coll.Database().RunCommand(ctx, bson.D{{"getMore", cursor.ID()}, {"collection", coll.Name()}})
	//require.NoError(t, more.Err())
	//
	//resDoc = bson.D{}
	//more.Decode(&resDoc)
	//
	//t.Log(resDoc)
}
