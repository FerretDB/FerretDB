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

// Package clientconn provides client connection implementation.
package clientconn

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/proxy"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Mode represents FerretDB mode of operation.
type Mode string

const (
	// NormalMode only handles requests.
	NormalMode Mode = "normal"
	// ProxyMode only proxies requests to another wire protocol compatible service.
	ProxyMode Mode = "proxy"
	// DiffNormalMode both handles requests and proxies them, then logs the diff.
	// Only the FerretDB response is sent to the client.
	DiffNormalMode Mode = "diff-normal"
	// DiffProxyMode both handles requests and proxies them, then logs the diff.
	// Only the proxy response is sent to the client.
	DiffProxyMode Mode = "diff-proxy"
)

// AllModes includes all operation modes, with the first one being the default.
var AllModes = []Mode{NormalMode, ProxyMode, DiffNormalMode, DiffProxyMode}

// conn represents client connection.
type conn struct {
	netConn       net.Conn
	mode          Mode
	l             *zap.SugaredLogger
	h             handlers.Interface
	m             *ConnMetrics
	proxy         *proxy.Router
	lastRequestID int32
}

// newConnOpts represents newConn options.
type newConnOpts struct {
	netConn     net.Conn
	mode        Mode
	l           *zap.Logger
	handler     handlers.Interface
	connMetrics *ConnMetrics
	proxyAddr   string
}

// newConn creates a new client connection for given net.Conn.
func newConn(opts *newConnOpts) (*conn, error) {
	if opts.handler == nil {
		panic("handler required")
	}

	var p *proxy.Router
	if opts.mode != NormalMode {
		var err error
		if p, err = proxy.New(opts.proxyAddr); err != nil {
			return nil, err
		}
	}

	return &conn{
		netConn: opts.netConn,
		mode:    opts.mode,
		l:       opts.l.Sugar(),
		h:       opts.handler,
		m:       opts.connMetrics,
		proxy:   p,
	}, nil
}

