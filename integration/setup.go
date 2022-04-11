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
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var (
	startupPort = flag.String("port", "ferretdb", "port to use")

	startupOnce sync.Once
)

// setupOpts represents setup options.
type setupOpts struct {
	// Database to use. If empty, temporary test-specific database is created.
	databaseName string

	// Data providers.
	providers []shareddata.Provider
}

// setupWithOpts setups the test according to given options,
// and returns test-specific context (that is cancelled when the test ends) and database collection.
func setupWithOpts(t *testing.T, opts *setupOpts) (context.Context, *mongo.Collection) {
	t.Helper()

	startupOnce.Do(func() { startup(t) })

	if opts == nil {
		opts = new(setupOpts)
	}

	var ownDatabase bool
	if opts.databaseName == "" {
		opts.databaseName = testutil.SchemaName(t)
		ownDatabase = true
	}

	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))

	port, err := strconv.Atoi(*startupPort)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	if port == 0 {
		port = setupListener(t, ctx, logger)
	}

	// register cleanup function after setupListener's internal registration
	t.Cleanup(cancel)

	uri := fmt.Sprintf("mongodb://127.0.0.1:%d", port)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)
	err = client.Ping(ctx, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(t, err)
	})

	db := client.Database(opts.databaseName)
	collectionName := testutil.TableName(t)
	collection := db.Collection(collectionName)

	// drop remnants of the previous failed run
	_ = collection.Drop(ctx)
	if ownDatabase {
		_ = db.Drop(ctx)
	}

	// create collection explicitly in case there are no docs to insert
	err = db.CreateCollection(ctx, collectionName)
	require.NoError(t, err)

	// delete collection and (possibly) database unless test failed
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("Keeping database %q and collection %q for debugging.", opts.databaseName, collectionName)
			return
		}

		err = collection.Drop(ctx)
		require.NoError(t, err)

		if ownDatabase {
			err = db.Drop(ctx)
			require.NoError(t, err)
		}
	})

	// insert all provided data
	for _, provider := range opts.providers {
		for _, doc := range provider.Docs() {
			_, err = collection.InsertOne(ctx, doc)
			require.NoError(t, err)
		}
	}

	return ctx, collection
}

// setup calls setupWithOpts with specified data providers.
func setup(t *testing.T, providers ...shareddata.Provider) (context.Context, *mongo.Collection) {
	t.Helper()

	return setupWithOpts(t, &setupOpts{
		providers: providers,
	})
}

// setupListener starts in-process FerretDB server that runs until ctx is cancelled,
// and returns listening port number.
func setupListener(t *testing.T, ctx context.Context, logger *zap.Logger) int {
	t.Helper()

	pgPool := testutil.Pool(ctx, t, nil, logger)

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr: "127.0.0.1:0",
		Mode:       clientconn.NormalMode,
		PgPool:     pgPool,
		Logger:     logger,
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
	t.Cleanup(func() {
		<-done
	})

	return l.Addr().(*net.TCPAddr).Port
}

// startup initializes things that should be initialized only once.
func startup(t *testing.T) {
	t.Helper()

	*startupPort = strings.ToLower(*startupPort)
	switch *startupPort {
	case "ferretdb":
		*startupPort = "0"
	case "default":
		*startupPort = "27017"
	case "mongodb":
		*startupPort = "37017"
	}

	if _, err := strconv.Atoi(*startupPort); err != nil {
		t.Fatal(err)
	}

	logging.Setup(zap.DebugLevel)

	ctx := context.Background()

	go debug.RunHandler(ctx, "127.0.0.1:0", zap.L().Named("debug"))
}
