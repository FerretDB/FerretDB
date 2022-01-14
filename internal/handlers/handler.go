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
	"fmt"
	"sort"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Handler data struct.
type Handler struct {
	// TODO replace those fields with opts *NewOpts
	pgPool        *pg.Pool
	peerAddr      string
	l             *zap.Logger
	sql           common.Storage
	jsonb1        common.Storage
	metrics       *Metrics
	lastRequestID int32
}

// NewOpts represents handler configuration.
type NewOpts struct {
	PgPool        *pg.Pool
	Logger        *zap.Logger
	SQLStorage    common.Storage
	JSONB1Storage common.Storage
	Metrics       *Metrics
	PeerAddr      string
}

// New returns a new handler.
func New(opts *NewOpts) *Handler {
	return &Handler{
		pgPool:   opts.PgPool,
		l:        opts.Logger,
		sql:      opts.SQLStorage,
		jsonb1:   opts.JSONB1Storage,
		metrics:  opts.Metrics,
		peerAddr: opts.PeerAddr,
	}
}

// Handle handles the message.
//
// Message handlers should:
//  * return normal response body;
//  * return protocol error (*common.Error) - it will be returned to the client;
//  * return any other error - it will be returned to the client as InternalError before terminating connection;
//  * panic - that will terminate the connection without a response.
//
//nolint:lll // arguments are long
func (h *Handler) Handle(ctx context.Context, reqHeader *wire.MsgHeader, reqBody wire.MsgBody) (resHeader *wire.MsgHeader, resBody wire.MsgBody, closeConn bool) {
	resHeader = new(wire.MsgHeader)
	var err error

	switch reqHeader.OpCode {
	case wire.OP_MSG:
		resHeader.OpCode = wire.OP_MSG
		resBody, err = h.handleOpMsg(ctx, reqBody.(*wire.OpMsg))
	case wire.OP_QUERY:
		resHeader.OpCode = wire.OP_REPLY
		resBody, err = h.handleOpQuery(ctx, reqBody.(*wire.OpQuery))
	case wire.OP_REPLY:
		fallthrough
	case wire.OP_UPDATE:
		fallthrough
	case wire.OP_INSERT:
		fallthrough
	case wire.OP_GET_BY_OID:
		fallthrough
	case wire.OP_GET_MORE:
		fallthrough
	case wire.OP_DELETE:
		fallthrough
	case wire.OP_KILL_CURSORS:
		fallthrough
	case wire.OP_COMPRESSED:
		fallthrough
	default:
		h.metrics.requests.WithLabelValues(reqHeader.OpCode.String(), "").Inc()
		panic(fmt.Sprintf("unexpected OpCode %s", reqHeader.OpCode))
	}

	if err != nil {
		if resHeader.OpCode != wire.OP_MSG {
			panic(err)
		}

		protoErr, recoverable := common.ProtocolError(err)
		closeConn = !recoverable
		var res wire.OpMsg
		err = res.SetSections(wire.OpMsgSection{
			Documents: []types.Document{protoErr.Document()},
		})
		if err != nil {
			panic(err)
		}
		resBody = &res
	}

	resHeader.ResponseTo = reqHeader.RequestID

	// FIXME don't call MarshalBinary there
	// Fix header in the caller?
	b, err := resBody.MarshalBinary()
	if err != nil {
		panic(err)
	}
	resHeader.MessageLength = int32(wire.MsgHeaderLen + len(b))

	if resHeader.RequestID != 0 {
		panic("resHeader.RequestID must not be set by handler")
	}
	resHeader.RequestID = atomic.AddInt32(&h.lastRequestID, 1)

	return
}

func (h *Handler) handleOpMsg(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	cmd := document.Command()

	h.metrics.requests.WithLabelValues(wire.OP_MSG.String(), cmd).Inc()

	if cmd == "listcommands" {
		return SupportedCommands(ctx, msg)
	}

	if cmd, ok := commands[cmd]; ok {
		if cmd.handler != nil {
			return cmd.handler(h, ctx, msg)
		}

		storage, err := h.msgStorage(ctx, msg)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		return cmd.storageHandler(storage, ctx, msg)
	}

	return nil, common.NewErrorMessage(common.ErrCommandNotFound, "no such command: '%s'", cmd)
}

func (h *Handler) handleOpQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	cmd := query.Query.Command()
	h.metrics.requests.WithLabelValues(wire.OP_QUERY.String(), cmd).Inc()

	if query.FullCollectionName == "admin.$cmd" {
		return h.QueryCmd(ctx, query)
	}

	return nil, common.NewErrorMessage(common.ErrNotImplemented, "handleOpQuery: unhandled collection %q", query.FullCollectionName)
}

func (h *Handler) msgStorage(ctx context.Context, msg *wire.OpMsg) (common.Storage, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, fmt.Errorf("Handler.msgStorage: %w", err)
	}

	m := document.Map()
	command := document.Command()

	if command == "createindexes" {
		// TODO https://github.com/FerretDB/FerretDB/issues/78
		return h.jsonb1, nil
	}

	collection := m[command].(string)
	db := m["$db"].(string)

	var jsonbTableExist bool
	sql := `SELECT COUNT(*) > 0 FROM information_schema.columns WHERE column_name = $1 AND table_schema = $2 AND table_name = $3`
	if err := h.pgPool.QueryRow(ctx, sql, "_jsonb", db, collection).Scan(&jsonbTableExist); err != nil {
		return nil, lazyerrors.Errorf("Handler.msgStorage: %w", err)
	}

	switch command {
	case "delete", "find", "count":
		if jsonbTableExist {
			return h.jsonb1, nil
		}
		return h.sql, nil

	case "insert", "update":
		if jsonbTableExist {
			return h.jsonb1, nil
		}

		// check if SQL table exist
		tables, err := h.pgPool.Tables(ctx, db)
		if err != nil {
			return nil, lazyerrors.Errorf("Handler.msgStorage: %w", err)
		}
		if i := sort.SearchStrings(tables, collection); i < len(tables) && tables[i] == collection {
			return h.sql, nil
		}

		// create schema if needed
		if err := h.pgPool.CreateSchema(ctx, db); err != nil && err != pg.ErrAlreadyExist {
			return nil, lazyerrors.Errorf("Handler.msgStorage: %w", err)
		}

		// create table
		if err := h.pgPool.CreateTable(ctx, db, collection); err != nil {
			return nil, lazyerrors.Errorf("Handler.msgStorage: %w", err)
		}

		h.l.Info("Created jsonb1 table.", zap.String("schema", db), zap.String("table", collection))
		return h.jsonb1, nil

	default:
		panic(fmt.Sprintf("unhandled command %q", command))
	}
}
