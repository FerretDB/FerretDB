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

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

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
func Listen(ctx context.Context, opts *ListenerOpts) (*Listener, error) {

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
	mcpSrv := server.NewMCPServer("FerretDB", version.Get().Version)

	for _, t := range lis.opts.ToolHandler.initTools() {
		mcpSrv.AddTool(t.tool, withConnInfo(withLog(t.handleFunc, lis.opts.L)))
	}

	mux := http.NewServeMux()
	mux.Handle("/mcp", server.NewStreamableHTTPServer(mcpSrv))

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

// withConnInfo wraps the next handler with [*conninfo.ConnInfo] context and closes it once the handler is executed.
func withConnInfo(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		connInfo := conninfo.New()

		defer connInfo.Close()

		return next(conninfo.Ctx(ctx, connInfo), request)
	}
}

// withLog wraps the next handler with logging of request, response and error.
func withLog(next server.ToolHandlerFunc, l *slog.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if l.Enabled(ctx, slog.LevelDebug) {
			l.DebugContext(ctx, "MCP request", slog.String("request", fmt.Sprintf("%+v", request)))
		}

		res, err := next(ctx, request)
		if err != nil {
			l.ErrorContext(ctx, "MCP error", logging.Error(err))

			return nil, err
		}

		if l.Enabled(ctx, slog.LevelDebug) {
			l.DebugContext(ctx, "MCP response", slog.String("response", fmt.Sprintf("%+v", res)))
		}

		return res, nil
	}
}
