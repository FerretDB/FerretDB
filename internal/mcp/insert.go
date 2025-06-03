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

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
)

// newInsertTool creates a new MCP tool for insert command.
func newInsertTool() mcp.Tool {
	rawSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"collection": {"type": "string", "description": "The collection to insert documents into"},
			"documents": {"type": "array", "description": "The documents contains documents to insert, represented in Extended JSON v2 format"},
			"database": {"type": "string", "description": "The database to use for inserting documents"}
		},
		"required": ["collection", "documents", "database"]
	}`)

	return mcp.NewToolWithRawSchema("insert", "Insert documents and return the number inserted documents", rawSchema)

	//	return mcp.NewTool("insert",
	//		mcp.WithDescription("Insert documents and return the number inserted documents"),
	//		mcp.WithString("database",
	//			mcp.Required(),
	//			mcp.Description("The database to use for inserting documents"),
	//		),
	//		mcp.WithString("collection",
	//			mcp.Required(),
	//			mcp.Description("The collection to insert documents into"),
	//		),
	//		mcp.WithArray("documents",
	//			mcp.Required(),
	//			mcp.Description("The documents contains documents to insert, represented in Extended JSON v2 format"),
	//			mcp.Items(`{
	//		"type": "object",
	//}`),
	//		),
	//	)
}

// insert executes insert command.
func (h *Handler) insert(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw := request.GetRawArguments().([]byte)

	var body api.InsertManyJSONBody
	err := json.Unmarshal(raw, &body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	req, err := prepareOpMsg(
		"insert", body.Collection,
		"$db", body.Database,
		"documents", body.Documents,
	)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	h.l.DebugContext(ctx, "OP_MSG request", slog.String("request", req.OpMsg.StringIndent()))

	res, err := h.h.Handle(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	h.l.DebugContext(ctx, "OP_MSG response", slog.String("response", res.OpMsg.StringIndent()))

	doc, err := res.OpMsg.DocumentDeep()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	n := doc.Get("n").(int32)

	return mcp.FormatNumberResult(float64(n)), nil
}
