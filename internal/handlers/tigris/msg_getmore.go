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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetMore implements handlers.Interface.
func (h *Handler) MsgGetMore(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = common.Unimplemented(document, "comment", "maxTimeMS"); err != nil {
		return nil, err
	}

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	cursorID, err := common.GetRequiredParam[int64](document, "getMore")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if cursorID != 1 {
		return nil, lazyerrors.Errorf("cursor not found")
	}

	collection, err := common.GetRequiredParam[string](document, "collection")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	batchSize, err := common.GetOptionalParam(document, "batchSize", int32(common.DefaultBatchSize))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	info := conninfo.Get(ctx)

	cur := info.Cursor(db + "." + collection)
	if cur == nil {
		return nil, lazyerrors.Errorf("cursor for collection %s not found", collection)
	}

	resDocs := types.MakeArray(0)

	var done bool

	for i := 0; i < int(batchSize); i++ {
		var doc any

		_, doc, err = cur.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				done = true
				break
			}

			return nil, lazyerrors.Error(err)
		}

		resDocs.Append(doc)
	}

	id := int64(1)
	if done {
		id = 0
	}

	var reply wire.OpMsg

	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"id", id,
				"ns", collection,
				"nextBatch", resDocs,
			)),
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
