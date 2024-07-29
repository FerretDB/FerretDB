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

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgCollStats implements `collStats` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgCollStats(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := bson.OpMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := common.GetRequiredParam[string](document, command)
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

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid database specified '%s'", dbName)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", collection)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}

	collectionParam := backends.ListCollectionsParams{Name: collection}
	collections, err := db.ListCollections(connCtx, &collectionParam)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var found bool
	var cInfo backends.CollectionInfo

	found = len(collections.Collections) > 0
	if found {
		cInfo = collections.Collections[0]
	}

	indexes, err := c.ListIndexes(connCtx, new(backends.ListIndexesParams))
	if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
		indexes = new(backends.ListIndexesResult)
		err = nil
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	stats, err := c.Stats(connCtx, &backends.CollectionStatsParams{Refresh: true})
	if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
		stats = new(backends.CollectionStatsResult)
		err = nil
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	pairs := []any{
		"ns", dbName + "." + collection,
		"size", stats.SizeCollection / scale,
		"count", stats.CountDocuments,
	}

	// If there are objects in the collection, calculate the average object size.
	if stats.CountDocuments > 0 {
		pairs = append(pairs, "avgObjSize", stats.SizeCollection/stats.CountDocuments)
	}

	indexSizes := types.MakeDocument(len(stats.IndexSizes))
	for _, indexSize := range stats.IndexSizes {
		indexSizes.Set(indexSize.Name, indexSize.Size/scale)
	}

	// MongoDB uses "numbers" that could be int32 or int64,
	// FerretDB always returns int64 for simplicity.
	pairs = append(pairs,
		"storageSize", stats.SizeCollection/scale,
	)

	if found {
		pairs = append(pairs,
			"freeStorageSize", stats.SizeFreeStorage/scale,
		)
	}

	pairs = append(pairs,
		"nindexes", int64(len(indexes.Indexes)),
		"totalIndexSize", stats.SizeIndexes/scale,
		"totalSize", stats.SizeTotal/scale,
		"indexSizes", indexSizes,
		"scaleFactor", int32(scale),
		"capped", cInfo.Capped(),
	)

	if cInfo.Capped() {
		pairs = append(pairs,
			"max", cInfo.CappedDocuments,
			"maxSize", cInfo.CappedSize/scale,
		)
	}

	pairs = append(pairs,
		"ok", float64(1),
	)

	return bson.NewOpMsg(
		must.NotFail(types.NewDocument(pairs...)),
	)
}
