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

// newListDatabases creates a new MCP tool for listDatabases command.
func newListDatabases() mcp.Tool {
	return mcp.NewTool("list databases",
		mcp.WithDescription("Returns list of all databases"),
	)
}

// listDatabases calls the listDatabases command and returns the results.
func (h *Handler) listDatabases(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	req := wire.MustOpMsg(
		"listDatabases", int32(1),
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
