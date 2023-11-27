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
	"sort"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// command represents a handler command.
type command struct {
	// Handler processes this command.
	//
	// The passed context is canceled when the client disconnects.
	Handler func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error)

	// Help is shown in the help function.
	// If empty, that command is skipped in `listCommands` output.
	Help string
}

// Commands maps commands names to descriptions and implementations.
var Commands = map[string]command{
	// sorted alphabetically
	"aggregate": {
		Handler: (*Handler).MsgAggregate,
		Help:    "Returns aggregated data.",
	},
	"buildInfo": {
		Handler: (*Handler).MsgBuildInfo,
		Help:    "Returns a summary of the build information.",
	},
	"buildinfo": { // old lowercase variant
		Handler: (*Handler).MsgBuildInfo,
	},
	"collMod": {
		Handler: (*Handler).MsgCollMod,
		Help:    "Adds options to a collection or modify view definitions.",
	},
	"collStats": {
		Handler: (*Handler).MsgCollStats,
		Help:    "Returns storage data for a collection.",
	},
	"compact": {
		Handler: (*Handler).MsgCompact,
		Help:    "Reduces the disk space collection takes and refreshes its statistics.",
	},
	"connectionStatus": {
		Handler: (*Handler).MsgConnectionStatus,
		Help: "Returns information about the current connection, " +
			"specifically the state of authenticated users and their available permissions.",
	},
	"count": {
		Handler: (*Handler).MsgCount,
		Help:    "Returns the count of documents that's matched by the query.",
	},
	"create": {
		Handler: (*Handler).MsgCreate,
		Help:    "Creates the collection.",
	},
	"createIndexes": {
		Handler: (*Handler).MsgCreateIndexes,
		Help:    "Creates indexes on a collection.",
	},
	"currentOp": {
		Handler: (*Handler).MsgCurrentOp,
		Help:    "Returns information about operations currently in progress.",
	},
	"dataSize": {
		Handler: (*Handler).MsgDataSize,
		Help:    "Returns the size of the collection in bytes.",
	},
	"dbStats": {
		Handler: (*Handler).MsgDBStats,
		Help:    "Returns the statistics of the database.",
	},
	"dbstats": { // old lowercase variant
		Handler: (*Handler).MsgDBStats,
	},
	"debugError": {
		Handler: (*Handler).MsgDebugError,
		Help:    "Returns error for debugging.",
	},
	"delete": {
		Handler: (*Handler).MsgDelete,
		Help:    "Deletes documents matched by the query.",
	},
	"distinct": {
		Handler: (*Handler).MsgDistinct,
		Help:    "Returns an array of distinct values for the given field.",
	},
	"drop": {
		Handler: (*Handler).MsgDrop,
		Help:    "Drops the collection.",
	},
	"dropDatabase": {
		Handler: (*Handler).MsgDropDatabase,
		Help:    "Drops production database.",
	},
	"dropIndexes": {
		Handler: (*Handler).MsgDropIndexes,
		Help:    "Drops indexes on a collection.",
	},
	"explain": {
		Handler: (*Handler).MsgExplain,
		Help:    "Returns the execution plan.",
	},
	"find": {
		Handler: (*Handler).MsgFind,
		Help:    "Returns documents matched by the query.",
	},
	"findAndModify": {
		Handler: (*Handler).MsgFindAndModify,
		Help:    "Docs, updates, or deletes, and returns a document matched by the query.",
	},
	"findandmodify": { // old lowercase variant
		Handler: (*Handler).MsgFindAndModify,
	},
	"getCmdLineOpts": {
		Handler: (*Handler).MsgGetCmdLineOpts,
		Help:    "Returns a summary of all runtime and configuration options.",
	},
	"getFreeMonitoringStatus": {
		Handler: (*Handler).MsgGetFreeMonitoringStatus,
		Help:    "Returns a status of the free monitoring.",
	},
	"getLog": {
		Handler: (*Handler).MsgGetLog,
		Help:    "Returns the most recent logged events from memory.",
	},
	"getMore": {
		Handler: (*Handler).MsgGetMore,
		Help:    "Returns the next batch of documents from a cursor.",
	},
	"getParameter": {
		Handler: (*Handler).MsgGetParameter,
		Help:    "Returns the value of the parameter.",
	},
	"hello": {
		Handler: (*Handler).MsgHello,
		Help:    "Returns the role of the FerretDB instance.",
	},
	"hostInfo": {
		Handler: (*Handler).MsgHostInfo,
		Help:    "Returns a summary of the system information.",
	},
	"insert": {
		Handler: (*Handler).MsgInsert,
		Help:    "Docs documents into the database.",
	},
	"isMaster": {
		Handler: (*Handler).MsgIsMaster,
		Help:    "Returns the role of the FerretDB instance.",
	},
	"ismaster": { // old lowercase variant
		Handler: (*Handler).MsgIsMaster,
	},
	"killCursors": {
		Handler: (*Handler).MsgKillCursors,
		Help:    "Closes server cursors.",
	},
	"listCollections": {
		Handler: (*Handler).MsgListCollections,
		Help:    "Returns the information of the collections and views in the database.",
	},
	// listCommands is added by the init() function below.
	"listDatabases": {
		Handler: (*Handler).MsgListDatabases,
		Help:    "Returns a summary of all the databases.",
	},
	"listIndexes": {
		Handler: (*Handler).MsgListIndexes,
		Help:    "Returns a summary of indexes of the specified collection.",
	},
	"logout": {
		Handler: (*Handler).MsgLogout,
		Help:    "Logs out from the current session.",
	},
	"ping": {
		Handler: (*Handler).MsgPing,
		Help:    "Returns a pong response.",
	},
	"renameCollection": {
		Handler: (*Handler).MsgRenameCollection,
		Help:    "Changes the name of an existing collection.",
	},
	"saslStart": {
		Handler: (*Handler).MsgSASLStart,
		Help:    "Starts a SASL conversation.",
	},
	"serverStatus": {
		Handler: (*Handler).MsgServerStatus,
		Help:    "Returns an overview of the databases state.",
	},
	"setFreeMonitoring": {
		Handler: (*Handler).MsgSetFreeMonitoring,
		Help:    "Toggles free monitoring.",
	},
	"update": {
		Handler: (*Handler).MsgUpdate,
		Help:    "Updates documents that are matched by the query.",
	},
	"validate": {
		Handler: (*Handler).MsgValidate,
		Help:    "Validate collection.",
	},
	"whatsmyuri": {
		Handler: (*Handler).MsgWhatsMyURI,
		Help:    "Returns peer information.",
	},
	// please keep sorted alphabetically
}

func init() {
	// to prevent the initialization cycle
	Commands["listCommands"] = command{
		Handler: (*Handler).MsgListCommands,
		Help:    "Returns a list of currently supported commands.",
	}
}

// MsgListCommands implements `listCommands` command.
func (h *Handler) MsgListCommands(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	cmdList := must.NotFail(types.NewDocument())
	names := maps.Keys(Commands)
	sort.Strings(names)

	for _, name := range names {
		cmd := Commands[name]
		if cmd.Help == "" {
			continue
		}

		cmdList.Set(name, must.NotFail(types.NewDocument(
			"help", cmd.Help,
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
