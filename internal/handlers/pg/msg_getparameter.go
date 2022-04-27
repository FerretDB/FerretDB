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

package pg

import (
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetParameter OpMsg used to get parameter.
func (h *Handler) MsgGetParameter(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	cmd, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	getPrm, err := cmd.Get("getParameter")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// SELECT * FROM 'admin'
	resDB := types.MustNewDocument(
		"acceptApiVersion2", false,
		"authSchemaVersion", int32(5),
		"quiet", false,
		"ok", float64(1),
	)

	var reply wire.OpMsg
	var errNoFound error
	var resDoc *types.Document
	if getPrm == "*" {
		resDoc = resDB

	} else {
		resDoc, errNoFound, err = selectParam(cmd, resDB)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{resDoc},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	comment, err := cmd.Get("comment")
	if err == nil {
		common.Ignored(cmd, h.l, fmt.Sprint(comment))
	}

	return &reply, errNoFound
}

// selectParam is makes a selection of the requested parameters received from database 'admin'.
func selectParam(cmd, resDB *types.Document) (doc *types.Document, errNoFound error, err error) {
	doc = types.MustNewDocument()
	keys := cmd.Keys()
	for _, k := range keys {
		if k == "getParameter" || k == "comment" || k == "$db" {
			continue
		}

		item, err := resDB.Get(k)
		if err != nil {
			continue
		}

		err = doc.Set(k, item)
		if err != nil {
			return nil, nil, err
		}
	}
	if doc.Len() < 1 {
		err := doc.Set("ok", float64(0))
		if err != nil {
			return nil, nil, err
		}
		errNoFound = common.NewErrorMsg(common.ErrorCode(0), "no option found to get")
	} else {
		err := doc.Set("ok", float64(1))
		if err != nil {
			return nil, nil, err
		}
	}

	return doc, errNoFound, nil
}
