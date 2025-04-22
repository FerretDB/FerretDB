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
	"fmt"
	"log/slog"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// command represents a handler for single command.
type command struct {
	// anonymous indicates that the command does not require authentication.
	anonymous bool

	// Handler processes this command.
	//
	// The passed context is canceled when the client disconnects.
	Handler middleware.HandleFunc

	// Help is shown in the `listCommands` command output.
	// If empty, that command is hidden, but still can be used.
	Help string
}

// initCommands initializes the commands map for that handler instance.
func (h *Handler) initCommands() {
	commands := map[string]*command{
		// sorted alphabetically
		"aggregate": {
			Handler: h.msgAggregate,
			Help:    "Returns aggregated data.",
		},
		"authenticate": {
			// TODO https://github.com/FerretDB/FerretDB/issues/1731
			anonymous: true,
			Help:      "", // hidden while not implemented
		},
		"buildInfo": {
			Handler:   h.msgBuildInfo,
			anonymous: true,
			Help:      "Returns a summary of the build information.",
		},
		"buildinfo": { // old lowercase variant
			Handler:   h.msgBuildInfo,
			anonymous: true,
			Help:      "", // hidden
		},
		"bulkWrite": {
			// TODO https://github.com/microsoft/documentdb/issues/108
			// TODO https://github.com/FerretDB/FerretDB/issues/4910
			Help: "", // hidden while not implemented
		},
		"collMod": {
			Handler: h.msgCollMod,
			Help:    "Adds options to a collection or modify view definitions.",
		},
		"collStats": {
			Handler: h.msgCollStats,
			Help:    "Returns storage data for a collection.",
		},
		"compact": {
			Handler: h.msgCompact,
			Help:    "Reduces the disk space collection takes and refreshes its statistics.",
		},
		"connPoolStats": {
			// TODO https://github.com/FerretDB/FerretDB/issues/4909
			anonymous: true,
			Help:      "", // hidden while not implemented
		},
		"connectionStatus": {
			Handler:   h.msgConnectionStatus,
			anonymous: true,
			Help: "Returns information about the current connection, " +
				"specifically the state of authenticated users and their available permissions.",
		},
		"count": {
			Handler: h.msgCount,
			Help:    "Returns the count of documents that's matched by the query.",
		},
		"create": {
			Handler: h.msgCreate,
			Help:    "Creates the collection.",
		},
		"createIndexes": {
			Handler: h.msgCreateIndexes,
			Help:    "Creates indexes on a collection.",
		},
		"createUser": {
			Handler: h.msgCreateUser,
			Help:    "Creates a new user.",
		},
		"currentOp": {
			Handler: h.msgCurrentOp,
			Help:    "Returns information about operations currently in progress.",
		},
		"dataSize": {
			Handler: h.msgDataSize,
			Help:    "Returns the size of the collection in bytes.",
		},
		"dbStats": {
			Handler: h.msgDBStats,
			Help:    "Returns the statistics of the database.",
		},
		"dbstats": { // old lowercase variant
			Handler: h.msgDBStats,
			Help:    "", // hidden
		},
		"delete": {
			Handler: h.msgDelete,
			Help:    "Deletes documents matched by the query.",
		},
		"distinct": {
			Handler: h.msgDistinct,
			Help:    "Returns an array of distinct values for the given field.",
		},
		"drop": {
			Handler: h.msgDrop,
			Help:    "Drops the collection.",
		},
		"dropAllUsersFromDatabase": {
			Handler: h.msgDropAllUsersFromDatabase,
			Help:    "Drops all user from database.",
		},
		"dropDatabase": {
			Handler: h.msgDropDatabase,
			Help:    "Drops production database.",
		},
		"dropIndexes": {
			Handler: h.msgDropIndexes,
			Help:    "Drops indexes on a collection.",
		},
		"dropUser": {
			Handler: h.msgDropUser,
			Help:    "Drops user.",
		},
		"endSessions": {
			Handler: h.msgEndSessions,
			Help:    "Marks sessions as expired.",
		},
		"explain": {
			Handler: h.msgExplain,
			Help:    "Returns the execution plan.",
		},
		"ferretDebugError": {
			Handler: h.msgFerretDebugError,
			Help:    "Returns error for debugging.",
		},
		"find": {
			Handler: h.msgFind,
			Help:    "Returns documents matched by the query.",
		},
		"findAndModify": {
			Handler: h.msgFindAndModify,
			Help:    "Updates or deletes, and returns a document matched by the query.",
		},
		"findandmodify": { // old lowercase variant
			Handler: h.msgFindAndModify,
			Help:    "", // hidden
		},
		"getCmdLineOpts": {
			Handler: h.msgGetCmdLineOpts,
			Help:    "Returns a summary of all runtime and configuration options.",
		},
		"getFreeMonitoringStatus": {
			Handler: h.msgGetFreeMonitoringStatus,
			Help:    "Returns a status of the free monitoring.",
		},
		"getLog": {
			Handler: h.msgGetLog,
			Help:    "Returns the most recent logged events from memory.",
		},
		"getMore": {
			Handler: h.msgGetMore,
			Help:    "Returns the next batch of documents from a cursor.",
		},
		"getParameter": {
			Handler: h.msgGetParameter,
			Help:    "Returns the value of the parameter.",
		},
		"hello": {
			Handler:   h.msgHello,
			anonymous: true,
			Help:      "Returns the role of the FerretDB instance.",
		},
		"hostInfo": {
			Handler: h.msgHostInfo,
			Help:    "Returns a summary of the system information.",
		},
		"insert": {
			Handler: h.msgInsert,
			Help:    "Inserts documents into the database.",
		},
		"isMaster": {
			Handler:   h.msgIsMaster,
			anonymous: true,
			Help:      "Returns the role of the FerretDB instance.",
		},
		"ismaster": { // old lowercase variant
			Handler:   h.msgIsMaster,
			anonymous: true,
			Help:      "", // hidden
		},
		"killAllSessions": {
			Handler: h.msgKillAllSessions,
			Help:    "Kills all sessions.",
		},
		"killAllSessionsByPattern": {
			Handler: h.msgKillAllSessionsByPattern,
			Help:    "Kills all sessions that match the pattern.",
		},
		"killCursors": {
			Handler: h.msgKillCursors,
			Help:    "Closes server cursors.",
		},
		"killSessions": {
			Handler: h.msgKillSessions,
			Help:    "Kills sessions.",
		},
		"listCollections": {
			Handler: h.msgListCollections,
			Help:    "Returns the information of the collections and views in the database.",
		},
		"listCommands": {
			Handler: h.msgListCommands,
			Help:    "Returns a list of currently supported commands.",
		},
		"listDatabases": {
			Handler: h.msgListDatabases,
			Help:    "Returns a summary of all the databases.",
		},
		"listIndexes": {
			Handler: h.msgListIndexes,
			Help:    "Returns a summary of indexes of the specified collection.",
		},
		"logout": {
			Handler:   h.msgLogout,
			anonymous: true,
			Help:      "Logs out from the current session.",
		},
		"ping": {
			Handler:   h.msgPing,
			anonymous: true,
			Help:      "Returns a pong response.",
		},
		"refreshSessions": {
			Handler: h.msgRefreshSessions,
			Help:    "Updates the last used time of sessions.",
		},
		"reIndex": {
			Handler: h.msgReIndex,
			Help:    "Drops and recreates all indexes except default _id index of a collection.",
		},
		"renameCollection": {
			Handler: h.msgRenameCollection,
			Help:    "Changes the name of an existing collection.",
		},
		"saslStart": {
			Handler:   h.msgSASLStart,
			anonymous: true,
			Help:      "", // hidden
		},
		"saslContinue": {
			Handler:   h.msgSASLContinue,
			anonymous: true,
			Help:      "", // hidden
		},
		"serverStatus": {
			Handler: h.msgServerStatus,
			Help:    "Returns an overview of the databases state.",
		},
		"setFreeMonitoring": {
			Handler: h.msgSetFreeMonitoring,
			Help:    "Toggles free monitoring.",
		},
		"startSession": {
			Handler: h.msgStartSession,
			Help:    "Returns a session.",
		},
		"update": {
			Handler: h.msgUpdate,
			Help:    "Updates documents that are matched by the query.",
		},
		"updateUser": {
			Handler: h.msgUpdateUser,
			Help:    "Updates user.",
		},
		"usersInfo": {
			Handler: h.msgUsersInfo,
			Help:    "Returns information about users.",
		},
		"validate": {
			Handler: h.msgValidate,
			Help:    "Validates collection.",
		},
		"whatsmyuri": {
			Handler:   h.msgWhatsMyURI,
			anonymous: true,
			Help:      "Returns peer information.",
		},
		// please keep sorted alphabetically
	}

	h.commands = make(map[string]*command, len(commands))

	for name, cmd := range commands {
		if cmd.Handler == nil {
			cmd.Handler = notImplemented(name)
		}

		if h.Auth && !cmd.anonymous {
			cmd.Handler = auth(cmd.Handler, logging.WithName(h.L, "auth"), name)
		}

		h.commands[name] = cmd
	}
}

