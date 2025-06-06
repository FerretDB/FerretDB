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
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

// openAPISchemaResource returns a resource that contains the OpenAPI schema.
func openAPISchemaResource() mcp.Resource {
	return mcp.NewResource(
		"openapi-schema",
		"OpenAPI",
		mcp.WithResourceDescription("OpenAPI schema for CRUD operations"),
		mcp.WithMIMEType("application/json"),
	)
}

// openAPISchema returns a resource that contains the OpenAPI schema.
func (h *ToolHandler) openAPISchema(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// FIXME
	f, err := os.ReadFile(" ../dataapi/api/openapi.json")
	if err != nil {
		return nil, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "openapi-schema",
			MIMEType: "application/json",
			Text:     string(f),
		},
	}, nil
}
