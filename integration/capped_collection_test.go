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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCappedCollectionMaxDocs(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		CappedCleanupInterval: 1 * time.Second,
	})

	coll, ctx := s.Collection, s.Ctx
	db := coll.Database()

	// drop and re-create the collection as capped collection
	require.NoError(t, coll.Drop(ctx))
	collOpts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1000000).SetMaxDocuments(10)
	err := db.CreateCollection(s.Ctx, coll.Name(), collOpts)
	require.NoError(t, err)

	bsonArr, docs := GenerateDocuments(0, 20)

	_, err = coll.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	t.Run("AfterInitialCleanup", func(t *testing.T) {
		time.Sleep(2 * time.Second)

		opts := options.Find().SetSort(bson.D{{"_id", 1}})
		cursor, err := coll.Find(ctx, bson.D{}, opts)
		require.NoError(t, err)

		result := FetchAll(t, ctx, cursor)
		// verify only last 10 documents are present in the collection after cleanup
		AssertEqualDocumentsSlice(t, docs[10:], result)
	})

	t.Run("AfterMultipleCycle", func(t *testing.T) {
		time.Sleep(4 * time.Second)

		docCount := getDocCount(t, ctx, coll)
		// verify no other documents are deleted after initial cleanup cycle
		assert.Equal(t, docCount, int64(10))
	})
}

func TestCappedCollectionMaxSize(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		CappedCleanupInterval:   1 * time.Second,
		CappedCleanupPercentage: 20,
	})

	coll, ctx := s.Collection, s.Ctx
	db := coll.Database()

	// drop and re-create the collection as capped collection
	require.NoError(t, coll.Drop(ctx))

	// setting collection size very low, so that in every cleanup cycle some documents will be deleted
	collOpts := options.CreateCollection().SetCapped(true).SetSizeInBytes(100)

	err := db.CreateCollection(s.Ctx, coll.Name(), collOpts)
	require.NoError(t, err)

	bsonArr, _ := GenerateDocuments(0, 100)

	_, err = coll.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	prevDocCount := int64(len(bsonArr))
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second)

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			docCount := getDocCount(t, ctx, coll)
			if docCount == prevDocCount {
				t.Fatalf("unexpeced collection stat, expected the document count to be less than %d after capped collection cleanup", docCount)
			}
		})
	}
}

// getDocCount returns the number of documents in a collection based on collStats.
func getDocCount(t *testing.T, ctx context.Context, coll *mongo.Collection) int64 {
	t.Helper()

	var result bson.D

	err := coll.Database().RunCommand(ctx, bson.D{{"collStats", coll.Name()}}).Decode(&result)
	require.NoError(t, err)

	stats := ConvertDocument(t, result)
	v, _ := stats.Get("count")
	require.NotNil(t, v)

	return v.(int64)
}
