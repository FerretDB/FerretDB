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
	"go.mongodb.org/mongo-driver/v2/bson"
)

// findArgs represents the arguments for the find tool.
type findArgs struct {
	Collection string `json:"collection"`
	Database   string `json:"database"`
	Limit      int64  `json:"limit"`
	// filter is hard, the tool does not know how to construct a bson filter,
	// also missing projection, skip and sort
}

// find returns documents from the collection.
func (s *server) find(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[findArgs]) (*mcp.CallToolResult, error) { //nolint:lll // for readability
	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, "MCP tool params", slog.Any("params", params))
	}

	req := wirebson.MustDocument(
		"find", params.Arguments.Collection,
		"limit", params.Arguments.Limit,
		"$db", params.Arguments.Database,
	)

	return s.handle(ctx, req)
}

// toWireBSON converts a JSON raw message to a wirebson.RawDocument or wirebson.RawArray.
func toWireBSON(b json.RawMessage) (any, error) {
	var raw any

	err := bson.UnmarshalExtJSON(b, false, &raw)
	if err != nil {
		return nil, err
	}

	bsonType, b, err := bson.MarshalValue(raw)
	if err != nil {
		return nil, err
	}

	switch bsonType {
	case bson.TypeArray:
		return wirebson.RawArray(b), nil
	case bson.TypeEmbeddedDocument:
		return wirebson.RawDocument(b), nil
	default:
		return nil, errors.New("invalid type")
	}
}
