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

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// server handles MCP request.
type server struct {
	l *slog.Logger
	m *middleware.Middleware
}

// newServer creates a new server with the given parameter.
func newServer(l *slog.Logger, m *middleware.Middleware) *server {
	return &server{
		l: l,
		m: m,
	}
}

// addTools adds available MCP tools for the given mcp server.
func (s *server) addTools(srv *mcp.Server) {
	// sorted alphabetically
	mcp.AddTool(
		srv,
		&mcp.Tool{
			Name:        "find",
			Description: "Returns documents matched by the query.",
		},
		s.find,
	)
	mcp.AddTool(
		srv,
		&mcp.Tool{
			Name:        "listCollections",
			Description: "Returns the information of the collections and views in the database.",
		},
		s.listCollections,
	)
	mcp.AddTool(
		srv,
		&mcp.Tool{
			Name:        "listDatabases",
			Description: "Returns a summary of all databases.",
		},
		s.listDatabases,
	)
}

// handle sends the request document to the middleware and returns result used by MCP tool.
func (s *server) handle(ctx context.Context, reqDoc *wirebson.Document) (*mcp.CallToolResult, error) {
	req, err := middleware.RequestDoc(reqDoc)
	if err != nil {
		return nil, err
	}

	resp := s.m.Handle(ctx, req)
	if resp == nil {
		return nil, errors.New("internal error")
	}

	doc, err := resp.DocumentRaw().DecodeDeep()
	if doc == nil {
		return nil, err
	}

	b, err := doc.MarshalJSON()
	if err != nil {
		return nil, err
	}

	res := &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
		IsError: !resp.OK(),
	}

	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, "MCP tool result", slog.Any("result", res))
	}

	return res, nil
}

// fromExtendedJSON converts raw encoded extended JSON v2 to a wirebson.RawDocument or wirebson.RawArray.
//
//nolint:unused // for now
func fromExtendedJSON(b json.RawMessage) (any, error) {
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
	case bson.TypeEmbeddedDocument:
		return wirebson.RawDocument(b), nil
	case bson.TypeArray:
		return wirebson.RawArray(b), nil
	case bson.TypeDouble,
		bson.TypeString,
		bson.TypeBinary,
		bson.TypeUndefined,
		bson.TypeObjectID,
		bson.TypeBoolean,
		bson.TypeDateTime,
		bson.TypeNull,
		bson.TypeRegex,
		bson.TypeDBPointer,
		bson.TypeJavaScript,
		bson.TypeSymbol,
		bson.TypeCodeWithScope,
		bson.TypeInt32,
		bson.TypeTimestamp,
		bson.TypeInt64,
		bson.TypeDecimal128,
		bson.TypeMinKey,
		bson.TypeMaxKey:
		fallthrough
	default:
		return nil, errors.New("unsupported type")
	}
}
