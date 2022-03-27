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

package pg

import (
	"context"
	"errors"
	"sort"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

type command struct {
	help    string
	handler func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error)
}

var commands = map[string]command{
	"listCommands": {
		help: "Returns information about the currently supported commands.",
		// no handler - special case
	},

	"buildInfo": {
		help:    "Returns a summary of the build information.",
		handler: (*Handler).MsgBuildInfo,
	},
	"collStats": {
		help:    "Storage data for a collection.",
		handler: (*Handler).MsgCollStats,
	},
	"create": {
		help:    "Creates the collection.",
		handler: (*Handler).MsgCreate,
	},
	"dataSize": {
		help:    "Returns the size of the collection in bytes.",
		handler: (*Handler).MsgDataSize,
	},
	"dbStats": {
		help:    "Returns the statistics of the database.",
		handler: (*Handler).MsgDBStats,
	},
	"drop": {
		help:    "Drops the collection.",
		handler: (*Handler).MsgDrop,
	},
	"dropDatabase": {
		help:    "Deletes the database.",
		handler: (*Handler).MsgDropDatabase,
	},
	"getCmdLineOpts": {
		help:    "Returns a summary of all runtime and configuration options.",
		handler: (*Handler).MsgGetCmdLineOpts,
	},
	"getLog": {
		help:    "Returns the most recent logged events from memory.",
		handler: (*Handler).MsgGetLog,
	},
	"getParameter": {
		help:    "Returns the value of the parameter.",
		handler: (*Handler).MsgGetParameter,
	},
	"hostInfo": {
		help:    "Returns a summary of the system information.",
		handler: (*Handler).MsgHostInfo,
	},
	"ismaster": {
		help:    "Returns the role of the FerretDB instance.",
		handler: (*Handler).MsgHello,
	},
	"isMaster": { // both `ismaster` and `isMaster` are valid
		help:    "Returns the role of the FerretDB instance.",
		handler: (*Handler).MsgHello,
	},
	"hello": {
		help:    "Returns the role of the FerretDB instance.",
		handler: (*Handler).MsgHello,
	},
	"listCollections": {
		help:    "Returns the information of the collections and views in the database.",
		handler: (*Handler).MsgListCollections,
	},
	"listDatabases": {
		help:    "Returns a summary of all the databases.",
		handler: (*Handler).MsgListDatabases,
	},
	"ping": {
		help:    "Returns a pong response. Used for testing purposes.",
		handler: (*Handler).MsgPing,
	},
	"serverStatus": {
		help:    "Returns an overview of the databases state.",
		handler: (*Handler).MsgServerStatus,
	},
	"whatsmyuri": {
		help:    "An internal command.",
		handler: (*Handler).MsgWhatsMyURI,
	},

	"count": {
		help:    "Returns the count of documents that's matched by the query.",
		handler: (*Handler).MsgCount,
	},
	"createIndexes": {
		help:    "Creates indexes on a collection.",
		handler: (*Handler).MsgCreateIndexes,
	},
	"delete": {
		help:    "Deletes documents matched by the query.",
		handler: (*Handler).MsgDelete,
	},
	"find": {
		help:    "Returns documents matched by the custom query.",
		handler: (*Handler).MsgFind,
	},
	"insert": {
		help:    "Inserts documents into the database.",
		handler: (*Handler).MsgInsert,
	},
	"update": {
		help:    "Updates documents that are matched by the query.",
		handler: (*Handler).MsgUpdate,
	},

	// internal commands
	"debug_error": {
		help: "Used for debugging purposes.",
		handler: func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
			return nil, errors.New("debug_error")
		},
	},
	"debug_panic": {
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
	names := maps.Keys(commands)
	sort.Strings(names)
	for _, name := range names {
		cmdList.Set(name, types.MustNewDocument(
			"help", commands[name].help,
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
