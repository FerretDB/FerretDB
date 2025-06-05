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
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// newListDatabases creates a new MCP tool for listDatabases command.
func newListDatabases() mcp.Tool {
	return mcp.NewTool("list-databases",
		mcp.WithDescription(
			"Returns a list of databases by running listDatabases command. "+
				"It uses Extended JSON v2 format for the response. "+
				"Use this tool if you need to retrieve a list of databases. "+
				"The response may be truncated if there are many databases, "+
				"which is indicated by the presence of non zero cursor."),
	)
}

// listDatabases returns the list of databases in a string containing Extended JSON v2 format.
func (h *ToolHandler) listDatabases(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	req := wire.MustOpMsg(
		"listDatabases", int32(1),
	)

	res, err := h.h.Handle(ctx, &middleware.Request{OpMsg: req})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to handle OP_MSG", err), nil
	}

	resDoc, err := res.OpMsg.DocumentDeep()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to decode OP_MSG", err), nil
	}

	resJson, err := resDoc.MarshalJSON()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to marshal", err), nil
	}

	return mcp.NewToolResultText(string(resJson)), nil
}
