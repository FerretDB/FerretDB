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
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// Listener is an MCP listener.
type Listener struct {
	opts *ListenerOpts
	lis  net.Listener
}

// ListenerOpts represents options configurable for [Listener].
type ListenerOpts struct {
	L           *slog.Logger
	Handler     *handler.Handler
	ToolHandler *ToolHandler
	TCPAddr     string
}

// Listen creates an MCP server and starts listener on the given TCP address.
func Listen(opts *ListenerOpts) (*Listener, error) {
	lis, err := net.Listen("tcp", opts.TCPAddr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Listener{
		opts: opts,
		lis:  lis,
	}, nil
}

// Run runs the MCP server until the context is done.
func (lis *Listener) Run(ctx context.Context) error {
	mcpSrv := mcp.NewServer(&mcp.Implementation{Name: "FerretDB", Version: version.Get().Version}, nil)
	lis.opts.ToolHandler.initTools(mcpSrv)

	mux := http.NewServeMux()

	h := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server { return mcpSrv }, nil)
	mux.Handle("/mcp", connInfoMiddleware(h))

	s := &http.Server{
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(lis.opts.L.Handler(), slog.LevelError),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	lis.opts.L.InfoContext(ctx, fmt.Sprintf("Starting MCP server on http://%s/", lis.opts.TCPAddr))

	go func() {
		if err := s.Serve(lis.lis); !errors.Is(err, http.ErrServerClosed) {
			lis.opts.L.LogAttrs(ctx, logging.LevelDPanic, "Serve exited with unexpected error", logging.Error(err))
		}
	}()

	<-ctx.Done()

	lis.opts.L.InfoContext(ctx, "MCP server stopped")

	return nil
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
