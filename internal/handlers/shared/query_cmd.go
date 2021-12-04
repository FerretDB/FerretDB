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

package shared

import (
	"context"
	"time"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/wire"
)

func (h *Handler) QueryCmd(ctx context.Context, msg *wire.OpQuery) (*wire.OpReply, error) {
	switch cmd := msg.Query.Command(); cmd {
	case "ismaster":
		// TODO merge with handleOpMsgIsMaster
		reply := &wire.OpReply{
			NumberReturned: 1,
			Documents: []types.Document{
				types.MustMakeDocument(
					"ismaster", true,
					// topologyVersion
					"maxBsonObjectSize", int32(bson.MaxDocumentLen),
					"maxMessageSizeBytes", int32(wire.MaxMsgLen),
					"maxWriteBatchSize", int32(100000),
					"localTime", time.Now(),
					// logicalSessionTimeoutMinutes
					// connectionId
					"minWireVersion", int32(13),
					"maxWireVersion", int32(13),
					"readOnly", false,
					"ok", float64(1),
				),
			},
		}
		return reply, nil

	default:
		return nil, common.NewErrorMessage(common.ErrNotImplemented, "QueryCmd: unhandled command %q", cmd)
	}
}
