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
	targetPortF = flag.Int("target-port", 0, "target system's port for tests; if 0, in-process FerretDB is used")
	targetTLSF  = flag.Bool("target-tls", false, "use TLS for target system")

	postgreSQLURLF    = flag.String("postgresql-url", "", "PostgreSQL URL for 'pg' handler.")
	tigrisURLSF       = flag.String("tigris-urls", "", "Tigris URLs for 'tigris' handler in comma separated list.")
	targetUnixSocketF = flag.Bool("target-unix-socket", false, "use Unix socket for in-process FerretDB if possible")
	proxyAddrF        = flag.String("proxy-addr", "", "proxy to use for in-process FerretDB")

	compatPortF = flag.Int("compat-port", 37017, "compat system's (MongoDB) port for compatibility tests; if 0, they are skipped")
	compatTLSF  = flag.Bool("compat-tls", false, "use TLS for compat system")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	recordsDirF = flag.String("records-dir", "", "directory for record files")

	// TODO https://github.com/FerretDB/FerretDB/issues/1912
	_ = flag.Bool("disable-pushdown", false, "disable query pushdown")
)

// Other globals.
var (
	// See docker-compose.yml.
	tigrisURLsIndex atomic.Uint32
)

// IsTigris returns if tests are running against the `tigris`handler.
//
// This function should not be used lightly.
func IsTigris(tb testing.TB) bool {
	tb.Helper()

	return getHandler() == "tigris"
}

// SkipForTigris is deprecated.
//
// Deprecated: use SkipForTigrisWithReason instead if you must.
func SkipForTigris(tb testing.TB) {
	SkipForTigrisWithReason(tb, "")
}

// SkipForTigrisWithReason skips the current test for `tigris` handler.
//
// This function should not be used lightly.
func SkipForTigrisWithReason(tb testing.TB, reason string) {
	tb.Helper()

	if getHandler() == "tigris" {
		if reason == "" {
			tb.Skipf("Skipping for Tigris")
		} else {
			tb.Skipf("Skipping for Tigris: %s", reason)
		}
	}
}

// TigrisOnlyWithReason skips the current test for handlers that are not `tigris`.
//
// This function should not be used lightly.
func TigrisOnlyWithReason(tb testing.TB, reason string) {
	tb.Helper()

	if getHandler() != "tigris" {
		tb.Skipf("Skipping for non-tigris: %s", reason)
	}
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

// nextTigrisUrl returns the next url for the Tigris handler.
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

	require.Zero(tb, *targetPortF, "-target-port must be 0 for in-process FerretDB")

	// only one of postgresql-url and tigris-urls should be set.
	if *tigrisURLSF != "" && *postgreSQLURLF != "" {
		tb.Fatalf("Both postgresql-url and tigris-urls are set, only one should be set.")
	}

	// one of postgresql-url or tigris-urls should be set.
	if *tigrisURLSF == "" && *postgreSQLURLF == "" {
		tb.Fatalf("Both postgresql-url and tigris-urls are empty, one should be set.")
	}

	p, err := state.NewProvider("")
	require.NoError(tb, err)

	metrics := connmetrics.NewListenerMetrics()

	handlerOpts := &registry.NewHandlerOpts{
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: *postgreSQLURLF,

		TigrisURL: nextTigrisUrl(),
	}
	h, err := registry.NewHandler(getHandler(), handlerOpts)
	require.NoError(tb, err)

	listenerOpts := &clientconn.NewListenerOpts{
		ProxyAddr:      *proxyAddrF,
		Mode:           clientconn.NormalMode,
		Metrics:        metrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: *recordsDirF,
	}

	if *proxyAddrF != "" {
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
		opts.unixSocketPath = l.UnixAddr().String()
	default:
		opts.hostPort = l.TCPAddr().String()
	}

	uri := buildMongoDBURI(tb, &opts)
	client := setupClient(tb, ctx, uri, *targetTLSF)

	logger.Info("Listener started", zap.String("handler", getHandler()), zap.String("uri", uri))

	return client, uri
}

// setupClient returns MongoDB client for database on given MongoDB URI.
func setupClient(tb testing.TB, ctx context.Context, uri string, isTLS bool) *mongo.Client {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupClient")
	defer span.End()

	defer trace.StartRegion(ctx, "setupClient").End()
	trace.Log(ctx, "setupClient", uri)

	tb.Logf("setupClient: %s", uri)

	clientOpts := options.Client().ApplyURI(uri).SetMonitor(otelmongo.NewMonitor())

	// set TLSConfig to the client option, this adds
	// RootCAs and Certificates.
	if isTLS {
		clientOpts.SetTLSConfig(GetClientTLSConfig(tb))
	}

	client, err := mongo.Connect(ctx, clientOpts)
	require.NoError(tb, err, "URI: %s", uri)

	tb.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(tb, err)
	})

	_, err = client.ListDatabases(ctx, bson.D{})
	require.NoError(tb, err)

	return client
}

// getHandler returns the handler based on the URL.
//
//   - when `-postgresql-url` flag is set, it is `pg` handler;
//   - when `tigris-urls` flag is set, it is `tigris` handler;
//   - and the handler is empty for MongoDB.
func getHandler() string {
	if *postgreSQLURLF != "" {
		return "pg"
	}

	if *tigrisURLSF != "" {
		return "tigris"
	}

	return ""
}

// getUser returns test user credential if TLS is enabled, nil otherwise.
func getUser(isTLS bool) *url.Userinfo {
	if isTLS {
		return url.UserPassword("username", "password")
	}

	return nil
}
