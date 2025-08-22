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
	"errors"
	"log/slog"

	"github.com/FerretDB/wire/wirebson"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.mongodb.org/mongo-driver/bson"
)

// insertArgs represents the arguments for the insert tool.
type insertArgs struct {
	Collection string          `json:"collection"`
	Database   string          `json:"database"`
	Documents  json.RawMessage `json:"documents"` // should be an array of documents
}

// insert inserts documents to a collection.
func (s *server) insert(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[insertArgs]) (*mcp.CallToolResult, error) { //nolint:lll // for readability
	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, "MCP tool params", slog.Any("params", params))
	}

	var raw any

	err := bson.UnmarshalExtJSON(params.Arguments.Documents, false, &raw)
	if err != nil {
		return nil, err
	}

	bsonType, b, err := bson.MarshalValue(raw)
	if err != nil {
		return nil, err
	}

	if bsonType != bson.TypeArray {
		return nil, errors.New("invalid type")
	}

	req := wirebson.MustDocument(
		"insert", params.Arguments.Collection,
		"documents", wirebson.RawArray(b),
		"$db", params.Arguments.Database,
	)

	return s.handle(ctx, req)
}
