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

// Server implements an MCP server.
type Server struct {
	opts       *ServerOpts
	httpServer *http.Server
	lis        net.Listener
}

// ServerOpts represents options configurable for [Server].
type ServerOpts struct {
	L           *slog.Logger
	Handler     *handler.Handler
	ToolHandler *ToolHandler
	TCPAddr     string
}

// New creates an MCP server.
func New(ctx context.Context, opts *ServerOpts) (*Server, error) {
	s := server.NewMCPServer("FerretDB", version.Get().Version)

	for _, t := range opts.ToolHandler.initTools() {
		s.AddTool(t.tool, withLog(withConnInfo(t.handleFunc), opts.L))
	}

	sseServer := server.NewSSEServer(s,
		server.WithBaseURL(opts.TCPAddr),
	)

	srv := NewAuthHandler(opts.Handler)
	mux := http.NewServeMux()

	// Is WithDynamicBasePath necessary?
	sseHandler := sseServer.SSEHandler()
	messageHandler := sseServer.MessageHandler()

	if opts.Handler.Auth {
		sseHandler = srv.AuthMiddleware(sseHandler)
		messageHandler = srv.AuthMiddleware(messageHandler)
	}

	mux.Handle("/mcp/sse", sseHandler)
	mux.Handle("/mcp/message", messageHandler)

	httpSrv := &http.Server{
		Handler:  mux,
		ErrorLog: slog.NewLogLogger(opts.L.Handler(), slog.LevelError),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return conninfo.Ctx(ctx, conninfo.New())
		},
	}

	lis, err := net.Listen("tcp", opts.TCPAddr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Server{
		opts:       opts,
		httpServer: httpSrv,
		lis:        lis,
	}, nil
}

// Serve runs the MCP server until the context is done.
func (s *Server) Serve(ctx context.Context) error {
	s.opts.L.InfoContext(ctx, fmt.Sprintf("Starting MCP server on http://%s/", s.opts.TCPAddr))

	go func() {
		if err := s.httpServer.Serve(s.lis); !errors.Is(err, http.ErrServerClosed) {
			s.opts.L.LogAttrs(ctx, logging.LevelDPanic, "Serve exited with unexpected error", logging.Error(err))
		}
	}()

	<-ctx.Done()

	s.opts.L.InfoContext(ctx, "MCP server stopped")

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
