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
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
)

// newInsertTool creates a new MCP tool for insert command.
func newInsertTool() mcp.Tool {
	return mcp.NewTool("insert",
		mcp.WithDescription("Insert documents and return the number inserted documents"),
		mcp.WithString("database",
			mcp.Required(),
			mcp.Description("The database to use for inserting documents"),
		),
		mcp.WithString("collection",
			mcp.Required(),
			mcp.Description("The collection to insert documents into"),
		),
		mcp.WithArray("documents",
			mcp.Required(),
			mcp.Description("The documents contains documents to insert, represented in Extended JSON v2 format"),
			mcp.Items(`{"type":"object"}`),
		),
	)
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

	var rawDocuments json.RawMessage

	// ideally, documents arguments should be json.RawMessage bytes, but so far []any is observed
	switch documents := request.GetArguments()["documents"].(type) {
	case []any:
		// marshal to json.RawMessage so we can use similar code as DataAPI to do json -> bson conversion
		if rawDocuments, err = json.Marshal(documents); err != nil {
			return mcp.NewToolResultErrorFromErr("cannot marshal insert request slice", err), nil
		}
	case []byte:
		rawDocuments = documents
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid argument type %T for insert command", documents)), nil
	}

	req, err := prepareOpMsg(
		"insert", collection,
		"$db", database,
		"documents", rawDocuments,
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

	doc, err := res.OpMsg.DocumentDeep()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonRes, err := doc.MarshalJSON()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(string(jsonRes)), nil
}
