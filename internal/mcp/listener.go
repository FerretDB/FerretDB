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
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// Listener implements an MCP server.
type Listener struct {
	opts *ListenerOpts
	s    *server.MCPServer
	h    *toolHandler
}

// ListenerOpts represents [Serve] options.
type ListenerOpts struct {
	Handler *handler.Handler
	L       *slog.Logger
	TCPAddr string
}

// New creates an MCP server.
func New(opts *ListenerOpts) *Listener {
	return &Listener{
		opts: opts,
		s:    server.NewMCPServer("FerretDB", version.Get().Version),
		h:    newToolHandler(opts.Handler),
	}
}

// Run runs the MCP server.
func (s *Listener) Run(ctx context.Context) error {
	for _, t := range s.h.initTools() {
		s.s.AddTool(t.tool, withLog(withConnInfo(t.handleFunc), s.opts.L))
	}

	s.opts.L.InfoContext(ctx, fmt.Sprintf("Starting MCP server on http://%s/", s.opts.TCPAddr))

	// can authentication be added?
	// TODO https://github.com/FerretDB/FerretDB/issues/5209
	sseServer := server.NewSSEServer(s.s, server.WithBaseURL(s.opts.TCPAddr))

	if err := sseServer.Start(s.opts.TCPAddr); err != nil {
		return lazyerrors.Error(err)
	}

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
