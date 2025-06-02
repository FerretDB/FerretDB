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
)

// Server implements services described by OpenAPI description file.
type Server struct {
	opts *ServerOpts
	s    *server.MCPServer
	h    *Handler
}

// Handler implements handling of MCP requests.
type Handler struct {
	l *slog.Logger
	h *handler.Handler
}

// ServerOpts represents [Serve] options.
type ServerOpts struct {
	L       *slog.Logger
	Handler *handler.Handler
	TCPAddr string
}

// New creates a MCP server.
func New(opts *ServerOpts) *Server {
	mcpServer := server.NewMCPServer(
		"Wire Protocol Server",
		"0.0.1",
	)

	return &Server{
		opts: opts,
		h: &Handler{
			h: opts.Handler,
			l: opts.L,
		},
		s: mcpServer,
	}
}

// mcpHandlers represents an MCP handler function.
type mcpHandlers struct {
	tool        mcp.Tool
	toolHandler server.ToolHandlerFunc
}

// initTools initializes the MCP tools for the server.
func (s *Server) initTools() map[string]mcpHandlers {
	return map[string]mcpHandlers{
		"find": {
			toolHandler: withLog(s.h.find, s.opts.L),
			tool:        newFindTool(),
		},
		"insert": {
			toolHandler: withLog(s.h.insert, s.opts.L),
			tool:        newInsertTool(),
		},
		"listCollections": {
			toolHandler: s.h.listCollections,
			tool:        newListCollections(),
		},
		"listDatabases": {
			toolHandler: s.h.listDatabases,
			tool:        newListDatabases(),
		},
	}
}

// Serve runs the MCP server.
func (s *Server) Serve(ctx context.Context) error {
	for _, t := range s.initTools() {
		if t.toolHandler != nil {
			s.s.AddTool(t.tool, t.toolHandler)
		}
	}

	s.opts.L.InfoContext(ctx, fmt.Sprintf("Starting MCP server on http://%s/", s.opts.TCPAddr))

	// FIXME add authentication
	sseServer := server.NewSSEServer(s.s, server.WithBaseURL(s.opts.TCPAddr), server.WithSSEContextFunc(withConnInfo))

	if err := sseServer.Start(s.opts.TCPAddr); err != nil {
		return err
	}

	return nil
}

// withConnInfo creates a new connection info and adds it to the context.
func withConnInfo(ctx context.Context, r *http.Request) context.Context {
	connInfo := conninfo.New()

	defer connInfo.Close()

	// FIXME this is not quite correct
	return conninfo.Ctx(r.Context(), connInfo)
}

// withLog wraps the next handler with logging functionality.
func withLog(next server.ToolHandlerFunc, l *slog.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		l.DebugContext(ctx, "MCP request", "request", fmt.Sprintf("%+v", request))

		res, err := next(ctx, request)
		if err != nil {
			l.ErrorContext(ctx, "MCP Error response", "response", fmt.Sprintf("%+v", res))
		}

		l.DebugContext(ctx, "MCP response", "response", fmt.Sprintf("%+v", res))
		return res, err
	}
}
