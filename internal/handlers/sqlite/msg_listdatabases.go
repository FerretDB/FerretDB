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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListDatabases implements HandlerInterface.
func (h *Handler) MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var filter *types.Document
	if filter, err = common.GetOptionalParam(document, "filter", filter); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "comment", "authorizedDatabases")

	var nameOnly bool

	if v, _ := document.Get("nameOnly"); v != nil {
		if nameOnly, err = commonparams.GetBoolOptionalParam("nameOnly", v); err != nil {
			return nil, err
		}
	}

	res, err := h.b.ListDatabases(ctx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var totalSize int64

	databases := types.MakeArray(len(res.Databases))

	for _, db := range res.Databases {
		d := must.NotFail(types.NewDocument(
			"name", db.Name,
			"sizeOnDisk", db.Size,
			"empty", db.Size == 0,
		))

		totalSize += db.Size

		matches, err := common.FilterDocument(d, filter)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if !matches {
			continue
		}

		if nameOnly {
			d = must.NotFail(types.NewDocument(
				"name", db.Name,
			))
		}

		databases.Append(d)
	}

	var reply wire.OpMsg

	switch {
	case nameOnly:
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"databases", databases,
				"ok", float64(1),
			))},
		}))
	default:
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"databases", databases,
				"totalSize", totalSize,
				"totalSizeMb", totalSize/1024/1024,
				"ok", float64(1),
			))},
		}))
	}

	return &reply, nil
}
