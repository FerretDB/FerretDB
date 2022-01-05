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
		name:    "buildinfo",
		help:    "",
		handler: (*Handler).MsgBuildInfo,
	},
	"collstats": {
		name:    "collstats",
		help:    "",
		handler: (*Handler).MsgCollStats,
	},
	"createindexes": {
		name:           "createindexes",
		help:           "",
		storageHandler: (common.Storage).MsgCreateIndexes,
	},
	"create": {
		name:    "create",
		help:    "",
		handler: (*Handler).MsgCreate,
	},
	"drop": {
		name:    "drop",
		help:    "",
		handler: (*Handler).MsgDrop,
	},
	"dropdatabase": {
		name:    "dropdatabase",
		help:    "",
		handler: (*Handler).MsgDropDatabase,
	},
	"getcmdlineopts": {
		name:    "getcmdlineopts",
		help:    "",
		handler: (*Handler).MsgGetCmdLineOpts,
	},
	"getlog": {
		name:    "getlog",
		help:    "",
		handler: (*Handler).MsgGetLog,
	},
	"getparameter": {
		name:    "getparameter",
		help:    "",
		handler: (*Handler).MsgGetParameter,
	},
	"hostinfo": {
		name:    "hostInfo",
		help:    "",
		handler: (*Handler).MsgHostInfo,
	},
	"ismaster": {
		name:    "ismaster",
		help:    "",
		handler: (*Handler).MsgHello,
	},
	"hello": {
		name:    "hello",
		help:    "",
		handler: (*Handler).MsgHello,
	},
	"listcollections": {
		name:    "listcollections",
		help:    "",
		handler: (*Handler).MsgListCollections,
	},
	"listdatabases": {
		name:    "listdatabases",
		help:    "",
		handler: (*Handler).MsgListDatabases,
	},
	"listcommands": {
		name: "listcommands",
		help: "",
	},
	"ping": {
		name:    "listcommands",
		help:    "",
		handler: (*Handler).MsgPing,
	},
	"whatsmyuri": {
		name:    "whatsmyuri",
		help:    "",
		handler: (*Handler).MsgWhatsMyURI,
	},
	"serverstatus": {
		name:    "serverstatus",
		help:    "",
		handler: (*Handler).MsgServerStatus,
	},
	"delete": {
		name:           "delete",
		help:           "",
		storageHandler: (common.Storage).MsgDelete,
	},
	"find": {
		name:           "find",
		help:           "",
		storageHandler: (common.Storage).MsgFindOrCount,
	},
	"count": {
		name:           "count",
		help:           "",
		storageHandler: (common.Storage).MsgFindOrCount,
	},
	"insert": {
		name:           "insert",
		help:           "",
		storageHandler: (common.Storage).MsgInsert,
	},
	"update": {
		name:           "update",
		help:           "",
		storageHandler: (common.Storage).MsgUpdate,
	},
	"debug_error": {
		name: "debug_error",
		help: "",
		handler: func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
			return nil, errors.New("debug_error")
		},
	},
	"debug_panic": {
		name: "debug_panic",
		help: "",
		handler: func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
			panic("debug_panic")
		},
	},
}

func SupportedCommands(context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
	var reply wire.OpMsg
	err := reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"commands", func() []types.Document {
				var commandList []types.Document
				for _, v := range commands {
					commandList = append(commandList, types.MustMakeDocument(
						v.name, types.MustMakeDocument(
							"help", v.help,
						),
					))
				}

				return commandList
			},
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
