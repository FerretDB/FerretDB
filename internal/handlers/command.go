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

package handlers

import (
	"context"
	"errors"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

type command struct {
	name           string
	help           string
	handler        func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error)
	storageHandler func(common.Storage, context.Context, *wire.OpMsg) (*wire.OpMsg, error)
}

var commands = map[string]command{
	"buildinfo": {
		name:    "buildInfo",
		help:    "Returns a summary of the build information.",
		handler: (*Handler).MsgBuildInfo,
	},
	"collstats": {
		// This command implements the follow database methods:
		// 	- db.collection.stats()
		// 	- db.collection.dataSize()
		name:    "collStats",
		help:    "Storage data for a collection.",
		handler: (*Handler).MsgCollStats,
	},
	"createindexes": {
		name:           "createIndexes",
		help:           "Creates indexes on a collection.",
		storageHandler: (common.Storage).MsgCreateIndexes,
	},
	"create": {
		name:    "create",
		help:    "Creates the collection.",
		handler: (*Handler).MsgCreate,
	},
	"datasize": {
		name:    "dataSize",
		help:    "Returns the size of the collection in bytes.",
		handler: (*Handler).MsgDataSize,
	},
	"dbstats": {
		name:    "dbStats",
		help:    "Returns the statistics of the database.",
		handler: (*Handler).MsgDBStats,
	},
	"drop": {
		name:    "drop",
		help:    "Drops the collection.",
		handler: (*Handler).MsgDrop,
	},
	"dropdatabase": {
		name:    "dropDatabase",
		help:    "Deletes the database.",
		handler: (*Handler).MsgDropDatabase,
	},
	"getcmdlineopts": {
		name:    "getCmdLineOpts",
		help:    "Returns a summary of all runtime and configuration options.",
		handler: (*Handler).MsgGetCmdLineOpts,
	},
	"getlog": {
		name:    "getLog",
		help:    "Returns the most recent logged events from memory.",
		handler: (*Handler).MsgGetLog,
	},
	"getparameter": {
		name:    "getParameter",
		help:    "Returns the value of the parameter.",
		handler: (*Handler).MsgGetParameter,
	},
	"hostinfo": {
		name:    "hostInfo",
		help:    "Returns a summary of the system information.",
		handler: (*Handler).MsgHostInfo,
	},
	"ismaster": {
		name:    "isMaster",
		help:    "Returns the role of the FerretDB instance.",
		handler: (*Handler).MsgHello,
	},
	"hello": {
		name:    "hello",
		help:    "Returns the role of the FerretDB instance.",
		handler: (*Handler).MsgHello,
	},
	"listcollections": {
		name:    "listCollections",
		help:    "Returns the information of the collections and views in the database.",
		handler: (*Handler).MsgListCollections,
	},
	"listdatabases": {
		name:    "listDatabases",
		help:    "Returns a summary of all the databases.",
		handler: (*Handler).MsgListDatabases,
	},
	"listcommands": {
		name: "listCommands",
		help: "Returns information about the currently supported commands.",
		// no handler - special case
	},
	"ping": {
		name:    "ping",
		help:    "Returns a pong response. Used for testing purposes.",
		handler: (*Handler).MsgPing,
	},
	"whatsmyuri": {
		name:    "whatsmyuri",
		help:    "An internal command.",
		handler: (*Handler).MsgWhatsMyURI,
	},
	"serverstatus": {
		name:    "serverStatus",
		help:    "Returns an overview of the databases state.",
		handler: (*Handler).MsgServerStatus,
	},
	"delete": {
		name:           "delete",
		help:           "Deletes documents matched by the query.",
		storageHandler: (common.Storage).MsgDelete,
	},
	"find": {
		name:           "find",
		help:           "Returns documents matched by the custom query.",
		storageHandler: (common.Storage).MsgFindOrCount,
	},
	"count": {
		name:           "count",
		help:           "Returns the count of documents that's matched by the query.",
		storageHandler: (common.Storage).MsgFindOrCount,
	},
	"insert": {
		name:           "insert",
		help:           "Inserts documents into the database. ",
		storageHandler: (common.Storage).MsgInsert,
	},
	"update": {
		name:           "update",
		help:           "Updates documents that are matched by the query.",
		storageHandler: (common.Storage).MsgUpdate,
	},
	"debug_error": {
		name: "debug_error",
		help: "Used for debugging purposes.",
		handler: func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
			return nil, errors.New("debug_error")
		},
	},
	"debug_panic": {
		name: "debug_panic",
		help: "Used for debugging purposes.",
		handler: func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
			panic("debug_panic")
		},
	},
}

// listCommands returns a list of currently supported commands.
func listCommands(context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
	var reply wire.OpMsg

	cmdList := types.MustNewDocument()
	for _, command := range commands {
		cmdList.Set(command.name, types.MustNewDocument(
			"help", command.help,
		))
	}

	err := reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"commands", cmdList,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
