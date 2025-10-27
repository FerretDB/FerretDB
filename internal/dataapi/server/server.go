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

// Package server provides a Data API server handlers.
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// New creates a new Server.
func New(l *slog.Logger, handler *middleware.Middleware) *Server {
	return &Server{
		l: l,
		m: handler,
	}
}

// Server implements services described by OpenAPI description file.
type Server struct {
	l *slog.Logger
	m *middleware.Middleware
}

// AuthMiddleware handles SCRAM authentication based on the username and password specified in request.
// After a successful handshake it calls the next handler.
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username, password, ok := r.BasicAuth()

		if !ok {
			writeError(rw, errorNoAuthenticationSpecified, http.StatusBadRequest)
			return
		}

		if password == "" || username == "" {
			writeError(rw, errorMissingAuthenticationParameter, http.StatusBadRequest)
			return
		}

		client, err := scram.SHA256.NewClient(username, password, "")
		if err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusBadRequest)
			return
		}

		conv := client.NewConversation()

		payload, err := conv.Step("")
		if err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusBadRequest)
			return
		}

		msg := must.NotFail(prepareRequest(
			"saslStart", int32(1),
			"mechanism", "SCRAM-SHA-256",
			"payload", wirebson.Binary{B: []byte(payload)},
			// use skipEmptyExchange to complete the handshake with one `saslStart` and one `saslContinue`
			"options", wirebson.MustDocument("skipEmptyExchange", true),
			"$db", "admin",
		))

		resp := s.m.Handle(ctx, msg)
		if resp == nil {
			http.Error(rw, "internal error", http.StatusInternalServerError)
			return
		}

		if !resp.OK() {
			s.writeJSONError(ctx, rw, resp)
			return
		}

		convID := resp.Document().Get("conversationId").(int32)
		payloadBytes := resp.Document().Get("payload").(wirebson.Binary).B

		payload, err = conv.Step(string(payloadBytes))
		if err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
			return
		}

		msg = must.NotFail(prepareRequest(
			"saslContinue", int32(1),
			"conversationId", convID,
			"payload", wirebson.Binary{B: []byte(payload)},
			"$db", "admin",
		))

		resp = s.m.Handle(ctx, msg)
		if resp == nil {
			http.Error(rw, "internal error", http.StatusInternalServerError)
			return
		}

		if !resp.OK() {
			s.writeJSONError(ctx, rw, resp)
			return
		}

		if !resp.Document().Get("done").(bool) {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		payloadBytes = resp.Document().Get("payload").(wirebson.Binary).B

		if _, err = conv.Step(string(payloadBytes)); err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
			return
		}

		if !conv.Valid() {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

// ConnInfoMiddleware returns a handler function that creates a new [*conninfo.ConnInfo],
// calls the next handler, and closes the connection info after the request is done.
func (s *Server) ConnInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ci := conninfo.New()

		defer ci.Close()

		next.ServeHTTP(rw, r.WithContext(conninfo.Ctx(r.Context(), ci)))
	})
}

// writeJSONResponse marshals provided res document into extended JSON and
// writes it to provided [http.ResponseWriter].
func (s *Server) writeJSONResponse(ctx context.Context, rw http.ResponseWriter, res api.Response) {
	rw.Header().Set("Content-Type", "application/json")

	var resWriter io.Writer = rw

	if s.l.Enabled(ctx, slog.LevelDebug) {
		buf := new(bytes.Buffer)

		resWriter = io.MultiWriter(rw, buf)

		defer func() {
			// extended JSON value writer always finish with '\n' character
			s.l.DebugContext(ctx, fmt.Sprintf("Results:\n%s", strings.TrimSpace(buf.String())))
		}()
	}

	err := json.NewEncoder(resWriter).Encode(res)
	if err != nil {
		s.l.ErrorContext(ctx, "marshalJSON failed", logging.Error(err))
	}
}

// TODO https://github.com/FerretDB/FerretDB/issues/4965
func (s *Server) writeJSONError(ctx context.Context, rw http.ResponseWriter, resp *middleware.Response) {
	doc := resp.Document()
	errmsg := doc.Get("errmsg").(string)
	codeName := doc.Get("codeName").(string)

	rw.Header().Set("Content-Type", "application/json")

	rw.WriteHeader(http.StatusInternalServerError)

	s.writeJSONResponse(ctx, rw, &api.Error{
		Error:     errmsg,
		ErrorCode: codeName,
	})
}

// prepareDocument creates a new bson document from the given pairs of
// field names and values, which can be used as handler command msg.
//
// If any of pair values is nil it's ignored.
func prepareDocument(pairs ...any) (*wirebson.Document, error) {
	l := len(pairs)

	if l%2 != 0 {
		return nil, lazyerrors.Errorf("invalid number of arguments: %d", l)
	}

	docPairs := make([]any, 0, l)

	for i := 0; i < l; i += 2 {
		var err error

		key := pairs[i]
		v := pairs[i+1]

		switch val := v.(type) {
		// json.RawMessage is the non-pointer exception.
		// Other non-pointer types don't need special handling.
		case json.RawMessage:
			v, err = unmarshalSingleJSON(&val)
			if err != nil {
				return nil, err
			}

		case *json.RawMessage:
			if val == nil {
				continue
			}

			v, err = unmarshalSingleJSON(val)
			if err != nil {
				return nil, err
			}
		case *float32:
			if val == nil {
				continue
			}

			v = float64(*val)
		case *bool:
			if val == nil {
				continue
			}

			v = *val
		}

		if v == nil {
			continue
		}

		docPairs = append(docPairs, key, v)
	}

	return wirebson.NewDocument(docPairs...)
}

// prepareRequest creates a new middleware request from the given pairs of field names and values,
// which can be used as handler command msg.
//
// If any of pair values is nil it's ignored.
func prepareRequest(pairs ...any) (*middleware.Request, error) {
	doc, err := prepareDocument(pairs...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return middleware.RequestDoc(doc)
}

// decodeJSONRequest takes request with JSON body and decodes it into
// provided oapi generated request struct.
func decodeJSONRequest(r *http.Request, out any) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return lazyerrors.New("Content-Type must be set to application/json")
	}

	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// check interfaces
var (
	_ api.ServerInterface = (*Server)(nil)
)
