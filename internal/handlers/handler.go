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
		jsonb1:    opts.JSONB1Storage,
		metrics:   opts.Metrics,
		peerAddr:  opts.PeerAddr,
		startTime: opts.StartTime,
	}
}

// Handle handles the message.
//
// Specific message handlers should do one of the following:
//  * return normal response body;
//  * return protocol error (*common.Error) - it will be returned to the client;
//  * return any other error - it will be returned to the client as InternalError before terminating connection.
//
// They should not panic on bad input, but may do so in "impossible" cases.
// They also should not use recover(). That allows us to use fuzzing.
// (Panics terminate the connection without a response on a different level.)
//
//nolint:lll // arguments are long
func (h *Handler) Handle(ctx context.Context, reqHeader *wire.MsgHeader, reqBody wire.MsgBody) (resHeader *wire.MsgHeader, resBody wire.MsgBody, closeConn bool) {
	var command string
	requests := h.metrics.requests.MustCurryWith(prometheus.Labels{"opcode": reqHeader.OpCode.String()})

	var result *string
	defer func() {
		if result == nil {
			result = pointer.ToString("panic")
		}
		h.metrics.responses.WithLabelValues(resHeader.OpCode.String(), command, *result).Inc()
	}()

	resHeader = new(wire.MsgHeader)
	var err error
	switch reqHeader.OpCode {
	case wire.OP_MSG:
		// count requests even if msg's document is invalid
		msg := reqBody.(*wire.OpMsg)
		var document *types.Document
		document, err = msg.Document()
		command = document.Command()
		requests.WithLabelValues(command).Inc()

		if err == nil {
			resHeader.OpCode = wire.OP_MSG
			resBody, err = h.handleOpMsg(ctx, msg, command)
		}

	case wire.OP_QUERY:
		query := reqBody.(*wire.OpQuery)
		command = query.Query.Command()
		requests.WithLabelValues(command).Inc()

		resHeader.OpCode = wire.OP_REPLY
		resBody, err = h.handleOpQuery(ctx, query, command)

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
		requests.WithLabelValues(command).Inc()

		err = lazyerrors.Errorf("unexpected OpCode %s", reqHeader.OpCode)
	}

	// set body for error
	if err != nil {
		switch resHeader.OpCode {
		case wire.OP_MSG:
			protoErr, recoverable := common.ProtocolError(err)
			closeConn = !recoverable
			var res wire.OpMsg
			err = res.SetSections(wire.OpMsgSection{
				Documents: []*types.Document{protoErr.Document()},
			})
			if err != nil {
				panic(err)
			}
			resBody = &res
			result = pointer.ToString(protoErr.Code().String())

		case wire.OP_QUERY:
			fallthrough
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
			// do not panic to make fuzzing easier
			h.l.Error("Handler error for unexpected response opcode", zap.Error(err), zap.Stringer("opcode", resHeader.OpCode))
			closeConn = true
			result = pointer.ToString("unexpected")
			return
		}
	}

	// TODO Don't call MarshalBinary there. Fix header in the caller?
	// https://github.com/FerretDB/FerretDB/issues/273
	b, err := resBody.MarshalBinary()
	if err != nil {
		result = nil
		panic(err)
	}
	resHeader.MessageLength = int32(wire.MsgHeaderLen + len(b))

	resHeader.RequestID = atomic.AddInt32(&h.lastRequestID, 1)
	resHeader.ResponseTo = reqHeader.RequestID

	if result == nil {
		result = pointer.ToString("ok")
	}

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
		h.l.Sugar().Debugf("Handling with storage %T", storage)
		return cmd.storageHandler(storage, ctx, msg)
	}

	errMsg := fmt.Sprintf("no such command: '%s'", cmd)
	return nil, common.NewErrorMsg(common.ErrCommandNotFound, errMsg)
}

func (h *Handler) handleOpQuery(ctx context.Context, query *wire.OpQuery, cmd string) (*wire.OpReply, error) {
	if query.FullCollectionName == "admin.$cmd" {
		return h.QueryCmd(ctx, query)
	}

	msg := fmt.Sprintf("handleOpQuery: unhandled collection %q", query.FullCollectionName)
	return nil, common.NewErrorMsg(common.ErrNotImplemented, msg)
}

func (h *Handler) msgStorage(ctx context.Context, msg *wire.OpMsg) (common.Storage, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, fmt.Errorf("Handler.msgStorage: %w", err)
	}

	command := document.Command()
	if command == "createindexes" {
		// TODO https://github.com/FerretDB/FerretDB/issues/78
		return h.jsonb1, nil
	}

	var db, collection string
	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}
	if collection, err = common.GetRequiredParam[string](document, command); err != nil {
		return nil, err
	}

	tables, storages, err := h.pgPool.Tables(ctx, db)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	storage := "default"
	for i, t := range tables {
		if t == collection {
			storage = storages[i]
			break
		}
	}

	h.l.Sugar().Debugf("Using storage %q for collection %q in database %q", storage, collection, db)

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
