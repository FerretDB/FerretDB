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

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var (
	startupPort = flag.String("port", "ferretdb", "port to use")

	startupOnce sync.Once
)

// setup returns test-specific context (that is cancelled when the test ends) and database client.
func setup(t *testing.T) (context.Context, *mongo.Database) {
	t.Helper()

	startupOnce.Do(func() { startup(t) })

	logger := zaptest.NewLogger(t, zaptest.Level(zap.DebugLevel))

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	port, err := strconv.Atoi(*startupPort)
	if err != nil {
		t.Fatal(err)
	}

	// start in-process FerretDB if port is not set
	if port == 0 {
		pgPool := testutil.Pool(ctx, t, nil, logger)

		l := clientconn.NewListener(&clientconn.NewListenerOpts{
			ListenAddr: "127.0.0.1:0",
			Mode:       clientconn.NormalMode,
			PgPool:     pgPool,
			Logger:     logger.Named("listener"),
		})

		wg.Add(1)
		go func() {
			defer wg.Done()

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

	databaseName := databaseName(t)
	db := client.Database(databaseName)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		if t.Failed() {
			t.Logf("Keeping database %q for debugging.", databaseName)
		} else {
			client.Database(databaseName)
			err = db.Drop(context.Background())
			require.NoError(t, err)
		}

		err = client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	err = db.Drop(context.Background())
	require.NoError(t, err)

	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	return ctx, db
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
