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
	"net/url"

	"github.com/FerretDB/wire"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// newListCollectionsResource creates a new MCP resource for listCollections command.
func newListCollectionsResource() mcp.ResourceTemplate {
	return mcp.NewResourceTemplate(
		"databases://{database}",
		"A list of all collections in a database",
		mcp.WithTemplateDescription("A list of all collections in the database"),
		mcp.WithTemplateMIMEType("application/json"),
	)
}

// handleListDatabases calls the listCollections command and returns the result as an MCP resource.
func (s *Server) handleListCollections(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	u, err := url.Parse(request.Params.URI)
	if err != nil {
		return nil, err
	}

	database := u.Path

	req := wire.MustOpMsg(
		"listCollections", int32(1),
		"$db", database,
	)

	s.opts.L.DebugContext(ctx, "OP_MSG request", "request", req.StringIndent())

	res, err := s.opts.Handler.Handle(ctx, &middleware.Request{OpMsg: req})
	if err != nil {
		return nil, err
	}

	s.opts.L.DebugContext(ctx, "OP_MSG response", "response", res.OpMsg.StringIndent())

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
			URI:      "databases://{database}",
			MIMEType: "application/json",
			Text:     string(jsonRes),
		},
	}, nil
}
