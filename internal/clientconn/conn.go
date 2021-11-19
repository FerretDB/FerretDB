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

package clientconn

import (
	"bufio"
	"context"
	"fmt"
	"net"

	"github.com/pmezard/go-difflib/difflib"
	"go.uber.org/zap"

	"github.com/MangoDB-io/MangoDB/internal/handlers"
	"github.com/MangoDB-io/MangoDB/internal/handlers/jsonb1"
	"github.com/MangoDB-io/MangoDB/internal/handlers/proxy"
	"github.com/MangoDB-io/MangoDB/internal/handlers/shared"
	"github.com/MangoDB-io/MangoDB/internal/handlers/sql"
	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

type Mode string

const (
	// NormalMode only handles requests.
	NormalMode Mode = "normal"
	// ProxyMode only proxies requests to another wire protocol compatible service.
	ProxyMode Mode = "proxy"
	// DiffNormalMode both handles requests and proxies them, then logs the diff.
	// Only the MangoDB response is sent to the client.
	DiffNormalMode Mode = "diff-normal"
	// DiffProxyMode both handles requests and proxies them, then logs the diff.
	// Only the proxy response is sent to the client.
	DiffProxyMode Mode = "diff-proxy"
)

var AllModes = []Mode{NormalMode, ProxyMode, DiffNormalMode, DiffProxyMode}

type conn struct {
	netConn net.Conn
	mode    Mode
	h       *handlers.Handler
	proxy   *proxy.Handler
	l       *zap.SugaredLogger
}

func newConn(netConn net.Conn, pgPool *pg.Pool, proxyAddr string, mode Mode) (*conn, error) {
	prefix := fmt.Sprintf("// %s -> %s ", netConn.RemoteAddr(), netConn.LocalAddr())
	l := zap.L().Named(prefix)

	peerAddr := netConn.RemoteAddr().String()
	shared := shared.NewHandler(pgPool, peerAddr)
	sqlH := sql.NewStorage(pgPool, l.Sugar())
	jsonb1H := jsonb1.NewStorage(pgPool, l)

	var p *proxy.Handler
	if mode != NormalMode {
		var err error
		if p, err = proxy.New(proxyAddr); err != nil {
			return nil, err
		}
	}

	return &conn{
		netConn: netConn,
		mode:    mode,
		h:       handlers.New(pgPool, l, shared, sqlH, jsonb1H),
		proxy:   p,
		l:       l.Sugar(),
	}, nil
}

func (c *conn) run(ctx context.Context) (err error) {
	defer func() {
		if p := recover(); p != nil {
			// Log human-readable stack trace there (included in the error level automatically).
			c.l.Errorf("panic:\n%v\n(err = %v)", p, err)
			err = fmt.Errorf("recovered from panic (err = %v): %v", err, p)
		}
	}()

	deadline, _ := ctx.Deadline()
	c.netConn.SetDeadline(deadline)

	bufr := bufio.NewReader(c.netConn)
	bufw := bufio.NewWriter(c.netConn)
	defer func() {
		e := bufw.Flush()
		if err == nil {
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

		c.l.Infof("Request header:\n%s", wire.DumpMsgHeader(reqHeader))
		c.l.Infof("Request message:\n%s\n\n\n", wire.DumpMsgBody(reqBody))

		// handle request unless we are in proxy mode
		var resHeader *wire.MsgHeader
		var resBody wire.MsgBody
		if c.mode != ProxyMode {
			resHeader, resBody, err = c.h.Handle(ctx, reqHeader, reqBody)
			if err != nil {
				c.l.Infof("Response error: %s.", err)

				// TODO write handler.Error
				if c.mode == NormalMode {
					return
				}
			} else {
				c.l.Infof("Response header:\n%s", wire.DumpMsgHeader(resHeader))
				c.l.Infof("Response message:\n%s\n\n\n", wire.DumpMsgBody(resBody))
			}
		}

		// send request to proxy unless we are in normal mode
		var proxyHeader *wire.MsgHeader
		var proxyBody wire.MsgBody
		if c.mode != NormalMode {
			if c.proxy == nil {
				panic("proxy addr was nil")
			}

			proxyHeader, proxyBody, err = c.proxy.Handle(ctx, reqHeader, reqBody)
			if err != nil {
				c.l.Infof("Proxy error: %s.", err)
				return
			}

			c.l.Infof("Proxy header:\n%s", wire.DumpMsgHeader(proxyHeader))
			c.l.Infof("Proxt message:\n%s\n\n\n", wire.DumpMsgBody(proxyBody))
		}

		// diff in diff mode
		if c.mode == DiffNormalMode || c.mode == DiffProxyMode {
			res := difflib.SplitLines(wire.DumpMsgHeader(resHeader) + "\n" + wire.DumpMsgBody(resBody))
			proxy := difflib.SplitLines(wire.DumpMsgHeader(proxyHeader) + "\n" + wire.DumpMsgBody(proxyBody))
			diff := difflib.UnifiedDiff{
				A:        res,
				FromFile: "res",
				B:        proxy,
				ToFile:   "proxy",
				Context:  1,
			}
			var s string
			s, err = difflib.GetUnifiedDiffString(diff)
			if err != nil {
				return
			}

			c.l.Infof("Diff:\n%s\n\n\n", s)
		}

		// replace response with one from proxy in proxy and diff-proxy modes
		if c.mode == ProxyMode || c.mode == DiffProxyMode {
			resHeader = proxyHeader
			resBody = proxyBody
		}

		if resHeader == nil {
			c.l.Info("no response to send to client")
			return
		}

		if err = wire.WriteMessage(bufw, resHeader, resBody); err != nil {
			return
		}

		if err = bufw.Flush(); err != nil {
			return
		}
	}
}
