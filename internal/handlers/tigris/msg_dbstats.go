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
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDBStats implements HandlerInterface.
func (h *Handler) MsgDBStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	db, err := commonparams.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	scale := int32(1)

	var s any
	if s, err = document.Get("scale"); err == nil {
		if scale, err = common.GetScaleParam(command, s); err != nil {
			return nil, err
		}
	}

	stats, err := dbPool.Driver.DescribeDatabase(ctx, db)
	switch err := err.(type) {
	case nil:
		// do nothing
	case *driver.Error:
		if !tigrisdb.IsNotFound(err) {
			return nil, lazyerrors.Error(err)
		}

		// If DB doesn't exist just return empty stats.
		stats = new(driver.DescribeDatabaseResponse)

	default:
		return nil, lazyerrors.Error(err)
	}

	// TODO We need a better way to get the number of documents in all collections.
	var objects int32

	for _, collection := range stats.Collections {
		querier := dbPool.Driver.UseDatabase(db)

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
	must.NoError(reply.SetSections(wire.OpMsgSection{
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
			"scaleFactor", float64(scale),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
