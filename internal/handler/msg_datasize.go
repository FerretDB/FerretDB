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
	"time"

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

// MsgDataSize implements `dataSize` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgDataSize(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := bson.OpMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = common.Unimplemented(document, "keyPattern", "min", "max"); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "estimate")

	var namespaceParam any

	if namespaceParam, err = document.Get(document.Command()); err != nil {
		return nil, err
	}

	namespace, ok := namespaceParam.(string)
	if !ok {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrTypeMismatch,
			fmt.Sprintf("collection name has invalid type %s", handlerparams.AliasFromType(namespaceParam)),
			document.Command(),
		)
	}

	dbName, cName, err := handlerparams.SplitNamespace(namespace, document.Command())
	if err != nil {
		return nil, err
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid database specified '%s'", dbName)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(cName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", cName)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}

	started := time.Now()

	stats, err := c.Stats(connCtx, new(backends.CollectionStatsParams))
	if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
		stats = new(backends.CollectionStatsResult)
		err = nil
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return bson.NewOpMsg(
		must.NotFail(types.NewDocument(
			"estimate", false,
			"size", stats.SizeTotal,
			"numObjects", stats.CountDocuments,
			"millis", int32(time.Since(started).Milliseconds()),
			"ok", float64(1),
		)),
	)
}
