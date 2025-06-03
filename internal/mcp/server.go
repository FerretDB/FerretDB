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
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// Server implements an MCP server.
type Server struct {
	opts *ServerOpts
	s    *server.MCPServer
}

// ServerOpts represents [Serve] options.
type ServerOpts struct {
	L       *slog.Logger
	Handler *Handler
	TCPAddr string
}

// Handler handles MCP request.
type Handler struct {
	h *handler.Handler
	l *slog.Logger
}

// NewHandler creates a new MCP handler with the given logger and handler.
func NewHandler(h *handler.Handler, l *slog.Logger) *Handler {
	return &Handler{
		h: h,
		l: l,
	}
}

// New creates an MCP server.
func New(opts *ServerOpts) *Server {
	mcpServer := server.NewMCPServer(
		"Wire Protocol Server",
		"0.0.1",
	)

	return &Server{
		opts: opts,
		s:    mcpServer,
	}
}

// mcpHandlers represents a tool and its handler to retrieve and return data.
type mcpHandler struct {
	tool        mcp.Tool
	toolHandler server.ToolHandlerFunc
}

// initHandlers returns available MCP handlers for the server.
func (s *Server) initHandlers() map[string]mcpHandler {
	return map[string]mcpHandler{
		"find": {
			toolHandler: s.opts.Handler.find,
			tool:        newFindTool(),
		},
		"insert": {
			toolHandler: s.opts.Handler.insert,
			tool:        newInsertTool(),
		},
		"listCollections": {
			toolHandler: s.opts.Handler.listCollections,
			tool:        newListCollections(),
		},
		"listDatabases": {
			toolHandler: s.opts.Handler.listDatabases,
			tool:        newListDatabases(),
		},
	}
}

// Serve runs the MCP server.
func (s *Server) Serve(ctx context.Context) error {
	for _, t := range s.initHandlers() {
		if t.toolHandler != nil {
			s.s.AddTool(t.tool, withLog(t.toolHandler, s.opts.L))
		}
	}

	s.opts.L.InfoContext(ctx, fmt.Sprintf("Starting MCP server on http://%s/", s.opts.TCPAddr))

	// can authentication be added?
	// TODO https://github.com/FerretDB/FerretDB/issues/5209
	sseServer := server.NewSSEServer(s.s, server.WithBaseURL(s.opts.TCPAddr), server.WithSSEContextFunc(withConnInfo))

	if err := sseServer.Start(s.opts.TCPAddr); err != nil {
		return err
	}

	return nil
}

// withConnInfo creates a new connection info and adds it to the context.
func withConnInfo(ctx context.Context, r *http.Request) context.Context {
	connInfo := conninfo.New()

	// improve handling of conninfo
	// TODO https://github.com/FerretDB/FerretDB/issues/5209
	defer connInfo.Close()

	return conninfo.Ctx(r.Context(), connInfo)
}

// withLog wraps the next handler with logging of request, response and error.
func withLog(next server.ToolHandlerFunc, l *slog.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		l.DebugContext(ctx, "MCP request", slog.String("request", fmt.Sprintf("%+v", request)))

		res, err := next(ctx, request)
		if err != nil {
			l.ErrorContext(ctx, "MCP error", logging.Error(err))

			return nil, err
		}

		l.DebugContext(ctx, "MCP response", slog.String("response", fmt.Sprintf("%+v", res)))

		return res, nil
	}
}
