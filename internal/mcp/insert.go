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
	"encoding/json"
	"log/slog"

	"github.com/FerretDB/wire"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// newInsertTool creates a new MCP tool for insert command.
func newInsertTool() mcp.Tool {
	rawSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"insert": {"type": "string", "description": "Name of collection to insert documents into"},
			"documents": {"type": "array", "description": "array of documents where each document is a JSON object"},
			"$db": {"type": "string", "description": "Name of database"}
		},
		"required": ["insert", "documents", "$db"]
	}`)

	return mcp.NewToolWithRawSchema("insert", "Insert documents and it returns the number of inserted documents", rawSchema)
}

// insert executes insert command.
func (h *Handler) insert(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	database, err := request.RequireString("database")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	collection, err := request.RequireString("collection")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	documents := request.GetStringSlice("documents", []string{})

	req := wire.MustOpMsg(
		"insert", collection,
		"$db", database,
		"documents", documents,
	)

	h.l.DebugContext(ctx, "OP_MSG request", slog.String("request", req.StringIndent()))

	res, err := h.h.Handle(ctx, &middleware.Request{OpMsg: req})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	h.l.DebugContext(ctx, "OP_MSG response", slog.String("response", res.OpMsg.StringIndent()))

	doc, err := res.OpMsg.DocumentDeep()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	n := doc.Get("n").(float64)

	return mcp.FormatNumberResult(n), nil
}
