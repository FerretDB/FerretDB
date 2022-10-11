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
	"sort"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// command represents a handler command.
type command struct {
	// Help is shown in the help function
	Help string

	// Handler processes command
	Handler func(handlers.Interface, context.Context, *wire.OpMsg) (*wire.OpMsg, error)
}

// Commands is a map of Commands that Handler interface can support.
// Order of entries matches the interface definition.
//
// Please keep help text in sync with handlers.Interface methods documentation.
var Commands = map[string]command{
	// sorted alphabetically
	"aggregate": {
		Help:    "Returns aggregated data.",
		Handler: (handlers.Interface).MsgAggregate,
	},
	"buildinfo": {
		Help:    "Returns a summary of the build information.",
		Handler: (handlers.Interface).MsgBuildInfo,
	},
	"buildInfo": { // both `buildinfo` and `buildInfo` are valid
		Help:    "Returns a summary of the build information.",
		Handler: (handlers.Interface).MsgBuildInfo,
	},
	"collMod": {
		Help:    "Adds options to a collection or modify view definitions.",
		Handler: (handlers.Interface).MsgCollMod,
	},
	"collStats": {
		Help:    "Returns storage data for a collection.",
		Handler: (handlers.Interface).MsgCollStats,
	},
	"connectionStatus": {
		Help: "Returns information about the current connection, " +
			"specifically the state of authenticated users and their available permissions.",
		Handler: (handlers.Interface).MsgConnectionStatus,
	},
	"count": {
		Help:    "Returns the count of documents that's matched by the query.",
		Handler: (handlers.Interface).MsgCount,
	},
	"create": {
		Help:    "Creates the collection.",
		Handler: (handlers.Interface).MsgCreate,
	},
	"createIndexes": {
		Help:    "Creates indexes on a collection.",
		Handler: (handlers.Interface).MsgCreateIndexes,
	},
	"dataSize": {
		Help:    "Returns the size of the collection in bytes.",
		Handler: (handlers.Interface).MsgDataSize,
	},
	"dbStats": {
		Help:    "Returns the statistics of the database.",
		Handler: (handlers.Interface).MsgDBStats,
	},
	"debugError": {
		Help:    "Returns error for debugging.",
		Handler: (handlers.Interface).MsgDebugError,
	},
	"delete": {
		Help:    "Deletes documents matched by the query.",
		Handler: (handlers.Interface).MsgDelete,
	},
	"drop": {
		Help:    "Drops the collection.",
		Handler: (handlers.Interface).MsgDrop,
	},
	"dropDatabase": {
		Help:    "Drops production database.",
		Handler: (handlers.Interface).MsgDropDatabase,
	},
	"explain": {
		Help:    "Returns the execution plan.",
		Handler: (handlers.Interface).MsgExplain,
	},
	"find": {
		Help:    "Returns documents matched by the query.",
		Handler: (handlers.Interface).MsgFind,
	},
	"findAndModify": {
		Help:    "Inserts, updates, or deletes, and returns a document matched by the query.",
		Handler: (handlers.Interface).MsgFindAndModify,
	},
	"getCmdLineOpts": {
		Help:    "Returns a summary of all runtime and configuration options.",
		Handler: (handlers.Interface).MsgGetCmdLineOpts,
	},
	"getFreeMonitoringStatus": {
		Help:    "Returns a status of the free monitoring.",
		Handler: (handlers.Interface).MsgGetFreeMonitoringStatus,
	},
	"getLog": {
		Help:    "Returns the most recent logged events from memory.",
		Handler: (handlers.Interface).MsgGetLog,
	},
	"getParameter": {
		Help:    "Returns the value of the parameter.",
		Handler: (handlers.Interface).MsgGetParameter,
	},
	"hello": {
		Help:    "Returns the role of the FerretDB instance.",
		Handler: (handlers.Interface).MsgHello,
	},
	"hostInfo": {
		Help:    "Returns a summary of the system information.",
		Handler: (handlers.Interface).MsgHostInfo,
	},
	"insert": {
		Help:    "Inserts documents into the database.",
		Handler: (handlers.Interface).MsgInsert,
	},
	"ismaster": {
		Help:    "Returns the role of the FerretDB instance.",
		Handler: (handlers.Interface).MsgIsMaster,
	},
	"isMaster": { // both `ismaster` and `isMaster` are valid
		Help:    "Returns the role of the FerretDB instance.",
		Handler: (handlers.Interface).MsgIsMaster,
	},
	"listCollections": {
		Help:    "Returns the information of the collections and views in the database.",
		Handler: (handlers.Interface).MsgListCollections,
	},
	"listCommands": {
		Help:    "Returns a list of currently supported commands.",
		Handler: (handlers.Interface).MsgListCommands,
	},
	"listDatabases": {
		Help:    "Returns a summary of all the databases.",
		Handler: (handlers.Interface).MsgListDatabases,
	},
	"ping": {
		Help:    "Returns a pong response.",
		Handler: (handlers.Interface).MsgPing,
	},
	"serverStatus": {
		Help:    "Returns an overview of the databases state.",
		Handler: (handlers.Interface).MsgServerStatus,
	},
	"setFreeMonitoring": {
		Help:    "Toggles free monitoring.",
		Handler: (handlers.Interface).MsgSetFreeMonitoring,
	},
	"update": {
		Help:    "Updates documents that are matched by the query.",
		Handler: (handlers.Interface).MsgUpdate,
	},
	"whatsmyuri": {
		Help:    "Returns peer information.",
		Handler: (handlers.Interface).MsgWhatsMyURI,
	},
}

// MsgListCommands is a common implementation of the listCommands command.
func MsgListCommands(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	cmdList := must.NotFail(types.NewDocument())
	names := maps.Keys(Commands)
	sort.Strings(names)
	for _, name := range names {
		cmdList.Set(name, must.NotFail(types.NewDocument(
			"help", Commands[name].Help,
		)))
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"commands", cmdList,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
