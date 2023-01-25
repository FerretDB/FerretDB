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

	// TODO https://github.com/FerretDB/FerretDB/issues/1568
	handlerF          = flag.String("handler", "pg", "handler to use for in-process FerretDB")
	targetUnixSocketF = flag.Bool("target-unix-socket", false, "use Unix socket for in-process FerretDB if possible")
	proxyAddrF        = flag.String("proxy-addr", "", "proxy to use for in-process FerretDB")

	compatPortF = flag.Int("compat-port", 37017, "compat system's port for compatibility tests; if 0, they are skipped")
	compatTLSF  = flag.Bool("compat-tls", false, "use TLS for compat system")

	postgreSQLURLF = flag.String("postgresql-url", "", "PostgreSQL URL for 'pg' handler.")

	tigrisURLF = flag.String("tigris-url", "127.0.0.1:8081", "Tigris URL for 'tigris' handler.")

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	recordsDirF = flag.String("records-dir", "", "directory for record files")

	startupOnce sync.Once
)

// Flags store flags used for test setup.
type Flags struct {
	TargetPort       *int
	TargetTLS        *bool
	Handler          *string
	TargetUnixSocket *bool
	ProxyAddr        *string
	CompatPort       *int
	CompatTLS        *bool
	PostgreSQLURL    *string
	TigrisURL        *string
}

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

// buildMongoDBURIOpts represents connectMongoDB's options.
type buildMongoDBURIOpts struct {
	host          string
	authMechanism string
	user          *url.Userinfo
	tls           bool
}

// connectMongoDB connects to mongoDB.
//
// TODO rework or remove this https://github.com/FerretDB/FerretDB/issues/1568
func connectMongoDB(tb testing.TB, ctx context.Context, opts *buildMongoDBURIOpts) *mongo.Client {
	tb.Helper()

	uri := buildMongoDBURI(opts)

	return setupClient(tb, ctx, uri, opts.tls)
}

// buildMongoDBURI builds MongoDB URI with given URI options and validates that it works.
func buildMongoDBURI(opts *buildMongoDBURIOpts) string {
	q := make(url.Values)
	q.Set("authMechanism", opts.authMechanism)

	// TODO https://github.com/FerretDB/FerretDB/issues/1507
	u := &url.URL{
		Scheme:   "mongodb",
		Host:     opts.host,
		Path:     "/",
		User:     opts.user,
		RawQuery: q.Encode(),
	}

	return u.String()
}

// setupListener starts in-process FerretDB server that runs until ctx is done.
// It returns MongoDB URI for that listener.
func setupListener(tb testing.TB, ctx context.Context, logger *zap.Logger, f Flags) *mongo.Client {
	tb.Helper()

	defer trace.StartRegion(ctx, "setupListener").End()

	require.Zero(tb, f.GetTargetPort(), "-target-port must be 0 for in-process FerretDB")

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
		Ctx:           ctx,
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: *postgreSQLURLF,

		TigrisURL: *tigrisURLF,
	}
	h, err := registry.NewHandler(f.GetHandler(), handlerOpts)
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
		// TODO: should this only be assigned for non unix
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

	opts := &buildMongoDBURIOpts{
		user:          url.UserPassword("username", "password"),
		authMechanism: "PLAIN",
	}

	switch {
	case f.IsTargetTLS():
		opts.host = l.TLSAddr().String()
	case f.IsTargetUnixSocket():
		opts.host = l.UnixAddr().String()
	default:
		opts.host = l.TCPAddr().String()
	}

	uri := buildMongoDBURI(opts)
	client := setupClient(tb, ctx, uri, f.IsTargetTLS())

	logger.Info("Listener started", zap.String("handler", *handlerF), zap.String("uri", uri))

	return client
}

