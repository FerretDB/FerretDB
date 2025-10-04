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

package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// Listener represents MCP listener.
type Listener struct {
	opts *ListenOpts
	lis  net.Listener
	srv  *server
}

// ListenOpts represents [Listen] options.
type ListenOpts struct { //nolint:vet // for readability
	L       *slog.Logger
	M       *middleware.Middleware
	TCPAddr string
	Auth    bool
}

// Listen creates a new MCP handler and starts listener on the given TCP address.
// [Listener.Run] must be called on the returned value.
func Listen(opts *ListenOpts) (*Listener, error) {
	lis, err := net.Listen("tcp", opts.TCPAddr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Listener{
		opts: opts,
		lis:  lis,
		srv:  newServer(opts.M),
	}, nil
}

// Run runs MCP handler until ctx is canceled.
//
// It exits when handler is stopped and listener closed.
func (lis *Listener) Run(ctx context.Context) {
	s := mcp.NewServer(&mcp.Implementation{Name: "FerretDB", Version: version.Get().Version}, nil)
	lis.srv.addTools(s)

	var mcpHandler http.Handler = mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server { return s }, nil)
	srvHandler := http.NewServeMux()

	if lis.opts.Auth {
		mcpHandler = lis.srv.authMiddleware(mcpHandler)
	}

	srvHandler.Handle("/mcp", connInfoMiddleware(mcpHandler))

	srv := &http.Server{
		Handler:  srvHandler,
		ErrorLog: slog.NewLogLogger(lis.opts.L.Handler(), slog.LevelError),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	lis.opts.L.InfoContext(ctx, fmt.Sprintf("Starting MCP server on http://%s/mcp", lis.lis.Addr()))

	go func() {
		if err := srv.Serve(lis.lis); !errors.Is(err, http.ErrServerClosed) {
			lis.opts.L.LogAttrs(ctx, logging.LevelDPanic, "Serve exited with unexpected error", logging.Error(err))
		}
	}()

	<-ctx.Done()

	// ctx is already canceled, but we want to inherit its values
	shutdownCtx, shutdownCancel := ctxutil.WithDelay(ctx)
	defer shutdownCancel(nil)

	if err := srv.Shutdown(shutdownCtx); err != nil {
		lis.opts.L.LogAttrs(ctx, logging.LevelDPanic, "Shutdown exited with unexpected error", logging.Error(err))
	}

	if err := srv.Close(); err != nil {
		lis.opts.L.LogAttrs(ctx, logging.LevelDPanic, "Close exited with unexpected error", logging.Error(err))
	}

	lis.opts.L.InfoContext(ctx, "MCP server stopped")
}

// connInfoMiddleware returns a handler function that creates a new [*conninfo.ConnInfo],
// calls the next handler, and closes the connection info after the request is done.
func connInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connInfo := conninfo.New()
		defer connInfo.Close()
		next.ServeHTTP(w, r.WithContext(conninfo.Ctx(r.Context(), connInfo)))
	})
}

// Addr returns TCP listener's address.
// It can be used to determine an actually used port, if it was zero.
func (lis *Listener) Addr() net.Addr {
	return lis.lis.Addr()
}

// Describe implements [prometheus.Collector].
func (lis *Listener) Describe(ch chan<- *prometheus.Desc) {
}

// Collect implements [prometheus.Collector].
func (lis *Listener) Collect(ch chan<- prometheus.Metric) {
}
