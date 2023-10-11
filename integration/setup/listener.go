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
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// sharedUnixSocketPath returns temporary Unix domain socket path for all tests.
func sharedUnixSocketPath() string {
	// do not use tb.TempDir() because generated path is too long on macOS
	f, err := os.CreateTemp("", "*-ferretdb.sock")
	must.NoError(err)

	// remove file so listener could create it (and remove it itself on stop)
	must.NoError(f.Close())
	must.NoError(os.Remove(f.Name()))

	zap.S().Infof("Using shared Unix socket: %s.", f.Name())

	return f.Name()
}

// privateUnixSocketPath returns temporary Unix domain socket path for that test.
func privateUnixSocketPath(tb testtb.TB) string {
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
func listenerMongoDBURI(hostPort, unixSocketPath string, tlsAndAuth bool) string {
	var host string

	if hostPort != "" {
		must.BeZero(unixSocketPath)
		host = hostPort
	} else {
		host = unixSocketPath
	}

	var user *url.Userinfo
	var q url.Values

	if tlsAndAuth {
		must.BeZero(unixSocketPath)

		// we don't separate TLS and auth just for simplicity of our test configurations
		q = url.Values{
			"tls":                   []string{"true"},
			"tlsCertificateKeyFile": []string{filepath.Join(CertsRoot, "client.pem")},
			"tlsCaFile":             []string{filepath.Join(CertsRoot, "rootCA-cert.pem")},
			"authMechanism":         []string{"PLAIN"},
		}
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

// makeListener returns non-running listener for flags.
//
// Shared and private listeners are constructed differently.
func makeListener(tb testtb.TB, ctx context.Context, logger *zap.Logger) (*clientconn.Listener, string) {
	must.BeTrue(flags.shareServer == (tb == nil))

	if tb != nil {
		tb.Helper()
	}

	_, span := otel.Tracer("").Start(ctx, "makeListener")
	defer span.End()

	defer observability.FuncCall(ctx)()

	must.BeZero(flags.targetURL)

	var handler string
	var postgreSQLURL, sqliteURL, hanaURL string

	switch flags.targetBackend {
	case "ferretdb-pg":
		must.NotBeZero(flags.postgreSQLURL)
		must.BeZero(flags.sqliteURL)
		must.BeZero(flags.hanaURL)

		handler = "pg"

		if flags.shareServer {
			postgreSQLURL = sharedPostgreSQLURL(flags.postgreSQLURL)
		} else {
			postgreSQLURL = privatePostgreSQLURL(tb, flags.postgreSQLURL)
		}

	case "ferretdb-sqlite":
		must.BeZero(flags.postgreSQLURL)
		must.NotBeZero(flags.sqliteURL)
		must.BeZero(flags.hanaURL)

		handler = "sqlite"

		if flags.shareServer {
			sqliteURL = sharedSQLiteURL(flags.sqliteURL)
		} else {
			sqliteURL = privateSQLiteURL(tb, flags.sqliteURL)
		}

	case "ferretdb-hana":
		must.BeZero(flags.postgreSQLURL)
		must.BeZero(flags.sqliteURL)
		must.NotBeZero(flags.hanaURL)

		handler = "hana"

		hanaURL = flags.hanaURL

	case "mongodb":
		panic("can't start in-process MongoDB")

	default:
		// that should be caught by Startup function
		panic("not reached")
	}

	p, err := state.NewProvider("")
	must.NoError(err)

	handlerOpts := &registry.NewHandlerOpts{
		Logger:        logger,
		ConnMetrics:   listenerMetrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: postgreSQLURL,
		SQLiteURL:     sqliteURL,
		HANAURL:       hanaURL,

		TestOpts: registry.TestOpts{
			DisableFilterPushdown: flags.disableFilterPushdown,
			EnableSortPushdown:    flags.enableSortPushdown,
			EnableOplog:           flags.enableOplog,

			UseNewPG:   flags.useNewPg,
			UseNewHana: flags.useNewHana,
		},
	}

	h, err := registry.NewHandler(handler, handlerOpts)
	must.NoError(err)

	listenerOpts := clientconn.NewListenerOpts{
		ProxyAddr:      flags.targetProxyAddr,
		Mode:           clientconn.NormalMode,
		Metrics:        listenerMetrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: filepath.Join("..", "tmp", "records"),
	}

	if flags.targetProxyAddr != "" {
		listenerOpts.Mode = clientconn.DiffNormalMode
	}

	must.BeTrue(!(flags.targetTLS && flags.targetUnixSocket))

	switch {
	case flags.targetTLS:
		listenerOpts.TLS = "127.0.0.1:0"
		listenerOpts.TLSCertFile = filepath.Join(CertsRoot, "server-cert.pem")
		listenerOpts.TLSKeyFile = filepath.Join(CertsRoot, "server-key.pem")
		listenerOpts.TLSCAFile = filepath.Join(CertsRoot, "rootCA-cert.pem")
	case flags.targetUnixSocket:
		if flags.shareServer {
			sqliteURL = sharedUnixSocketPath()
		} else {
			sqliteURL = privateUnixSocketPath(tb)
		}
	default:
		listenerOpts.TCP = "127.0.0.1:0"
	}

	l := clientconn.NewListener(&listenerOpts)

	return l, handler
}

// setupListener starts in-process FerretDB server that runs until ctx is canceled.
// It returns basic MongoDB URI for that listener.
//
// Shared and private listeners are constructed and run differently.
func setupListener(tb testtb.TB, ctx context.Context) string {
	must.BeTrue(flags.shareServer == (tb == nil))

	if tb != nil {
		tb.Helper()
	}

	_, span := otel.Tracer("").Start(ctx, "runPrivateListener")
	defer span.End()

	defer observability.FuncCall(ctx)()

	level := zap.NewAtomicLevelAt(zap.ErrorLevel)
	if flags.debugSetup {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	var logger *zap.Logger
	if flags.shareServer {
		logger = zap.L()
	} else {
		logger = testutil.LevelLogger(tb, level)
	}

	l, handler := makeListener(tb, ctx, logger)

	runDone := make(chan struct{})

	go func() {
		defer close(runDone)

		err := l.Run(ctx)
		if err == nil || errors.Is(err, context.Canceled) {
			logger.Info("Listener stopped without error")
		} else {
			logger.Error("Listener stopped", zap.Error(err))
		}
	}()

	// ensure that all listener's and handler's logs are written before test ends
	if flags.shareServer {
		tb.Cleanup(func() {
			select {
			case <-runDone:
			case <-time.After(10 * time.Second):
				panic("listener didn't stop in 10 seconds")
			}
		})
	}

	var hostPort, unixSocketPath string
	var tlsAndAuth bool

	switch {
	case flags.targetTLS:
		hostPort = l.TLSAddr().String()
		tlsAndAuth = true
	case flags.targetUnixSocket:
		unixSocketPath = l.UnixAddr().String()
	default:
		hostPort = l.TCPAddr().String()
	}

	uri := listenerMongoDBURI(hostPort, unixSocketPath, tlsAndAuth)

	msg := "Private listener started"
	if flags.shareServer {
		msg = "Shared listener started"
	}
	logger.Info(msg, zap.String("handler", handler), zap.String("uri", uri))

	return uri
}
