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

package handlers

import (
	"context"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgHello returns a document that describes the role of the instance.
func (h *Handler) MsgHello(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	var reply wire.OpMsg
	err := reply.SetSections(wire.OpMsgSection{
		// TODO merge with QueryCmd
		Documents: []*types.Document{types.MustNewDocument(
			"helloOk", true,
			"ismaster", true,
			// topologyVersion
			"maxBsonObjectSize", int32(types.MaxDocumentLen),
			"maxMessageSizeBytes", int32(wire.MaxMsgLen),
			"maxWriteBatchSize", int32(100000),
			"localTime", time.Now(),
			// logicalSessionTimeoutMinutes
			// connectionId
			"minWireVersion", int32(13),
			"maxWireVersion", int32(13),
			"readOnly", false,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
