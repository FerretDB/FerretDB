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

// Find implements [ServerInterface].
func (s *Server) Find(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.l.Enabled(ctx, slog.LevelDebug) {
		s.l.DebugContext(ctx, fmt.Sprintf("Request:\n%s", must.NotFail(httputil.DumpRequest(r, true))))
	}

	var req api.FindManyRequestBody
	if err := decodeJSONRequest(r, &req); err != nil {
		http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	msg, err := prepareRequest(
		"find", req.Collection,
		"$db", req.Database,
		"filter", req.Filter,
		"limit", req.Limit,
		"projection", req.Projection,
		"skip", req.Skip,
		"sort", req.Sort,
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

	b, err := marshalSingleJSON(firstBatch)
	if err != nil {
		http.Error(rw, lazyerrors.Error(err).Error(), http.StatusInternalServerError)
		return
	}

	res := api.FindManyResponseBody{
		Documents: &b,
	}

	s.writeJSONResponse(ctx, rw, &res)
}
