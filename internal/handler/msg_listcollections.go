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
	"fmt"

	"github.com/FerretDB/wire"
	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgListCollections implements `listCollections` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgListCollections(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := bson.OpMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var filter *types.Document
	if filter, err = common.GetOptionalParam(document, "filter", filter); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "comment")

	// TODO https://github.com/FerretDB/FerretDB/issues/3770
	common.Ignored(document, h.L, "authorizedCollections")

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	var nameOnly bool

	if v, _ := document.Get("nameOnly"); v != nil {
		if nameOnly, err = handlerparams.GetBoolOptionalParam("nameOnly", v); err != nil {
			return nil, err
		}
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s'", dbName)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "listCollections")
		}

		return nil, lazyerrors.Error(err)
	}

	res, err := db.ListCollections(connCtx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	collections := types.MakeArray(len(res.Collections))

	for _, collection := range res.Collections {
		d := must.NotFail(types.NewDocument(
			"name", collection.Name,
			"type", "collection",
			"idIndex", must.NotFail(types.NewDocument(
				"v", int32(2),
				"key", must.NotFail(types.NewDocument("_id", int32(1))),
				"name", "_id_",
			)),
		))

		options := must.NotFail(types.NewDocument())
		info := must.NotFail(types.NewDocument("readOnly", false))

		if collection.Capped() {
			options.Set("capped", true)
		}

		if collection.CappedSize > 0 {
			options.Set("size", collection.CappedSize)
		}

		if collection.CappedDocuments > 0 {
			options.Set("max", collection.CappedDocuments)
		}

		d.Set("options", options)

		if collection.UUID != "" {
			uuid, err := uuid.Parse(collection.UUID)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			uuidBinary := types.Binary{
				Subtype: types.BinaryUUID,
				B:       must.NotFail(uuid.MarshalBinary()),
			}

			info.Set("uuid", uuidBinary)
		}

		d.Set("info", info)

		matches, err := common.FilterDocument(d, filter)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if !matches {
			continue
		}

		if nameOnly {
			d = must.NotFail(types.NewDocument(
				"name", collection.Name,
			))
		}

		collections.Append(d)
	}

	return bson.NewOpMsg(
		must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"id", int64(0),
				"ns", dbName+".$cmd.listCollections",
				"firstBatch", collections,
			)),
			"ok", float64(1),
		)),
	)
}
