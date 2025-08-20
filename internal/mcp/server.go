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

// initTools initializes available MCP tools for the given mcp server.
func (srv *server) initTools(s *mcp.Server) {
	listDatabasesTool := &mcp.Tool{
		Name:        "listDatabases",
		Description: "Returns a summary of all databases.",
	}
	mcp.AddTool(s, listDatabasesTool, srv.listDatabases)
}

// request sends a request document to the middleware and returns decoded response document.
func (srv *server) request(ctx context.Context, reqDoc *wirebson.Document) (*wirebson.Document, error) {
	req, err := middleware.RequestDoc(reqDoc)
	if err != nil {
		return nil, err
	}

	resp := srv.m.Handle(ctx, req)
	if resp == nil {
		return nil, errors.New("internal error")
	}

	return resp.DocumentRaw().DecodeDeep()
}
