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

// Package proxy sends requests to another wire protocol compatible service.
package proxy

import (
	"bufio"
	"context"
	"net"

	"github.com/FerretDB/FerretDB/internal/wire"
)

// Router "handles" messages by sending them to another wire protocol compatible service.
type Router struct {
	conn net.Conn
	bufr *bufio.Reader
	bufw *bufio.Writer
}

// New creates a new Router for a service with given address.
func New(addr string) (*Router, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Router{
		conn: conn,
		bufr: bufio.NewReader(conn),
		bufw: bufio.NewWriter(conn),
	}, nil
}

// Close stops the handler.
func (r *Router) Close() {
	r.conn.Close()
}

// Route routes the message by sending it to another wire protocol compatible service.
func (r *Router) Route(ctx context.Context, header *wire.MsgHeader, body wire.MsgBody) (*wire.MsgHeader, wire.MsgBody) {
	deadline, _ := ctx.Deadline()
	r.conn.SetDeadline(deadline)

	if err := wire.WriteMessage(r.bufw, header, body); err != nil {
		panic(err)
	}

	if err := r.bufw.Flush(); err != nil {
		panic(err)
	}

	resHeader, resBody, err := wire.ReadMessage(r.bufr)
	if err != nil {
		panic(err)
	}

	return resHeader, resBody
}