// run runs the client connection until ctx is done, client disconnects,
// or fatal error or panic is encountered.
//
// The caller is responsible for closing the underlying net.Conn.
func (c *conn) run(ctx context.Context) (err error) {
	done := make(chan struct{})

	// handle ctx cancellation
	go func() {
		select {
		case <-done:
			// nothing, let goroutine exit
		case <-ctx.Done():
			// unblocks ReadMessage below; any non-zero past value will do
			if e := c.netConn.SetDeadline(time.Unix(0, 0)); e != nil {
				c.l.Warnf("Failed to set deadline: %s", e)
			}
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			// Log human-readable stack trace there (included in the error level automatically).
			c.l.DPanicf("%v\n(err = %v)", p, err)
			err = errors.New("panic")
		}

		if err == nil {
			err = ctx.Err()
		}

		// let goroutine above exit
		close(done)
	}()

	bufr := bufio.NewReader(c.netConn)
	bufw := bufio.NewWriter(c.netConn)
	defer func() {
		if e := bufw.Flush(); err == nil {
			err = e
		}

		if c.proxy != nil {
			c.proxy.Close()
		}

		// c.netConn is closed by the caller
	}()

	for {
		var reqHeader *wire.MsgHeader
		var reqBody wire.MsgBody
		reqHeader, reqBody, err = wire.ReadMessage(bufr)
		if err != nil {
			return
		}

		c.l.Debugf("Request header: %s", reqHeader)
		c.l.Debugf("Request message:\n%s\n\n\n", reqBody)

		// diffLogLevel provides the level of logging for the diff between the "normal" and "proxy" responses.
		// It is set to the highest level of logging used to log response.
		var diffLogLevel zapcore.Level

		// handle request unless we are in proxy mode
		var resHeader *wire.MsgHeader
		var resBody wire.MsgBody
		var resCloseConn bool
		if c.mode != ProxyMode {
			resHeader, resBody, resCloseConn = c.route(ctx, reqHeader, reqBody)
			diffLogLevel = c.logResponse("Response", resHeader, resBody, resCloseConn)
		}

		// send request to proxy unless we are in normal mode
		var proxyHeader *wire.MsgHeader
		var proxyBody wire.MsgBody
		if c.mode != NormalMode {
			if c.proxy == nil {
				panic("proxy addr was nil")
			}

			proxyHeader, proxyBody, _ = c.proxy.Route(ctx, reqHeader, reqBody)
			if level := c.logResponse("Proxy response", proxyHeader, proxyBody, resCloseConn); level != diffLogLevel {
				// In principle, normal and proxy responses should be logged with the same level
				// as they behave the same way. If it's not true, there is a bug somewhere, so
				// we should log the diff as an error.
				diffLogLevel = zap.ErrorLevel
			}
		}

		// diff in diff mode
		if c.mode == DiffNormalMode || c.mode == DiffProxyMode {
			var diffHeader string
			diffHeader, err = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(resHeader.String()),
				FromFile: "res header",
				B:        difflib.SplitLines(proxyHeader.String()),
				ToFile:   "proxy header",
				Context:  1,
			})
			if err != nil {
				return
			}

			var diffBody string
			diffBody, err = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(resBody.String()),
				FromFile: "res body",
				B:        difflib.SplitLines(proxyBody.String()),
				ToFile:   "proxy body",
				Context:  1,
			})
			if err != nil {
				return
			}

			c.l.Desugar().Check(diffLogLevel, fmt.Sprintf("Header diff:\n%s\nBody diff:\n%s\n\n", diffHeader, diffBody)).Write()
		}

		// replace response with one from proxy in proxy and diff-proxy modes
		if c.mode == ProxyMode || c.mode == DiffProxyMode {
			resHeader = proxyHeader
			resBody = proxyBody
		}

		if resHeader == nil || resBody == nil {
			panic("no response to send to client")
		}

		if err = wire.WriteMessage(bufw, resHeader, resBody); err != nil {
			return
		}

		if err = bufw.Flush(); err != nil {
			return
		}

		if resCloseConn {
			err = errors.New("fatal error")
			return
		}
	}
}

