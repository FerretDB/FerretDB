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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListDatabases command provides a list of all existing databases along with basic statistics about them.
func (h *Handler) MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = common.UnimplementedNonDefault(document, "filter", func(v any) bool {
		d, ok := v.(*types.Document)
		return ok && d.Len() == 0
	}); err != nil {
		return nil, err
	}
	common.Ignored(document, h.l, "comment", "authorizedDatabases")

	databaseNames, err := h.client.conn.ListDatabases(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: it's agreed that for the beginning a zero value can be sent.
	// after the data will be available via tigris client, it can be added to the
	// FerretDB response
	databases := types.MakeArray(len(databaseNames))
	for _, n := range databaseNames {
		sizeOnDisk := int32(0)
		d := must.NotFail(types.NewDocument(
			"name", n,
			"sizeOnDisk", sizeOnDisk,
			"empty", sizeOnDisk == 0,
		))
		if err = databases.Append(d); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"databases", databases,
			"ok", float64(1),
		))},
	})

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
