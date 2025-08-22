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

	"github.com/FerretDB/wire/wirebson"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// dropDatabaseArgs represents the arguments for the dropDatabase tool.
type dropDatabaseArgs struct {
	Database string `json:"database"`
}

// dropDatabase deletes the database.
func (s *server) dropDatabase(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[dropDatabaseArgs]) (*mcp.CallToolResult, error) { //nolint:lll // for readability
	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, "MCP tool params", slog.Any("params", params))
	}

	req := wirebson.MustDocument(
		"dropDatabase", int32(1),
		"$db", params.Arguments.Database,
	)

	return s.handle(ctx, req)
}
