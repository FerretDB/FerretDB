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

// newListCollections creates a new MCP tool for listCollections command.
func newListCollections() mcp.Tool {
	return mcp.NewTool("list collections",
		mcp.WithDescription("Returns list of all collections in a database"),
		mcp.WithString("database",
			mcp.Required(),
			mcp.Description("The database to query"),
		),
	)
}

// listDatabases calls the listCollections command with the given parameters.
func (h *Handler) listCollections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	database, err := request.RequireString("database")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	req := wire.MustOpMsg(
		"listCollections", int32(1),
		"$db", database,
	)

	h.l.DebugContext(ctx, "OP_MSG request", slog.String("request", req.StringIndent()))

	res, err := h.h.Handle(ctx, &middleware.Request{OpMsg: req})
	if err != nil {
		return nil, err
	}

	h.l.DebugContext(ctx, "OP_MSG response", slog.String("response", res.OpMsg.StringIndent()))

	doc, err := res.OpMsg.DocumentDeep()
	if err != nil {
		return nil, err
	}

	jsonRes, err := doc.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(jsonRes)), nil
}
