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
	"fmt"
	"runtime/trace"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

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

	// Flags overrides the flags set from cli.
	Flags map[string]any
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

	s := startup(tb)
	f := getFlags(s)

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	defer trace.StartRegion(ctx, "SetupCompatWithOpts").End()

	// skip tests for MongoDB as soon as possible
	if f.GetCompatPort() == 0 {
		tb.Skip("compatibility tests require second system")
	}

	if opts == nil {
		opts = new(SetupCompatOpts)
	}

	validateFlags(tb)
	f.ApplyOpts(tb, opts.Flags)

	// When we use `task all` to run `pg` and `tigris` compat tests in parallel,
	// they both use the same MongoDB instance.
	// Add the handler's name to prevent the usage of the same database.
	opts.databaseName = testutil.DatabaseName(tb) + "_" + f.GetHandler()

	opts.baseCollectionName = testutil.CollectionName(tb)

	level := zap.NewAtomicLevelAt(zap.ErrorLevel)
	if *debugSetupF {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger := testutil.Logger(tb, level)

	var targetClient *mongo.Client
	if f.GetTargetPort() == 0 {
		targetClient, _ = setupListener(tb, ctx, logger, s, f)
	} else {
		// When TLS is enabled, RootCAs and Certificates are fetched
		// upon creating client. Target uses PLAIN for authMechanism.
		targetURI := buildMongoDBURI(tb, &buildMongoDBURIOpts{
			host: fmt.Sprintf("127.0.0.1:%d", f.GetTargetPort()),
			tls:  f.IsTargetTLS(),
			user: getUser(f.IsTargetTLS()),
		})
		targetClient = setupClient(tb, ctx, targetURI, f.IsTargetTLS())
	}

	// register cleanup function after setupListener registers its own to preserve full logs
	tb.Cleanup(cancel)

	// When TLS is enabled, RootCAs and Certificates are fetched
	// upon creating client. Compat leaves authMechanism empty which defaults to SCRAM.
	uri := buildMongoDBURI(tb, &buildMongoDBURIOpts{
		host: fmt.Sprintf("127.0.0.1:%d", f.GetCompatPort()),
		tls:  f.IsCompatTLS(),
		user: getUser(f.IsCompatTLS()),
	})
	compatClient := setupClient(tb, ctx, uri, f.IsCompatTLS())

	targetCollections := setupCompatCollections(tb, ctx, targetClient, opts, f, true)
	compatCollections := setupCompatCollections(tb, ctx, compatClient, opts, f, false)

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
func setupCompatCollections(tb testing.TB, ctx context.Context, client *mongo.Client, opts *SetupCompatOpts, f flags, isTarget bool) []*mongo.Collection {
	tb.Helper()

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

		if f.GetTargetPort() == 0 && !slices.Contains(provider.Handlers(), f.GetHandler()) {
			tb.Logf(
				"Provider %q is not compatible with handler %q, skipping creating %q.",
				provider.Name(), f.GetHandler(), fullName,
			)
			continue
		}

		region := trace.StartRegion(ctx, fmt.Sprintf("setupCompatCollections/%s", collectionName))

		collection := database.Collection(collectionName)

		// drop remnants of the previous failed run
		_ = collection.Drop(ctx)

		// Validators are only applied to target. Compat is compatible with all provider.
		if isTarget {
			// if validators are set, create collection with them (otherwise collection will be created on first insert)
			if validators := provider.Validators(f.GetHandler(), collectionName); len(validators) > 0 {
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
		}

		docs := shareddata.Docs(provider)
		require.NotEmpty(tb, docs)

		res, err := collection.InsertMany(ctx, docs)

		require.NoError(tb, err, "%s: handler %q, collection %s", provider.Name(), f.GetHandler(), fullName)
		require.Len(tb, res.InsertedIDs, len(docs))

		// delete collection unless test failed
		tb.Cleanup(func() {
			if tb.Failed() {
				tb.Logf("Keeping %s for debugging.", fullName)
				return
			}

			err := collection.Drop(ctx)
			require.NoError(tb, err)
		})

		collections = append(collections, collection)

		region.End()
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
