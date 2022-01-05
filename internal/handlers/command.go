package handlers

import (
	"context"
	"errors"

	"github.com/FerretDB/FerretDB/internal/wire"
)

type command struct {
	name    string
	help    string
	handler func(*Handler, context.Context, *wire.OpMsg) (*wire.OpMsg, error)
}

var commands = map[string]command{
	"buildinfo": {
		name:    "buildinfo",
		help:    "",
		handler: (*Handler).shared.MsgBuildInfo,
	},
	"collstats": {
		name:    "collstats",
		help:    "",
		handler: (*Handler).shared.MsgCollStats,
	},
	"createindexes": {
		name:    "createindexes",
		help:    "",
		handler: storage.MsgCreateIndexes,
	},
	"create": {
		name:    "create",
		help:    "",
		handler: h.shared.MsgCreate,
	},
	"drop": {
		name:    "drop",
		help:    "",
		handler: h.shared.MsgHostInfo,
	},
	"dropdatabase": {
		name:    "dropdatabase",
		help:    "",
		handler: h.shared.MsgHostInfo,
	},
	"getcmdlineopts": {
		name:    "getcmdlineopts",
		help:    "",
		handler: h.shared.MsgGetCmdLineOpts,
	},
	"getlog": {
		name:    "getlog",
		help:    "",
		handler: h.shared.MsgGetLog,
	},
	"getparameter": {
		name:    "getparameter",
		help:    "",
		handler: h.shared.MsgGetParameter,
	},
	"hostinfo": {
		name:    "hostInfo",
		help:    "",
		handler: h.shared.MsgHostInfo,
	},
	"ismaster": {
		name:    "ismaster",
		help:    "",
		handler: h.shared.MsgHello,
	},
	"hello": {
		name:    "hello",
		help:    "",
		handler: h.shared.MsgHello,
	},
	"listcollections": {
		name:    "listcollections",
		help:    "",
		handler: h.shared.MsgListCollections,
	},
	"listdatabases": {
		name:    "listdatabases",
		help:    "",
		handler: h.shared.MsgListDatabases,
	},
	"listcommands": {
		name:    "listcommands",
		help:    "",
		handler: h.shared.MsgListCommands,
	},
	"ping": {
		name:    "listcommands",
		help:    "",
		handler: h.shared.MsgPing,
	},
	"whatsmyuri": {
		name:    "whatsmyuri",
		help:    "",
		handler: h.shared.MsgWhatsMyURI,
	},
	"serverstatus": {
		name:    "serverstatus",
		help:    "",
		handler: h.shared.MsgServerStatus,
	},
	"delete": {
		name:    "delete",
		help:    "",
		handler: storage.MsgDelete,
	},
	"find": {
		name:    "find",
		help:    "",
		handler: storage.MsgFindOrCount,
	},
	"count": {
		name:    "count",
		help:    "",
		handler: storage.MsgFindOrCount,
	},
	"insert": {
		name:    "insert",
		help:    "",
		handler: storage.MsgInsert,
	},
	"update": {
		name:    "update",
		help:    "",
		handler: storage.MsgUpdate,
	},
	"debug_error": {
		name: "debug_error",
		help: "",
		handler: func(context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
			return nil, errors.New("debug_error")
		},
	},
	"debug_panic": {
		name: "debug_panic",
		help: "",
		handler: func(context.Context, *wire.OpMsg) (*wire.OpMsg, error) {
			panic("debug_panic")
		},
	},
}
