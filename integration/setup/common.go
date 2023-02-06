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
	"net/url"
	"runtime/trace"
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

	// TODO https://github.com/FerretDB/FerretDB/issues/1568
	handlerF          = flag.String("handler", "pg", "handler to use for in-process FerretDB")
	targetUnixSocketF = flag.Bool("target-unix-socket", false, "use Unix socket for in-process FerretDB if possible")
	proxyAddrF        = flag.String("proxy-addr", "", "proxy to use for in-process FerretDB")

	compatPortF = flag.Int("compat-port", 37017, "compat system's port for compatibility tests; if 0, they are skipped")
	compatTLSF  = flag.Bool("compat-tls", false, "use TLS for compat system")

	postgreSQLURLF = flag.String("postgresql-url", "", "PostgreSQL URL for 'pg' handler.")

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
	tigrisPorts      = []uint16{8081, 8091, 8092, 8093, 8094}
	tigrisPortsIndex atomic.Uint32
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

// IsTigris returns if tests are running against the Tigris handler.
func IsTigris(tb testing.TB) bool {
	tb.Helper()
	return *handlerF == "tigris"
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

// checkMongoDBURI returns true if given MongoDB URI is working.
func checkMongoDBURI(tb testing.TB, ctx context.Context, uri string) bool {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "checkMongoDBURI")
	defer span.End()

	defer trace.StartRegion(ctx, "checkMongoDBURI").End()
	trace.Log(ctx, "checkMongoDBURI", uri)

	clientOpts := options.Client().ApplyURI(uri).SetMonitor(otelmongo.NewMonitor())

	if *targetTLSF {
		clientOpts.SetTLSConfig(GetClientTLSConfig(tb))
	}

	client, err := mongo.Connect(ctx, clientOpts)

	if err == nil {
		defer client.Disconnect(ctx)

		_, err = client.ListDatabases(ctx, bson.D{})
	}

	if err != nil {
		tb.Logf("checkMongoDBURI: %s: %s", uri, err)

		return false
	}

	tb.Logf("checkMongoDBURI: %s: connected", uri)

	return true
}

// buildMongoDBURIOpts represents buildMongoDBURI's options.
type buildMongoDBURIOpts struct {
	hostPort       string
	unixSocketPath string
	tls            bool
}

// buildMongoDBURI builds MongoDB URI with given URI options and validates that it works.
//
// TODO rework or remove this https://github.com/FerretDB/FerretDB/issues/1568
func buildMongoDBURI(tb testing.TB, ctx context.Context, opts *buildMongoDBURIOpts) string {
	tb.Helper()

	var host string

	if opts.hostPort != "" {
		require.Empty(tb, opts.unixSocketPath, "both hostPort and unixSocketPath are set")
		host = opts.hostPort
	} else {
		require.NotEmpty(tb, opts.unixSocketPath, "neither hostPort nor unixSocketPath are set")
		host = opts.unixSocketPath
	}

	q := make(url.Values)

	if opts.tls {
		require.Empty(tb, opts.unixSocketPath, "unixSocketPath cannot be used with TLS")
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1507
	u := &url.URL{
		Scheme: "mongodb",
		Host:   host,
		Path:   "/",
	}

	// we don't know if that's FerretDB or MongoDB, so try different auth mechanisms
	for _, c := range []struct {
		user          *url.Userinfo
		authMechanism string
	}{{
		user:          nil,
		authMechanism: "",
	}, {
		user:          url.UserPassword("username", "password"),
		authMechanism: "PLAIN",
	}, {
		user:          url.UserPassword("username", "password"),
		authMechanism: "", // defaults to SCRAM when username is set
	}} {
		u.User = c.user

		q.Del("authMechanism")

		if c.authMechanism != "" {
			q.Set("authMechanism", c.authMechanism)
		}

		u.RawQuery = q.Encode()
		res := u.String()

		if checkMongoDBURI(tb, ctx, res) {
			return res
		}
	}

	tb.Fatalf("buildMongoDBURI: failed for %+v", opts)

	panic("not reached")
}

// nextTigrisPort returns the next port for the Tigris handler.
func nextTigrisPort() uint16 {
	i := int(tigrisPortsIndex.Add(1)) - 1
	return tigrisPorts[i%len(tigrisPorts)]
}

// setupListener starts in-process FerretDB server that runs until ctx is done.
// It returns MongoDB URI for that listener.
func setupListener(tb testing.TB, ctx context.Context, logger *zap.Logger) string {
	tb.Helper()

	_, span := otel.Tracer("").Start(ctx, "setupListener")
	defer span.End()

	defer trace.StartRegion(ctx, "setupListener").End()

	require.Zero(tb, *targetPortF, "-target-port must be 0 for in-process FerretDB")

	// that's already checked by handlers constructors,
	// but here we could produce a better error message
	switch *handlerF {
	case "pg":
		require.NotEmpty(tb, *postgreSQLURLF, "-postgresql-url must be set for 'pg' handler")
	}

	p, err := state.NewProvider("")
	require.NoError(tb, err)

	metrics := connmetrics.NewListenerMetrics()

	handlerOpts := &registry.NewHandlerOpts{
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: *postgreSQLURLF,

		TigrisURL: fmt.Sprintf("127.0.0.1:%d", nextTigrisPort()),
	}
	h, err := registry.NewHandler(*handlerF, handlerOpts)
	require.NoError(tb, err)

	listenerOpts := &clientconn.NewListenerOpts{
		ProxyAddr:      *proxyAddrF,
		Mode:           clientconn.NormalMode,
		Metrics:        metrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: *recordsDirF,
	}

	if listenerOpts.ProxyAddr != "" {
		listenerOpts.Mode = clientconn.DiffNormalMode
	}

	if *targetUnixSocketF {
		listenerOpts.Unix = unixSocketPath(tb)
	}

	if *targetTLSF {
		listenerOpts.TLS = "127.0.0.1:0"
		fp := GetTLSFilesPaths(tb, ServerSide)
		listenerOpts.TLSCertFile, listenerOpts.TLSKeyFile, listenerOpts.TLSCAFile = fp.Cert, fp.Key, fp.CA
	} else {
		listenerOpts.TCP = "127.0.0.1:0"
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
	case listenerOpts.TLS != "":
		opts.hostPort = l.TLSAddr().String()
		opts.tls = true
	case listenerOpts.Unix != "":
		opts.unixSocketPath = l.UnixAddr().String()
	default:
		opts.hostPort = l.TCPAddr().String()
	}

	uri := buildMongoDBURI(tb, ctx, &opts)
	logger.Info("Listener started", zap.String("handler", *handlerF), zap.String("uri", uri))

	return uri
}

// setupClient returns MongoDB client for database on given MongoDB URI.
func setupClient(tb testing.TB, ctx context.Context, uri string) *mongo.Client {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupClient")
	defer span.End()

	defer trace.StartRegion(ctx, "setupClient").End()
	trace.Log(ctx, "setupClient", uri)

	tb.Logf("setupClient: %s", uri)

	clientOpts := options.Client().ApplyURI(uri).SetMonitor(otelmongo.NewMonitor())

	if *targetTLSF {
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
