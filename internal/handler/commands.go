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

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// commandHandler represents a function/method that processes a single request.
//
// The passed context is canceled when the client disconnects.
//
// Response is a normal response or an error.
// TODO https://github.com/FerretDB/FerretDB/issues/4965
type commandHandler func(context.Context, *middleware.Request) (*middleware.Response, error)

// command represents a handler for single command.
type command struct {
	// anonymous indicates that the command does not require authentication.
	anonymous bool

	// handler processes this command.
	//
	// The passed context is canceled when the client disconnects.
	handler commandHandler

	// Help is shown in the `listCommands` command output.
	// If empty, that command is hidden, but still can be used.
	Help string
}

// initCommands initializes the commands map for that handler instance.
func (h *Handler) initCommands() {
	h.commands = map[string]*command{
		// sorted alphabetically
		"aggregate": {
			handler: h.msgAggregate,
			Help:    "Returns aggregated data.",
		},
		"authenticate": {
			// TODO https://github.com/FerretDB/FerretDB/issues/1731
			anonymous: true,
			Help:      "", // hidden while not implemented
		},
		"buildInfo": {
			handler:   h.msgBuildInfo,
			anonymous: true,
			Help:      "Returns a summary of the build information.",
		},
		"buildinfo": { // old lowercase variant
			handler:   h.msgBuildInfo,
			anonymous: true,
			Help:      "", // hidden
		},
		"bulkWrite": {
			// TODO https://github.com/documentdb/documentdb/issues/108
			// TODO https://github.com/FerretDB/FerretDB/issues/4910
			Help: "", // hidden while not implemented
		},
		"collMod": {
			handler: h.msgCollMod,
			Help:    "Adds options to a collection or modify view definitions.",
		},
		"collStats": {
			handler: h.msgCollStats,
			Help:    "Returns storage data for a collection.",
		},
		"compact": {
			handler: h.msgCompact,
			Help:    "Reduces the disk space collection takes and refreshes its statistics.",
		},
		"connPoolStats": {
			// TODO https://github.com/FerretDB/FerretDB/issues/4909
			anonymous: true,
			Help:      "", // hidden while not implemented
		},
		"connectionStatus": {
			handler:   h.msgConnectionStatus,
			anonymous: true,
			Help: "Returns information about the current connection, " +
				"specifically the state of authenticated users and their available permissions.",
		},
		"count": {
			handler: h.msgCount,
			Help:    "Returns the count of documents that's matched by the query.",
		},
		"create": {
			handler: h.msgCreate,
			Help:    "Creates the collection.",
		},
		"createIndexes": {
			handler: h.msgCreateIndexes,
			Help:    "Creates indexes on a collection.",
		},
		"createUser": {
			handler: h.msgCreateUser,
			Help:    "Creates a new user.",
		},
		"currentOp": {
			handler: h.msgCurrentOp,
			Help:    "Returns information about operations currently in progress.",
		},
		"dataSize": {
			handler: h.msgDataSize,
			Help:    "Returns the size of the collection in bytes.",
		},
		"dbStats": {
			handler: h.msgDBStats,
			Help:    "Returns the statistics of the database.",
		},
		"dbstats": { // old lowercase variant
			handler: h.msgDBStats,
			Help:    "", // hidden
		},
		"delete": {
			handler: h.msgDelete,
			Help:    "Deletes documents matched by the query.",
		},
		"distinct": {
			handler: h.msgDistinct,
			Help:    "Returns an array of distinct values for the given field.",
		},
		"drop": {
			handler: h.msgDrop,
			Help:    "Drops the collection.",
		},
		"dropAllUsersFromDatabase": {
			handler: h.msgDropAllUsersFromDatabase,
			Help:    "Drops all user from database.",
		},
		"dropDatabase": {
			handler: h.msgDropDatabase,
			Help:    "Drops production database.",
		},
		"dropIndexes": {
			handler: h.msgDropIndexes,
			Help:    "Drops indexes on a collection.",
		},
		"dropUser": {
			handler: h.msgDropUser,
			Help:    "Drops user.",
		},
		"endSessions": {
			handler: h.msgEndSessions,
			Help:    "Marks sessions as expired.",
		},
		"explain": {
			handler: h.msgExplain,
			Help:    "Returns the execution plan.",
		},
		"ferretDebugError": {
			handler: h.msgFerretDebugError,
			Help:    "Returns error for debugging.",
		},
		"find": {
			handler: h.msgFind,
			Help:    "Returns documents matched by the query.",
		},
		"findAndModify": {
			handler: h.msgFindAndModify,
			Help:    "Updates or deletes, and returns a document matched by the query.",
		},
		"findandmodify": { // old lowercase variant
			handler: h.msgFindAndModify,
			Help:    "", // hidden
		},
		"getCmdLineOpts": {
			handler: h.msgGetCmdLineOpts,
			Help:    "Returns a summary of all runtime and configuration options.",
		},
		"getFreeMonitoringStatus": {
			handler: h.msgGetFreeMonitoringStatus,
			Help:    "Returns a status of the free monitoring.",
		},
		"getLog": {
			handler: h.msgGetLog,
			Help:    "Returns the most recent logged events from memory.",
		},
		"getMore": {
			handler: h.msgGetMore,
			Help:    "Returns the next batch of documents from a cursor.",
		},
		"getParameter": {
			handler: h.msgGetParameter,
			Help:    "Returns the value of the parameter.",
		},
		"hello": {
			handler:   h.msgHello,
			anonymous: true,
			Help:      "Returns the role of the FerretDB instance.",
		},
		"hostInfo": {
			handler: h.msgHostInfo,
			Help:    "Returns a summary of the system information.",
		},
		"insert": {
			handler: h.msgInsert,
			Help:    "Inserts documents into the database.",
		},
		"isMaster": {
			handler:   h.msgIsMaster,
			anonymous: true,
			Help:      "Returns the role of the FerretDB instance.",
		},
		"ismaster": { // old lowercase variant
			handler:   h.msgIsMaster,
			anonymous: true,
			Help:      "", // hidden
		},
		"killAllSessions": {
			handler: h.msgKillAllSessions,
			Help:    "Kills all sessions.",
		},
		"killAllSessionsByPattern": {
			handler: h.msgKillAllSessionsByPattern,
			Help:    "Kills all sessions that match the pattern.",
		},
		"killCursors": {
			handler: h.msgKillCursors,
			Help:    "Closes server cursors.",
		},
		"killSessions": {
			handler: h.msgKillSessions,
			Help:    "Kills sessions.",
		},
		"listCollections": {
			handler: h.msgListCollections,
			Help:    "Returns the information of the collections and views in the database.",
		},
		"listCommands": {
			handler: h.msgListCommands,
			Help:    "Returns a list of currently supported commands.",
		},
		"listDatabases": {
			handler: h.msgListDatabases,
			Help:    "Returns a summary of all databases.",
		},
		"listIndexes": {
			handler: h.msgListIndexes,
			Help:    "Returns a summary of indexes of the specified collection.",
		},
		"logout": {
			handler:   h.msgLogout,
			anonymous: true,
			Help:      "Logs out from the current session.",
		},
		"ping": {
			handler:   h.msgPing,
			anonymous: true,
			Help:      "Returns a pong response.",
		},
		"refreshSessions": {
			handler: h.msgRefreshSessions,
			Help:    "Updates the last used time of sessions.",
		},
		"reIndex": {
			handler: h.msgReIndex,
			Help:    "Drops and recreates all indexes except default _id index of a collection.",
		},
		"renameCollection": {
			handler: h.msgRenameCollection,
			Help:    "Changes the name of an existing collection.",
		},
		"saslStart": {
			handler:   h.msgSASLStart,
			anonymous: true,
			Help:      "", // hidden
		},
		"saslContinue": {
			handler:   h.msgSASLContinue,
			anonymous: true,
			Help:      "", // hidden
		},
		"serverStatus": {
			handler: h.msgServerStatus,
			Help:    "Returns an overview of the databases state.",
		},
		"setFreeMonitoring": {
			handler: h.msgSetFreeMonitoring,
			Help:    "Toggles free monitoring.",
		},
		"startSession": {
			handler: h.msgStartSession,
			Help:    "Returns a session.",
		},
		"update": {
			handler: h.msgUpdate,
			Help:    "Updates documents that are matched by the query.",
		},
		"updateUser": {
			handler: h.msgUpdateUser,
			Help:    "Updates user.",
		},
		"usersInfo": {
			handler: h.msgUsersInfo,
			Help:    "Returns information about users.",
		},
		"validate": {
			handler: h.msgValidate,
			Help:    "Validates collection.",
		},
		"whatsmyuri": {
			handler:   h.msgWhatsMyURI,
			anonymous: true,
			Help:      "Returns peer information.",
		},
		// please keep sorted alphabetically
	}
}