// setupClient returns MongoDB client for database on given MongoDB URI.
func setupClient(tb testing.TB, ctx context.Context, uri string, isTLS bool) *mongo.Client {
	tb.Helper()

	defer trace.StartRegion(ctx, "setupClient").End()
	trace.Log(ctx, "setupClient", uri)

	tb.Logf("setupClient: %s", uri)

	clientOpts := options.Client().ApplyURI(uri)

	if isTLS {
		clientOpts.SetTLSConfig(GetClientTLSConfig(tb))
	}

	client, err := mongo.Connect(ctx, clientOpts)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(tb, err)
	})

	_, err = client.ListDatabases(ctx, bson.D{})
	require.NoError(tb, err)

	return client
}

// startup initializes things that should be initialized only once.
// It returns flag values.
func startup() Flags {
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

	return Flags{
		TargetPort:       targetPortF,
		TargetTLS:        targetTLSF,
		Handler:          handlerF,
		TargetUnixSocket: targetUnixSocketF,
		ProxyAddr:        proxyAddrF,
		CompatPort:       compatPortF,
		CompatTLS:        compatTLSF,
		PostgreSQLURL:    postgreSQLURLF,
		TigrisURL:        tigrisURLF,
	}
}

// ApplyOpts applies non nil value of opts.
func (c *Flags) ApplyOpts(opts Flags) *Flags {
	if opts.TargetPort != nil {
		c.TargetPort = opts.TargetPort
	}

	if opts.TargetTLS != nil {
		c.TargetTLS = opts.TargetTLS
	}

	if opts.Handler != nil {
		c.Handler = opts.Handler
	}

	if opts.TargetUnixSocket != nil {
		c.TargetUnixSocket = opts.TargetUnixSocket
	}

	if opts.ProxyAddr != nil {
		c.ProxyAddr = opts.ProxyAddr
	}

	if opts.CompatPort != nil {
		c.CompatPort = opts.CompatPort
	}

	if opts.CompatTLS != nil {
		c.CompatTLS = opts.CompatTLS
	}

	if opts.PostgreSQLURL != nil {
		c.PostgreSQLURL = opts.PostgreSQLURL
	}

	if opts.TigrisURL != nil {
		c.TigrisURL = opts.TigrisURL
	}

	return c
}

// IsTargetTLS returns true if TargetTLS is set.
func (c *Flags) IsTargetTLS() bool {
	if c.TargetTLS == nil {
		return false
	}

	return *c.TargetTLS
}

// GetTargetPort returns target port number.
func (c *Flags) GetTargetPort() int {
	if c.TargetPort == nil {
		return 0
	}

	return *c.TargetPort
}

// GetHandler returns the handler name.
func (c *Flags) GetHandler() string {
	if c.Handler == nil {
		return ""
	}

	return *c.Handler
}

// IsTargetUnixSocket returns true if TargetUnixSocket is set.
func (c *Flags) IsTargetUnixSocket() bool {
	if c.TargetUnixSocket == nil {
		return false
	}

	return *c.TargetUnixSocket
}

// GetProxyAddr returns proxy address.
func (c *Flags) GetProxyAddr() string {
	if c.ProxyAddr == nil {
		return ""
	}

	return *c.ProxyAddr
}

// IsCompatTLS returns true if CompatTLS is set.
func (c *Flags) IsCompatTLS() bool {
	if c.CompatTLS == nil {
		return false
	}

	return *c.CompatTLS
}

// GetCompatPort returns compat port number.
func (c *Flags) GetCompatPort() int {
	if c.CompatPort == nil {
		return 0
	}

	return *c.CompatPort
}

// GetPostgreSQLURL returns postgreSQL url.
func (c *Flags) GetPostgreSQLURL() string {
	if c.PostgreSQLURL == nil {
		return ""
	}

	return *c.PostgreSQLURL
}

// GetTigrisURL returns tigris url.
func (c *Flags) GetTigrisURL() string {
	if c.TigrisURL == nil {
		return ""
	}

	return *c.TigrisURL
}
