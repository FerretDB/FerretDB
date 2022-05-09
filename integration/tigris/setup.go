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

package tigris

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var startupOnce sync.Once

// setupOpts represents setup options.
type setupOpts struct {
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

	databaseName := testutil.SchemaName(t)

	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))
	ctx, cancel := context.WithCancel(context.Background())

	port := setupListener(t, ctx, logger)
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

	tigrisAddress := "127.0.0.1:8081"

	conf := &config.Driver{
		URL: tigrisAddress,
	}
	c, err := driver.NewDriver(ctx, conf)
	require.NoError(t, err)

	t.Log("recreate db", databaseName)
	_ = c.DropDatabase(ctx, databaseName)
	err = c.CreateDatabase(ctx, databaseName)
	require.NoError(t, err)

	collectionName := "values"
	// create a new collection: tigris requies strict collection definition
	collectionSchema := driver.Schema(`{
	"title" : "values",
	"properties": {
		"_id": {
			"type": "string"
		},
		"string": {
			"type": "string"
		},
		"number": {
			"type": "number"
		}
	},
	"primary_key": ["_id"]
  }`)

	t.Log("create collection", databaseName, collectionName)
	err = c.UseDatabase(databaseName).CreateOrUpdateCollection(ctx, collectionName, collectionSchema)
	require.NoError(t, err)

	db := client.Database(databaseName)
	collection := db.Collection(collectionName)

	t.Cleanup(func() {
		c.Close()
	})

	// insert all provided data
	for _, provider := range opts.providers {
		for _, doc := range provider.Docs() {
			_, err = collection.InsertOne(ctx, doc.Map())
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

	tgConn, err := tigris.NewConn("localhost:8081", logger, false)
	if err != nil {
		logger.Fatal(err.Error())
	}
	tgHandlerOpts := &tigris.NewOpts{
		Conn:      tgConn,
		L:         logger,
		StartTime: time.Now(),
	}
	handler := tigris.New(tgHandlerOpts)

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr: "127.0.0.1:0",
		Mode:       clientconn.NormalMode,
		Handler:    handler,
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

	logging.Setup(zap.DebugLevel)

	ctx := context.Background()

	go debug.RunHandler(ctx, "127.0.0.1:0", zap.L().Named("debug"))
}
