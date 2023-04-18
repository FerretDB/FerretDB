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
	"os"
	"path/filepath"
	"runtime/trace"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers/registry"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// See docker-compose.yml.
var tigrisURLsIndex atomic.Uint32

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
		Logger:        logger,
		Metrics:       metrics.ConnMetrics,
		StateProvider: p,

		PostgreSQLURL: *postgreSQLURLF,

		TigrisURL: nextTigrisUrl(),

		TestOpts: registry.TestOpts{
			DisableFilterPushdown: *disableFilterPushdownF,
			EnableCursors:         *enableCursorsF,
		},
	}
	h, err := registry.NewHandler(handler, handlerOpts)
	require.NoError(tb, err)

	listenerOpts := clientconn.NewListenerOpts{
		ProxyAddr:      *targetProxyAddrF,
		Mode:           clientconn.NormalMode,
		Metrics:        metrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: filepath.Join("..", "tmp", "records"),
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
		listenerOpts.TLSCertFile = filepath.Join(CertsRoot, "server-cert.pem")
		listenerOpts.TLSKeyFile = filepath.Join(CertsRoot, "server-key.pem")
		listenerOpts.TLSCAFile = filepath.Join(CertsRoot, "rootCA-cert.pem")
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

	var clientOpts mongoDBURIOpts

	switch {
	case *targetTLSF:
		clientOpts.hostPort = l.TLSAddr().String()
		clientOpts.tlsAndAuth = true
	case *targetUnixSocketF:
		clientOpts.unixSocketPath = l.UnixAddr().String()
	default:
		clientOpts.hostPort = l.TCPAddr().String()
	}

	uri := mongoDBURI(tb, &clientOpts)
	client := setupClient(tb, ctx, uri)

	logger.Info("Listener started", zap.String("handler", handler), zap.String("uri", uri))

	return client, uri
}
