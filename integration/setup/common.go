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
	"net/url"
	"os"
	"path/filepath"
	"runtime/trace"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Flags.
var (
	targetURLF     = flag.String("target-url", "", "target system's URL; if empty, in-process FerretDB is used")
	targetBackendF = flag.String("target-backend", "", "target system's backend: '%s'"+strings.Join(allBackends, "', '"))

	targetProxyAddrF  = flag.String("target-proxy-addr", "", "in-process FerretDB: use given proxy")
	targetTLSF        = flag.Bool("target-tls", false, "in-process FerretDB: use TLS")
	targetUnixSocketF = flag.Bool("target-unix-socket", false, "in-process FerretDB: use Unix socket")

	postgreSQLURLF = flag.String("postgresql-url", "", "in-process FerretDB: PostgreSQL URL for 'pg' handler.")
	tigrisURLSF    = flag.String("tigris-urls", "", "in-process FerretDB: Tigris URLs for 'tigris' handler (comma separated)")

	compatURLF = flag.String("compat-url", "", "compat system's (MongoDB) URL for compatibility tests; if empty, they are skipped")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	recordsDirF = flag.String("records-dir", "", "directory for record files")

	disablePushdownF = flag.Bool("disable-pushdown", false, "disable query pushdown")
)

// Other globals.
var (
	// See docker-compose.yml.
	tigrisURLsIndex atomic.Uint32

	allBackends = []string{"ferretdb-pg", "ferretdb-tigris", "mongodb"}

	tlsClientFile = filepath.Join("..", "build", "certs", "client.pem")
	tlsServerFile = filepath.Join("..", "build", "certs", "server.pem")
	tlsRootCAFile = filepath.Join("..", "build", "certs", "rootCA-cert.pem")
)

// IsTigris returns true if tests are running against FerretDB with `tigris` handler.
//
// This function should not be used lightly.
func IsTigris(tb testing.TB) bool {
	tb.Helper()

	return *targetBackendF == "ferretdb-tigris"
}

// SkipForTigris is deprecated.
//
// Deprecated: use SkipForTigrisWithReason instead if you must.
func SkipForTigris(tb testing.TB) {
	SkipForTigrisWithReason(tb, "empty, please update this test")
}

// SkipForTigrisWithReason skips the current test for FerretDB with `tigris` handler.
//
// This function should not be used lightly.
func SkipForTigrisWithReason(tb testing.TB, reason string) {
	tb.Helper()

	require.NotEmpty(tb, reason, "reason must not be empty")

	tb.Skipf("Skipping for Tigris: %s.", reason)
}

// TigrisOnlyWithReason skips the current test except for FerretDB with `tigris` handler.
//
// This function should not be used lightly.
func TigrisOnlyWithReason(tb testing.TB, reason string) {
	tb.Helper()

	require.NotEmpty(tb, reason, "reason must not be empty")

	if !IsTigris(tb) {
		tb.Skipf("Skipping for non-tigris: %s", reason)
	}
}

// IsPushdownDisabled returns if FerretDB pushdowns are disabled.
func IsPushdownDisabled() bool {
	return *disablePushdownF
}

// buildMongoDBURIOpts represents buildMongoDBURI's options.
type buildMongoDBURIOpts struct {
	hostPort       string // for TCP and TLS
	unixSocketPath string
	tlsAndAuth     bool
}

