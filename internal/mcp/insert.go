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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
			"documents": {"type": "array", "description": "The documents to insert, represented in Extended JSON v2 format"},
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
	var raw json.RawMessage
	var err error

	args := request.GetRawArguments()

	// ideally request arguments should be json.RawMessage according to tools initialization, but so far map[string]any is observed
	switch args := args.(type) {
	case map[string]any:
		if raw, err = json.Marshal(args); err != nil {
			return mcp.NewToolResultErrorFromErr("cannot marshal insert request map", err), nil
		}
	case []byte:
		raw = args
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid argument type %T for insert command", args)), nil
	}

	var body api.InsertManyJSONBody

	if err = json.Unmarshal(raw, &body); err != nil {
		return mcp.NewToolResultErrorFromErr("cannot unmarshal insert body", err), nil
	}

	req, err := prepareOpMsg(
		"insert", body.Collection,
		"$db", body.Database,
		"documents", body.Documents,
	)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("cannot create OP_MSG", err), nil
	}

	h.l.DebugContext(ctx, "OP_MSG request", slog.String("request", req.OpMsg.StringIndent()))

	res, err := h.h.Handle(ctx, req)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to handle OP_MSG", err), nil
	}

	h.l.DebugContext(ctx, "OP_MSG response", slog.String("response", res.OpMsg.StringIndent()))

	rawRes, err := res.OpMsg.DocumentRaw()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get raw document", err), nil
	}

	buf := new(bytes.Buffer)

	if err = marshalJSON(rawRes, buf); err != nil {
		return mcp.NewToolResultErrorFromErr("cannot marshal to extend JSON", err), nil
	}

	return mcp.NewToolResultText(buf.String()), nil
}
