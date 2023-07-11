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

package common

import (
	"context"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// IsMaster is a common implementation of the isMaster command used by deprecated OP_QUERY message.
func IsMaster(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	return &wire.OpReply{
		NumberReturned: 1,
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ismaster", true, // only lowercase
			// topologyVersion
			"maxBsonObjectSize", int32(types.MaxDocumentLen),
			"maxMessageSizeBytes", int32(wire.MaxMsgLen),
			"maxWriteBatchSize", int32(100000),
			"localTime", time.Now(),
			// logicalSessionTimeoutMinutes
			"connectionId", int32(42),
			"minWireVersion", MinWireVersion,
			"maxWireVersion", MaxWireVersion,
			"readOnly", false,
			"ok", float64(1),
		))},
	}, nil
}
