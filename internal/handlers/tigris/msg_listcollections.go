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
	"sort"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListCollections implements HandlerInterface.
func (h *Handler) MsgListCollections(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var filter *types.Document
	if filter, err = common.GetOptionalParam(document, "filter", filter); err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/301
	// if err = common.UnimplementedNonDefault(document, "nameOnly", func(v any) bool {
	// 	nameOnly, ok := v.(bool)
	// 	return ok && !nameOnly
	// }); err != nil {
	// 	return nil, err
	// }

	common.Ignored(document, h.L, "comment", "authorizedCollections")

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	names, err := dbPool.Driver.UseDatabase(db).ListCollections(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sort.Strings(names)

	collections := types.MakeArray(len(names))
	for _, n := range names {
		d := must.NotFail(types.NewDocument(
			"name", n,
			"type", "collection",
		))

		matches, err := common.FilterDocument(d, filter)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if !matches {
			continue
		}

		collections.Append(d)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"id", int64(0),
				"ns", db+".$cmd.listCollections",
				"firstBatch", collections,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
