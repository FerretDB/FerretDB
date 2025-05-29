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

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Server implements services described by OpenAPI description file.
type Server struct {
	opts *ServerOpts
	s    *server.MCPServer
}

// ServerOpts represents [Serve] options.
type ServerOpts struct {
	L       *slog.Logger
	Handler *handler.Handler
	TCPAddr string
}

// authKey is a custom context key for storing the auth token.
type authKey struct{}

// withAuthKey adds an auth key to the context.
func withAuthKey(ctx context.Context, auth string) context.Context {
	return context.WithValue(ctx, authKey{}, auth)
}

// authFromRequest extracts the auth token from the request headers.
func authFromRequest(ctx context.Context, r *http.Request) context.Context {
	return withAuthKey(ctx, r.Header.Get("Authorization"))
}

// New creates a MCP server.
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

// Serve runs the MCP server.
func (s *Server) Serve(ctx context.Context) error {
	dbStats := mcp.NewTool("find",
		mcp.WithDescription("Find the documents"),
		mcp.WithString("database",
			mcp.Required(),
			mcp.Description("The database to query"),
		),
		mcp.WithString("collection",
			mcp.Required(),
			mcp.Description("The collection to query"),
		),
	)

	s.s.AddTool(dbStats, s.handleFind)

	s.opts.L.InfoContext(ctx, fmt.Sprintf("Starting MCP server on http://%s/", s.opts.TCPAddr))

	httpServer := server.NewStreamableHTTPServer(s.s, server.WithHTTPContextFunc(authFromRequest))
	if err := httpServer.Start(s.opts.TCPAddr); err != nil {
		return err
	}

	return nil
}

// handleFind is the handler for the "find" MCP tool.
func (s *Server) handleFind(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	database, err := request.RequireString("database")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	collection, err := request.RequireString("collection")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	req := wire.MustOpMsg(
		"find", collection,
		"$db", database,
	)

	res, err := s.opts.Handler.Handle(ctx, &middleware.Request{OpMsg: req})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	resRaw := must.NotFail(res.OpMsg.DocumentRaw())
	results := must.NotFail(must.NotFail(resRaw.Decode()).Get("cursor").(wirebson.AnyDocument).Decode()).
		Get("firstBatch").(wirebson.AnyDocument)

	return mcp.NewToolResultText(results.LogMessage()), nil
}
