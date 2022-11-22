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
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var (
	targetPortF       = flag.Int("target-port", 0, "target system's port for tests; if 0, in-process FerretDB is used")
	targetUnixSocketF = flag.Bool("target-unix-socket", false, "use Unix socket for in-process FerretDB if possible")
	proxyAddrF        = flag.String("proxy-addr", "", "proxy to use for in-process FerretDB")
	handlerF          = flag.String("handler", "pg", "handler to use for in-process FerretDB")
	compatPortF       = flag.Int("compat-port", 37017, "second system's port for compatibility tests; if 0, they are skipped")

	postgreSQLURLF = flag.String("postgresql-url", "postgres://postgres@127.0.0.1:5432/ferretdb?pool_min_conns=1", "PostgreSQL URL for 'pg' handler.")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	recordsDirF = flag.String("records-dir", "", "directory for record files")

	startupOnce sync.Once
)

// SkipForTigris skips the current test for Tigris handler.
//
// This function should not be used lightly in new tests and should eventually be removed.
//
// Deprecated: use SkipForTigrisWithReason instead if you must.
func SkipForTigris(tb testing.TB) {
	SkipForTigrisWithReason(tb, "")
}

// SkipForTigrisWithReason skips the current test for Tigris handler.
//
// This function should not be used lightly in new tests and should eventually be removed.
func SkipForTigrisWithReason(tb testing.TB, reason string) {
	tb.Helper()

	if *handlerF == "tigris" {
		if reason == "" {
			tb.Skipf("Skipping for Tigris")
		} else {
			tb.Skipf("Skipping for Tigris: %s", reason)
		}
	}
}

// SkipForPostgresWithReason skips the current test for Postgres (pg) handler.
//
// Ideally, this function should not be used. It is allowed to use it in Tigris-specific tests only.
func SkipForPostgresWithReason(tb testing.TB, reason string) {
	tb.Helper()

	if *handlerF == "pg" {
		tb.Skipf("Skipping for Postgres: %s", reason)
	}
}

// setupListener starts in-process FerretDB server that runs until ctx is done,
// and returns listening MongoDB URI.
func setupListener(tb testing.TB, ctx context.Context, logger *zap.Logger, preferUnixSocket bool) (*state.Provider, string) {
	tb.Helper()

	p, err := state.NewProvider("")
	require.NoError(tb, err)

	u, err := url.Parse(*postgreSQLURLF)
	require.NoError(tb, err)

	metrics := connmetrics.NewListenerMetrics()

	h, err := registry.NewHandler(*handlerF, &registry.NewHandlerOpts{
		Ctx:           ctx,
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: u.String(),

		TigrisURL: testutil.TigrisURL(tb),
	})
	require.NoError(tb, err)

	proxyAddr := *proxyAddrF
	mode := clientconn.NormalMode
	if proxyAddr != "" {
		mode = clientconn.DiffNormalMode
	}

	listenUnix := listenUnix(tb)
	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr:     "127.0.0.1:0",
		ListenUnix:     listenUnix,
		ProxyAddr:      proxyAddr,
		Mode:           mode,
		Metrics:        metrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: *recordsDirF,
	})

	done := make(chan struct{})
	go func() {
		defer close(done)

		err := l.Run(ctx)
		if err == nil || errors.Is(err, context.Canceled) {
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

	// use Unix socket if preferred and possible
	if preferUnixSocket && listenUnix != "" {
		// TODO https://github.com/FerretDB/FerretDB/issues/1507
		u := &url.URL{
			Scheme: "mongodb",
			Host:   l.Unix().String(),
		}

		uri := u.String()
		logger.Info("Listener started", zap.String("handler", *handlerF), zap.String("uri", uri))

		return p, uri
	}

	port := l.Addr().(*net.TCPAddr).Port
	uri := buildMongoDBURI(tb, port)
	logger.Info("Listener started", zap.String("handler", *handlerF), zap.String("uri", uri))

	return p, uri
}

// buildMongoDBURI builds MongoDB URI with given TCP port number.
func buildMongoDBURI(tb testing.TB, port int) string {
	require.Greater(tb, port, 0)
	require.Less(tb, port, 65536)

	// TODO https://github.com/FerretDB/FerretDB/issues/1507
	u := &url.URL{
		Scheme: "mongodb",
		Host:   fmt.Sprintf("127.0.0.1:%d", port),
		Path:   "/",
	}

	return u.String()
}

// setupClient returns MongoDB client for database on given MongoDB URI.
func setupClient(tb testing.TB, ctx context.Context, uri string) *mongo.Client {
	tb.Helper()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(tb, err)

	tb.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(tb, err)
	})

	err = client.Ping(ctx, nil)
	require.NoError(tb, err)

	return client
}

// startup initializes things that should be initialized only once.
func startup() {
	startupOnce.Do(func() {
		logging.Setup(zap.DebugLevel, "")

		go debug.RunHandler(context.Background(), "127.0.0.1:0", prometheus.DefaultRegisterer, zap.L().Named("debug"))

		if p := *targetPortF; p == 0 {
			zap.S().Infof("Target system: in-process FerretDB with %q handler.", *handlerF)
		} else {
			zap.S().Infof("Target system: port %d.", p)
		}

		if p := *compatPortF; p == 0 {
			zap.S().Infof("Compat system: none, compatibility tests will be skipped.")
		} else {
			zap.S().Infof("Compat system: port %d.", p)
		}
	})
}
