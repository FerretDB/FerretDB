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

	"github.com/AlekSi/lazyerrors"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// msgGetMore implements `getMore` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgGetMore(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	userID, sessionID, err := h.s.CreateOrUpdateByLSID(connCtx, doc)
	if err != nil {
		return nil, err
	}

	command := doc.Command()

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	v := doc.Get(command)

	cursorID, ok := v.(int64)
	if !ok {
		m := fmt.Sprintf("BSON field 'getMore.getMore' is the wrong type '%T', expected type 'long'", v)
		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, m, command)
	}

	if err = h.s.ValidateCursor(userID, sessionID, cursorID); err != nil {
		return nil, err
	}

	page, err := h.p.GetMore(connCtx, dbName, req.DocumentRaw(), cursorID)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return middleware.ResponseDoc(req, page)
}
