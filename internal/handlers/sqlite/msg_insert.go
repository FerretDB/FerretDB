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
	"path"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/sqlite/sqlitedb"
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

	common.Ignored(document, h.L, "writeConcern", "bypassDocumentValidation", "comment")

	var qp sqlitedb.QueryParams

	if qp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	qp.DB = path.Join(h.SQLiteDBPath, qp.DB+".db")

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if qp.Collection, ok = collectionParam.(string); !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	var docs *types.Array
	if docs, err = common.GetOptionalParam(document, "documents", docs); err != nil {
		return nil, err
	}

	ordered := true
	if ordered, err = common.GetOptionalParam(document, "ordered", ordered); err != nil {
		return nil, err
	}

	inserted, insErrors := insertMany(ctx, &qp, docs, ordered)

	replyDoc := must.NotFail(types.NewDocument(
		"ok", float64(1),
	))

	if insErrors.Len() > 0 {
		replyDoc = insErrors.Document()
	}

	replyDoc.Set("n", inserted)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	}))

	return &reply, nil
}

func insertMany(ctx context.Context, qp *sqlitedb.QueryParams, docs *types.Array, ordered bool) (int32, *common.WriteErrors) { //nolint:lll // argument list is too long
	var inserted int32
	var insErrors commonerrors.WriteErrors

	for i := 0; i < docs.Len(); i++ {
		doc := must.NotFail(docs.Get(i))

		err := insertDocument(ctx, qp, doc)

		var we *commonerrors.WriteErrors

		switch {
		case err == nil:
			inserted++
			continue
		case errors.As(err, &we):
			insErrors.Merge(we, int32(i))
		default:
			insErrors.Append(err, int32(i))
		}

		if ordered {
			return inserted, &insErrors
		}
	}

	return inserted, &insErrors
}

// insertDocument prepares and executes actual INSERT request to SQLite.
func insertDocument(ctx context.Context, qp *sqlitedb.QueryParams, doc any) error {
	d, ok := doc.(*types.Document)
	if !ok {
		return commonerrors.NewCommandErrorMsg(
			commonerrors.ErrBadValue,
			fmt.Sprintf("document has invalid type %s", common.AliasFromType(doc)),
		)
	}

	err := sqlitedb.InsertDocument(ctx, qp.DB, qp.Collection, d)

	switch {
	case err == nil:
		return nil

	default:
		return commonerrors.CheckError(err)
	}
}
