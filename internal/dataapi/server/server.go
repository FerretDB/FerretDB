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
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/FerretDB/wire/wirebson"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// bearerPrefix is the prefix for bearer token in the Authorization header.
const bearerPrefix = "Bearer "

// New creates a new Server.
func New(l *slog.Logger, handler *handler.Handler) *Server {
	return &Server{
		l:       l,
		handler: handler,
	}
}

// Server implements services described by OpenAPI description file.
type Server struct {
	l       *slog.Logger
	handler *handler.Handler

	m sync.Map
}

// AuthMiddleware authenticates the request using bearer token or basic authentication.
// If bearer token is provided, it authenticates using bearer token.
// Otherwise, basic auth is used and upon successful authentication
// it sets a bearer token in the response header to be used for subsequent requests.
// After a successful authentication, it calls the next handler.
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if bearer := strings.HasPrefix(r.Header.Get("Authorization"), bearerPrefix); bearer {
			if ok := s.bearerAuth(w, r); !ok {
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		if ok := s.basicAuth(w, r); !ok {
			return
		}

		if err := s.setBearerTokenHeader(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setBearerTokenHeader generates a new bearer token, stores it,
// and sets it in the response header.
func (s *Server) setBearerTokenHeader(w http.ResponseWriter) error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return err
	}

	token := fmt.Sprintf("%x", b)
	s.m.Store(token, true)
	w.Header().Set("Authorization", bearerPrefix+token)

	return nil
}

// bearerAuth checks if the request has a valid bearer token.
// If the token is valid, it returns true. Otherwise, it writes an error response
// and returns false.
func (s *Server) bearerAuth(w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, bearerPrefix)

	if _, ok := s.m.Load(token); !ok {
		http.Error(w, "token is not valid", http.StatusUnauthorized)
		return false
	}

	return true
}

// basicAuth handles basic authentication using SCRAM-SHA-256 mechanism.
// It returns true if authentication is successful, otherwise it writes an error response
// and returns false.
func (s *Server) basicAuth(w http.ResponseWriter, r *http.Request) bool {
	username, password, ok := r.BasicAuth()

	if !ok {
		writeError(w, errorNoAuthenticationSpecified, http.StatusBadRequest)
		return false
	}

	if password == "" || username == "" {
		writeError(w, errorMissingAuthenticationParameter, http.StatusBadRequest)
		return false
	}

	client, err := scram.SHA256.NewClient(username, password, "")
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusBadRequest)
		return false
	}

	conv := client.NewConversation()

	payload, err := conv.Step("")
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusBadRequest)
		return false
	}

	msg := must.NotFail(prepareRequest(
		"saslStart", int32(1),
		"mechanism", "SCRAM-SHA-256",
		"payload", wirebson.Binary{B: []byte(payload)},
		// use skipEmptyExchange to complete the handshake with one `saslStart` and one `saslContinue`
		"options", wirebson.MustDocument("skipEmptyExchange", true),
		"$db", "admin",
	))

	res, err := s.handler.Handle(r.Context(), msg)
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
		return false
	}

	resDoc := must.NotFail(must.NotFail(res.OpMsg.DocumentRaw()).Decode())
	convID := resDoc.Get("conversationId").(int32)

	payloadBytes := resDoc.Get("payload").(wirebson.Binary).B

	payload, err = conv.Step(string(payloadBytes))
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
		return false
	}

	msg = must.NotFail(prepareRequest(
		"saslContinue", int32(1),
		"conversationId", convID,
		"payload", wirebson.Binary{B: []byte(payload)},
		"$db", "admin",
	))

	res, err = s.handler.Handle(r.Context(), msg)
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
		return false
	}

	resDoc = must.NotFail(must.NotFail(res.OpMsg.DocumentRaw()).Decode())
	if !resDoc.Get("done").(bool) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return false
	}

	payloadBytes = resDoc.Get("payload").(wirebson.Binary).B

	if _, err = conv.Step(string(payloadBytes)); err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
		return false
	}

	if !conv.Valid() {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return false
	}

	return true
}

// writeJSONResponse marshals provided res document into extended JSON and
// writes it to provided [http.ResponseWriter].
func (s *Server) writeJSONResponse(ctx context.Context, w http.ResponseWriter, res wirebson.AnyDocument) {
	l := s.l

	resRaw, err := res.Encode()
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var resWriter io.Writer = w

	if l.Enabled(ctx, slog.LevelDebug) {
		buf := new(bytes.Buffer)

		resWriter = io.MultiWriter(w, buf)

		defer func() {
			// extended JSON value writer always finish with '\n' character
			l.DebugContext(ctx, fmt.Sprintf("Results:\n%s", strings.TrimSpace(buf.String())))
		}()
	}

	if err = marshalJSON(resRaw, resWriter); err != nil {
		l.ErrorContext(ctx, "marshalJSON failed", logging.Error(err))
	}
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
		return nil, err
	}

	return middleware.RequestDoc(doc), nil
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
