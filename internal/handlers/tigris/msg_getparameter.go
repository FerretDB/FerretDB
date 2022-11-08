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

package tigris

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetParameter implements HandlerInterface.
func (h *Handler) MsgGetParameter(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	getParameter, err := document.Get(command)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resDB := must.NotFail(types.NewDocument(
		"acceptApiVersion2", false,
		"authSchemaVersion", int32(5),
		"quiet", false,
		"ok", float64(1),
	))

	var reply wire.OpMsg
	resDoc := resDB
	if getParameter != "*" {
		resDoc, err = selectParam(document, resDB)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	err = reply.SetSections(wire.OpMsgSection{Documents: []*types.Document{resDoc}})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "comment")

	if resDoc.Len() < 2 {
		return &reply, common.NewCommandErrorMsg(common.ErrorCode(0), "no option found to get")
	}

	return &reply, nil
}

// selectParam is makes a selection of requested parameters.
func selectParam(document, resDB *types.Document) (doc *types.Document, err error) {
	doc = must.NotFail(types.NewDocument())
	keys := document.Keys()

	for _, k := range keys {
		if k == "getParameter" || k == "comment" || k == "$db" {
			continue
		}

		item, err := resDB.Get(k)
		if err != nil {
			continue
		}

		doc.Set(k, item)
	}

	if doc.Len() < 1 {
		doc.Set("ok", float64(0))
		return doc, nil
	}

	doc.Set("ok", float64(1))

	return doc, nil
}
