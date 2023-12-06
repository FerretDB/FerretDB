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

	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListCollections implements `listCollections` command.
func (h *Handler) MsgListCollections(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
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

	res, err := db.ListCollections(ctx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	collections := types.MakeArray(len(res.Collections))

	for _, collection := range res.Collections {
		options := []any{"capped", collection.Capped()}
		info := []any{"readOnly", false}

		if collection.CappedSize > 0 {
			options = append(options, "size", collection.CappedSize)
		}
		if collection.CappedDocuments > 0 {
			options = append(options, "max", collection.CappedDocuments)
		}

		// Add indices to the collection info.
		// TODO(henvic): fix getting the indices. Also, check for performance impact.
		col, err := db.Collection(collection.Name)
		if err != nil {
			return nil, fmt.Errorf("cannot get collection %s: %w", collection.UUID, err)
		}
		indexes, err := col.ListIndexes(ctx, &backends.ListIndexesParams{})
		if err != nil {
			return nil, fmt.Errorf("cannot get indexes for collection %s: %w", collection.UUID, err)
		}
		idIndex := types.MakeArray(len(indexes.Indexes))
		for _, index := range indexes.Indexes {
			keys := types.MakeArray(len(index.Key))
			for ki, k := range index.Key {
				keys.Append(k.Field, ki)
			}
			idIndex.Append(must.NotFail(types.NewDocument(
				"v", 2,
				"key", keys,
				"name", index.Name,
			)))
		}

		if collection.UUID != "" {
			uuid, err := uuid.Parse(collection.UUID)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			uuidBinary := types.Binary{
				Subtype: types.BinaryUUID,
				B:       must.NotFail(uuid.MarshalBinary()),
			}

			info = append(info, "uuid", uuidBinary)
		}

		d := must.NotFail(types.NewDocument(
			"name", collection.Name,
			"type", "collection",
			"options", must.NotFail(types.NewDocument(options...)),
			"info", must.NotFail(types.NewDocument(info...)),
			"idIndex", idIndex,
		))

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

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"id", int64(0),
				"ns", dbName+".$cmd.listCollections",
				"firstBatch", collections,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
