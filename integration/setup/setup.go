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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// SetupOpts represents setup options.
//
// TODO add option to use read-only user: https://github.com/FerretDB/FerretDB/issues/914.
type SetupOpts struct {
	// Database to use. If empty, temporary test-specific database is created.
	DatabaseName string

	// Data providers.
	Providers []shareddata.Provider
}

// SetupResult represents setup results.
type SetupResult struct {
	Ctx        context.Context
	Collection *mongo.Collection
	Port       uint16
}

// SetupWithOpts setups the test according to given options.
func SetupWithOpts(tb testing.TB, opts *SetupOpts) *SetupResult {
	tb.Helper()

	startup()

	if opts == nil {
		opts = new(SetupOpts)
	}

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	level := zap.NewAtomicLevelAt(zap.ErrorLevel)
	if *debugSetupF {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger := zaptest.NewLogger(tb, zaptest.Level(level), zaptest.WrapOptions(zap.AddCaller()))

	port := *targetPortF
	if port == 0 {
		port = setupListener(tb, ctx, logger)
	}

	// register cleanup function after setupListener registers its own to preserve full logs
	tb.Cleanup(cancel)

	client := setupClient(tb, ctx, port)

	collection := setupCollection(tb, ctx, &setupCollectionOpts{
		client:    client,
		db:        opts.DatabaseName,
		providers: opts.Providers,
	})

	level.SetLevel(*logLevelF)

	return &SetupResult{
		Ctx:        ctx,
		Collection: collection,
		Port:       uint16(port),
	}
}

// Setup setups test with specified data providers.
func Setup(tb testing.TB, providers ...shareddata.Provider) (context.Context, *mongo.Collection) {
	tb.Helper()

	s := SetupWithOpts(tb, &SetupOpts{
		Providers: providers,
	})
	return s.Ctx, s.Collection
}

// setupCollection setups a single collection for all providers, if the are present.
func setupCollection(tb testing.TB, ctx context.Context, opts *setupCollectionOpts) *mongo.Collection {
	tb.Helper()

	var ownDatabase bool
	db := opts.db
	if db == "" {
		db = testutil.DatabaseName(tb)
		ownDatabase = true
	}

	database := opts.client.Database(db)
	collectionName := testutil.CollectionName(tb)
	collection := database.Collection(collectionName)

	// drop remnants of the previous failed run
	_ = collection.Drop(ctx)
	if ownDatabase {
		_ = database.Drop(ctx)
	}

	for _, provider := range opts.providers {
		docs := shareddata.Docs(provider)
		require.NotEmpty(tb, docs)

		res, err := collection.InsertMany(ctx, docs)
		require.NoError(tb, err, "provider %q", provider.Name())
		require.Len(tb, res.InsertedIDs, len(docs))
	}

	// delete collection and (possibly) database unless test failed
	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping %s.%s for debugging.", db, collectionName)
			return
		}

		err := collection.Drop(ctx)
		require.NoError(tb, err)

		if ownDatabase {
			err = database.Drop(ctx)
			require.NoError(tb, err)
		}
	})

	return collection
}
