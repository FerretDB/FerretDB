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
	compatPortF = flag.Int("compat-port", 37017, "second system's port for compatibility tests; if 0, they are skipped")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	startupOnce sync.Once
)

// SkipForTigris skips the current test for Tigris handler.
//
// This function should be removed soon. It should not be used in new tests.
func SkipForTigris(tb testing.TB) {
	tb.Helper()

	if *handlerF == "tigris" {
		tb.Skip("Skipping for Tigris")
	}
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

// setupClient returns MongoDB client for database on 127.0.0.1:port.
func setupClient(tb testing.TB, ctx context.Context, port int) *mongo.Client {
	tb.Helper()

	require.Greater(tb, port, 0)
	require.Less(tb, port, 65536)

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

// startup initializes things that should be initialized only once.
func startup() {
	startupOnce.Do(func() {
		logging.Setup(zap.DebugLevel)

		go debug.RunHandler(context.Background(), "127.0.0.1:0", zap.L().Named("debug"))
	})
}
