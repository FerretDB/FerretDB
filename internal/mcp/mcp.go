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

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
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

	// FIXME add authentication
	sseServer := server.NewSSEServer(s.s, server.WithBaseURL(s.opts.TCPAddr), server.WithSSEContextFunc(withConnInfo))

	if err := sseServer.Start(s.opts.TCPAddr); err != nil {
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
		Get("firstBatch").(wirebson.AnyArray)

	return mcp.NewToolResultText(results.LogMessage()), nil
}

// withConnInfo creates a new connection info and adds it to the context.
func withConnInfo(ctx context.Context, r *http.Request) context.Context {
	connInfo := conninfo.New()

	defer connInfo.Close()

	// FIXME this is not quite correct
	return conninfo.Ctx(r.Context(), connInfo)
}
