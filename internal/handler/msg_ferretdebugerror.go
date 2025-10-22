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
	"errors"
	"strconv"
	"strings"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// msgFerretDebugError implements `ferretDebugError` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgFerretDebugError(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	arg, err := getRequiredParam[string](doc, doc.Command())
	if err != nil {
		return nil, err
	}

	if code, _ := strconv.ParseInt(arg, 10, 32); code > 0 {
		return nil, mongoerrors.New(mongoerrors.Code(code), "debug error message")
	}

	switch {
	case arg == "ok":
		return middleware.ResponseDoc(req, wirebson.MustDocument(
			"ok", float64(1),
		))

	case strings.HasPrefix(arg, "panic"):
		panic("debugError " + arg)

	case strings.HasPrefix(arg, "lazy"):
		return nil, lazyerrors.New(arg)

	default:
		return nil, errors.New(arg)
	}
}
