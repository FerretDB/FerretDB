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

package sqlite

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgInsert implements HandlerInterface.
func (h *Handler) MsgInsert(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetInsertParams(document, h.L)
	if err != nil {
		return nil, err
	}

	db := h.b.Database(params.DB)
	defer db.Close()

	iter := params.Docs.Iterator()
	defer iter.Close()

	res, err := db.Collection(params.Collection).Insert(ctx, &backends.InsertParams{
		Iter: iter,
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	replyDoc := must.NotFail(types.NewDocument(
		"n", int32(res.Inserted),
		"ok", float64(1),
	))

	// TODO https://github.com/FerretDB/FerretDB/issues/2750
	//
	// if len(res.Errors) > 0 {
	// 	var errs *commonerrors.WriteErrors
	//
	// 	for i := 0; i < len(res.Errors); i++ {
	// 		errs.Append(err, int32(i))
	// 	}
	//
	// 	replyDoc = errs.Document()
	// }

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	}))

	return &reply, nil
}
