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
		del, err := execDelete(ctx, &execDeleteParams{
			coll,
			deleteParams.Filter,
			params.Collection,
			deleteParams.Limited,
		})
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

// execDeleteParams contains parameters for execDelete function.
type execDeleteParams struct {
	coll       backends.Collection
	filter     *types.Document
	collection string
	limited    bool
}

// execDelete fetches documents, filters them out, limits them (if needed) and deletes them.
// If limit is true, only the first matched document is chosen for deletion, otherwise all matched documents are chosen.
// It returns the number of deleted documents or an error.
func execDelete(ctx context.Context, dp *execDeleteParams) (int32, error) {
	var deleted int32

	// filter is used to filter documents on the FerretDB side,
	// qp.Filter is used to filter documents on the PostgreSQL side (query pushdown).
	filter := dp.filter

	// query documents here
	res, err := dp.coll.Query(ctx, nil)
	if err != nil {
		return 0, err
	}

	iter := res.Iter

	defer iter.Close()

	ids := make([]any, 0, 16)

	for {
		var doc *types.Document

		if _, doc, err = iter.Next(); err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return 0, err
		}

		var matches bool

		if matches, err = common.FilterDocument(doc, filter); err != nil {
			return 0, err
		}

		if !matches {
			continue
		}

		ids = append(ids, must.NotFail(doc.Get("_id")))

		// if limit is set, no need to fetch all the documents
		if dp.limited {
			break
		}
	}

	// if no documents matched, there is nothing to delete
	if len(ids) == 0 {
		return 0, nil
	}

	// close iterator to free db connection.
	iter.Close()

	deleteRes, err := dp.coll.Delete(ctx, &backends.DeleteParams{IDs: ids})
	if err != nil {
		return 0, err
	}

	deleted = int32(deleteRes.Deleted)

	return deleted, nil
}
