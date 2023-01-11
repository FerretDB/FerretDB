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
	"go.mongodb.org/mongo-driver/mongo/options"

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

	opts := options.Find().SetBatchSize(50)

	cursor, err := coll.Find(ctx, bson.D{}, opts)
	require.NoError(t, err)

	var results []any

	for cursor.Next(ctx) {
		var resDoc bson.D

		err = cursor.Decode(&resDoc)
		require.NoError(t, err)

		results = append(results, resDoc)
	}

	require.Equal(t, len(docs), len(results))
}
