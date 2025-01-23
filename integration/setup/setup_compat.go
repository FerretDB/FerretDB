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
	"log/slog"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

// SetupCompatOpts represents setup options for compatibility test.
//
// Add option to use read-only user.
// TODO https://github.com/FerretDB/FerretDB/issues/1025
type SetupCompatOpts struct {
	// Data providers.
	Providers []shareddata.Provider

	// If true, a non-existent collection will be added to the list of collections.
	// This is useful to test the behavior when a collection is not found.
	//
	// This flag is not needed, always add a non-existent collection.
	// TODO https://github.com/FerretDB/FerretDB/issues/1545
	AddNonExistentCollection bool

	databaseName       string
	baseCollectionName string
}

// SetupCompatResult represents compatibility test setup results.
type SetupCompatResult struct {
	Ctx               context.Context
	TargetCollections []*mongo.Collection
	CompatCollections []*mongo.Collection
}

// SetupCompatWithOpts setups the compatibility test according to given options.
func SetupCompatWithOpts(tb testing.TB, opts *SetupCompatOpts) *SetupCompatResult {
	tb.Helper()

	if *compatURLF == "" {
		tb.Skip("-compat-url is empty, skipping compatibility test")
	}

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	setupCtx, span := otel.Tracer("").Start(ctx, "SetupCompatWithOpts")
	defer span.End()

	if opts == nil {
		opts = new(SetupCompatOpts)
	}

	opts.databaseName = testutil.DatabaseName(tb)

	// When database name is too long, database is created but inserting documents
	// fail with InvalidNamespace error.
	require.Less(tb, len(opts.databaseName), 64, "database name %q is too long", opts.databaseName)

	opts.baseCollectionName = testutil.CollectionName(tb)

	var levelVar slog.LevelVar
	levelVar.Set(slog.LevelError)
	if *debugSetupF {
		levelVar.Set(slog.LevelDebug)
	}

	logger := testutil.LevelLogger(tb, &levelVar)

	var targetClient *mongo.Client

	uri := *targetURLF
	if uri == "" {
		uri = setupListener(tb, setupCtx, nil, logger)
	}

	targetClient = setupClient(tb, setupCtx, uri, false)

	// register cleanup function after setupListener registers its own to preserve full logs
	tb.Cleanup(cancel)

	targetCollections := setupCompatCollections(tb, setupCtx, targetClient, opts, *targetBackendF)

	compatClient := setupClient(tb, setupCtx, *compatURLF, false)
	compatCollections := setupCompatCollections(tb, setupCtx, compatClient, opts, "mongodb")

	err := levelVar.UnmarshalText([]byte(*logLevelF))
	require.NoError(tb, err)

	return &SetupCompatResult{
		Ctx:               ctx,
		TargetCollections: targetCollections,
		CompatCollections: compatCollections,
	}
}

// SetupCompat setups compatibility test.
func SetupCompat(tb testing.TB) (context.Context, []*mongo.Collection, []*mongo.Collection) {
	tb.Helper()

	s := SetupCompatWithOpts(tb, &SetupCompatOpts{
		Providers: shareddata.AllProviders(),
	})
	return s.Ctx, s.TargetCollections, s.CompatCollections
}

// setupCompatCollections setups a single database with one collection per provider for compatibility tests.
func setupCompatCollections(tb testing.TB, ctx context.Context, client *mongo.Client, opts *SetupCompatOpts, backend string) []*mongo.Collection {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupCompatCollections")
	defer span.End()

	database := client.Database(opts.databaseName)

	// drop remnants of the previous failed run
	_ = database.RunCommand(ctx, bson.D{{"dropAllUsersFromDatabase", 1}})
	_ = database.Drop(ctx)

	// drop database unless test failed
	tb.Cleanup(func() {
		if tb.Failed() {
			return
		}

		err := database.RunCommand(ctx, bson.D{{"dropAllUsersFromDatabase", 1}}).Err()
		require.NoError(tb, err)

		err = database.Drop(ctx)
		require.NoError(tb, err)
	})

	providers := slices.Clone(opts.Providers)

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/825
	// rand.Shuffle(len(providers), func(i, j int) { providers[i], providers[j] = providers[j], providers[i] })

	collections := make([]*mongo.Collection, 0, len(providers))

	for _, provider := range providers {
		collectionName := opts.baseCollectionName + "_" + provider.Name()
		fullName := opts.databaseName + "." + collectionName

		collection := database.Collection(collectionName)

		docs := shareddata.Docs(provider)
		require.NotEmpty(tb, docs)

		res, err := collection.InsertMany(ctx, docs)
		require.NoError(tb, err, "%s: backend %q, collection %s", provider.Name(), backend, fullName)
		require.Len(tb, res.InsertedIDs, len(docs))

		collections = append(collections, collection)
	}

	// opts.AddNonExistentCollection is not needed, always add a non-existent collection
	// TODO https://github.com/FerretDB/FerretDB/issues/1545
	if opts.AddNonExistentCollection {
		nonExistedCollectionName := opts.baseCollectionName + "-non-existent"
		collection := database.Collection(nonExistedCollectionName)
		collections = append(collections, collection)
	}

	require.NotEmpty(tb, collections)

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/825
	// rand.Shuffle(len(collections), func(i, j int) { collections[i], collections[j] = collections[j], collections[i] })

	return collections
}
