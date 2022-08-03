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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// SetupCompatOpts represents setup options for compatibility test.
type SetupCompatOpts struct {
	// Database to use. If empty, temporary test-specific database is created.
	DatabaseName string

	KeepData bool

	// Data providers.
	Providers []shareddata.Provider
}

// SetupResult represents compatibility test setup results.
type SetupCompatResult struct {
	Ctx               context.Context
	TargetCollections []*mongo.Collection
	TargetPort        uint16
	CompatCollections []*mongo.Collection
	CompatPort        uint16
}

// SetupCompatWithOpts setups the compatibility test according to given options.
func SetupCompatWithOpts(tb testing.TB, opts *SetupCompatOpts) *SetupCompatResult {
	tb.Helper()

	startup()

	if opts == nil {
		opts = new(SetupCompatOpts)
	}

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	level := zap.NewAtomicLevelAt(zap.ErrorLevel)
	if *debugSetupF {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger := zaptest.NewLogger(tb, zaptest.Level(level), zaptest.WrapOptions(zap.AddCaller()))

	targetPort := *targetPortF
	if targetPort == 0 {
		targetPort = setupListener(tb, ctx, logger)
	}

	// register cleanup function after setupListener registers its own to preserve full logs
	tb.Cleanup(cancel)

	compatPort := *compatPortF
	if compatPort == 0 {
		tb.Skip("compatibility tests require second system")
	}

	targetCollections := setupCompatCollections(tb, ctx, setupClient(tb, ctx, targetPort), opts)
	compatCollections := setupCompatCollections(tb, ctx, setupClient(tb, ctx, compatPort), opts)

	level.SetLevel(*logLevelF)

	return &SetupCompatResult{
		Ctx:               ctx,
		TargetCollections: targetCollections,
		TargetPort:        uint16(targetPort),
		CompatCollections: compatCollections,
		CompatPort:        uint16(compatPort),
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
func setupCompatCollections(tb testing.TB, ctx context.Context, client *mongo.Client, opts *SetupCompatOpts) []*mongo.Collection {
	tb.Helper()

	var ownDatabase bool
	databaseName := opts.DatabaseName
	if databaseName == "" {
		databaseName = testutil.DatabaseName(tb)
		ownDatabase = true
	}

	database := client.Database(databaseName)

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

	collections := make([]*mongo.Collection, 0, len(opts.Providers))
	for _, provider := range opts.Providers {
		collectionName := testutil.CollectionName(tb) + "_" + provider.Name()
		if opts.KeepData {
			collectionName = strings.ToLower(provider.Name())
		}
		fullName := databaseName + "." + collectionName

		if *targetPortF == 0 && !slices.Contains(provider.Handlers(), *handlerF) {
			tb.Logf(
				"Provider %q is not compatible with handler %q, skipping creating %q.",
				provider.Name(), *handlerF, fullName,
			)
			continue
		}

		collection := database.Collection(collectionName)

		// drop remnants of the previous failed run
		_ = collection.Drop(ctx)

		docs := shareddata.Docs(provider)
		require.NotEmpty(tb, docs)

		res, err := collection.InsertMany(ctx, docs)
		require.NoError(tb, err, "%s: handler %q, collection %s", provider.Name(), *handlerF, fullName)
		require.Len(tb, res.InsertedIDs, len(docs))

		if !opts.KeepData {
			// delete collection unless test failed
			tb.Cleanup(func() {
				if tb.Failed() {
					tb.Logf("Keeping %s for debugging.", fullName)
					return
				}

				err := collection.Drop(ctx)
				require.NoError(tb, err)
			})
		}

		collections = append(collections, collection)
	}

	require.NotEmpty(tb, collections)
	return collections
}
