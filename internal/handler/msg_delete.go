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

package handler

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handler/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDelete implements `delete` command.
func (h *Handler) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := GetDeleteParams(document, h.L)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	db, err := h.b.Database(params.DB)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "delete")
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(params.Collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "delete")
		}

		return nil, lazyerrors.Error(err)
	}

	var deleted int32
	writeErrors := types.MakeArray(0)

	for i, p := range params.Deletes {
		d, err := h.execDelete(ctx, c, &p)

		deleted += d

		if err != nil {
			var ce *commonerrors.CommandError
			if errors.As(err, &ce) {
				we := &mongo.WriteError{
					Index:   i,
					Code:    int(ce.Code()),
					Message: ce.Err().Error(),
				}

				writeErrors.Append(WriteErrorDocument(we))

				if params.Ordered {
					break
				}

				continue
			}

			return nil, lazyerrors.Error(err)
		}
	}

	res := must.NotFail(types.NewDocument(
		"n", deleted,
	))

	if writeErrors.Len() > 0 {
		res.Set("writeErrors", writeErrors)
	}

	res.Set("ok", float64(1))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{res},
	}))

	return &reply, nil
}

// execDelete performs a single delete operation.
//
// It returns a number of deleted documents or error.
// The error is either a (wrapped) *commonerrors.CommandError or something fatal.
func (h *Handler) execDelete(ctx context.Context, c backends.Collection, p *Delete) (int32, error) {
	var qp backends.QueryParams
	if !h.DisableFilterPushdown {
		qp.Filter = p.Filter
	}

	q, err := c.Query(ctx, &qp)
	if err != nil {
		return 0, lazyerrors.Error(err)
	}

	var ids []any
	for {
		var doc *types.Document

		if _, doc, err = q.Iter.Next(); err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			q.Iter.Close()
			return 0, lazyerrors.Error(err)
		}

		var matches bool

		if matches, err = common.FilterDocument(doc, p.Filter); err != nil {
			q.Iter.Close()
			return 0, lazyerrors.Error(err)
		}

		if !matches {
			continue
		}

		ids = append(ids, must.NotFail(doc.Get("_id")))

		if p.Limited {
			break
		}
	}

	// close read transaction before starting write transaction
	q.Iter.Close()

	if len(ids) == 0 {
		return 0, nil
	}

	d, err := c.DeleteAll(ctx, &backends.DeleteAllParams{IDs: ids})
	if err != nil {
		return 0, lazyerrors.Error(err)
	}

	return d.Deleted, nil
}

// DeleteParams represents parameters for the delete command.
//
//nolint:vet // for readability
type DeleteParams struct {
	DB         string `ferretdb:"$db"`
	Collection string `ferretdb:"delete,collection"`

	Deletes []Delete `ferretdb:"deletes,opt"`
	Comment string   `ferretdb:"comment,opt"`
	Ordered bool     `ferretdb:"ordered,opt"`

	Let *types.Document `ferretdb:"let,unimplemented"`

	WriteConcern *types.Document `ferretdb:"writeConcern,ignored"`
	LSID         any             `ferretdb:"lsid,ignored"`
}

// Delete represents single delete operation parameters.
//
//nolint:vet // for readability
type Delete struct {
	Filter  *types.Document `ferretdb:"q"`
	Limited bool            `ferretdb:"limit,zeroOrOneAsBool"`

	Collation *types.Document `ferretdb:"collation,unimplemented"`

	Hint string `ferretdb:"hint,ignored"`
}

// GetDeleteParams returns parameters for delete operation.
func GetDeleteParams(document *types.Document, l *zap.Logger) (*DeleteParams, error) {
	params := DeleteParams{
		Ordered: true,
	}

	err := commonparams.ExtractParams(document, "delete", &params, l)
	if err != nil {
		return nil, err
	}

	return &params, nil
}
