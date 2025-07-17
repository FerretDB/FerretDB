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

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// Handler represents MCP handler.
type Handler struct {
	opts    *ListenOpts
	lis     net.Listener
	handler http.Handler
}

// ListenOpts represents [Listen] options.
type ListenOpts struct { //nolint:vet // for readability
	Handler     *handler.Handler
	ToolHandler *ToolHandler
	TCPAddr     string

	L *slog.Logger
}

// Listen creates a new MCP handler and starts listener on the given TCP address.
func Listen(opts *ListenOpts) (*Handler, error) {
	s := mcp.NewServer(&mcp.Implementation{Name: "FerretDB", Version: version.Get().Version}, nil)
	opts.ToolHandler.initTools(s)

	mcpHandler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server { return s }, nil)

	mux := http.NewServeMux()
	mux.Handle("/mcp", connInfoMiddleware(mcpHandler))

	lis, err := net.Listen("tcp", opts.TCPAddr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Handler{
		opts:    opts,
		lis:     lis,
		handler: mux,
	}, nil
}

// Serve runs MCP handler until ctx is canceled.
//
// It exits when the handler is stopped and the listener is closed.
func (h *Handler) Serve(ctx context.Context) {
	l := h.opts.L

	s := &http.Server{
		Handler:  h.handler,
		ErrorLog: slog.NewLogLogger(l.Handler(), slog.LevelError),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	l.InfoContext(ctx, fmt.Sprintf("Starting MCP server on http://%s/", h.opts.TCPAddr))

	go func() {
		if err := s.Serve(h.lis); !errors.Is(err, http.ErrServerClosed) {
			l.LogAttrs(ctx, logging.LevelDPanic, "Serve exited with unexpected error", logging.Error(err))
		}
	}()

	<-ctx.Done()

	// ctx is already canceled, but we want to inherit its values
	shutdownCtx, shutdownCancel := ctxutil.WithDelay(ctx)
	defer shutdownCancel(nil)

	if err := s.Shutdown(shutdownCtx); err != nil {
		l.LogAttrs(ctx, logging.LevelDPanic, "Shutdown exited with unexpected error", logging.Error(err))
	}

	if err := s.Close(); err != nil {
		l.LogAttrs(ctx, logging.LevelDPanic, "Close exited with unexpected error", logging.Error(err))
	}

	l.InfoContext(ctx, "MCP server stopped")
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
