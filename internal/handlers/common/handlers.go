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

// Handler interface  commands handlers.
package common

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/wire"
)

type Handler interface {
	MsgBuildInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgCollStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgCreate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgDataSize(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgDBStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgDrop(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgDropDatabase(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgGetCmdLineOpts(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgGetLog(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgGetParameter(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgHostInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgIsMaster(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgHello(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgListCollections(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgPing(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgServerStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgWhatsMyURI(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgCreateIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgFind(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgInsert(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	MsgDebugError(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)
	MsgDebugPanic(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	MsgQueryCmd(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error)
}
