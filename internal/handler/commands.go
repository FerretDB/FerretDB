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

package handler

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/wire"
)

// command represents a handler for single command.
type command struct {
	// Handler processes this command.
	//
	// The passed context is canceled when the client disconnects.
	Handler func(context.Context, *wire.OpMsg) (*wire.OpMsg, error)

	// Help is shown in the `listCommands` command output.
	// If empty, that command is hidden, but still can be used.
	Help string
}

// initCommands initializes the commands map for that handler instance.
func (h *Handler) initCommands() {
	h.commands = map[string]command{
		// sorted alphabetically
		"aggregate": {
			Handler: h.MsgAggregate,
			Help:    "Returns aggregated data.",
		},
		"buildInfo": {
			Handler: h.MsgBuildInfo,
			Help:    "Returns a summary of the build information.",
		},
		"buildinfo": { // old lowercase variant
			Handler: h.MsgBuildInfo,
			Help:    "", // hidden
		},
		"collMod": {
			Handler: h.MsgCollMod,
			Help:    "Adds options to a collection or modify view definitions.",
		},
		"collStats": {
			Handler: h.MsgCollStats,
			Help:    "Returns storage data for a collection.",
		},
		"compact": {
			Handler: h.MsgCompact,
			Help:    "Reduces the disk space collection takes and refreshes its statistics.",
		},
		"connectionStatus": {
			Handler: h.MsgConnectionStatus,
			Help: "Returns information about the current connection, " +
				"specifically the state of authenticated users and their available permissions.",
		},
		"count": {
			Handler: h.MsgCount,
			Help:    "Returns the count of documents that's matched by the query.",
		},
		"create": {
			Handler: h.MsgCreate,
			Help:    "Creates the collection.",
		},
		"createIndexes": {
			Handler: h.MsgCreateIndexes,
			Help:    "Creates indexes on a collection.",
		},
		"currentOp": {
			Handler: h.MsgCurrentOp,
			Help:    "Returns information about operations currently in progress.",
		},
		"dataSize": {
			Handler: h.MsgDataSize,
			Help:    "Returns the size of the collection in bytes.",
		},
		"dbStats": {
			Handler: h.MsgDBStats,
			Help:    "Returns the statistics of the database.",
		},
		"dbstats": { // old lowercase variant
			Handler: h.MsgDBStats,
			Help:    "", // hidden
		},
		"debugError": {
			Handler: h.MsgDebugError,
			Help:    "Returns error for debugging.",
		},
		"delete": {
			Handler: h.MsgDelete,
			Help:    "Deletes documents matched by the query.",
		},
		"distinct": {
			Handler: h.MsgDistinct,
			Help:    "Returns an array of distinct values for the given field.",
		},
		"drop": {
			Handler: h.MsgDrop,
			Help:    "Drops the collection.",
		},
		"dropDatabase": {
			Handler: h.MsgDropDatabase,
			Help:    "Drops production database.",
		},
		"dropIndexes": {
			Handler: h.MsgDropIndexes,
			Help:    "Drops indexes on a collection.",
		},
		"explain": {
			Handler: h.MsgExplain,
			Help:    "Returns the execution plan.",
		},
		"find": {
			Handler: h.MsgFind,
			Help:    "Returns documents matched by the query.",
		},
		"findAndModify": {
			Handler: h.MsgFindAndModify,
			Help:    "Updates or deletes, and returns a document matched by the query.",
		},
		"findandmodify": { // old lowercase variant
			Handler: h.MsgFindAndModify,
			Help:    "", // hidden
		},
		"getCmdLineOpts": {
			Handler: h.MsgGetCmdLineOpts,
			Help:    "Returns a summary of all runtime and configuration options.",
		},
		"getFreeMonitoringStatus": {
			Handler: h.MsgGetFreeMonitoringStatus,
			Help:    "Returns a status of the free monitoring.",
		},
		"getLog": {
			Handler: h.MsgGetLog,
			Help:    "Returns the most recent logged events from memory.",
		},
		"getMore": {
			Handler: h.MsgGetMore,
			Help:    "Returns the next batch of documents from a cursor.",
		},
		"getParameter": {
			Handler: h.MsgGetParameter,
			Help:    "Returns the value of the parameter.",
		},
		"hello": {
			Handler: h.MsgHello,
			Help:    "Returns the role of the FerretDB instance.",
		},
		"hostInfo": {
			Handler: h.MsgHostInfo,
			Help:    "Returns a summary of the system information.",
		},
		"insert": {
			Handler: h.MsgInsert,
			Help:    "Inserts documents into the database.",
		},
		"isMaster": {
			Handler: h.MsgIsMaster,
			Help:    "Returns the role of the FerretDB instance.",
		},
		"ismaster": { // old lowercase variant
			Handler: h.MsgIsMaster,
			Help:    "", // hidden
		},
		"killCursors": {
			Handler: h.MsgKillCursors,
			Help:    "Closes server cursors.",
		},
		"listCollections": {
			Handler: h.MsgListCollections,
			Help:    "Returns the information of the collections and views in the database.",
		},
		"listCommands": {
			Handler: h.MsgListCommands,
			Help:    "Returns a list of currently supported commands.",
		},
		"listDatabases": {
			Handler: h.MsgListDatabases,
			Help:    "Returns a summary of all the databases.",
		},
		"listIndexes": {
			Handler: h.MsgListIndexes,
			Help:    "Returns a summary of indexes of the specified collection.",
		},
		"logout": {
			Handler: h.MsgLogout,
			Help:    "Logs out from the current session.",
		},
		"ping": {
			Handler: h.MsgPing,
			Help:    "Returns a pong response.",
		},
		"renameCollection": {
			Handler: h.MsgRenameCollection,
			Help:    "Changes the name of an existing collection.",
		},
		"saslStart": {
			Handler: h.MsgSASLStart,
			Help:    "Starts a SASL conversation.",
		},
		"serverStatus": {
			Handler: h.MsgServerStatus,
			Help:    "Returns an overview of the databases state.",
		},
		"setFreeMonitoring": {
			Handler: h.MsgSetFreeMonitoring,
			Help:    "Toggles free monitoring.",
		},
		"update": {
			Handler: h.MsgUpdate,
			Help:    "Updates documents that are matched by the query.",
		},
		"validate": {
			Handler: h.MsgValidate,
			Help:    "Validates collection.",
		},
		"whatsmyuri": {
			Handler: h.MsgWhatsMyURI,
			Help:    "Returns peer information.",
		},
		// please keep sorted alphabetically
	}

	if h.EnableNewAuth {
		// sorted alphabetically
		h.commands["createUser"] = command{
			Handler: h.MsgCreateUser,
			Help:    "Creates a new user.",
		}
		h.commands["dropAllUsersFromDatabase"] = command{
			Handler: h.MsgDropAllUsersFromDatabase,
			Help:    "Drops all user from database.",
		}
		h.commands["dropUser"] = command{
			Handler: h.MsgDropUser,
			Help:    "Drops user.",
		}
		h.commands["updateUser"] = command{
			Handler: h.MsgUpdateUser,
			Help:    "Updates user.",
		}
		h.commands["usersInfo"] = command{
			Handler: h.MsgUsersInfo,
			Help:    "Returns information about users.",
		}
		// please keep sorted alphabetically
	}
}

// Commands returns a map of enabled commands.
func (h *Handler) Commands() map[string]command {
	return h.commands
}
