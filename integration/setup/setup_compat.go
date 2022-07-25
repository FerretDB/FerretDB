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

package setup

import (
	"context"
	"testing"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slices"
)

// setupCompatCollections setups a single database with one collection per provider for compatibility tests.
func setupCompatCollections(tb testing.TB, ctx context.Context, client *mongo.Client, db string, providers []shareddata.Provider) []*mongo.Collection {
	tb.Helper()

	require.NotEmpty(tb, providers)

	var ownDatabase bool
	if db == "" {
		db = testutil.DatabaseName(tb)
		ownDatabase = true
	}

	database := client.Database(db)

	if ownDatabase {
		// drop remnants of the previous failed run
		_ = database.Drop(ctx)

		// delete database unless test failed
		tb.Cleanup(func() {
			if tb.Failed() {
				return
			}

			err := database.Drop(ctx)
			require.NoError(tb, err)
		})
	}

	collections := make([]*mongo.Collection, 0, len(providers))
	for _, provider := range providers {
		if *targetPortF == 0 && !slices.Contains(provider.Handlers(), *handlerF) {
			tb.Logf("Provider %q is not compatible with handler %q, skipping it.", provider.Name(), *handlerF)
			continue
		}

		name := testutil.CollectionName(tb) + "_" + provider.Name()
		collection := database.Collection(name)

		// drop remnants of the previous failed run
		_ = collection.Drop(ctx)

		docs := shareddata.Docs(provider)
		require.NotEmpty(tb, docs)

		res, err := collection.InsertMany(ctx, docs)
		require.NoError(tb, err, "provider %q, handler %q, colleciton %s.%s", provider.Name(), *handlerF, db, name)
		require.Len(tb, res.InsertedIDs, len(docs))

		// delete collection unless test failed
		tb.Cleanup(func() {
			if tb.Failed() {
				tb.Logf("Keeping %s.%s for debugging.", db, name)
				return
			}

			err := collection.Drop(ctx)
			require.NoError(tb, err)
		})

		collections = append(collections, collection)
	}

	require.NotEmpty(tb, collections)
	return collections
}
