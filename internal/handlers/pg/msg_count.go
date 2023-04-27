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

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

type CountParams struct {
	DB          string          `name:"$db"`
	Collection  string          `name:"collection"`
	Collation   any             `name:"collation,unimplemented"`
	Hint        any             `name:"hint,ignored"`
	ReadConcern any             `name:"readConcern,ignored"`
	Comment     any             `name:"comment,ignored"`
	Filter      *types.Document `name:"query"`
	Skip        int64           `name:"skip"`
	Limit       int64           `name:"limit"`
}

// MsgCount implements HandlerInterface.
func (h *Handler) MsgCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := commonparams.Unimplemented(document, "collation"); err != nil {
		return nil, err
	}

	commonparams.Ignored(document, h.L, "hint", "readConcern", "comment")

	var filter *types.Document
	if filter, err = commonparams.GetOptionalParam(document, "query", filter); err != nil {
		return nil, err
	}

	var skip, limit int64

	if s, _ := document.Get("skip"); s != nil {
		if skip, err = commonparams.GetSkipParam("count", s); err != nil {
			return nil, err
		}
	}

	if l, _ := document.Get("limit"); l != nil {
		if limit, err = commonparams.GetLimitParam("count", l); err != nil {
			return nil, err
		}
	}

	var qp pgdb.QueryParams

	if qp.DB, err = commonparams.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if qp.Collection, ok = collectionParam.(string); !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidNamespace,
			fmt.Sprintf("collection name has invalid type %s", commonparams.AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	qp.Filter = filter

	var resDocs []*types.Document
	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		resDocs, err = fetchAndFilterDocs(ctx, &fetchParams{tx, &qp, h.DisableFilterPushdown})
		return err
	})

	if err != nil {
		return nil, err
	}

	if resDocs, err = common.SkipDocuments(resDocs, skip); err != nil {
		return nil, lazyerrors.Error(err)
	}

	if resDocs, err = common.LimitDocuments(resDocs, limit); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"n", int32(len(resDocs)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
