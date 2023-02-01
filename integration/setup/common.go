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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
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

	// Disable noisy setup logs by default.
	debugSetupF = flag.Bool("debug-setup", false, "enable debug logs for tests setup")
	logLevelF   = zap.LevelFlag("log-level", zap.DebugLevel, "log level for tests")

	recordsDirF = flag.String("records-dir", "", "directory for record files")

	shareListenerF = flag.Bool("share-listener", false, "share listener between tests")

	startupOnce sync.Once

	setupListenerOnce sync.Once
	sharedListenerURI string
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

// mongoClient returns a new connected MongoDB client for the given Mongodb URI.
func mongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(uri)

	if *targetTLSF {
		cfg, err := GetClientTLSConfig()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		opts.SetTLSConfig(cfg)
	}

	// TODO tweak these options?
	// https://github.com/FerretDB/FerretDB/issues/1507

	opts.SetRetryReads(false)
	opts.SetRetryWrites(false)

	opts.SetMinPoolSize(5)
	opts.SetMaxPoolSize(0)
	opts.SetMaxConnecting(100)
	opts.SetMaxConnIdleTime(0)
	opts.SetDirect(true)

	if err := opts.Validate(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return mongo.Connect(ctx, opts)
}

// checkMongoDBURI returns true if given MongoDB URI is working.
func checkMongoDBURI(ctx context.Context, uri string) bool {
	defer trace.StartRegion(ctx, "checkMongoDBURI").End()
	trace.Log(ctx, "checkMongoDBURI", uri)

	client, err := mongoClient(ctx, uri)

	if err == nil {
		defer client.Disconnect(ctx)

		_, err = client.ListDatabases(ctx, bson.D{})
	}

	if err != nil {
		// FIXME
		// tb.Logf("checkMongoDBURI: %s: %s", uri, err)

		return false
	}

	// FIXME
	// tb.Logf("checkMongoDBURI: %s: connected", uri)

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
func buildMongoDBURI(ctx context.Context, opts *buildMongoDBURIOpts) (string, error) {
	var host string

	if opts.hostPort != "" {
		if opts.unixSocketPath != "" {
			return "", lazyerrors.Errorf("both hostPort and unixSocketPath are set")
		}
		host = opts.hostPort
	} else {
		if opts.unixSocketPath == "" {
			return "", lazyerrors.Errorf("neither hostPort nor unixSocketPath are set")
		}
		host = opts.unixSocketPath
	}

	if opts.tls && opts.unixSocketPath != "" {
		return "", lazyerrors.Errorf("unixSocketPath cannot be used with TLS")
	}

	u := &url.URL{
		Scheme: "mongodb",
		Host:   host,
		Path:   "/",
	}

	q := make(url.Values)

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

		if checkMongoDBURI(ctx, res) {
			return res, nil
		}
	}

	return "", lazyerrors.Errorf("buildMongoDBURI: failed for %+v", opts)
}

func getListener(ctx context.Context, logger *zap.Logger) (string, func(), error) {
	defer trace.StartRegion(ctx, "getListener").End()

	if *shareListenerF {
		var err error

		setupListenerOnce.Do(func() {
			sharedListenerURI, _, err = setupListener(context.Background(), zap.NewNop())
		})

		return sharedListenerURI, nil, err
	}

	return setupListener(ctx, logger)
}

// setupListener starts in-process FerretDB server that runs until ctx is done.
// It returns MongoDB URI for that listener and cleanup function.
func setupListener(ctx context.Context, logger *zap.Logger) (string, func(), error) {
	p, err := state.NewProvider("")
	if err != nil {
		return "", nil, lazyerrors.Error(err)
	}

	metrics := connmetrics.NewListenerMetrics()

	h, err := registry.NewHandler(*handlerF, &registry.NewHandlerOpts{
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: *postgreSQLURLF,

		TigrisURL: "127.0.0.1:8081",
	})
	if err != nil {
		return "", nil, lazyerrors.Error(err)
	}

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
		if listenerOpts.Unix, err = unixSocketPath(); err != nil {
			return "", nil, lazyerrors.Error(err)
		}
	}

	if *targetTLSF {
		listenerOpts.TLS = "127.0.0.1:0"
		fp, err := GetTLSFilesPaths(ServerSide)
		if err != nil {
			return "", nil, lazyerrors.Error(err)
		}
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

	uri, err := buildMongoDBURI(ctx, &opts)
	if err != nil {
		return "", nil, lazyerrors.Error(err)
	}

	logger.Info("Listener started", zap.String("handler", *handlerF), zap.String("uri", uri))

	// ensure that all listener's logs are written before test ends
	cleanup := func() {
		<-done
		h.Close()
	}

	return uri, cleanup, nil
}

// setupClient returns MongoDB client for database on given MongoDB URI.
func setupClient(tb testing.TB, ctx context.Context, uri string) *mongo.Client {
	tb.Helper()

	defer trace.StartRegion(ctx, "setupClient").End()
	trace.Log(ctx, "setupClient", uri)

	tb.Logf("setupClient: %s", uri)

	client, err := mongoClient(ctx, uri)
	require.NoError(tb, err, "URI: %s", uri)

	tb.Cleanup(func() {
		err = client.Disconnect(ctx)
		require.NoError(tb, err)
	})

	_, err = client.ListDatabases(ctx, bson.D{})
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
