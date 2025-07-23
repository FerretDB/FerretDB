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
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// listDatabases returns a list of databases in a string containing Extended JSON v2 format.
func (h *ToolHandler) listDatabases(ctx context.Context, _ *mcp.ServerSession, _ *mcp.CallToolParamsFor[any]) (*mcp.CallToolResultFor[any], error) { //nolint:lll // for readability
	req := wirebson.MustDocument(
		"listDatabases", int32(1),
	)

	res, err := h.request(ctx, req)
	if err != nil {
		return nil, err
	}

	resJson, err := res.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(resJson)}},
	}, nil
}
