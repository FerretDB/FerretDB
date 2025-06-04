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
	"log/slog"

	"github.com/FerretDB/FerretDB/v2/internal/handler"
)

// ToolHandler handles MCP request.
type ToolHandler struct {
	h *handler.Handler
	l *slog.Logger
}

// NewToolHandler creates a new MCP handler with the given parameters.
func NewToolHandler(h *handler.Handler, l *slog.Logger) *ToolHandler {
	return &ToolHandler{
		h: h,
		l: l,
	}
}

// initTools returns available MCP tools.
func (h *ToolHandler) initTools() map[string]tool {
	return map[string]tool{
		"find": {
			handleFunc: h.find,
			tool:       newFindTool(),
		},
		"insert": {
			handleFunc: h.insert,
			tool:       newInsertTool(),
		},
		"listCollections": {
			handleFunc: h.listCollections,
			tool:       newListCollections(),
		},
		"listDatabases": {
			handleFunc: h.listDatabases,
			tool:       newListDatabases(),
		},
	}
}
