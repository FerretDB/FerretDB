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
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// SetupOpts represents setup options.
//
// TODO add option to use read-only user: https://github.com/FerretDB/FerretDB/issues/914.
type SetupOpts struct {
	// If true, returns two client connections to different systems for compatibility test.
	CompatTest bool

	// Database to use. If empty, temporary test-specific database is created.
	DatabaseName string

	// Data providers.
	Providers []shareddata.Provider
}

// SetupResult represents setup results.
type SetupResult struct {
	Ctx               context.Context
	TargetCollections []*mongo.Collection
	TargetPort        uint16
	CompatCollections []*mongo.Collection
	CompatPort        uint16
}

// SetupWithOpts setups the test according to given options.
func SetupWithOpts(tb testing.TB, opts *SetupOpts) *SetupResult {
	tb.Helper()

	startupOnce.Do(func() {
		logging.Setup(zap.DebugLevel)

		go debug.RunHandler(testutil.Ctx(tb), "127.0.0.1:0", zap.L().Named("debug"))
	})

	if opts == nil {
		opts = new(SetupOpts)
	}

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	level := zap.NewAtomicLevelAt(zap.WarnLevel)
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

	collection := setupCollection(tb, ctx, client, opts.DatabaseName, opts.Providers)

	var compatPort int
	var compatCollection *mongo.Collection
	if opts.CompatTest {
		compatPort = *compatPortF
		if compatPort == 0 {
			tb.Skip("compatibility tests require second system")
		}

		client = setupClient(tb, ctx, compatPort)
		compatCollection = setupCollection(tb, ctx, client, opts.DatabaseName, opts.Providers)
	}

	level.SetLevel(*logLevelF)

	return &SetupResult{
		Ctx:               ctx,
		TargetCollections: []*mongo.Collection{collection},
		TargetPort:        uint16(port),
		CompatCollections: []*mongo.Collection{compatCollection},
		CompatPort:        uint16(compatPort),
	}
}

// Setup setups test with specified data providers.
func Setup(tb testing.TB, providers ...shareddata.Provider) (context.Context, *mongo.Collection) {
	tb.Helper()

	s := SetupWithOpts(tb, &SetupOpts{
		Providers: providers,
	})
	return s.Ctx, s.TargetCollection
}

// SetupCompat setups compatibility test with all data providers.
func SetupCompat(tb testing.TB) (context.Context, *mongo.Collection, *mongo.Collection) {
	tb.Helper()

	s := SetupWithOpts(tb, &SetupOpts{
		CompatTest: true,
		Providers:  shareddata.AllProviders(),
	})
	return s.Ctx, s.TargetCollection, s.CompatCollection
}

// setupListener starts in-process FerretDB server that runs until ctx is done,
// and returns listening port number.
func setupListener(tb testing.TB, ctx context.Context, logger *zap.Logger) int {
	tb.Helper()

	h, err := registry.NewHandler(*handlerF, &registry.NewHandlerOpts{
		Ctx:           ctx,
		Logger:        logger,
		PostgreSQLURL: testutil.PoolConnString(tb, nil),
		TigrisURL:     testutil.TigrisURL(tb),
	})
	require.NoError(tb, err)

	proxyAddr := *proxyAddrF
	mode := clientconn.NormalMode
	if proxyAddr != "" {
		mode = clientconn.DiffNormalMode
	}

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr:         "127.0.0.1:0",
		ProxyAddr:          proxyAddr,
		Mode:               mode,
		Handler:            h,
		Logger:             logger,
		TestRunCancelDelay: time.Hour, // make it easier to notice missing client's disconnects
	})

	done := make(chan struct{})
	go func() {
		defer close(done)

		err := l.Run(ctx)
		if err == nil || err == context.Canceled {
			logger.Info("Listener stopped without error")
		} else {
			logger.Error("Listener stopped", zap.Error(err))
		}
	}()

	// ensure that all listener's logs are written before test ends
	tb.Cleanup(func() {
		<-done
		h.Close()
	})

	port := l.Addr().(*net.TCPAddr).Port
	logger.Info("Listener started", zap.String("handler", *handlerF), zap.Int("port", port))

	return port
}

// setupCollection setups a single collection.
func setupCollection(tb testing.TB, ctx context.Context, client *mongo.Client, db string, providers []shareddata.Provider) *mongo.Collection {
	tb.Helper()

	require.NotEmpty(tb, providers)

	var ownDatabase bool
	if db == "" {
		db = testutil.DatabaseName(tb)
		ownDatabase = true
	}

	database := client.Database(db)
	collectionName := testutil.CollectionName(tb)
	collection := database.Collection(collectionName)

	// drop remnants of the previous failed run
	_ = collection.Drop(ctx)
	if ownDatabase {
		_ = database.Drop(ctx)
	}

	for _, provider := range providers {
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
