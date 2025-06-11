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

	"github.com/FerretDB/wire/wirebson"
	"github.com/mark3labs/mcp-go/mcp"
)

// newListDatabases creates a new MCP tool for listDatabases command.
func newListDatabases() mcp.Tool {
	return mcp.NewTool(
		"listDatabases",
		mcp.WithDescription("Returns a summary of all databases."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

// listDatabases returns a list of databases in a string containing Extended JSON v2 format.
func (h *toolHandler) listDatabases(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	req := wirebson.MustDocument(
		"listDatabases", int32(1),
	)

	res, err := h.request(ctx, req)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("request failed", err), nil
	}

	resJson, err := res.MarshalJSON()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("marshal failed", err), nil
	}

	return mcp.NewToolResultText(string(resJson)), nil
}
