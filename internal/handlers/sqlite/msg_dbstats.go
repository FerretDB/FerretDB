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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDBStats implements HandlerInterface.
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
		if scale, err = commonparams.GetValidatedNumberParamWithMinValue(command, "scale", s, 1); err != nil {
			return nil, err
		}
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid database specified '%s'", dbName)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}
	defer db.Close()

	stats, err := db.Stats(ctx, new(backends.DatabaseStatsParams))
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
		"collections", stats.CountCollections,
		// TODO https://github.com/FerretDB/FerretDB/issues/176
		"views", int32(0),
		"objects", stats.CountObjects,
	}

	if stats.CountObjects > 0 {
		pairs = append(pairs, "avgObjSize", stats.SizeCollections/stats.CountObjects)
	}

	pairs = append(pairs,
		"dataSize", stats.SizeCollections/scale,
		"storageSize", stats.SizeCollections/scale,
		"indexes", stats.CountIndexes,
		"indexSize", stats.SizeIndexes/scale,
		"totalSize", stats.SizeTotal/scale,
		"scaleFactor", float64(scale),
		"ok", float64(1),
	)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(pairs...))},
	}))

	return &reply, nil
}
