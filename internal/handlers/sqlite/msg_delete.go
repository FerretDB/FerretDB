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
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDelete implements HandlerInterface.
func (h *Handler) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetDeleteParams(document, h.L)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var deleted int64
	var delErrors commonerrors.WriteErrors

	db := h.b.Database(params.DB)
	defer db.Close()

	coll := db.Collection(params.Collection)

	// process every delete filter
	for i, deleteParams := range params.Deletes {
		res, err := coll.Delete(ctx, &backends.DeleteParams{
			Filter:  deleteParams.Filter,
			Limited: deleteParams.Limited,
		})

		if err == nil {
			deleted += res.Deleted
			continue
		}

		delErrors.Append(err, int32(i))

		if params.Ordered {
			break
		}
	}

	replyDoc := must.NotFail(types.NewDocument(
		"ok", float64(1),
	))

	if delErrors.Len() > 0 {
		replyDoc = delErrors.Document()
	}

	replyDoc.Set("n", deleted)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	}))

	return &reply, nil
}
