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

// FindOne implements [ServerInterface].
func (s *Server) FindOne(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, fmt.Sprintf("Request:\n%s", must.NotFail(httputil.DumpRequest(r, true))))
	}

	var req api.FindOneRequestBody
	if err := decodeJSONRequest(r, &req); err != nil {
		http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	msg, err := prepareRequest(
		"find", req.Collection,
		"$db", req.Database,
		"filter", req.Filter,
		"projection", req.Projection,
		"limit", float64(1),
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

	cursor := resp.Document().Get("cursor").(wirebson.AnyDocument)
	firstBatch := must.NotFail(cursor.Decode()).Get("firstBatch").(wirebson.AnyArray)

	var res api.FindOneResponseBody

	docs := must.NotFail(firstBatch.Decode())
	if docs.Len() == 0 {
		s.writeJSONResponse(ctx, rw, &res)
		return
	}

	doc := docs.Get(0).(wirebson.AnyDocument)

	b, err := marshalSingleJSON(doc)
	if err != nil {
		http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	res.Document = &b

	s.writeJSONResponse(ctx, rw, &res)
}
