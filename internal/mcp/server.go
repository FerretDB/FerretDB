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
	"log/slog"

	"github.com/FerretDB/wire/wirebson"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// server handles MCP request.
type server struct {
	l *slog.Logger
	m *middleware.Middleware
}

// newServer creates a new server with the given parameter.
func newServer(l *slog.Logger, m *middleware.Middleware) *server {
	return &server{
		l: l,
		m: m,
	}
}

// addTools adds available MCP tools for the given mcp server.
func (s *server) addTools(srv *mcp.Server) {
	// sorted alphabetically
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "dropDatabase",
		Description: "Deletes the database.",
	}, s.dropDatabase)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "find",
		Description: "Search documents from a collection.",
	}, s.find)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "insert",
		Description: "Inserts multiple documents into a collection.",
	}, s.insert)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "listCollections",
		Description: "Returns a summary of all collections in a database.",
	}, s.listCollections)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "listDatabases",
		Description: "Returns a summary of all databases.",
	}, s.listDatabases)
}

// handle sends the request document to the middleware and returns result used by MCP tool.
func (s *server) handle(ctx context.Context, reqDoc *wirebson.Document) (*mcp.CallToolResult, error) {
	req, err := middleware.RequestDoc(reqDoc)
	if err != nil {
		return nil, err
	}

	resp := s.m.Handle(ctx, req)
	if resp == nil {
		return nil, errors.New("internal error")
	}

	doc, err := resp.DocumentRaw().DecodeDeep()
	if doc == nil {
		return nil, err
	}

	b, err := doc.MarshalJSON()
	if err != nil {
		return nil, err
	}

	res := &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
		IsError: !resp.OK(),
	}

	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, "MCP tool result", slog.Any("result", res))
	}

	return res, nil
}
