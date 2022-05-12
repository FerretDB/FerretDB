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

// Package dummy implements a dummy handler.
package dummy

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Handler data struct.
type Handler struct{}

// New returns a new handler.
func New() common.Handler {
	return new(Handler)
}

// MsgBuildInfo returns a summary of the build information.
func (h *Handler) MsgBuildInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgCollStats returns storage data for a collection.
func (h *Handler) MsgCollStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgCreate creates a collection.
func (h *Handler) MsgCreate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgDataSize returns the size of the collection in bytes.
func (h *Handler) MsgDataSize(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgDBStats Returns the statistics of the database.
func (h *Handler) MsgDBStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgDrop drops the collection.
func (h *Handler) MsgDrop(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgDropDatabase deletes the database.
func (h *Handler) MsgDropDatabase(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgGetCmdLineOpts returns a summary of all runtime and configuration options.
func (h *Handler) MsgGetCmdLineOpts(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgGetLog returns the most recent logged events from memory.
func (h *Handler) MsgGetLog(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgGetParameter returns the value of the parameter.
func (h *Handler) MsgGetParameter(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgHostInfo returns a summary of the system information.
func (h *Handler) MsgHostInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgIsMaster returns the role of the FerretDB instance.
func (h *Handler) MsgIsMaster(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgHello returns the role of the FerretDB instance.
func (h *Handler) MsgHello(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgListCollections returns the information of the collections and views in the database.
func (h *Handler) MsgListCollections(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgListDatabases returns a summary of all the databases.
func (h *Handler) MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgPing returns a pong response. Used for testing purposes.
func (h *Handler) MsgPing(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgServerStatus returns an overview of the databases state.
func (h *Handler) MsgServerStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgWhatsMyURI an internal command.
func (h *Handler) MsgWhatsMyURI(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgCount returns the count of documents that's matched by the query.
func (h *Handler) MsgCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgCreateIndexes creates indexes on a collection.
func (h *Handler) MsgCreateIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgDelete deletes documents matched by the query.
func (h *Handler) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgFind returns documents matched by the query.
func (h *Handler) MsgFind(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgFindAndModify inserts, updates, or deletes, and returns a document matched by the query.
func (h *Handler) MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgInsert inserts documents into the database.
func (h *Handler) MsgInsert(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgUpdate updates documents that are matched by the query.
func (h *Handler) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// CmdQuery runs query operation command.
func (h *Handler) CmdQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// MsgConnectionStatus information about the current connection,
// specifically the state of authenticated users and their available permissions.
func (h *Handler) MsgConnectionStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
}

// Close prepares handler for graceful shutdown: closes connections, channels etc.
func (h *Handler) Close() {}

// check interfaces
var (
	_ common.Handler = (*Handler)(nil)
)
