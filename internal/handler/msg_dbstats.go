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

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDBStats implements `dbStats` command.
func (h *Handler) MsgDBStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	scale := int64(1)

	var s any
	if s, err = document.Get("scale"); err == nil {
		if scale, err = handlerparams.GetValidatedNumberParamWithMinValue(command, "scale", s, 1); err != nil {
			return nil, err
		}
	}

	var freeStorage bool

	if v, _ := document.Get("freeStorage"); v != nil {
		if freeStorage, err = handlerparams.GetBoolOptionalParam("freeStorage", v); err != nil {
			return nil, err
		}
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid database specified '%s'", dbName)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}

	list, err := db.ListCollections(ctx, new(backends.ListCollectionsParams))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var nIndexes int64

	for _, cInfo := range list.Collections {
		var coll backends.Collection

		coll, err = db.Collection(cInfo.Name)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		var iList *backends.ListIndexesResult

		iList, err = coll.ListIndexes(ctx, new(backends.ListIndexesParams))
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
			iList = new(backends.ListIndexesResult)
			err = nil
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		nIndexes += int64(len(iList.Indexes))
	}

	stats, err := db.Stats(ctx, &backends.DatabaseStatsParams{Refresh: true})
	if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseDoesNotExist) {
		stats = new(backends.DatabaseStatsResult)
		err = nil
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// MongoDB uses "numbers" that could be int32 or int64,
	// FerretDB always returns int64 for simplicity.
	pairs := []any{
		"db", dbName,
		"collections", int64(len(list.Collections)),
		// TODO https://github.com/FerretDB/FerretDB/issues/176
		"views", int32(0),
		"objects", stats.CountDocuments,
	}

	if stats.CountDocuments > 0 {
		pairs = append(pairs, "avgObjSize", stats.SizeCollections/stats.CountDocuments)
	}

	pairs = append(pairs,
		"dataSize", stats.SizeCollections/scale,
		"storageSize", stats.SizeCollections/scale,
	)

	if freeStorage {
		pairs = append(pairs,
			"freeStorageSize", stats.SizeFreeStorage/scale,
		)
	}

	pairs = append(pairs,
		"indexes", nIndexes,
		"indexSize", stats.SizeIndexes/scale,
	)

	// add indexFreeStorageSize
	// TODO https://github.com/FerretDB/FerretDB/issues/2447

	pairs = append(pairs,
		"totalSize", stats.SizeTotal/scale,
	)

	if freeStorage {
		pairs = append(pairs,
			"totalFreeStorageSize", (stats.SizeFreeStorage)/scale,
		)
	}

	pairs = append(pairs,
		"scaleFactor", float64(scale),
		"ok", float64(1),
	)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(pairs...))},
	}))

	return &reply, nil
}
