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
	Collection string          `json:"collection"`
	Database   string          `json:"database"`
	Filter     json.RawMessage `json:"filter"`
	Limit      int64           `json:"limit"`
	Projection json.RawMessage `json:"projection"`
	Skip       int64           `json:"skip"`
	Sort       json.RawMessage `json:"sort"`
}

// find returns documents from the collection.
func (s *server) find(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[findArgs]) (*mcp.CallToolResult, error) { //nolint:lll // for readability
	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, "MCP tool params", slog.Any("params", params))
	}

	filter, err := toWireBSON(params.Arguments.Filter)
	if err != nil {
		return nil, err
	}

	projection, err := toWireBSON(params.Arguments.Projection)
	if err != nil {
		return nil, err
	}

	sort, err := toWireBSON(params.Arguments.Sort)
	if err != nil {
		return nil, err
	}

	req := wirebson.MustDocument(
		"find", params.Arguments.Collection,
		"filter", filter,
		"limit", params.Arguments.Limit,
		"projection", projection,
		"skip", params.Arguments.Skip,
		"sort", sort,
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
