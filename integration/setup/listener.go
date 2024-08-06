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

package setup

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handler/registry"
	"github.com/FerretDB/FerretDB/internal/util/password"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// unixSocketPath returns temporary Unix domain socket path for that test.
func unixSocketPath(tb testtb.TB) string {
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

// listenerMongoDBURI builds MongoDB URI for in-process FerretDB.
func listenerMongoDBURI(tb testtb.TB, hostPort, unixSocketPath, newAuthDB string, tlsAndAuth bool) string {
	tb.Helper()

	var host string

	if hostPort != "" {
		require.Empty(tb, unixSocketPath, "both hostPort and unixSocketPath are set")
		host = hostPort
	} else {
		host = unixSocketPath
	}

	var user *url.Userinfo
	q := url.Values{}

	if tlsAndAuth {
		require.Empty(tb, unixSocketPath, "unixSocketPath cannot be used with TLS")

		// we don't separate TLS and auth just for simplicity of our test configurations
		q = url.Values{
			"tls":                   []string{"true"},
			"tlsCertificateKeyFile": []string{filepath.Join(testutil.BuildCertsDir, "client.pem")},
			"tlsCaFile":             []string{filepath.Join(testutil.BuildCertsDir, "rootCA-cert.pem")},
			"authMechanism":         []string{"PLAIN"},
		}
		user = url.UserPassword("username", "password")
	}

	path := "/"

	if newAuthDB != "" {
		q.Set("authMechanism", "SCRAM-SHA-256")
		path += newAuthDB
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1507
	u := &url.URL{
		Scheme:   "mongodb",
		Host:     host,
		Path:     path,
		User:     user,
		RawQuery: q.Encode(),
	}

	return u.String()
}

// setupListener starts in-process FerretDB server that runs until ctx is canceled.
// It returns basic MongoDB URI for that listener.
func setupListener(tb testtb.TB, ctx context.Context, logger *slog.Logger, opts *BackendOpts) string {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupListener")
	defer span.End()

	require.Empty(tb, *targetURLF, "-target-url must be empty for in-process FerretDB")

	var handler string

	switch *targetBackendF {
	case "ferretdb-postgresql":
		require.NotEmpty(tb, *postgreSQLURLF, "-postgresql-url must be set for %q", *targetBackendF)
		require.Empty(tb, *sqliteURLF, "-sqlite-url must be empty for %q", *targetBackendF)
		require.Empty(tb, *mysqlURLF, "-mysql-url must be empty for %q", *targetBackendF)
		require.Empty(tb, *hanaURLF, "-hana-url must be empty for %q", *targetBackendF)
		handler = "postgresql"

	case "ferretdb-sqlite":
		require.Empty(tb, *postgreSQLURLF, "-postgresql-url must be empty for %q", *targetBackendF)
		require.NotEmpty(tb, *sqliteURLF, "-sqlite-url must be set for %q", *targetBackendF)
		require.Empty(tb, *mysqlURLF, "-mysql-url must be empty for %q", *targetBackendF)
		require.Empty(tb, *hanaURLF, "-hana-url must be empty for %q", *targetBackendF)
		handler = "sqlite"

	case "ferretdb-mysql":
		require.Empty(tb, *postgreSQLURLF, "-postgresql-url must be empty for %q", *targetBackendF)
		require.Empty(tb, *sqliteURLF, "-sqlite-url must be empty for %q", *targetBackendF)
		require.NotEmpty(tb, *mysqlURLF, "-mysql-url must be empty for %q", *targetBackendF)
		require.Empty(tb, *hanaURLF, "-hana-url must be set for %q", *targetBackendF)
		handler = "mysql"

	case "ferretdb-hana":
		require.Empty(tb, *postgreSQLURLF, "-postgresql-url must be empty for %q", *targetBackendF)
		require.Empty(tb, *sqliteURLF, "-sqlite-url must be empty for %q", *targetBackendF)
		require.Empty(tb, *mysqlURLF, "-mysql-url must be empty for %q", *targetBackendF)
		require.NotEmpty(tb, *hanaURLF, "-hana-url must be set for %q", *targetBackendF)
		handler = "hana"

	case "mongodb":
		tb.Fatal("can't start in-process MongoDB")

	default:
		// that should be caught by Startup function
		panic("not reached")
	}

	// use per-test PostgreSQL database to prevent problems with parallel tests
	postgreSQLURLF := *postgreSQLURLF
	if postgreSQLURLF != "" {
		postgreSQLURLF = testutil.TestPostgreSQLURI(tb, ctx, postgreSQLURLF)
	}

	// use per-test directory to prevent handler's/backend's metadata registry
	// read databases owned by concurrent tests
	sqliteURL := *sqliteURLF
	if sqliteURL != "" {
		sqliteURL = testutil.TestSQLiteURI(tb, sqliteURL)
	}

	// user per-test MySQL database to prevent handler's/backend's metadata registry
	// read databases owned by concurrent tests
	mysqlURL := *mysqlURLF
	if mysqlURL != "" {
		mysqlURL = testutil.TestMySQLURI(tb, ctx, mysqlURL)
	}

	sp, err := state.NewProvider("")
	require.NoError(tb, err)

	if opts == nil {
		opts = new(BackendOpts)
	}

	handlerOpts := &registry.NewHandlerOpts{
		Logger:        logger,
		ConnMetrics:   listenerMetrics.ConnMetrics,
		StateProvider: sp,

		PostgreSQLURL: postgreSQLURLF,
		SQLiteURL:     sqliteURL,
		MySQLURL:      mysqlURL,
		HANAURL:       *hanaURLF,

		TestOpts: registry.TestOpts{
			DisablePushdown:         *disablePushdownF,
			CappedCleanupPercentage: opts.CappedCleanupPercentage,
			CappedCleanupInterval:   opts.CappedCleanupInterval,
			EnableNewAuth:           !opts.DisableNewAuth,
			BatchSize:               *batchSizeF,
			MaxBsonObjectSizeBytes:  opts.MaxBsonObjectSizeBytes,
		},
	}

	if !opts.DisableNewAuth {
		handlerOpts.SetupDatabase = "test"
		handlerOpts.SetupUsername = "username"
		handlerOpts.SetupPassword = password.WrapPassword("password")
		handlerOpts.SetupTimeout = 20 * time.Second // CI may be slow for many parallel tests
	}

	h, closeBackend, err := registry.NewHandler(handler, handlerOpts)

	if closeBackend != nil {
		tb.Cleanup(closeBackend)
	}

	require.NoError(tb, err)

	listenerOpts := clientconn.NewListenerOpts{
		ProxyAddr:      *targetProxyAddrF,
		Mode:           clientconn.NormalMode,
		Metrics:        listenerMetrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: testutil.TmpRecordsDir,
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
		listenerOpts.TLSCertFile = filepath.Join(testutil.BuildCertsDir, "server-cert.pem")
		listenerOpts.TLSKeyFile = filepath.Join(testutil.BuildCertsDir, "server-key.pem")
		listenerOpts.TLSCAFile = filepath.Join(testutil.BuildCertsDir, "rootCA-cert.pem")
	case *targetUnixSocketF:
		listenerOpts.Unix = unixSocketPath(tb)
	default:
		listenerOpts.TCP = "127.0.0.1:0"
	}

	l, err := clientconn.Listen(&listenerOpts)
	require.NoError(tb, err)

	runDone := make(chan struct{})

	go func() {
		defer close(runDone)

		runCtx, runSpan := otel.Tracer("").Start(ctx, "setupListener.Run")
		defer runSpan.End()

		l.Run(runCtx)
	}()

	// ensure that all listener's and handler's logs are written before test ends
	tb.Cleanup(func() {
		<-runDone
	})

	var hostPort, unixSocketPath string
	var tlsAndAuth bool

	switch {
	case *targetTLSF:
		hostPort = l.TLSAddr().String()
		tlsAndAuth = true
	case *targetUnixSocketF:
		unixSocketPath = l.UnixAddr().String()
	default:
		hostPort = l.TCPAddr().String()
	}

	uri := listenerMongoDBURI(tb, hostPort, unixSocketPath, handlerOpts.SetupDatabase, tlsAndAuth)

	logger.InfoContext(ctx, "Listener started", slog.String("handler", handler), slog.String("uri", uri))

	return uri
}