// route sends request to a handler's command based on the op code provided in the request header.
//
// The possible resBody returns:
//   - normal response  - to be returned to the client, closeConn is false;
//   - protocol error (*common.Error, possibly wrapped) - to be returned to the client, closeConn is false;
//   - any other error - to be returned to the client as InternalError before terminating connection, closeConn is true.
//
// Handlers to which it routes, should not panic on bad input, but may do so in "impossible" cases.
// They also should not use recover(). That allows us to use fuzzing.
func (c *conn) route(ctx context.Context, reqHeader *wire.MsgHeader, reqBody wire.MsgBody) (resHeader *wire.MsgHeader, resBody wire.MsgBody, closeConn bool) { //nolint:lll // argument list is too long
	requests := c.m.requests.MustCurryWith(prometheus.Labels{"opcode": reqHeader.OpCode.String()})
	var command string
	var result *string
	defer func() {
		if result == nil {
			result = pointer.ToString("panic")
		}
		c.m.responses.WithLabelValues(resHeader.OpCode.String(), command, *result).Inc()
	}()

	connInfo := &conninfo.ConnInfo{
		PeerAddr:          c.netConn.RemoteAddr(),
		AggregationStages: c.m.aggregationStages,
	}
	ctx, cancel := context.WithCancel(conninfo.WithConnInfo(ctx, connInfo))
	defer cancel()

	resHeader = new(wire.MsgHeader)
	var err error
	switch reqHeader.OpCode {
	case wire.OpCodeMsg:
		var document *types.Document
		msg := reqBody.(*wire.OpMsg)
		document, err = msg.Document()

		command = document.Command()
		if err == nil {
			resHeader.OpCode = wire.OpCodeMsg
			resBody, err = c.handleOpMsg(ctx, msg, command)
		}

	case wire.OpCodeQuery:
		query := reqBody.(*wire.OpQuery)
		resHeader.OpCode = wire.OpCodeReply
		resBody, err = c.h.CmdQuery(ctx, query)

	case wire.OpCodeReply:
		fallthrough
	case wire.OpCodeUpdate:
		fallthrough
	case wire.OpCodeInsert:
		fallthrough
	case wire.OpCodeGetByOID:
		fallthrough
	case wire.OpCodeGetMore:
		fallthrough
	case wire.OpCodeDelete:
		fallthrough
	case wire.OpCodeKillCursors:
		fallthrough
	case wire.OpCodeCompressed:
		err = lazyerrors.Errorf("unhandled OpCode %s", reqHeader.OpCode)

	default:
		err = lazyerrors.Errorf("unexpected OpCode %s", reqHeader.OpCode)
	}

	requests.WithLabelValues(command).Inc()

	// set body for error
	if err != nil {
		switch resHeader.OpCode {
		case wire.OpCodeMsg:
			protoErr, recoverable := common.ProtocolError(err)
			closeConn = !recoverable
			var res wire.OpMsg
			must.NoError(res.SetSections(wire.OpMsgSection{
				Documents: []*types.Document{protoErr.Document()},
			}))
			if resBody != nil {
				// TODO: find a way to somehow append it to response
				// use it if possible!
				// resBody.(wire.OpMsg).Document()
			} else {
				resBody = &res
			}
			result = pointer.ToString(protoErr.Code().String())

		case wire.OpCodeQuery:
			fallthrough
		case wire.OpCodeReply:
			fallthrough
		case wire.OpCodeUpdate:
			fallthrough
		case wire.OpCodeInsert:
			fallthrough
		case wire.OpCodeGetByOID:
			fallthrough
		case wire.OpCodeGetMore:
			fallthrough
		case wire.OpCodeDelete:
			fallthrough
		case wire.OpCodeKillCursors:
			fallthrough
		case wire.OpCodeCompressed:
			// do not panic to make fuzzing easier
			closeConn = true
			result = pointer.ToString("unhandled")
			c.l.Error(
				"Handler error for unhandled response opcode",
				zap.Error(err), zap.Stringer("opcode", resHeader.OpCode),
			)
			return

		default:
			// do not panic to make fuzzing easier
			closeConn = true
			result = pointer.ToString("unexpected")
			c.l.Error(
				"Handler error for unexpected response opcode",
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

// Describe implements prometheus.Collector.
func (c *conn) Describe(ch chan<- *prometheus.Desc) {
	c.m.Describe(ch)
}

// Collect implements prometheus.Collector.
func (c *conn) Collect(ch chan<- prometheus.Metric) {
	c.m.Collect(ch)
}

// logResponse logs response's header and body and returns the log level that was used.
//
// The param `who` will be used in logs and should represent the type of the response,
// for example "Response" or "Proxy Response".
//
// If response op code is not `OP_MSG`, it always logs as a debug.
// For the `OP_MSG` code, the level depends on the type of error.
// If there is no errors in the response, it will be logged as a debug.
// If there is an error in the response, and connection is closed, it will be logged as an error.
// If there is an error in the response, and connection is not closed, it will be logged as a warning.
func (c *conn) logResponse(who string, resHeader *wire.MsgHeader, resBody wire.MsgBody, closeConn bool) zapcore.Level {
	level := zap.DebugLevel

	if resHeader.OpCode == wire.OpCodeMsg {
		doc := must.NotFail(resBody.(*wire.OpMsg).Document())

		ok := must.NotFail(doc.Get("ok"))

		if ok.(float64) != 1 {
			if closeConn {
				level = zap.ErrorLevel
			} else {
				level = zap.WarnLevel
			}
		}
	}

	c.l.Desugar().Check(level, fmt.Sprintf("%s header: %s", who, resHeader)).Write()
	c.l.Desugar().Check(level, fmt.Sprintf("%s message:\n%s\n\n\n", who, resBody)).Write()

	return level
}
