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
	"fmt"
	"runtime/trace"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// SetupCompatOpts represents setup options for compatibility test.
//
// TODO Add option to use read-only user. https://github.com/FerretDB/FerretDB/issues/1025
type SetupCompatOpts struct {
	// Data providers.
	Providers []shareddata.Provider

	// If true, a non-existent collection will be added to the list of collections.
	// This is useful to test the behavior when a collection is not found.
	// TODO This flag is not needed, always add a non-existent collection https://github.com/FerretDB/FerretDB/issues/1545
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

	defer trace.StartRegion(setupCtx, "SetupCompatWithOpts").End()

	if opts == nil {
		opts = new(SetupCompatOpts)
	}

	// When we use `task all` to run `pg` and `tigris` compat tests in parallel,
	// they both use the same MongoDB instance.
	// Add the backend's name to prevent the usage of the same database.
	opts.databaseName = testutil.DatabaseName(tb) + "_" + strings.ReplaceAll(*targetBackendF, "-", "_")

	opts.baseCollectionName = testutil.CollectionName(tb)

	level := zap.NewAtomicLevelAt(zap.ErrorLevel)
	if *debugSetupF {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger := testutil.Logger(tb, level)

	var targetClient *mongo.Client
	if *targetURLF == "" {
		targetClient, _ = setupListener(tb, setupCtx, logger)
	} else {
		targetClient = setupClient(tb, setupCtx, *targetURLF)
	}

	// register cleanup function after setupListener registers its own to preserve full logs
	tb.Cleanup(cancel)

	targetCollections := setupCompatCollections(tb, setupCtx, targetClient, opts, *targetBackendF)

	compatClient := setupClient(tb, setupCtx, *compatURLF)
	compatCollections := setupCompatCollections(tb, setupCtx, compatClient, opts, "mongodb")

	level.SetLevel(*logLevelF)

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

	defer trace.StartRegion(ctx, "setupCompatCollections").End()

	database := client.Database(opts.databaseName)

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

	collections := make([]*mongo.Collection, 0, len(opts.Providers))
	for _, provider := range opts.Providers {
		collectionName := opts.baseCollectionName + "_" + provider.Name()
		fullName := opts.databaseName + "." + collectionName

		if *targetURLF == "" && !provider.IsCompatible(*targetBackendF) {
			// Skip creating collection for both target and compat if
			// target is not compatible. Target and compat must have same collection.
			tb.Logf(
				"Provider %q is not compatible with target backend %q, skipping creating %q.",
				provider.Name(),
				*targetBackendF,
				fullName,
			)
			continue
		}

		spanName := fmt.Sprintf("setupCompatCollections/%s", collectionName)
		collCtx, span := otel.Tracer("").Start(ctx, spanName)
		region := trace.StartRegion(collCtx, spanName)

		collection := database.Collection(collectionName)

		// drop remnants of the previous failed run
		_ = collection.Drop(collCtx)

		// if validators are set, create collection with them (otherwise collection will be created on first insert)
		if validators := provider.Validators(backend, collectionName); len(validators) > 0 {
			var opts options.CreateCollectionOptions
			for key, value := range validators {
				opts.SetValidator(bson.D{{key, value}})
			}

			err := database.CreateCollection(ctx, collectionName, &opts)
			require.NoError(tb, err)
		}

		docs := shareddata.Docs(provider)
		require.NotEmpty(tb, docs)

		res, err := collection.InsertMany(collCtx, docs)
		require.NoError(tb, err, "%s: backend %q, collection %s", provider.Name(), backend, fullName)
		require.Len(tb, res.InsertedIDs, len(docs))

		// delete collection unless test failed
		tb.Cleanup(func() {
			if tb.Failed() {
				tb.Logf("Keeping %s for debugging.", fullName)
				return
			}

			err := collection.Drop(collCtx)
			require.NoError(tb, err)
		})

		collections = append(collections, collection)

		region.End()
		span.End()
	}

	// TODO opts.AddNonExistentCollection is not needed, always add a non-existent collection
	// https://github.com/FerretDB/FerretDB/issues/1545
	if opts.AddNonExistentCollection {
		nonExistedCollectionName := opts.baseCollectionName + "-non-existent"
		collection := database.Collection(nonExistedCollectionName)
		collections = append(collections, collection)
	}

	require.NotEmpty(tb, collections, "all providers were not compatible")
	return collections
}
