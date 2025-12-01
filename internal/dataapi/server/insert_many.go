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

	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// InsertMany implements [ServerInterface].
func (s *Server) InsertMany(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, fmt.Sprintf("Request:\n%s", must.NotFail(httputil.DumpRequest(r, true))))
	}

	var req api.InsertManyRequestBody
	if err := decodeJSONRequest(r, &req); err != nil {
		http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	docsArr, err := unmarshalSingleJSON(&req.Documents)
	if err != nil {
		http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	documents, err := docsArr.(wirebson.RawArray).Decode()
	if err != nil {
		http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	var insertedIds []any

	for i, v := range documents.All() {
		v, ok := v.(wirebson.AnyDocument)
		if !ok {
			http.Error(rw, fmt.Sprintf("document %d is not a valid BSON document", i), http.StatusBadRequest)
			return
		}

		var doc *wirebson.Document

		doc, err = ensureID(v)
		if err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
			return
		}

		var insertedId any

		insertedId, err = wirebson.ToDriver(doc.Get("_id"))
		if err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
			return
		}

		if err = documents.Replace(i, doc); err != nil {
			http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
			return
		}

		insertedIds = append(insertedIds, insertedId)
	}

	msg, err := prepareRequest(
		"insert", req.Collection,
		"$db", req.Database,
		"documents", documents,
	)
	if err != nil {
		http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	resp := s.m.Handle(ctx, msg)
	if resp == nil {
		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}

	if !resp.OK() {
		s.writeJSONError(ctx, rw, resp)
		return
	}

	res := api.InsertManyResponseBody{
		InsertedIds: &insertedIds,
	}

	s.writeJSONResponse(ctx, rw, &res)
}
