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

package httpauth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/FerretDB/wire/wirebson"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// AuthHandler handles authentication.
type AuthHandler struct { //nolint:vet // ok for struct instantiated once
	handler *handler.Handler
	m       sync.Map

	l *slog.Logger
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(handler *handler.Handler, l *slog.Logger) *AuthHandler {
	return &AuthHandler{
		handler: handler,
		l:       l,
	}
}

// AuthMiddleware authenticates the request using bearer token or basic authentication.
// If bearer token is provided, it authenticates using bearer token.
// Otherwise, basic auth is used and upon successful authentication
// it sets a bearer token in the response header.
// After a successful authentication, it calls the next handler if not nil.
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if bearer := strings.HasPrefix(r.Header.Get("Authorization"), "Bearer "); bearer {
			ci := h.bearerAuth(w, r)
			if ci == nil {
				return
			}

			h.l.DebugContext(r.Context(), "Authenticated with bearer token")

			if next != nil {
				next.ServeHTTP(w, r.WithContext(conninfo.Ctx(r.Context(), ci)))
			}

			return
		}

		if ok := h.basicAuth(w, r); !ok {
			return
		}

		h.l.DebugContext(r.Context(), "Authenticated with SCRAM")

		if err := h.setBearerTokenHeader(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}

// setBearerTokenHeader generates a new bearer token, stores it,
// and sets it in the response header.
func (h *AuthHandler) setBearerTokenHeader(ctx context.Context, w http.ResponseWriter) error {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return err
	}

	token := fmt.Sprintf("%x", b)
	h.m.Store(token, conninfo.Get(ctx))
	w.Header().Set("Authorization", "Bearer "+token)

	return nil
}

// bearerAuth checks if the request has a valid bearer token.
// If the token is valid, it returns [*conninfo.ConnInfo] associated with token.
// Otherwise, it writes an error response and returns nil.
func (h *AuthHandler) bearerAuth(w http.ResponseWriter, r *http.Request) *conninfo.ConnInfo {
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")

	ci, ok := h.m.Load(token)
	if !ok {
		http.Error(w, "token is not valid", http.StatusUnauthorized)
	}

	return ci.(*conninfo.ConnInfo)
}

// basicAuth handles basic authentication using SCRAM-SHA-256 mechanism.
// It returns true if authentication is successful, otherwise it writes an error response
// and returns false.
func (h *AuthHandler) basicAuth(w http.ResponseWriter, r *http.Request) bool {
	username, password, ok := r.BasicAuth()

	if !ok {
		http.Error(w, "no authentication methods were specified", http.StatusUnauthorized)
		return false
	}

	if password == "" || username == "" {
		msg := "must specify some form of authentication (either email+password, api-key, or jwt) in the request header or body"
		http.Error(w, msg, http.StatusUnauthorized)

		return false
	}

	client, err := scram.SHA256.NewClient(username, password, "")
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
		return false
	}

	conv := client.NewConversation()

	payload, err := conv.Step("")
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
		return false
	}

	req := middleware.RequestDoc(wirebson.MustDocument(
		"saslStart", int32(1),
		"mechanism", "SCRAM-SHA-256",
		"payload", wirebson.Binary{B: []byte(payload)},
		// use skipEmptyExchange to complete the handshake with one `saslStart` and one `saslContinue`
		"options", wirebson.MustDocument("skipEmptyExchange", true),
		"$db", "admin",
	))

	res, err := h.handler.Handle(r.Context(), req)
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

	req = middleware.RequestDoc(wirebson.MustDocument(
		"saslContinue", int32(1),
		"conversationId", convID,
		"payload", wirebson.Binary{B: []byte(payload)},
		"$db", "admin",
	))

	res, err = h.handler.Handle(r.Context(), req)
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
