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

package tigris

import (
	"context"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
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

	var db string

	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	m := document.Map()
	scale, ok := m["scale"].(float64)

	if !ok {
		scale = 1
	}

	stats, err := h.db.Driver.DescribeDatabase(ctx, db)
	switch err := err.(type) {
	case nil:
		// do nothing
	case *driver.Error:
		if !tigrisdb.IsNotFound(err) {
			return nil, lazyerrors.Error(err)
		}

		// If DB doesn't exist just return empty stats.
		stats = &driver.DescribeDatabaseResponse{
			Db:   db,
			Size: 0,
		}

	default:
		return nil, lazyerrors.Error(err)
	}

	// TODO We need a better way to get the number of documents in all collections.
	var objects int32

	for _, collection := range stats.Collections {
		querier := h.db.Driver.UseDatabase(db)

		stats, err := tigrisdb.FetchStats(ctx, querier, collection.Collection)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		objects += stats.NumObjects
	}

	var avgObjSize float64
	if objects > 0 {
		avgObjSize = float64(stats.Size) / float64(objects)
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"db", db,
			"collections", int32(len(stats.Collections)),
			// TODO https://github.com/FerretDB/FerretDB/issues/176
			"views", int32(0),
			"objects", int32(objects),
			"avgObjSize", float64(avgObjSize),
			"dataSize", float64(stats.Size),
			// Tigris indexes all the fields https://docs.tigrisdata.com/apidocs/#operation/Tigris_Read
			"indexes", int32(0),
			"indexSize", int32(0),
			"totalSize", int32(0),
			"scaleFactor", scale,
			"ok", float64(1),
		))},
	})

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
