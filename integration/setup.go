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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var (
	startupPort = flag.Int("port", 0, "port to use; 0 will start in-process FerretDB on a random port")

	startupOnce sync.Once
)

// setup returns test-specific context (that is cancelled when the test ends) and database client.
func setup(t *testing.T) (context.Context, *mongo.Database) {
	t.Helper()

	startupOnce.Do(func() { startup(t) })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))

	port := *startupPort
	if port == 0 {
		pgPool := testutil.Pool(ctx, t, nil, logger)

		l := clientconn.NewListener(&clientconn.NewListenerOpts{
			ListenAddr: "127.0.0.1:0",
			Mode:       clientconn.NormalMode,
			PgPool:     pgPool,
			Logger:     logger.Named("listener"),
		})

		go func() {
			err := l.Run(ctx)
			if err == nil || err == context.Canceled {
				logger.Info("Listener stopped")
			} else {
				logger.Error("Listener stopped", zap.Error(err))
			}
		}()

		port = l.Addr().(*net.TCPAddr).Port
	}

	uri := fmt.Sprintf("mongodb://127.0.0.1:%d", port)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)
	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(t, err)
	})

	db := client.Database(databaseName(t))
	err = db.Drop(context.Background())
	require.NoError(t, err)

	return ctx, db
}

// startup initializes things that should be initialized only once.
func startup(t *testing.T) {
	t.Helper()

	logging.Setup(zap.DebugLevel)

	ctx := context.Background()

	go debug.RunHandler(ctx, "127.0.0.1:0", zap.L().Named("debug"))
}
