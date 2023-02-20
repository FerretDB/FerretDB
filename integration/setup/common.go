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

	postgreSQLURLF    = flag.String("postgresql-url", "", "in-process FerretDB: PostgreSQL URL for 'pg' handler.")
	tigrisURLSF       = flag.String("tigris-urls", "", "in-process FerretDB: Tigris URLs for 'tigris' handler (comma separated)")
	targetTLSF        = flag.Bool("target-tls", false, "in-process FerretDB: use TLS")
	targetUnixSocketF = flag.Bool("target-unix-socket", false, "in-process FerretDB: use Unix socket (if possible)")
	targetProxyAddrF  = flag.String("target-proxy-addr", "", "in-process FerretDB: use given proxy")

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
	SkipForTigrisWithReason(tb, "")
}

// SkipForTigrisWithReason skips the current test for FerretDB with `tigris` handler.
//
// This function should not be used lightly.
func SkipForTigrisWithReason(tb testing.TB, reason string) {
	tb.Helper()

	if IsTigris(tb) {
		if reason == "" {
			tb.Skipf("Skipping for Tigris")
		}

		tb.Skipf("Skipping for Tigris: %s", reason)
	}
}

// TigrisOnlyWithReason skips the current test except for FerretDB with `tigris` handler.
//
// This function should not be used lightly.
func TigrisOnlyWithReason(tb testing.TB, reason string) {
	tb.Helper()

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
	tls            bool
	authMechanism  string
	user           *url.Userinfo
}

// buildMongoDBURI builds MongoDB URI with given URI options.
func buildMongoDBURI(tb testing.TB, opts *buildMongoDBURIOpts) string {
	tb.Helper()

	q := make(url.Values)

	if opts.tls {
		require.Empty(tb, opts.unixSocketPath, "unixSocketPath cannot be used with TLS")
		q.Set("tls", "true")
		// certificates are set by setupClient
	}

	var host string
	if opts.hostPort != "" {
		require.Empty(tb, opts.unixSocketPath, "both hostPort and unixSocketPath are set")
		host = opts.hostPort
	} else {
		host = opts.unixSocketPath
	}

	if opts.authMechanism != "" {
		q.Set("authMechanism", opts.authMechanism)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1507
	u := &url.URL{
		Scheme:   "mongodb",
		Host:     host,
		Path:     "/",
		User:     opts.user,
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

	listenerOpts := &clientconn.NewListenerOpts{
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

	if *targetUnixSocketF {
		listenerOpts.Unix = unixSocketPath(tb)
	}

	addr := "127.0.0.1:0"

	if *targetTLSF {
		listenerOpts.TLS = addr
		fp := GetTLSFilesPaths(tb, ServerSide)
		listenerOpts.TLSCertFile, listenerOpts.TLSKeyFile, listenerOpts.TLSCAFile = fp.Cert, fp.Key, fp.CA
	} else {
		listenerOpts.TCP = addr
	}

	l := clientconn.NewListener(listenerOpts)

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

	opts := buildMongoDBURIOpts{
		user: getUser(*targetTLSF),
	}

	switch {
	case *targetTLSF:
		opts.hostPort = l.TLSAddr().String()
		opts.authMechanism = "PLAIN"
		opts.tls = true
	case *targetUnixSocketF:
		opts.unixSocketPath = l.UnixAddr().String() // TODO
	default:
		opts.hostPort = l.TCPAddr().String()
	}

	uri := buildMongoDBURI(tb, &opts)
	client := setupClient(tb, ctx, uri)

	logger.Info("Listener started", zap.String("handler", getHandler()), zap.String("uri", uri))

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

// TODO
// getUser returns test user credential if TLS is enabled, nil otherwise.
func getUser(isTLS bool) *url.Userinfo {
	if isTLS {
		return url.UserPassword("username", "password")
	}

	return nil
}
