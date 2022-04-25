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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetParameter OpMsg used to get parameter.
func (h *Handler) MsgGetParameter(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	getPrm, err := document.Get("getParameter")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// SELECT * FROM 'admin'
	resDB := types.MustNewDocument(
		"quiet", false,
		"ok", float64(1),
	)

	var reply wire.OpMsg
	if getPrm == "*" {
		err = reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{resDB},
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	} else {
		keys := document.Keys()
		res := types.MustNewDocument()
		for _, k := range keys {
			if k == "getParameter" || k == "comment" || k == "$db" {
				continue
			}
			item, err := resDB.Get(k)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			err = res.Set(k, item)

			if err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		err = reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{res},
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	// comment, err := document.Get("comment")
	// if err == nil {
	// 	common.Ignored(document, h.l, fmt.Sprint(comment))
	// }

	return &reply, nil

	// fmt.Printf("h: %+v\n", h)
	//	fmt.Printf("msg: %+v\n", msg)
	//	fmt.Printf("document: %+v\n", document)
	// TODO https://github.com/FerretDB/FerretDB/issues/449
}
