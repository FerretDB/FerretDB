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

	"github.com/FerretDB/FerretDB/internal/wire"
)

// Handler interface represents common commands handlers.
type Handler interface {
	// MsgBuildInfo returns a summary of the build information.
	MsgBuildInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCollStats returns storage data for a collection.
	MsgCollStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCreate creates a collection.
	MsgCreate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDataSize returns the size of the collection in bytes.
	MsgDataSize(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDBStats returns the statistics of the database.
	MsgDBStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDrop drops the collection.
	MsgDrop(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDropDatabase deletes the database.
	MsgDropDatabase(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgGetCmdLineOpts returns a summary of all runtime and configuration options.
	MsgGetCmdLineOpts(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgGetLog returns the most recent logged events from memory.
	MsgGetLog(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgGetParameter returns the value of the parameter.
	MsgGetParameter(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgHostInfo returns a summary of the system information.
	MsgHostInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgIsMaster returns the role of the FerretDB instance.
	MsgIsMaster(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgHello returns the role of the FerretDB instance.
	MsgHello(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgListCollections returns the information of the collections and views in the database.
	MsgListCollections(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgListDatabases returns a summary of all the databases.
	MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgPing returns a pong response. Used for testing purposes.
	MsgPing(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgServerStatus returns an overview of the databases state.
	MsgServerStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgWhatsMyURI an internal command.
	MsgWhatsMyURI(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCount returns the count of documents that's matched by the query.
	MsgCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCreateIndexes creates indexes on a collection.
	MsgCreateIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDelete deletes documents matched by the query.
	MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgFind returns documents matched by the query.
	MsgFind(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgFindAndModify inserts, updates, or deletes, and returns a document matched by the query.
	MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgInsert inserts documents into the database.
	MsgInsert(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgUpdate updates documents that are matched by the query.
	MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// CmdQuery runs query operation command.
	CmdQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error)

	// MsgConnectionStatus information about the current connection,
	// specifically the state of authenticated users and their available permissions.
	MsgConnectionStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// Close prepares handler for graceful shutdown: closes connections, channels etc.
	Close()
}
