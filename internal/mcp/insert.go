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

	"github.com/FerretDB/wire/wirebson"
	"github.com/mark3labs/mcp-go/mcp"
)

// newInsertTool creates a new MCP tool for insert command.
func newInsertTool() mcp.Tool {
	return mcp.NewTool("insert",
		mcp.WithDescription("Insert documents to the collection and return the response in Extended JSON v2 format. "+
			"If the response contains ok field with value 1, the operation was successful. "+
			"If the response contains ok field with value 0, the operation failed and the error message is in the errmsg field."+
			"The documents to insert are represented in Extended JSON v2 format."),
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
			mcp.Description("The documents contains documents to insert, represented in Extended JSON v2 format."+
				"Extended JSON v2 format also used in openapi-schema.",
			),
			mcp.Items(`{"type":"object"}`),
		),
	)
}

// insert adds documents to the given collection in the database and returns the result of the insert command
// in a string containing Extended JSON v2 format.
// Each document to insert is map[string]any.
func (h *ToolHandler) insert(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	database, err := request.RequireString("database")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get database name", err), nil
	}

	collection, err := request.RequireString("collection")
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get collection name", err), nil
	}

	documents, ok := request.GetArguments()["documents"].([]any)
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("invalid insert documents type %T", documents)), nil
	}

	jsonReq, err := json.Marshal(documents)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("cannot marshal insert request slice", err), nil
	}

	bsonDocs := wirebson.MustArray()

	if err = bsonDocs.UnmarshalJSON(jsonReq); err != nil {
		return mcp.NewToolResultErrorFromErr("cannot create document", err), nil
	}

	reqDoc, err := wirebson.NewDocument(
		"insert", collection,
		"$db", database,
		"documents", bsonDocs,
	)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("create document failed", err), nil
	}

	resDoc, err := h.request(ctx, reqDoc)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("request failed", err), nil
	}

	resJson, err := resDoc.MarshalJSON()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("marshal failed", err), nil
	}

	return mcp.NewToolResultText(string(resJson)), nil
}
