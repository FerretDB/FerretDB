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
	"path"
	"path/filepath"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// listenerMongoDBURI builds MongoDB URI for in-process FerretDB.
func listenerMongoDBURI(tb testtb.TB, hostPort, unixSocketPath string, tlsAndAuth bool) string {
	tb.Helper()

	var host string

	if hostPort != "" {
		require.Empty(tb, unixSocketPath, "both hostPort and unixSocketPath are set")
		host = hostPort
	} else {
		host = unixSocketPath
	}

	var user *url.Userinfo
	var q url.Values

	if tlsAndAuth {
		require.Empty(tb, unixSocketPath, "unixSocketPath cannot be used with TLS")

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

// setupListener starts in-process FerretDB server that runs until ctx is canceled.
// It returns basic MongoDB URI for that listener.
func setupListener(tb testtb.TB, ctx context.Context, logger *zap.Logger) string {
	tb.Helper()

	_, span := otel.Tracer("").Start(ctx, "setupListener")
	defer span.End()

	defer observability.FuncCall(ctx)()
	var l *clientconn.Listener
	var handler string

	if *shareServerF {
		l = listener
		handler = handlerType
	} else {
		handler, l = initListener(ctx)

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
		tb.Cleanup(func() {
			<-runDone
		})
	}

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

	uri := listenerMongoDBURI(tb, hostPort, unixSocketPath, tlsAndAuth)

	logger.Info("Listener started", zap.String("handler", handler), zap.String("uri", uri))

	return uri
}

func initListener(ctx context.Context) (string, *clientconn.Listener) {
	_, span := otel.Tracer("").Start(ctx, "initListener")
	defer span.End()

	defer observability.FuncCall(ctx)()

	if *targetURLF != "" {
		zap.S().Fatal("-target-url must be empty for in-process FerretDB")
	}

	var handler string
	switch *targetBackendF {
	case "ferretdb-pg":
		if *postgreSQLURLF == "" {
			zap.S().Fatalf("-postgresql-url must be set for %q", *targetBackendF)
		}

		if *sqliteURLF != "" {
			zap.S().Fatalf("-sqlite-url must be empty for %q", *targetBackendF)
		}

		if *hanaURLF != "" {
			zap.S().Fatalf("-hana-url must be empty for %q", *targetBackendF)
		}

		handler = "pg"

	case "ferretdb-sqlite":
		if *postgreSQLURLF != "" {
			zap.S().Fatalf("-postgresql-url must be empty for %q", *targetBackendF)
		}

		if *sqliteURLF == "" {
			zap.S().Fatalf("-sqlite-url must be set for %q", *targetBackendF)
		}

		if *hanaURLF != "" {
			zap.S().Fatalf("-hana-url must be empty for %q", *targetBackendF)
		}

		handler = "sqlite"

	case "ferretdb-hana":
		if *postgreSQLURLF != "" {
			zap.S().Fatalf("-postgresql-url must be empty for %q", *targetBackendF)
		}

		if *sqliteURLF != "" {
			zap.S().Fatalf("-sqlite-url must be empty for %q", *targetBackendF)
		}

		if *hanaURLF == "" {
			zap.S().Fatalf("-hana-url must be set for %q", *targetBackendF)
		}

		handler = "hana"

	case "mongodb":
		zap.S().Fatal("can't start in-process MongoDB")

	default:
		zap.S().Fatal("not reached")
	}

	// use per-test directory to prevent handler's/backend's metadata registry
	// read databases owned by concurrent tests
	sqliteURL := *sqliteURLF
	if sqliteURL != "" {
		u, err := url.Parse(sqliteURL)
		if err != nil {
			zap.S().Fatal(err)
		}

		u.Opaque = path.Join(u.Opaque, "test") + "/"
		sqliteURL = u.String()

		dir, err := filepath.Abs(u.Opaque)
		if err != nil {
			zap.S().Fatal(err)
		}

		if os.RemoveAll(dir) != nil {
			zap.S().Fatal(err)
		}

		if os.MkdirAll(dir, 0o777) != nil {
			zap.S().Fatal(err)
		}

		defer func() {
			if os.RemoveAll(dir) != nil {
				zap.S().Fatal(err)
			}
		}()
	}

	p, err := state.NewProvider("")
	if err != nil {
		zap.S().Fatal(err)
	}

	handlerOpts := &registry.NewHandlerOpts{
		Logger:        zap.L(),
		ConnMetrics:   listenerMetrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: *postgreSQLURLF,
		SQLiteURL:     sqliteURL,
		HANAURL:       *hanaURLF,

		TestOpts: registry.TestOpts{
			DisableFilterPushdown: *disableFilterPushdownF,
			EnableSortPushdown:    *enableSortPushdownF,
			EnableOplog:           *enableOplogF,

			UseNewPG:   *useNewPgF,
			UseNewHana: *useNewHanaF,
		},
	}

	h, err := registry.NewHandler(handler, handlerOpts)
	if err != nil {
		zap.S().Fatal(err)
	}

	listenerOpts := clientconn.NewListenerOpts{
		ProxyAddr:      *targetProxyAddrF,
		Mode:           clientconn.NormalMode,
		Metrics:        listenerMetrics,
		Handler:        h,
		Logger:         zap.L(),
		TestRecordsDir: filepath.Join("..", "tmp", "records"),
	}

	if *targetProxyAddrF != "" {
		listenerOpts.Mode = clientconn.DiffNormalMode
	}

	if *targetTLSF && *targetUnixSocketF {
		zap.S().Fatal("Both -target-tls and -target-unix-socket are set.")
	}

	switch {
	case *targetTLSF:
		listenerOpts.TLS = "127.0.0.1:0"
		listenerOpts.TLSCertFile = filepath.Join(CertsRoot, "server-cert.pem")
		listenerOpts.TLSKeyFile = filepath.Join(CertsRoot, "server-key.pem")
		listenerOpts.TLSCAFile = filepath.Join(CertsRoot, "rootCA-cert.pem")
	case *targetUnixSocketF:
		// do not use tb.TempDir() because generated path is too long on macOS
		f, err := os.CreateTemp("", "ferretdb-*.sock")
		if err != nil {
			zap.S().Fatal(err)
		}

		// remove file so listener could create it (and remove it itself on stop)
		err = f.Close()
		if err != nil {
			zap.S().Fatal(err)
		}

		err = os.Remove(f.Name())
		if err != nil {
			zap.S().Fatal(err)
		}
		listenerOpts.Unix = f.Name()
	default:
		listenerOpts.TCP = "127.0.0.1:0"
	}

	l := clientconn.NewListener(&listenerOpts)

	return handler, l
}
