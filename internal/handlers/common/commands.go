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

// Command represents a handler command.
type Command struct {
	// Help is shown in the help function
	Help string

	// Handler processes command
	Handler func(Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error)
}

// Commands is a map of commands that common.Handler interface can support.
// Order of entries matches the interface definition.
var Commands = map[string]Command{
	"listCommands": {
		Help:    "Returns a list of currently supported commands.",
		Handler: (Handler).MsgListCommands,
	},
	"buildInfo": {
		Help:    "Returns a summary of the build information.",
		Handler: (Handler).MsgBuildInfo,
	},
	"collStats": {
		Help:    "Returns storage data for a collection.",
		Handler: (Handler).MsgCollStats,
	},
	"create": {
		Help:    "Creates the collection.",
		Handler: (Handler).MsgCreate,
	},
	"dataSize": {
		Help:    "Returns the size of the collection in bytes.",
		Handler: (Handler).MsgDataSize,
	},
	"dbStats": {
		Help:    "Returns the statistics of the database.",
		Handler: (Handler).MsgDBStats,
	},
	"drop": {
		Help:    "Drops the collection.",
		Handler: (Handler).MsgDrop,
	},
	"dropDatabase": {
		Help:    "Deletes the database.",
		Handler: (Handler).MsgDropDatabase,
	},
	"getCmdLineOpts": {
		Help:    "Returns a summary of all runtime and configuration options.",
		Handler: (Handler).MsgGetCmdLineOpts,
	},
	"getLog": {
		Help:    "Returns the most recent logged events from memory.",
		Handler: (Handler).MsgGetLog,
	},
	"getParameter": {
		Help:    "Returns the value of the parameter.",
		Handler: (Handler).MsgGetParameter,
	},
	"hostInfo": {
		Help:    "Returns a summary of the system information.",
		Handler: (Handler).MsgHostInfo,
	},
	"ismaster": {
		Help:    "Returns the role of the FerretDB instance.",
		Handler: (Handler).MsgIsMaster,
	},
	"isMaster": { // both `ismaster` and `isMaster` are valid
		Help:    "Returns the role of the FerretDB instance.",
		Handler: (Handler).MsgIsMaster,
	},
	"hello": {
		Help:    "Returns the role of the FerretDB instance.",
		Handler: (Handler).MsgHello,
	},
	"listCollections": {
		Help:    "Returns the information of the collections and views in the database.",
		Handler: (Handler).MsgListCollections,
	},
	"listDatabases": {
		Help:    "Returns a summary of all the databases.",
		Handler: (Handler).MsgListDatabases,
	},
	"ping": {
		Help:    "Returns a pong response. Used for testing purposes.",
		Handler: (Handler).MsgPing,
	},
	"serverStatus": {
		Help:    "Returns an overview of the databases state.",
		Handler: (Handler).MsgServerStatus,
	},
	"whatsmyuri": {
		Help:    "An internal command.",
		Handler: (Handler).MsgWhatsMyURI,
	},
	"count": {
		Help:    "Returns the count of documents that's matched by the query.",
		Handler: (Handler).MsgCount,
	},
	"createIndexes": {
		Help:    "Creates indexes on a collection.",
		Handler: (Handler).MsgCreateIndexes,
	},
	"delete": {
		Help:    "Deletes documents matched by the query.",
		Handler: (Handler).MsgDelete,
	},
	"find": {
		Help:    "Returns documents matched by the query.",
		Handler: (Handler).MsgFind,
	},
	"findAndModify": {
		Help:    "Inserts, updates, or deletes, and returns a document matched by the query.",
		Handler: (Handler).MsgFindAndModify,
	},
	"insert": {
		Help:    "Inserts documents into the database.",
		Handler: (Handler).MsgInsert,
	},
	"update": {
		Help:    "Updates documents that are matched by the query.",
		Handler: (Handler).MsgUpdate,
	},
	"debug_error": {
		Help:    "Used for debugging purposes.",
		Handler: (Handler).MsgDebugError,
	},
	"debug_panic": {
		Help:    "Used for debugging purposes.",
		Handler: (Handler).MsgDebugPanic,
	},
}
