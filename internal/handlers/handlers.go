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

// Package handlers provides an interface for all handlers.
package handlers

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/wire"
)

// Interface contains methods that should be implemented by all handlers.
//
// Those methods are called to handle clients' requests sent over wire protocol.
// MsgXXX methods handle OP_MSG commands.
// CmdQuery handles a limited subset of OP_QUERY messages.
//
// Handlers are shared between all connections! Be careful when you need connection-specific information.
// Currently, we pass connection information through context, see `ConnInfo` and its usage.
//
// Please keep methods documentation in sync with commands help text in the handlers/common package.
type Interface interface {
	// Close gracefully shutdowns handler.
	Close()

	// CmdQuery queries collections for documents.
	// Used by deprecated OP_QUERY message during connection handshake with an old client.
	CmdQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error)

	// OP_MSG commands, sorted alphabetically

	// MsgAggregate returns aggregated data.
	MsgAggregate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgBuildInfo returns a summary of the build information.
	MsgBuildInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCollMod adds options to a collection or modify view definitions.
	MsgCollMod(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCollStats returns storage data for a collection.
	MsgCollStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgConnectionStatus returns information about the current connection,
	// specifically the state of authenticated users and their available permissions.
	MsgConnectionStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCount returns the count of documents that's matched by the query.
	MsgCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCreate creates the collection.
	MsgCreate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCreateIndexes creates indexes on a collection.
	MsgCreateIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgCurrentOp returns information about operations currently in progress.
	MsgCurrentOp(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDataSize returns the size of the collection in bytes.
	MsgDataSize(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDBStats returns the statistics of the database.
	MsgDBStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDebugError returns error for debugging
	MsgDebugError(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDelete deletes documents matched by the query.
	MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDistinct returns an array of distinct values for the given field.
	MsgDistinct(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDrop drops the collection.
	MsgDrop(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDropIndexes drops indexes on a collection.
	MsgDropIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgDropDatabase drops production database.
	MsgDropDatabase(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgExplain returns the execution plan.
	MsgExplain(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgFind returns documents matched by the query.
	MsgFind(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgFindAndModify inserts, updates, or deletes, and returns a document matched by the query.
	MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgGetCmdLineOpts returns a summary of all runtime and configuration options.
	MsgGetCmdLineOpts(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgGetFreeMonitoringStatus returns a status of the free monitoring.
	MsgGetFreeMonitoringStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgGetLog returns the most recent logged events from memory.
	MsgGetLog(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgGetMore returns the next batch of documents from a cursor.
	MsgGetMore(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgGetParameter returns the value of the parameter.
	MsgGetParameter(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgHello returns the role of the FerretDB instance.
	MsgHello(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgHostInfo returns a summary of the system information.
	MsgHostInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgInsert inserts documents into the database.
	MsgInsert(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgIsMaster returns the role of the FerretDB instance.
	MsgIsMaster(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgListCollections returns the information of the collections and views in the database.
	MsgListCollections(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgListCommands returns a list of supported commands.
	MsgListCommands(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgListDatabases returns a summary of all the databases.
	MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgListIndexes returns a summary of indexes of the specified collection.
	MsgListIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgPing returns a pong response.
	MsgPing(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgRenameCollection replaces collection name.
	MsgRenameCollection(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgSASLStart starts the SASL authentication process.
	MsgSASLStart(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgServerStatus returns an overview of the databases state.
	MsgServerStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgSetFreeMonitoring toggles free monitoring.
	MsgSetFreeMonitoring(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgUpdate updates documents that are matched by the query.
	MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgValidate validates collection.
	MsgValidate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// MsgWhatsMyURI returns peer information.
	MsgWhatsMyURI(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error)

	// please keep OP_MSG commands sorted alphabetically
}
