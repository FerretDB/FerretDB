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

package pg

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListIndexes implements HandlerInterface.
func (h *Handler) MsgListIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/278

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "comment", "cursor")

	firstBatch := must.NotFail(types.NewArray(
		must.NotFail(types.NewDocument(
			"v", float64(2),
			"key", must.NotFail(types.NewDocument(
				"_id", float64(1),
			)),
			"name", "_id_",
		)),
	))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				// TODO "ns" field
				"id", int64(0),
				"firstBatch", firstBatch,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
