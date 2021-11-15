// Copyright 2021 Baltoro OÜ.
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
	"sync/atomic"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/MangoDB-io/MangoDB/internal/handlers/common"
	"github.com/MangoDB-io/MangoDB/internal/handlers/shared"
	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

type Handler struct {
	pgPool *pg.Pool
	l      *zap.Logger
	shared *shared.Handler
	sql    common.Storage
	jsonb1 common.Storage

	lastRequestID int32
}

func New(pgPool *pg.Pool, l *zap.Logger, shared *shared.Handler, sql, jsonb1 common.Storage) *Handler {
	return &Handler{
		pgPool: pgPool,
		l:      l,
		shared: shared,
		sql:    sql,
		jsonb1: jsonb1,
	}
}

func (h *Handler) Handle(ctx context.Context, header *wire.MsgHeader, msg wire.MsgBody) (*wire.MsgHeader, wire.MsgBody, error) {
	resHeader := new(wire.MsgHeader)
	var resMsg wire.MsgBody
	var err error

	switch header.OpCode {
	case wire.OP_MSG:
		resHeader.OpCode = wire.OP_MSG
		resMsg, err = h.handleOpMsg(ctx, msg.(*wire.OpMsg))
	case wire.OP_QUERY:
		resHeader.OpCode = wire.OP_REPLY
		resMsg, err = h.handleOpQuery(ctx, msg.(*wire.OpQuery))
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
		panic(fmt.Sprintf("unexpected OpCode %s", header.OpCode))
	}

	if err != nil {
		e, ok := err.(common.Error)
		if !ok {
			return nil, nil, err
		}

		// FIXME use correct type for OP_QUERY
		var res wire.OpMsg
		err := res.SetSections(wire.OpMsgSection{
			Documents: []types.Document{e.Document()},
		})
		if err != nil {
			return nil, nil, err
		}
		resMsg = &res
	}

	resHeader.ResponseTo = header.RequestID

	// FIXME don't call MarshalBinary there
	// Fix header in the caller?
	b, err := resMsg.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	resHeader.MessageLength = int32(wire.MsgHeaderLen + len(b))

	if resHeader.RequestID != 0 {
		panic("resHeader.RequestID must not be set by handler")
	}
	resHeader.RequestID = atomic.AddInt32(&h.lastRequestID, 1)

	return resHeader, resMsg, nil
}

//nolint:goconst // good enough
func (h *Handler) handleOpMsg(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, common.NewError(common.ErrInternalError, err)
	}

	cmd := document.Command()

	switch cmd {
	case "buildinfo":
		return h.shared.MsgBuildInfo(ctx, msg)
	case "drop":
		return h.shared.MsgDrop(ctx, msg)
	case "getcmdlineopts":
		return h.shared.MsgGetCmdLineOpts(ctx, msg)
	case "getlog":
		return h.shared.MsgGetLog(ctx, msg)
	case "ismaster":
		return h.shared.MsgIsMaster(ctx, msg)
	case "listcollections":
		return h.shared.MsgListCollections(ctx, msg)
	case "listdatabases":
		return h.shared.MsgListDatabases(ctx, msg)
	case "ping":
		return h.shared.MsgPing(ctx, msg)
	case "whatsmyuri":
		return h.shared.MsgWhatsMyURI(ctx, msg)

	case "delete", "find", "insert", "update":
		storage, err := h.msgStorage(ctx, msg)
		if err != nil {
			return nil, common.NewError(common.ErrInternalError, err)
		}

		switch cmd {
		case "delete":
			return storage.MsgDelete(ctx, msg)
		case "find":
			return storage.MsgFind(ctx, msg)
		case "insert":
			return storage.MsgInsert(ctx, msg)
		case "update":
			return storage.MsgUpdate(ctx, msg)
		default:
			panic("not reached")
		}

	default:
		return nil, common.NewError(common.ErrNotImplemented, fmt.Errorf("unhandled msg %q", cmd))
	}
}

func (h *Handler) handleOpQuery(ctx context.Context, msg *wire.OpQuery) (*wire.OpReply, error) {
	if msg.FullCollectionName == "admin.$cmd" {
		return h.shared.QueryCmd(ctx, msg)
	}

	return nil, common.NewError(common.ErrNotImplemented, fmt.Errorf("unhandled collection %q", msg.FullCollectionName))
}

func (h *Handler) msgStorage(ctx context.Context, msg *wire.OpMsg) (common.Storage, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Errorf("Handler.msgStorage: %w", err)
	}

	m := document.Map()
	command := document.Command()
	collection := m[command].(string)
	db := m["$db"].(string)

	sql := `SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2`
	var tableExist bool
	if err := h.pgPool.QueryRow(ctx, sql, db, collection).Scan(&tableExist); err != nil {
		return nil, lazyerrors.Errorf("Handler.msgStorage: %w", err)
	}

	sql = `SELECT COUNT(*) > 0 FROM information_schema.columns WHERE column_name = $1 AND table_schema = $2 AND table_name = $3`
	var jsonbExist bool
	if err := h.pgPool.QueryRow(ctx, sql, "_jsonb", db, collection).Scan(&jsonbExist); err != nil {
		return nil, lazyerrors.Errorf("Handler.msgStorage: %w", err)
	}

	switch command {
	case "delete":
		if jsonbExist {
			return h.jsonb1, nil
		}
	case "find":
		if jsonbExist {
			return h.jsonb1, nil
		}
	case "insert":
		if jsonbExist {
			return h.jsonb1, nil
		}
		if !tableExist {
			sql = `CREATE TABLE ` + pgx.Identifier{db, collection}.Sanitize() + ` (_jsonb jsonb)`
			_, err := h.pgPool.Exec(ctx, sql)
			fields := []zap.Field{zap.String("schema", db), zap.String("table", collection)}
			if err != nil {
				fields = append(fields, zap.Error(err))
				h.l.Warn("Failed to create jsonb1 table.", fields...)
			} else {
				h.l.Info("Created jsonb1 table.", fields...)
			}

			return h.jsonb1, nil
		}
	case "update":
		if jsonbExist {
			return h.jsonb1, nil
		}
	default:
		panic(fmt.Sprintf("unhandled command %q", command))
	}

	return h.sql, nil
}
