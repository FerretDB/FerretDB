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
	"github.com/MangoDB-io/MangoDB/internal/handlers/shared"
	"github.com/MangoDB-io/MangoDB/internal/handlers/sql"
	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/shadow"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

type Mode string

const (
	normalMode Mode = "normal"
	proxyMode  Mode = "proxy"
	diffMode   Mode = "diff"
)

var AllModes = []Mode{normalMode, proxyMode, diffMode}

type conn struct {
	netConn net.Conn
	mode    Mode
	h       *handlers.Handler
	s       *shadow.Handler
	l       *zap.SugaredLogger
}

func newConn(netConn net.Conn, pgPool *pg.Pool, shadowAddr string, mode Mode) (*conn, error) {
	prefix := fmt.Sprintf("// %s -> %s ", netConn.RemoteAddr(), netConn.LocalAddr())
	l := zap.L().Named(prefix)

	peerAddr := netConn.RemoteAddr().String()
	shared := shared.NewHandler(pgPool, peerAddr)
	sqlH := sql.NewStorage(pgPool, l.Sugar())
	jsonb1H := jsonb1.NewStorage(pgPool, l)

	var s *shadow.Handler
	if mode != normalMode {
		var err error
		if s, err = shadow.New(shadowAddr); err != nil {
			return nil, err
		}
	}

	return &conn{
		netConn: netConn,
		mode:    mode,
		h:       handlers.New(pgPool, l, shared, sqlH, jsonb1H),
		s:       s,
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

	bufr := bufio.NewReader(c.netConn)
	bufw := bufio.NewWriter(c.netConn)
	defer func() {
		e := bufw.Flush()
		if err == nil {
			err = e
		}
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

		var resHeader *wire.MsgHeader
		var resBody wire.MsgBody
		if c.mode != proxyMode {
			resHeader, resBody, err = c.h.Handle(ctx, reqHeader, reqBody)
			if err != nil {
				c.l.Infof("Response error: %s.", err)

				// TODO write handler.Error
				if c.mode == normalMode {
					return
				}
			} else {
				c.l.Infof("Response header:\n%s", wire.DumpMsgHeader(resHeader))
				c.l.Infof("Response message:\n%s\n\n\n", wire.DumpMsgBody(resBody))
			}
		}

		var shadowHeader *wire.MsgHeader
		var shadowBody wire.MsgBody
		if c.mode != normalMode {
			if c.s == nil {
				panic("shadow addr was nil")
			}

			shadowHeader, shadowBody, err = c.s.Handle(ctx, reqHeader, reqBody)
			if err != nil {
				c.l.Infof("Shadow error: %s.", err)
				return
			}

			c.l.Infof("Shadow header:\n%s", wire.DumpMsgHeader(shadowHeader))
			c.l.Infof("Shadow message:\n%s\n\n\n", wire.DumpMsgBody(shadowBody))
		}

		if c.mode == diffMode {
			res := difflib.SplitLines(wire.DumpMsgHeader(resHeader) + "\n" + wire.DumpMsgBody(resBody))
			shadow := difflib.SplitLines(wire.DumpMsgHeader(shadowHeader) + "\n" + wire.DumpMsgBody(shadowBody))
			diff := difflib.UnifiedDiff{
				A:        res,
				FromFile: "res",
				B:        shadow,
				ToFile:   "shadow",
				Context:  1,
			}
			var s string
			s, err = difflib.GetUnifiedDiffString(diff)
			if err != nil {
				return
			}

			c.l.Infof("Diff:\n%s\n\n\n", s)
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
