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

// Package httpauth provides utilities for handling HTTP authentication.
package httpauth

import (
	"net/http"

	"github.com/FerretDB/wire/wirebson"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// AuthHandler handles authentication.
type AuthHandler struct { //nolint:vet // ok for struct instantiated once
	handler *handler.Handler
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(handler *handler.Handler) *AuthHandler {
	return &AuthHandler{
		handler: handler,
	}
}

// AuthMiddleware handles SCRAM authentication based on the username and password specified in request.
// After a successful handshake it calls the next handler.
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username, password, ok := r.BasicAuth()

		if !ok {
			writeError(w, errorNoAuthenticationSpecified, http.StatusBadRequest)
			return
		}

		if password == "" || username == "" {
			writeError(w, errorMissingAuthenticationParameter, http.StatusBadRequest)
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

		msg := middleware.RequestDoc(wirebson.MustDocument(
			"saslStart", int32(1),
			"mechanism", "SCRAM-SHA-256",
			"payload", wirebson.Binary{B: []byte(payload)},
			// use skipEmptyExchange to complete the handshake with one `saslStart` and one `saslContinue`
			"options", wirebson.MustDocument("skipEmptyExchange", true),
			"$db", "admin",
		))

		res, err := h.handler.Handle(ctx, msg)
		if err != nil {
			http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
			return
		}

		resDoc := must.NotFail(must.NotFail(res.OpMsg.DocumentRaw()).Decode())
		convID := resDoc.Get("conversationId").(int32)

		payloadBytes := resDoc.Get("payload").(wirebson.Binary).B

		payload, err = conv.Step(string(payloadBytes))
		if err != nil {
			http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
			return
		}

		msg = middleware.RequestDoc(wirebson.MustDocument(
			"saslContinue", int32(1),
			"conversationId", convID,
			"payload", wirebson.Binary{B: []byte(payload)},
			"$db", "admin",
		))

		res, err = h.handler.Handle(ctx, msg)
		if err != nil {
			http.Error(w, lazyerrors.Error(err).Error(), http.StatusUnauthorized)
			return
		}

		resDoc = must.NotFail(must.NotFail(res.OpMsg.DocumentRaw()).Decode())
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
