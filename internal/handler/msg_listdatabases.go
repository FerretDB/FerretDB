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

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgListDatabases implements `listDatabases` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgListDatabases(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := bson.OpMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var filter *types.Document
	if filter, err = common.GetOptionalParam(document, "filter", filter); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "comment")

	// TODO https://github.com/FerretDB/FerretDB/issues/3769
	common.Ignored(document, h.L, "authorizedDatabases")

	var nameOnly bool

	if v, _ := document.Get("nameOnly"); v != nil {
		if nameOnly, err = handlerparams.GetBoolOptionalParam("nameOnly", v); err != nil {
			return nil, err
		}
	}

	res, err := h.b.ListDatabases(connCtx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var totalSize int64

	databases := types.MakeArray(len(res.Databases))

	for _, dbInfo := range res.Databases {
		db, err := h.b.Database(dbInfo.Name)
		if err != nil {
			h.L.WarnContext(connCtx, "Failed to get database", logging.Error(err))
			continue
		}

		stats, err := db.Stats(connCtx, nil)
		if err != nil {
			h.L.WarnContext(connCtx, "Failed to get database stats", logging.Error(err))
			continue
		}

		d := must.NotFail(types.NewDocument(
			"name", dbInfo.Name,
			"sizeOnDisk", stats.SizeTotal,
			"empty", stats.SizeTotal == 0,
		))

		matches, err := common.FilterDocument(d, filter)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if matches {
			if nameOnly {
				d = must.NotFail(types.NewDocument(
					"name", dbInfo.Name,
				))
			} else {
				totalSize += stats.SizeTotal
			}

			databases.Append(d)
		}
	}

	switch {
	case nameOnly:
		return bson.NewOpMsg(
			must.NotFail(types.NewDocument(
				"databases", databases,
				"ok", float64(1),
			)),
		)
	default:
		return bson.NewOpMsg(
			must.NotFail(types.NewDocument(
				"databases", databases,
				"totalSize", totalSize,
				"totalSizeMb", totalSize/1024/1024,
				"ok", float64(1),
			)),
		)
	}
}