// auth is a middleware that wraps the command handler with authentication check.
//
// Context must contain [*conninfo.ConnInfo].
func auth(next middleware.HandleFunc, l *slog.Logger, command string) middleware.HandleFunc {
	return func(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
		conv := conninfo.Get(ctx).Conv()
		succeed := conv.Succeed()
		username := conv.Username()

		switch {
		case conv == nil:
			l.WarnContext(ctx, "No existing conversation")

		case !succeed:
			l.WarnContext(ctx, "Conversation did not succeed", slog.String("username", username))

		default:
			l.DebugContext(ctx, "Authentication passed", slog.String("username", username))

			return next(ctx, req)
		}

		return nil, mongoerrors.New(
			mongoerrors.ErrUnauthorized,
			fmt.Sprintf("Command %s requires authentication", command),
		)
	}
}

// notImplemented returns a handler that returns an error indicating that the command is not implemented.
func notImplemented(command string) middleware.HandleFunc {
	return func(context.Context, *middleware.Request) (*middleware.Response, error) {
		return nil, mongoerrors.New(
			mongoerrors.ErrNotImplemented,
			fmt.Sprintf("Command %s is not implemented", command),
		)
	}
}

// notFound returns a handler that returns not found error.
func notFound(command string) middleware.HandleFunc {
	return func(context.Context, *middleware.Request) (*middleware.Response, error) {
		return nil, mongoerrors.New(
			mongoerrors.ErrCommandNotFound,
			fmt.Sprintf("no such command: '%s'", command),
		)
	}
}
