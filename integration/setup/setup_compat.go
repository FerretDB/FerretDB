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
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// SetupCompatOpts represents setup options for compatibility test.
//
// TODO Add option to use read-only user. https://github.com/FerretDB/FerretDB/issues/1025
type SetupCompatOpts struct {
	// Database to use. If empty, temporary test-specific database is created and dropped after test.
	// Most tests should keep this empty.
	DatabaseName string

	// Data providers.
	Providers []shareddata.Provider

	ownDatabase        bool
	databaseName       string
	baseCollectionName string
}

// SetupCompatResult represents compatibility test setup results.
type SetupCompatResult struct {
	Ctx               context.Context
	TargetCollections []*mongo.Collection
	TargetPort        uint16
	CompatCollections []*mongo.Collection
	CompatPort        uint16
	StateProvider     *state.Provider
}

// SetupCompatWithOpts setups the compatibility test according to given options.
func SetupCompatWithOpts(tb testing.TB, opts *SetupCompatOpts) *SetupCompatResult {
	tb.Helper()

	startup()

	// skip tests for MongoDB as soon as possible
	compatPort := *compatPortF
	if compatPort == 0 {
		tb.Skip("compatibility tests require second system")
	}

	if opts == nil {
		opts = new(SetupCompatOpts)
	}

	opts.databaseName = opts.DatabaseName
	if opts.DatabaseName == "" {
		// When we use `task all` to run `pg` and `tigris` compat tests in parallel,
		// they both use the same MongoDB instance.
		// Add the handler's name to prevent the usage of the same database.
		// The only test that uses DatabaseName is `TestEnvData`,
		// but it is protected by a build tag; see comment there.
		opts.databaseName = testutil.DatabaseName(tb) + "_" + *handlerF
		opts.ownDatabase = true
	}

	opts.baseCollectionName = testutil.CollectionName(tb)

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	level := zap.NewAtomicLevelAt(zap.ErrorLevel)
	if *debugSetupF {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger := testutil.Logger(tb, level)

	var stateProvider *state.Provider
	targetPort := *targetPortF
	if targetPort == 0 {
		stateProvider, targetPort = setupListener(tb, ctx, logger)
	}

	// register cleanup function after setupListener registers its own to preserve full logs
	tb.Cleanup(cancel)

	targetCollections := setupCompatCollections(tb, ctx, setupClient(tb, ctx, targetPort), opts)
	compatCollections := setupCompatCollections(tb, ctx, setupClient(tb, ctx, compatPort), opts)

	level.SetLevel(*logLevelF)

	return &SetupCompatResult{
		Ctx:               ctx,
		TargetCollections: targetCollections,
		TargetPort:        uint16(targetPort),
		CompatCollections: compatCollections,
		CompatPort:        uint16(compatPort),
		StateProvider:     stateProvider,
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

	database := client.Database(opts.databaseName)

	if opts.ownDatabase {
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
		collectionName := opts.baseCollectionName + "_" + provider.Name()
		if !opts.ownDatabase {
			collectionName = strings.ToLower(provider.Name())
		}
		fullName := opts.databaseName + "." + collectionName

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

		// if validators are set, create collection with them (otherwise collection will be created on first insert)
		if validators := provider.Validators(*handlerF, collectionName); len(validators) > 0 {
			var opts options.CreateCollectionOptions
			for key, value := range validators {
				opts.SetValidator(bson.D{{key, value}})
			}

			err := database.CreateCollection(ctx, collectionName, &opts)
			if err != nil {
				var cmdErr *mongo.CommandError
				if errors.As(err, &cmdErr) {
					// If collection can't be created in MongoDB because MongoDB has a different validator format, it's ok:
					require.Contains(tb, cmdErr.Message, `unknown top level operator: $tigrisSchemaString`)
				}
			}
		}

		docs := shareddata.Docs(provider)
		require.NotEmpty(tb, docs)

		res, err := collection.InsertMany(ctx, docs)
		require.NoError(tb, err, "%s: handler %q, collection %s", provider.Name(), *handlerF, fullName)
		require.Len(tb, res.InsertedIDs, len(docs))

		if opts.ownDatabase {
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

	require.NotEmpty(tb, collections, "all providers were not compatible")
	return collections
}
