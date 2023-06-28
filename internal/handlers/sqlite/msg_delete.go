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
	"errors"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
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

	var deleted int32
	var delErrors commonerrors.WriteErrors

	db := h.b.Database(params.DB)
	defer db.Close()

	coll := db.Collection(params.Collection)

	// process every delete filter
	for i, deleteParams := range params.Deletes {
		del, err := execDelete(ctx, coll, deleteParams.Filter, deleteParams.Limited)
		if err == nil {
			deleted += del
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

// execDelete fetches documents, filters them out, limits them (if needed) and deletes them.
// If limited is true, only the first matched document is chosen for deletion, otherwise all matched documents are chosen.
// It returns the number of deleted documents or an error.
func execDelete(ctx context.Context, coll backends.Collection, filter *types.Document, limited bool) (int32, error) {
	// query documents here
	res, err := coll.Query(ctx, nil)
	if err != nil {
		return 0, lazyerrors.Error(err)
	}

	defer res.Iter.Close()

	var ids []any
	var deleted int32

	for {
		_, doc, err := res.Iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return 0, lazyerrors.Error(err)
		}

		matches, err := common.FilterDocument(doc, filter)
		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		if !matches {
			continue
		}

		ids = append(ids, must.NotFail(doc.Get("_id")))

		// if limit is set, no need to fetch all the documents
		if limited {
			res.Iter.Close() // good comment

			break
		}
	}

	// if no documents matched, there is nothing to delete
	if len(ids) == 0 {
		return 0, nil
	}

	deleteRes, err := coll.Delete(ctx, &backends.DeleteParams{IDs: ids})
	if err != nil {
		return 0, err
	}

	deleted = int32(deleteRes.Deleted)

	return deleted, nil
}
