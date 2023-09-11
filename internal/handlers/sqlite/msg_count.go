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

// MsgCount implements HandlerInterface.
func (h *Handler) MsgCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetCountParams(document, h.L)
	if err != nil {
		return nil, err
	}

	db, err := h.b.Database(params.DB)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "count")
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(params.Collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "count")
		}

		return nil, lazyerrors.Error(err)
	}

	var qp backends.QueryParams
	if !h.DisableFilterPushdown {
		qp.Filter = params.Filter
	}

	queryRes, err := c.Query(ctx, &qp)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	iter := queryRes.Iter

	closer := iterator.NewMultiCloser(iter)
	defer closer.Close()

	iter = common.FilterIterator(iter, closer, params.Filter)

	iter = common.SkipIterator(iter, closer, params.Skip)

	iter = common.LimitIterator(iter, closer, params.Limit)

	iter = common.CountIterator(iter, closer, "count")

	_, res, err := iter.Next()
	if errors.Is(err, iterator.ErrIteratorDone) {
		err = nil
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	count, _ := res.Get("count")
	n, _ := count.(int32)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"n", n,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
