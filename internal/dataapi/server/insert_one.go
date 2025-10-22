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

package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// InsertOne implements [ServerInterface].
func (s *Server) InsertOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, fmt.Sprintf("Request:\n%s", must.NotFail(httputil.DumpRequest(r, true))))
	}

	var req api.InsertOneRequestBody
	if err := decodeJSONRequest(r, &req); err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	insert, err := unmarshalSingleJSON(&req.Document)
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	insertDoc, ok := insert.(wirebson.AnyDocument)
	if !ok {
		http.Error(w, lazyerrors.New("document must be a BSON document").Error(), http.StatusInternalServerError)
		return
	}

	var doc *wirebson.Document

	doc, err = ensureID(insertDoc)
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	documents := wirebson.MustArray(doc)

	msg, err := prepareRequest(
		"insert", req.Collection,
		"$db", req.Database,
		"documents", documents,
	)
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	resp := s.m.Handle(ctx, msg)
	if resp == nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !resp.OK() {
		s.writeJSONError(ctx, w, resp)
		return
	}

	insertedId, err := wirebson.ToDriver(doc.Get("_id"))
	if err != nil {
		http.Error(w, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	res := api.InsertOneResponseBody{
		InsertedId: &insertedId,
	}

	s.writeJSONResponse(ctx, w, &res)
}

// ensureID ensures that inserted document has an "_id" field.
func ensureID(doc wirebson.AnyDocument) (*wirebson.Document, error) {
	decoded, err := doc.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	for field := range decoded.Fields() {
		if field == "_id" {
			return decoded, nil
		}
	}

	id, err := wirebson.FromDriver(bson.NewObjectID())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	err = decoded.Add("_id", id)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return decoded, err
}
