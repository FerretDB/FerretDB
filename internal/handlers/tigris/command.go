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

package tigris

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
	"ismaster": {
		help:    "Returns the role of the FerretDB instance.",
		handler: (*Handler).MsgIsMaster,
	},
	"isMaster": { // both `ismaster` and `isMaster` are valid
		help:    "Returns the role of the FerretDB instance.",
		handler: (*Handler).MsgIsMaster,
	},
	"hello": {
		help:    "Returns the role of the FerretDB instance.",
		handler: (*Handler).MsgHello,
	},
	"create": {
		help:    "Creates the collection.",
		handler: (*Handler).MsgCreate,
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
	"drop": {
		help:    "Drops the collection.",
		handler: (*Handler).MsgDrop,
	},
	"dropDatabase": {
		help:    "Deletes the database.",
		handler: (*Handler).MsgDropDatabase,
	},
	"find": {
		help:    "Returns documents matched by the custom query.",
		handler: (*Handler).MsgFind,
	},
	"insert": {
		help:    "Inserts documents into the database.",
		handler: (*Handler).MsgInsert,
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
