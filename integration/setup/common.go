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
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

var (
	targetPortF = flag.Int("target-port", 0, "target system's port for tests; if 0, in-process FerretDB is used")
	targetTLSF  = flag.Bool("target-tls", false, "use TLS for target system")

	targetUnixSocketF = flag.Bool("target-unix-socket", false, "use Unix socket for in-process FerretDB if possible")
	proxyAddrF        = flag.String("proxy-addr", "", "proxy to use for in-process FerretDB")

	compatPortF = flag.Int("compat-port", 37017, "compat system's port for compatibility tests; if 0, they are skipped")
	compatTLSF  = flag.Bool("compat-tls", false, "use TLS for compat system")

	postgreSQLURLF = flag.String("postgresql-url", "", "PostgreSQL URL for 'pg' handler.")

	tigrisURLsF = flag.String("tigris-urls", "", "Tigris URLs for 'tigris' handler in comma separated list.")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	recordsDirF = flag.String("records-dir", "", "directory for record files")

	startupOnce sync.Once
	startupEnv  *startupInitializer
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

	if getHandler() == "tigris" {
		if reason == "" {
			tb.Skipf("Skipping for Tigris")
		} else {
			tb.Skipf("Skipping for Tigris: %s", reason)
		}
	}
}

// IsTigris returns if tests are running against the Tigris handler.
func IsTigris(tb testing.TB) bool {
	tb.Helper()
	return getHandler() == "tigris"
}

// ApplyForTigrisOnlyWithReason applies the current test for Tigris handler only.
// It skips on PosgreSQL and MongoDB.
//
// Ideally, this function should not be used. It is allowed to use it in Tigris-specific tests only.
func ApplyForTigrisOnlyWithReason(tb testing.TB, reason string) {
	tb.Helper()

	if getHandler() == "pg" {
		tb.Skipf("Skipping for Postgres: %s", reason)
	}

	if getHandler() == "" {
		tb.Skipf("Skipping for MongoDB: %s", reason)
	}
}

// buildMongoDBURIOpts represents buildMongoDBURI's options.
type buildMongoDBURIOpts struct {
	host           string
	unixSocketPath string
	tls            bool
	authMechanism  string
	user           *url.Userinfo
}

// buildMongoDBURI builds MongoDB URI with given URI options.
func buildMongoDBURI(tb testing.TB, opts *buildMongoDBURIOpts) string {
	q := make(url.Values)

	if opts.tls {
		require.Empty(tb, opts.unixSocketPath, "unixSocketPath cannot be used with TLS")
		q.Set("tls", "true")
	}

	var host string
	if opts.host != "" {
		require.Empty(tb, opts.unixSocketPath, "unixSocketPath and TCP/TLS cannot be both set")
		host = opts.host
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

// setupListener starts in-process FerretDB server that runs until ctx is done.
// It returns client and MongoDB URI of that listener.
func setupListener(tb testing.TB, ctx context.Context, logger *zap.Logger, s *startupInitializer, f flags) (*mongo.Client, string) {
	tb.Helper()

	defer trace.StartRegion(ctx, "setupListener").End()

	require.Zero(tb, f.GetTargetPort(), "-target-port must be 0 for in-process FerretDB")

	p, err := state.NewProvider("")
	require.NoError(tb, err)

	metrics := connmetrics.NewListenerMetrics()

	handlerOpts := &registry.NewHandlerOpts{
		Ctx:           ctx,
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: f.GetPostgreSQLURL(),

		TigrisURL: s.getNextTigrisURL(),
	}
	h, err := registry.NewHandler(getHandler(), handlerOpts)
	require.NoError(tb, err)

	listenerOpts := &clientconn.NewListenerOpts{
		ProxyAddr:      f.GetProxyAddr(),
		Mode:           clientconn.NormalMode,
		Metrics:        metrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: *recordsDirF,
	}

	if f.GetProxyAddr() != "" {
		listenerOpts.Mode = clientconn.DiffNormalMode
	}

	if f.IsTargetUnixSocket() {
		listenerOpts.Unix = unixSocketPath(tb)
	}

	addr := "127.0.0.1:0"

	if f.IsTargetTLS() {
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

	var opts buildMongoDBURIOpts

	switch {
	case f.IsTargetTLS():
		opts.host = l.TLSAddr().String()
		opts.user = url.UserPassword("username", "password")
		opts.authMechanism = "PLAIN"
		opts.tls = true
	case f.IsTargetUnixSocket():
		opts.unixSocketPath = l.UnixAddr().String()
	default:
		opts.host = l.TCPAddr().String()
	}

	// When TLS is enabled, RootCAs and Certificates are fetched
	// upon creating client. Target uses PLAIN for authMechanism.
	// Listeners are created only for Target since targetPort must be 0,
	// for listener creation, so we use f.IsTargetTLS() to find TLS.
	uri := buildMongoDBURI(tb, &opts)
	client := setupClient(tb, ctx, uri, f.IsTargetTLS())

	logger.Info("Listener started", zap.String("handler", f.GetHandler()), zap.String("uri", uri))

	return client, uri
}

// validateFlags validates flag values.
func validateFlags(tb testing.TB) {
	tb.Helper()

	// only one of postgresql-url and tigris-urls should be set.
	if *tigrisURLsF != "" && *postgreSQLURLF != "" {
		tb.Fatalf("postgresql-url and tigris-urls must not be both set, only one should be set.")
	}

	// target-port is required when neither postgresql-url nor tigris-urls is set
	if *tigrisURLsF == "" && *postgreSQLURLF == "" && *targetPortF == 0 {
		tb.Fatalf("target-port must be non-zero for empty postgresql-url and tigris-urls.")
	}
}

// getHandler gets pg, tigris or empty for mongoDB.
func getHandler() string {
	if *tigrisURLsF != "" {
		return "tigris"
	}

	if *postgreSQLURLF != "" {
		return "pg"
	}

	return ""
}

// setupClient returns MongoDB client for database on given MongoDB URI.
func setupClient(tb testing.TB, ctx context.Context, uri string, isTLS bool) *mongo.Client {
	tb.Helper()

	defer trace.StartRegion(ctx, "setupClient").End()
	trace.Log(ctx, "setupClient", uri)

	tb.Logf("setupClient: %s", uri)

	clientOpts := options.Client().ApplyURI(uri)

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

// getUser returns test user credential if TLS is enabled, nil otherwise.
func getUser(isTLS bool) *url.Userinfo {
	if isTLS {
		return url.UserPassword("username", "password")
	}

	return nil
}

// startup initializes things that should be initialized only once.
func startup(tb testing.TB) *startupInitializer {
	startupOnce.Do(func() {
		logging.Setup(zap.DebugLevel, "")

		go debug.RunHandler(context.Background(), "127.0.0.1:0", prometheus.DefaultRegisterer, zap.L().Named("debug"))

		if p := *targetPortF; p == 0 {
			zap.S().Infof("Target system: in-process FerretDB with %q handler.", getHandler())
		} else {
			zap.S().Infof("Target system: port %d.", p)
		}

		if p := *compatPortF; p == 0 {
			zap.S().Infof("Compat system: none, compatibility tests will be skipped.")
		} else {
			zap.S().Infof("Compat system: port %d.", p)
		}
		startupEnv = newStartupInitializer(tb, *tigrisURLsF)
	})

	return startupEnv
}
