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

	"github.com/FerretDB/wire"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// newTool creates a new MCP resource for listDatabases command.
func newListDatabasesResource() mcp.Resource {
	return mcp.NewResource(
		"databases",
		"A list of all databases",
		mcp.WithResourceDescription("A list of all databases"),
		mcp.WithMIMEType("application/json"),
	)
}

// handleListDatabases calls the listDatabases command and returns the result as an MCP resource.
func (s *Server) handleListDatabases(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	req := wire.MustOpMsg(
		"listDatabase", int32(1),
	)

	res, err := s.opts.Handler.Handle(ctx, &middleware.Request{OpMsg: req})
	if err != nil {
		return nil, err
	}

	doc, err := res.OpMsg.DocumentDeep()
	if err != nil {
		return nil, err
	}

	jsonRes, err := doc.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "databases",
			MIMEType: "application/json",
			Text:     string(jsonRes),
		},
	}, nil
}
