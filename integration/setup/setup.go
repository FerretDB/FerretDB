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

// Package setup provides integration tests setup helpers.
package setup

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var (
	targetPortF = flag.Int("target-port", 0, "target system's port for tests; if 0, in-process FerretDB is used")
	proxyAddrF  = flag.String("proxy-addr", "", "proxy to use for in-process FerretDB")
	handlerF    = flag.String("handler", "pg", "handler to use for in-process FerretDB")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	startupOnce sync.Once
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

	collection := setupCollection(tb, ctx, port, opts.DatabaseName, opts.Providers)

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
			logger.Info("Listener stopped")
		} else {
			logger.Error("Listener stopped", zap.Error(err))
		}
	}()

	// ensure that all listener's logs are written before test ends
	tb.Cleanup(func() {
		<-done
		h.Close()
	})

	return l.Addr().(*net.TCPAddr).Port
}

// setupCollection setups a single collection.
// If there are no providers, we don't create a database and collection.
// That is intentional:
//   * for those tests where no collection and database are needed.
//   * for Tigris: we can't create a collection without a schema, and we don't know schema without documents.
func setupCollection(tb testing.TB, ctx context.Context, port int, db string, providers []shareddata.Provider) *mongo.Collection {
	tb.Helper()

	require.Greater(tb, port, 0)
	require.Less(tb, port, 65536)

	var ownDatabase bool
	if db == "" {
		db = testutil.DatabaseName(tb)
		ownDatabase = true
	}

	client := setupClient(tb, ctx, uint16(port))
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
		require.NoError(tb, err)
		require.Len(tb, res.InsertedIDs, len(docs))
	}

	// delete collection and (possibly) database unless test failed
	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping database %q and collection %q for debugging.", db, collectionName)
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

// setupClient returns MongoDB client for database on 127.0.0.1:port.
func setupClient(tb testing.TB, ctx context.Context, port uint16) *mongo.Client {
	tb.Helper()

	// those options should not affect anything except tests speed
	v := url.Values{
		// TODO: Test fails occured on some platforms due to i/o timeout.
		// Needs more investigation.
		//
		//"connectTimeoutMS":         []string{"5000"},
		//"serverSelectionTimeoutMS": []string{"5000"},
		//"socketTimeoutMS":          []string{"5000"},
		//"heartbeatFrequencyMS":     []string{"30000"},

		//"minPoolSize":   []string{"1"},
		//"maxPoolSize":   []string{"1"},
		//"maxConnecting": []string{"1"},
		//"maxIdleTimeMS": []string{"0"},

		//"directConnection": []string{"true"},
		//"appName":          []string{tb.Name()},
	}

	u := url.URL{
		Scheme:   "mongodb",
		Host:     fmt.Sprintf("127.0.0.1:%d", port),
		Path:     "/",
		RawQuery: v.Encode(),
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(u.String()))
	require.NoError(tb, err)

	err = client.Ping(ctx, nil)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(tb, err)
	})

	return client
}
