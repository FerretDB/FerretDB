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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListDatabases implements HandlerInterface.
func (h *Handler) MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	response, err := h.b.ListDatabases(ctx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	databases := types.MakeArray(len(response.Databases))

	for _, db := range response.Databases {
		databases.Append(types.NewDocument(
			"name", db.Name,
			"sizeOnDisk", int64(0),
			"empty", false,
		))
	}

	var reply wire.OpMsg

	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"databases", databases,
			"totalSize", int64(0),
			"totalSizeMb", int64(0),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