// buildMongoDBURI builds MongoDB URI with given URI options.
func buildMongoDBURI(tb testing.TB, opts *buildMongoDBURIOpts) string {
	tb.Helper()

	var user *url.Userinfo
	q := make(url.Values)

	var host string
	if opts.hostPort != "" {
		require.Empty(tb, opts.unixSocketPath, "both hostPort and unixSocketPath are set")
		host = opts.hostPort
	} else {
		host = opts.unixSocketPath
	}

	if opts.tlsAndAuth {
		require.Empty(tb, opts.unixSocketPath, "unixSocketPath cannot be used with TLS")

		q.Set("tls", "true")
		q.Set("tlsCertificateKeyFile", tlsClientFile)
		// FIXME q.Set("tlsCaFile", filepath.Join("..", "build", "certs", "rootCA-cert.pem"))

		// use two combinations (no TLS, no auth) and (TLS, auth)
		// instead of four just for simplicity
		q.Set("authMechanism", "PLAIN")
		user = url.UserPassword("username", "password")
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1507
	u := &url.URL{
		Scheme:   "mongodb",
		Host:     host,
		Path:     "/",
		User:     user,
		RawQuery: q.Encode(),
	}

	return u.String()
}

// nextTigrisUrl returns the next url for the `tigris` handler.
func nextTigrisUrl() string {
	i := int(tigrisURLsIndex.Add(1)) - 1
	urls := strings.Split(*tigrisURLSF, ",")

	return urls[i%len(urls)]
}

// unixSocketPath returns temporary Unix domain socket path for that test.
func unixSocketPath(tb testing.TB) string {
	tb.Helper()

	// do not use tb.TempDir() because generated path is too long on macOS
	f, err := os.CreateTemp("", "ferretdb-*.sock")
	require.NoError(tb, err)

	// remove file so listener could create it (and remove it itself on stop)
	err = f.Close()
	require.NoError(tb, err)
	err = os.Remove(f.Name())
	require.NoError(tb, err)

	return f.Name()
}

// setupListener starts in-process FerretDB server that runs until ctx is done.
// It returns client and MongoDB URI of that listener.
func setupListener(tb testing.TB, ctx context.Context, logger *zap.Logger) (*mongo.Client, string) {
	tb.Helper()

	_, span := otel.Tracer("").Start(ctx, "setupListener")
	defer span.End()

	defer trace.StartRegion(ctx, "setupListener").End()

	require.Empty(tb, *targetURLF, "-target-url must be empty for in-process FerretDB")

	var handler string
	switch *targetBackendF {
	case "ferretdb-pg":
		require.NotEmpty(tb, *postgreSQLURLF, "-postgresql-url must be set for %q", *targetBackendF)
		require.Empty(tb, *tigrisURLSF, "-tigris-urls must be empty for %q", *targetBackendF)
		handler = "pg"
	case "ferretdb-tigris":
		require.Empty(tb, *postgreSQLURLF, "-postgresql-url must be empty for %q", *targetBackendF)
		require.NotEmpty(tb, *tigrisURLSF, "-tigris-urls must be set for %q", *targetBackendF)
		handler = "tigris"
	case "mongodb":
		tb.Fatal("can't start in-process MongoDB")
	default:
		// that should be caught by Startup function
		panic("not reached")
	}

	p, err := state.NewProvider("")
	require.NoError(tb, err)

	metrics := connmetrics.NewListenerMetrics()

	handlerOpts := &registry.NewHandlerOpts{
		Logger:          logger,
		Metrics:         metrics.ConnMetrics,
		StateProvider:   p,
		DisablePushdown: *disablePushdownF,

		PostgreSQLURL: *postgreSQLURLF,

		TigrisURL: nextTigrisUrl(),
	}
	h, err := registry.NewHandler(handler, handlerOpts)
	require.NoError(tb, err)

	listenerOpts := clientconn.NewListenerOpts{
		ProxyAddr:      *targetProxyAddrF,
		Mode:           clientconn.NormalMode,
		Metrics:        metrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: *recordsDirF,
	}

	if *targetProxyAddrF != "" {
		listenerOpts.Mode = clientconn.DiffNormalMode
	}

	if *targetTLSF && *targetUnixSocketF {
		tb.Fatal("Both -target-tls and -target-unix-socket are set.")
	}

	switch {
	case *targetTLSF:
		listenerOpts.TLS = "127.0.0.1:0"
		listenerOpts.TLSCertFile = filepath.Join("..", "build", "certs", "server-cert.pem")
		listenerOpts.TLSKeyFile = filepath.Join("..", "build", "certs", "server-key.pem")
		listenerOpts.TLSCAFile = filepath.Join("..", "build", "certs", "rootCA-cert.pem")
	case *targetUnixSocketF:
		listenerOpts.Unix = unixSocketPath(tb)
	default:
		listenerOpts.TCP = "127.0.0.1:0"
	}

	l := clientconn.NewListener(&listenerOpts)

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

	var clientOpts buildMongoDBURIOpts

	switch {
	case *targetTLSF:
		clientOpts.hostPort = l.TLSAddr().String()
		clientOpts.tlsAndAuth = true
	case *targetUnixSocketF:
		clientOpts.unixSocketPath = l.UnixAddr().String()
	default:
		clientOpts.hostPort = l.TCPAddr().String()
	}

	uri := buildMongoDBURI(tb, &clientOpts)
	client := setupClient(tb, ctx, uri)

	logger.Info("Listener started", zap.String("handler", handler), zap.String("uri", uri))

	return client, uri
}

// makeClient returns new client for the given MongoDB URI.
func makeClient(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(uri)

	clientOpts.SetMonitor(otelmongo.NewMonitor())

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	_, err = client.ListDatabases(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// setupClient returns test-specific client for the given MongoDB URI.
//
// It disconnects automatically when test ends.
func setupClient(tb testing.TB, ctx context.Context, uri string) *mongo.Client {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupClient")
	defer span.End()

	defer trace.StartRegion(ctx, "setupClient").End()

	client, err := makeClient(ctx, uri)
	require.NoError(tb, err, "URI: %s", uri)

	tb.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(tb, err)
	})

	return client
}
