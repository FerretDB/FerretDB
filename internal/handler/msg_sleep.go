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
	"fmt"
	"time"

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
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

	if dbName != "admin" {
		return nil, mongoerrors.New(mongoerrors.ErrUnauthorized, "sleep may only be run against the admin database.")
	}

	sleepDur := 10000 * time.Millisecond

	if v := doc.Get("millis"); v != nil {
		millis, ok := v.(int32)

		if !ok {
			return nil, lazyerrors.Error(mongoerrors.New(mongoerrors.ErrBadValue, "parameter millis has invalid type"))
		}

		sleepDur = time.Duration(max(millis, 0)) * time.Millisecond
	}

	lock, err := getRequiredParam[string](doc, "lock")
	if err != nil {
		return nil, err
	}

	if lock != "w" {
		return nil, lazyerrors.Error(
			mongoerrors.New(mongoerrors.ErrBadValue, fmt.Sprintf("parameter lock %s value is not supported", lock)),
		)
	}

	h.runM.Lock()
	ctxutil.Sleep(connCtx, sleepDur)
	h.runM.Unlock()

	return middleware.ResponseDoc(req, wirebson.MustDocument("ok", float64(1)))
}
