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

package handler

import (
	"context"
	"time"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
)

func (h *Handler) msgSleep(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	// TODO error
	if dbName != "admin" {
	}

	useDefault := true
	sleepDur := 10000 * time.Millisecond

	if v := doc.Get("millis"); v != nil {
		useDefault = false
		millis, ok := v.(int32)
		// TODO error
		if !ok {
			return nil, lazyerrors.Error(mongoerrors.New(mongoerrors.ErrBadValue, "parameter millis has invalid type"))
		}

		sleepDur = time.Duration(max(millis, 0)) * time.Millisecond
	}

	if v := doc.Get("secs"); v != nil {
		secs, ok := v.(int32)
		// TODO error
		if !ok {
			return nil, lazyerrors.Error(mongoerrors.New(mongoerrors.ErrBadValue, "parameter secs has invalid type"))
		}

		secs = max(secs, 0)

		if useDefault {
			sleepDur = time.Duration(secs) * time.Second
		} else {
			sleepDur = time.Duration(max(sleepDur, time.Duration(secs)*time.Second))
		}
	}

	ctxutil.Sleep(connCtx, sleepDur)

	return middleware.ResponseDoc(req, wirebson.MustDocument("ok", float64(1)))
}
