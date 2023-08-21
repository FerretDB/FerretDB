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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
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
		return nil, lazyerrors.Error(err)
	}

	db, err := h.b.Database(params.DB)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "insert")
		}

		return nil, lazyerrors.Error(err)
	}
	defer db.Close()

	c, err := db.Collection(params.Collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "insert")
		}

		return nil, lazyerrors.Error(err)
	}

	closer := iterator.NewMultiCloser()
	defer closer.Close()

	allDocs := make([]any, 0, params.Docs.Len())

	allDocsIter := params.Docs.Iterator()
	closer.Add(allDocsIter)

	for {
		_, d, err := allDocsIter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		doc := d.(*types.Document)

		if !doc.Has("_id") {
			doc.Set("_id", types.NewObjectID())
		}

		if err = doc.ValidateData(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		allDocs = append(allDocs, doc)
	}

	insertIter := iterator.ForSlice(allDocs)
	closer.Add(insertIter)

	res, err := c.InsertAll(ctx, &backends.InsertAllParams{
		Iter: insertIter,
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
