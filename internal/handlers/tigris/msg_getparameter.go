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
	"errors"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetParameter implements HandlerInterface.
func (h *Handler) MsgGetParameter(_ context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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
		"authSchemaVersion", int32(5),
		"quiet", false,
		"ok", float64(1),
	))

	resDoc := resDB
	if getParameter != "*" {
		resDoc, err = selectParam(resDB)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{resDoc},
	}))

	commonparams.Ignored(document, h.L, "comment")

	if resDoc.Len() < 2 {
		return &reply, commonerrors.NewCommandErrorMsg(commonerrors.ErrorCode(0), "no option found to get")
	}

	return &reply, nil
}

// selectParam is makes a selection of requested parameters.
func selectParam(resDB *types.Document) (*types.Document, error) {
	doc := must.NotFail(types.NewDocument())

	iter := resDB.Iterator()
	defer iter.Close()

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, err
		}

		if k == "getParameter" || k == "comment" || k == "$db" {
			continue
		}

		doc.Set(k, v)
	}

	if doc.Len() < 1 {
		doc.Set("ok", float64(0))
		return doc, nil
	}

	doc.Set("ok", float64(1))

	return doc, nil
}
