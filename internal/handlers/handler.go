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
	"sync/atomic"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/prometheus/client_golang/prometheus"
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
	startTime     time.Time
}

// NewOpts represents handler configuration.
type NewOpts struct {
	PgPool        *pg.Pool
	Logger        *zap.Logger
	SQLStorage    common.Storage
	JSONB1Storage common.Storage
	Metrics       *Metrics
	PeerAddr      string
	StartTime     time.Time
}

// New returns a new handler.
func New(opts *NewOpts) *Handler {
	return &Handler{
		pgPool:    opts.PgPool,
		l:         opts.Logger,
		sql:       opts.SQLStorage,
		jsonb1:    opts.JSONB1Storage,
		metrics:   opts.Metrics,
		peerAddr:  opts.PeerAddr,
		startTime: opts.StartTime,
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
	var cmdLabel string
	requests := h.metrics.requests.MustCurryWith(prometheus.Labels{"opcode": reqHeader.OpCode.String()})

	var resLabel *string
	defer func() {
		if resLabel == nil {
			resLabel = pointer.To("panic")
		}
		h.metrics.responses.WithLabelValues(resHeader.OpCode.String(), cmdLabel, *resLabel).Inc()
	}()

	resHeader = new(wire.MsgHeader)
	var err error
	switch reqHeader.OpCode {
	case wire.OP_MSG:
		// count requests even if msg's document is invalid
		msg := reqBody.(*wire.OpMsg)
		var document *types.Document
		document, err = msg.Document()
		if document != nil {
			cmdLabel = document.Command()
		}
		requests.WithLabelValues(cmdLabel).Inc()

		if err == nil {
			resHeader.OpCode = wire.OP_MSG
			resBody, err = h.handleOpMsg(ctx, msg, cmdLabel)
		}

	case wire.OP_QUERY:
		query := reqBody.(*wire.OpQuery)
		cmdLabel = query.Query.Command()
		requests.WithLabelValues(cmdLabel).Inc()

		resHeader.OpCode = wire.OP_REPLY
		resBody, err = h.handleOpQuery(ctx, query, cmdLabel)

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
		requests.WithLabelValues(cmdLabel).Inc()
		panic(fmt.Sprintf("unexpected OpCode %s", reqHeader.OpCode))
	}

	if err != nil {
		if resHeader.OpCode != wire.OP_MSG {
			panic(err)
		}

		protoErr, recoverable := common.ProtocolError(err)
		resLabel = pointer.To(protoErr.Error())
		closeConn = !recoverable
		var res wire.OpMsg
		err = res.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{protoErr.Document()},
		})
		if err != nil {
			panic(err)
		}
		resBody = &res
	}

	resHeader.ResponseTo = reqHeader.RequestID

	// TODO Don't call MarshalBinary there. Fix header in the caller?
	// https://github.com/FerretDB/FerretDB/issues/273
	b, err := resBody.MarshalBinary()
	if err != nil {
		panic(err)
	}
	resHeader.MessageLength = int32(wire.MsgHeaderLen + len(b))

	if resHeader.RequestID != 0 {
		panic("resHeader.RequestID must not be set by handler")
	}
	resHeader.RequestID = atomic.AddInt32(&h.lastRequestID, 1)

	resLabel = pointer.To("ok")
	return
}

func (h *Handler) handleOpMsg(ctx context.Context, msg *wire.OpMsg, cmd string) (*wire.OpMsg, error) {
	// special case to avoid circular dependency
	if cmd == "listcommands" {
		return listCommands(ctx, msg)
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

	return nil, common.NewError(common.ErrCommandNotFound, fmt.Errorf("no such command: '%s'", cmd))
}

func (h *Handler) handleOpQuery(ctx context.Context, query *wire.OpQuery, cmd string) (*wire.OpReply, error) {
	if query.FullCollectionName == "admin.$cmd" {
		return h.QueryCmd(ctx, query)
	}

	err := fmt.Errorf("handleOpQuery: unhandled collection %q", query.FullCollectionName)
	return nil, common.NewError(common.ErrNotImplemented, err)
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

	tables, storages, err := h.pgPool.Tables(ctx, db)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var storage string
	for i, t := range tables {
		if t == collection {
			storage = storages[i]
			break
		}
	}

	switch command {
	case "delete", "find", "count":
		switch storage {
		case pg.JSONB1Table:
			return h.jsonb1, nil
		case pg.SQLTable:
			return h.sql, nil
		default:
			// does not matter much what we return there
			return h.sql, nil
		}

	case "insert", "update":
		switch storage {
		case pg.JSONB1Table:
			return h.jsonb1, nil
		case pg.SQLTable:
			return h.sql, nil
		}

		// Table (or even schema) does not exist. Try to create it,
		// but keep in mind that it can be created in concurrent connection.

		if err := h.pgPool.CreateSchema(ctx, db); err != nil && err != pg.ErrAlreadyExist {
			return nil, lazyerrors.Errorf("Handler.msgStorage: %w", err)
		}

		if err := h.pgPool.CreateTable(ctx, db, collection); err != nil {
			if err == pg.ErrAlreadyExist {
				return h.jsonb1, nil
			}
			return nil, lazyerrors.Errorf("Handler.msgStorage: %w", err)
		}

		h.l.Info("Created jsonb1 table.", zap.String("schema", db), zap.String("table", collection))
		return h.jsonb1, nil

	default:
		panic(fmt.Sprintf("unhandled command %q", command))
	}
}
