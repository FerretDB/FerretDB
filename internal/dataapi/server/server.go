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

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

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
		return nil, lazyerrors.Error(err)
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
