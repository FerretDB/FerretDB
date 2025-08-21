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

	"github.com/FerretDB/wire/wirebson"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// server handles MCP request.
type server struct {
	m *middleware.Middleware
}

// newServer creates a new server with the given parameter.
func newServer(m *middleware.Middleware) *server {
	return &server{
		m: m,
	}
}

// addTools adds available MCP tools for the given mcp server.
func (s *server) addTools(srv *mcp.Server) {
	listDatabasesTool := &mcp.Tool{
		Name:        "listDatabases",
		Description: "Returns a summary of all databases.",
	}
	mcp.AddTool(srv, listDatabasesTool, s.listDatabases)
}

// handle sends the request document to the middleware and returns result used by MCP tool.
//
// Log MCP tool result for debug level.
// TODO https://github.com/FerretDB/FerretDB/issues/5277
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

	json, err := doc.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(json)}},
		IsError: !resp.OK(),
	}, nil
}
