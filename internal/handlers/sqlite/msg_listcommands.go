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

package sqlite

import (
	"context"
	"sort"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
	"golang.org/x/exp/maps"
)

// command represents a handler command.
type command struct {
	// Help is shown in the help function.
	// If empty, that command is skipped in `listCommands` output.
	Help string

	// Handler processes this command.
	//
	// The passed context is canceled when the client disconnects.
	Handler func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error)
}

// Commands maps commands names to descriptions and implementations.
var Commands = map[string]command{
	// sorted alphabetically
	"aggregate": {
		Help:    "Returns aggregated data.",
		Handler: (*Handler).MsgAggregate,
	},
	"buildInfo": {
		Help:    "Returns a summary of the build information.",
		Handler: (*Handler).MsgBuildInfo,
	},
	"buildinfo": { // old lowercase variant
		Handler: (*Handler).MsgBuildInfo,
	},
	"collMod": {
		Help:    "Adds options to a collection or modify view definitions.",
		Handler: (*Handler).MsgCollMod,
	},
	"collStats": {
		Help:    "Returns storage data for a collection.",
		Handler: (*Handler).MsgCollStats,
	},
	"compact": {
		Help:    "Reduces the disk space collection takes and refreshes its statistics.",
		Handler: (*Handler).MsgCompact,
	},
	"connectionStatus": {
		Help: "Returns information about the current connection, " +
			"specifically the state of authenticated users and their available permissions.",
		Handler: (*Handler).MsgConnectionStatus,
	},
	"count": {
		Help:    "Returns the count of documents that's matched by the query.",
		Handler: (*Handler).MsgCount,
	},
	"create": {
		Help:    "Creates the collection.",
		Handler: (*Handler).MsgCreate,
	},
	"createIndexes": {
		Help:    "Creates indexes on a collection.",
		Handler: (*Handler).MsgCreateIndexes,
	},
	"currentOp": {
		Help:    "Returns information about operations currently in progress.",
		Handler: (*Handler).MsgCurrentOp,
	},
	"dataSize": {
		Help:    "Returns the size of the collection in bytes.",
		Handler: (*Handler).MsgDataSize,
	},
	"dbStats": {
		Help:    "Returns the statistics of the database.",
		Handler: (*Handler).MsgDBStats,
	},
	"dbstats": { // old lowercase variant
		Handler: (*Handler).MsgDBStats,
	},
	"debugError": {
		Help:    "Returns error for debugging.",
		Handler: (*Handler).MsgDebugError,
	},
	"delete": {
		Help:    "Deletes documents matched by the query.",
		Handler: (*Handler).MsgDelete,
	},
	"distinct": {
		Help:    "Returns an array of distinct values for the given field.",
		Handler: (*Handler).MsgDistinct,
	},
	"drop": {
		Help:    "Drops the collection.",
		Handler: (*Handler).MsgDrop,
	},
	"dropDatabase": {
		Help:    "Drops production database.",
		Handler: (*Handler).MsgDropDatabase,
	},
	"dropIndexes": {
		Help:    "Drops indexes on a collection.",
		Handler: (*Handler).MsgDropIndexes,
	},
	"explain": {
		Help:    "Returns the execution plan.",
		Handler: (*Handler).MsgExplain,
	},
	"find": {
		Help:    "Returns documents matched by the query.",
		Handler: (*Handler).MsgFind,
	},
	"findAndModify": {
		Help:    "Docs, updates, or deletes, and returns a document matched by the query.",
		Handler: (*Handler).MsgFindAndModify,
	},
	"findandmodify": { // old lowercase variant
		Handler: (*Handler).MsgFindAndModify,
	},
	"getCmdLineOpts": {
		Help:    "Returns a summary of all runtime and configuration options.",
		Handler: (*Handler).MsgGetCmdLineOpts,
	},
	"getFreeMonitoringStatus": {
		Help:    "Returns a status of the free monitoring.",
		Handler: (*Handler).MsgGetFreeMonitoringStatus,
	},
	"getLog": {
		Help:    "Returns the most recent logged events from memory.",
		Handler: (*Handler).MsgGetLog,
	},
	"getMore": {
		Help:    "Returns the next batch of documents from a cursor.",
		Handler: (*Handler).MsgGetMore,
	},
	"getParameter": {
		Help:    "Returns the value of the parameter.",
		Handler: (*Handler).MsgGetParameter,
	},
	"hello": {
		Help:    "Returns the role of the FerretDB instance.",
		Handler: (*Handler).MsgHello,
	},
	"hostInfo": {
		Help:    "Returns a summary of the system information.",
		Handler: (*Handler).MsgHostInfo,
	},
	"insert": {
		Help:    "Docs documents into the database.",
		Handler: (*Handler).MsgInsert,
	},
	"isMaster": {
		Help:    "Returns the role of the FerretDB instance.",
		Handler: (*Handler).MsgIsMaster,
	},
	"ismaster": { // old lowercase variant
		Handler: (*Handler).MsgIsMaster,
	},
	"killCursors": {
		Help:    "Closes server cursors.",
		Handler: (*Handler).MsgKillCursors,
	},
	"listCollections": {
		Help:    "Returns the information of the collections and views in the database.",
		Handler: (*Handler).MsgListCollections,
	},
	// listCommands is added by the init() function below.
	"listDatabases": {
		Help:    "Returns a summary of all the databases.",
		Handler: (*Handler).MsgListDatabases,
	},
	"listIndexes": {
		Help:    "Returns a summary of indexes of the specified collection.",
		Handler: (*Handler).MsgListIndexes,
	},
	"logout": {
		Help:    "Logs out from the current session.",
		Handler: (*Handler).MsgLogout,
	},
	"ping": {
		Help:    "Returns a pong response.",
		Handler: (*Handler).MsgPing,
	},
	"renameCollection": {
		Help:    "Changes the name of an existing collection.",
		Handler: (*Handler).MsgRenameCollection,
	},
	"saslStart": {
		Help:    "Starts a SASL conversation.",
		Handler: (*Handler).MsgSASLStart,
	},
	"serverStatus": {
		Help:    "Returns an overview of the databases state.",
		Handler: (*Handler).MsgServerStatus,
	},
	"setFreeMonitoring": {
		Help:    "Toggles free monitoring.",
		Handler: (*Handler).MsgSetFreeMonitoring,
	},
	"update": {
		Help:    "Updates documents that are matched by the query.",
		Handler: (*Handler).MsgUpdate,
	},
	"validate": {
		Help:    "Validate collection.",
		Handler: (*Handler).MsgValidate,
	},
	"whatsmyuri": {
		Help:    "Returns peer information.",
		Handler: (*Handler).MsgWhatsMyURI,
	},
	// please keep sorted alphabetically
}

func init() {
	// to prevent the initialization cycle
	Commands["listCommands"] = command{
		Help:    "Returns a list of currently supported commands.",
		Handler: (*Handler).MsgListCommands,
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
