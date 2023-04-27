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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListDatabases implements HandlerInterface.
func (h *Handler) MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var filter *types.Document
	if filter, err = commonparams.GetOptionalParam(document, "filter", filter); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "comment", "authorizedDatabases")

	databaseNames, err := dbPool.Driver.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	nameOnly, err := commonparams.GetBoolOptionalParam(document, "nameOnly")
	if err != nil {
		return nil, err
	}

	var totalSize int64
	databases := types.MakeArray(len(databaseNames))
	for _, databaseName := range databaseNames {
		res, err := dbPool.Driver.DescribeDatabase(ctx, databaseName)
		if err != nil {
			// check if database was removed between ListDatabases and DescribeDatabase calls
			if tigrisdb.IsNotFound(err) {
				continue
			}
			return nil, lazyerrors.Error(err)
		}

		totalSize += res.Size

		d := must.NotFail(types.NewDocument(
			"name", databaseName,
			"sizeOnDisk", res.Size,
			"empty", res.Size == 0,
		))

		matches, err := common.FilterDocument(d, filter)
		if err != nil {
			return nil, err
		}

		if !matches {
			continue
		}

		if nameOnly {
			d = must.NotFail(types.NewDocument(
				"name", databaseName,
			))
		}

		databases.Append(d)
	}

	if nameOnly {
		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"databases", databases,
				"ok", float64(1),
			))},
		}))

		return &reply, nil
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"databases", databases,
			"totalSize", totalSize,
			"totalSizeMb", totalSize/1024/1024,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
