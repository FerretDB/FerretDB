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
	"github.com/FerretDB/wire/wirebson"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// tool represents MCP tool which clients can call to retrieve data or perform actions.
type tool struct {
	tool       mcp.Tool
	handleFunc server.ToolHandlerFunc
}

// ToolHandler handles MCP request.
type ToolHandler struct {
	h *handler.Handler
}

// NewToolHandler creates a new handler with the given parameters.
func NewToolHandler(h *handler.Handler) *ToolHandler {
	return &ToolHandler{
		h: h,
	}
}

// initTools returns available MCP tools.
func (h *ToolHandler) initTools() []tool {
	return []tool{
		{
			handleFunc: h.insert,
			tool:       newInsertTool(),
		},
	}
}

// request sends a request document to the handler and returns decoded response document.
func (h *ToolHandler) request(ctx context.Context, reqDoc *wirebson.Document) (*wirebson.Document, error) {
	req, err := wire.NewOpMsg(reqDoc)
	if err != nil {
		return nil, err
	}

	res, err := h.h.Handle(ctx, &middleware.Request{OpMsg: req})
	if err != nil {
		return nil, err
	}

	resDoc, err := res.OpMsg.DocumentDeep()
	if err != nil {
		return nil, err
	}

	return resDoc, nil
}
