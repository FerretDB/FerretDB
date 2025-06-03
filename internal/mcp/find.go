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
	"log/slog"

	"github.com/FerretDB/wire"
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

// find calls find command with the given parameters.
func (h *Handler) find(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	database, err := request.RequireString("database")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get database name", err), nil
	}

	collection, err := request.RequireString("collection")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get collection name", err), nil
	}

	req := wire.MustOpMsg(
		"find", collection,
		"$db", database,
	)

	h.l.DebugContext(ctx, "OP_MSG request", slog.String("request", req.StringIndent()))

	res, err := h.h.Handle(ctx, &middleware.Request{OpMsg: req})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to handle OP_MSG", err), nil
	}

	h.l.DebugContext(ctx, "OP_MSG response", slog.String("response", res.OpMsg.StringIndent()))

	doc, err := res.OpMsg.DocumentDeep()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to decode OP_MSG", err), nil
	}

	jsonRes, err := doc.MarshalJSON()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to marshal", err), nil
	}

	return mcp.NewToolResultText(string(jsonRes)), nil
}
