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
	"github.com/MangoDB-io/MangoDB/internal/pgconn"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

type Handler struct {
	pgPool *pgconn.Pool
	l      *zap.Logger
	shared *shared.Handler
	sql    common.Storage
	jsonb1 common.Storage

	lastRequestID int32
}

func New(pgPool *pgconn.Pool, l *zap.Logger, shared *shared.Handler, sql, jsonb1 common.Storage) *Handler {
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
		resMsg, err = h.handleOpMsg(ctx, header, msg.(*wire.OpMsg))
	case wire.OP_QUERY:
		resHeader.OpCode = wire.OP_REPLY
		resMsg, err = h.handleOpQuery(ctx, header, msg.(*wire.OpQuery))
	case wire.OP_REPLY:
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
		resMsg = &wire.OpMsg{
			Documents: []types.Document{e.Document()},
		}
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

func (h *Handler) handleOpMsg(ctx context.Context, header *wire.MsgHeader, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document := msg.Documents[0]

	cmd := document.Command()

	switch cmd {
	case "buildinfo":
		return h.shared.MsgBuildInfo(ctx, header, msg)
	case "drop":
		return h.shared.MsgDrop(ctx, header, msg)
	case "getcmdlineopts":
		return h.shared.MsgGetCmdLineOpts(ctx, header, msg)
	case "getlog":
		return h.shared.MsgGetLog(ctx, header, msg)
	case "ismaster":
		return h.shared.MsgIsMaster(ctx, header, msg)
	case "listcollections":
		return h.shared.MsgListCollections(ctx, header, msg)
	case "listdatabases":
		return h.shared.MsgListDatabases(ctx, header, msg)
	case "ping":
		return h.shared.MsgPing(ctx, header, msg)
	case "whatsmyuri":
		return h.shared.MsgWhatsMyURI(ctx, header, msg)

	case "delete":
		return h.msgStorage(ctx, msg).MsgDelete(ctx, header, msg)
	case "find":
		return h.msgStorage(ctx, msg).MsgFind(ctx, header, msg)
	case "insert":
		return h.msgStorage(ctx, msg).MsgInsert(ctx, header, msg)
	case "update":
		return h.msgStorage(ctx, msg).MsgUpdate(ctx, header, msg)

	default:
		return nil, common.NewError(common.ErrNotImplemented, fmt.Errorf("unhandled msg %q", cmd), header, msg)
	}
}

func (h *Handler) handleOpQuery(ctx context.Context, header *wire.MsgHeader, msg *wire.OpQuery) (*wire.OpReply, error) {
	if msg.FullCollectionName == "admin.$cmd" {
		return h.shared.QueryCmd(ctx, header, msg)
	}

	return nil, common.NewError(common.ErrNotImplemented, nil, header, msg)
}

func (h *Handler) msgStorage(ctx context.Context, msg *wire.OpMsg) common.Storage {
	document := msg.Documents[0]

	m := document.Map()
	command := document.Command()
	collection := m[command].(string)
	db := m["$db"].(string)

	sql := `SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2`
	var tableExist bool
	if err := h.pgPool.QueryRow(ctx, sql, db, collection).Scan(&tableExist); err != nil {
		h.l.Warn(err.Error())
		return h.sql
	}

	sql = `SELECT COUNT(*) > 0 FROM information_schema.columns WHERE column_name = $1 AND table_schema = $2 AND table_name = $3`
	var jsonbExist bool
	if err := h.pgPool.QueryRow(ctx, sql, "_jsonb", db, collection).Scan(&jsonbExist); err != nil {
		h.l.Warn(err.Error())
		return h.sql
	}

	switch command {
	case "delete":
		if jsonbExist {
			return h.jsonb1
		}
	case "find":
		if jsonbExist {
			return h.jsonb1
		}
	case "insert":
		if jsonbExist {
			return h.jsonb1
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

			return h.jsonb1
		}
	case "update":
		if jsonbExist {
			return h.jsonb1
		}
	default:
		panic(fmt.Sprintf("unhandled command %q", command))
	}

	return h.sql
}
