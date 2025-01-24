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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// ListenerOpts represents setup options for in-process FerretDB listener.
type ListenerOpts struct {
	// SessionCleanupInterval is a duration between expired session deletion runs.
	SessionCleanupInterval time.Duration
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

// listenerMongoDBURI builds MongoDB URI for in-process FerretDB.
func listenerMongoDBURI(tb testing.TB, hostPort, unixSocketPath string) string {
	tb.Helper()

	var host string

	if hostPort != "" {
		require.Empty(tb, unixSocketPath, "both hostPort and unixSocketPath are set")
		host = hostPort
	} else {
		host = unixSocketPath
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1507
	u := &url.URL{
		Scheme: "mongodb",
		Host:   host,
		Path:   "/",
		User:   url.UserPassword("username", "password"),
	}

	return u.String()
}

// setupListener starts in-process FerretDB server that runs until ctx is canceled.
// It returns basic MongoDB URI for that listener.
func setupListener(tb testing.TB, ctx context.Context, opts *ListenerOpts, logger *slog.Logger) string {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupListener")
	defer span.End()

	require.Empty(tb, *targetURLF, "-target-url must be empty for in-process FerretDB")

	switch *targetBackendF {
	case "ferretdb":
		require.NotEmpty(tb, *postgreSQLURLF, "-postgresql-url must be set for %q", *targetBackendF)

	case "mongodb":
		tb.Fatal("can't start in-process MongoDB")

	default:
		// that should be caught by Startup function
		panic("not reached")
	}

	sp, err := state.NewProvider("")
	require.NoError(tb, err)

	if opts == nil {
		opts = new(ListenerOpts)
	}

	p, err := documentdb.NewPool(*postgreSQLURLF, logging.WithName(logger, "pool"), sp)
	require.NoError(tb, err)

	tb.Cleanup(p.Close)

	handlerOpts := &handler.NewOpts{
		Pool: p,
		Auth: true,

		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/566
		TCPHost:     "",
		ReplSetName: "",

		L:             logging.WithName(logger, "handler"),
		ConnMetrics:   listenerMetrics.ConnMetrics,
		StateProvider: sp,

		SessionCleanupInterval: opts.SessionCleanupInterval,
	}

	h, err := handler.New(handlerOpts)
	require.NoError(tb, err)

	listenerOpts := clientconn.NewListenerOpts{
		ProxyAddr:      *targetProxyAddrF,
		Mode:           clientconn.NormalMode,
		Metrics:        listenerMetrics,
		Handler:        h,
		Logger:         logger,
		TestRecordsDir: filepath.Join(Dir(tb), "..", "..", "tmp", "records"),
	}

	if *targetProxyAddrF != "" {
		listenerOpts.Mode = clientconn.DiffNormalMode
	}

	switch {
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

	switch {
	case *targetUnixSocketF:
		unixSocketPath = l.UnixAddr().String()
	default:
		hostPort = l.TCPAddr().String()
	}

	uri := listenerMongoDBURI(tb, hostPort, unixSocketPath)

	logger.InfoContext(ctx, "Listener started", slog.String("uri", uri))

	return uri
}
