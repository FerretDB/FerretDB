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

package clientconn

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/AlekSi/pointer"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Route routes the message.
//
// Specific message handlers should do one of the following:
//  * return normal response body;
//  * return protocol error (*common.Error, possibly wrapped) - it will be returned to the client;
//  * return any other error - it will be returned to the client as InternalError before terminating connection.
//
// They should not panic on bad input, but may do so in "impossible" cases.
// They also should not use recover(). That allows us to use fuzzing.
// (Panics terminate the connection without a response on a different level.)
func (c *conn) route(ctx context.Context, reqHeader *wire.MsgHeader, reqBody wire.MsgBody,
) (resHeader *wire.MsgHeader, resBody wire.MsgBody, closeConn bool) {
	requests := c.m.requests.MustCurryWith(prometheus.Labels{"opcode": reqHeader.OpCode.String()})
	var command string
	var result *string
	defer func() {
		if result == nil {
			result = pointer.ToString("panic")
		}
		c.m.responses.WithLabelValues(resHeader.OpCode.String(), command, *result).Inc()
	}()

	resHeader = new(wire.MsgHeader)
	var err error
	switch reqHeader.OpCode {
	case wire.OP_MSG:
		msg := reqBody.(*wire.OpMsg)
		var document *types.Document
		document, err = msg.Document()
		command = document.Command()

		if err == nil {
			resHeader.OpCode = wire.OP_MSG
			resBody, err = c.handleOpMsg(ctx, msg, command)
		}

	case wire.OP_QUERY:
		query := reqBody.(*wire.OpQuery)
		resHeader.OpCode = wire.OP_REPLY
		resBody, err = c.h.CmdQuery(ctx, query)

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
		err = lazyerrors.Errorf("unexpected OpCode %s", reqHeader.OpCode)
	}
	requests.WithLabelValues(command).Inc()

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
			closeConn = true
			result = pointer.ToString("unexpected")
			c.l.Error("Handler error for unexpected response opcode",
				zap.Error(err), zap.Stringer("opcode", resHeader.OpCode),
			)
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

	resHeader.RequestID = atomic.AddInt32(&c.lastRequestID, 1)
	resHeader.ResponseTo = reqHeader.RequestID

	if result == nil {
		result = pointer.ToString("ok")
	}

	// do not spend time dumping if we are not going to log it
	if c.l.Desugar().Core().Enabled(zap.DebugLevel) {
		c.l.Debugf("Response header: %s", resHeader)
		c.l.Debugf("Response message:\n%s\n\n\n", resBody)
	}
	return
}

func (c *conn) handleOpMsg(ctx context.Context, msg *wire.OpMsg, cmd string) (*wire.OpMsg, error) {
	if cmd, ok := common.Commands[cmd]; ok {
		if cmd.Handler != nil {
			return cmd.Handler(c.h, ctx, msg)
		}
	}

	errMsg := fmt.Sprintf("no such command: '%s'", cmd)
	return nil, common.NewErrorMsg(common.ErrCommandNotFound, errMsg)
}
