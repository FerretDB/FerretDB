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
	"errors"
	"net/http"

	"github.com/FerretDB/wire/wirebson"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// server handles MCP request.
type server struct {
	m *middleware.Middleware
}

// newServer creates a new server with the given parameter.
func newServer(m *middleware.Middleware) *server {
	return &server{
		m: m,
	}
}

// addTools adds available MCP tools for the given mcp server.
func (s *server) addTools(srv *mcp.Server) {
	listDatabasesTool := &mcp.Tool{
		Name:        "listDatabases",
		Description: "Returns a summary of all databases.",
	}
	mcp.AddTool(srv, listDatabasesTool, s.listDatabases)
}

// handle sends the request document to the middleware and returns result used by MCP tool.
//
// Log MCP tool result for debug level.
// TODO https://github.com/FerretDB/FerretDB/issues/5277
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

	json, err := doc.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(json)}},
		IsError: !resp.OK(),
	}, nil
}

// authMiddleware handles SCRAM authentication based on the username and password specified in request.
// After a successful handshake it calls the next handler.
func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username, password, ok := r.BasicAuth()

		if !ok {
			http.Error(w, "no authentication specified", http.StatusBadRequest)
			return
		}

		if password == "" || username == "" {
			http.Error(w, "missing username or password", http.StatusBadRequest)
			return
		}

		client, err := scram.SHA256.NewClient(username, password, "")
		if err != nil {
			http.Error(w, lazyerrors.Error(err).Error(), http.StatusBadRequest)
			return
		}

		conv := client.NewConversation()

		payload, err := conv.Step("")
		if err != nil {
			http.Error(w, lazyerrors.Error(err).Error(), http.StatusBadRequest)
			return
		}

		msg, err := middleware.RequestDoc(wirebson.MustDocument(
			"saslStart", int32(1),
			"mechanism", "SCRAM-SHA-256",
			"payload", wirebson.Binary{B: []byte(payload)},
			// use skipEmptyExchange to complete the handshake with one `saslStart` and one `saslContinue`
			"options", wirebson.MustDocument("skipEmptyExchange", true),
			"$db", "admin",
		))
		if err != nil {
			http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
			return
		}

		res := s.m.Handle(ctx, msg)
		if res == nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		resDoc := must.NotFail(res.DocumentRaw().Decode())
		convID := resDoc.Get("conversationId").(int32)

		payloadBytes := resDoc.Get("payload").(wirebson.Binary).B

		payload, err = conv.Step(string(payloadBytes))
		if err != nil {
			http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
			return
		}

		msg, err = middleware.RequestDoc(wirebson.MustDocument(
			"saslContinue", int32(1),
			"conversationId", convID,
			"payload", wirebson.Binary{B: []byte(payload)},
			"$db", "admin",
		))
		if err != nil {
			http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
			return
		}

		res = s.m.Handle(ctx, msg)
		if res == nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		resDoc = must.NotFail(res.DocumentRaw().Decode())
		if !resDoc.Get("done").(bool) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		payloadBytes = resDoc.Get("payload").(wirebson.Binary).B

		if _, err = conv.Step(string(payloadBytes)); err != nil {
			http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
			return
		}

		if !conv.Valid() {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
