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
	"github.com/FerretDB/wire/wirebson"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// newFindTool creates a new MCP tool for find command.
func newFindTool() mcp.Tool {
	return mcp.NewTool("find",
		mcp.WithDescription("Find queries to get documents"),
		mcp.WithString("database",
			mcp.Required(),
			mcp.Description("The database to query"),
		),
		mcp.WithString("collection",
			mcp.Required(),
			mcp.Description("The collection to query"),
		),
	)
}

// handleFind calls find command with the given parameters.
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

	s.opts.L.DebugContext(ctx, "OP_MSG request", "request", req.StringIndent())

	res, err := s.opts.Handler.Handle(ctx, &middleware.Request{OpMsg: req})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.opts.L.DebugContext(ctx, "OP_MSG response", "response", res.OpMsg.StringIndent())

	doc, err := res.OpMsg.DocumentDeep()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var jsonRes []byte

	if doc.Get("ok").(float64) != 1 {
		if jsonRes, err = doc.MarshalJSON(); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultError(string(jsonRes)), nil
	}

	results := doc.Get("cursor").(*wirebson.Document).Get("firstBatch").(*wirebson.Array)

	jsonRes, err = results.MarshalJSON()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.opts.L.DebugContext(ctx, "Find response", "json", string(jsonRes))

	return mcp.NewToolResultText(string(jsonRes)), nil
}
