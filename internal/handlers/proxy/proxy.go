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

// Package proxy sends requests to another wire protocol compatible service.
package proxy

import (
	"bufio"
	"context"
	"net"

	"github.com/MangoDB-io/MangoDB/internal/wire"
)

// Handler "handles" messages by sending them to another wire protocol compatible service.
type Handler struct {
	conn net.Conn
	bufr *bufio.Reader
	bufw *bufio.Writer
}

// New creates a new Handler for a service with given address.
func New(addr string) (*Handler, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Handler{
		conn: conn,
		bufr: bufio.NewReader(conn),
		bufw: bufio.NewWriter(conn),
	}, nil
}

// Close stops the handler.
func (h *Handler) Close() {
	h.bufw.Flush()
	h.conn.Close()
}

// Handle "handles" the message by sending it to another wire protocol compatible service.
func (h *Handler) Handle(ctx context.Context, header *wire.MsgHeader, msg wire.MsgBody) (*wire.MsgHeader, wire.MsgBody, error) {
	deadline, _ := ctx.Deadline()
	h.conn.SetDeadline(deadline)

	if err := wire.WriteMessage(h.bufw, header, msg); err != nil {
		return nil, nil, err
	}

	if err := h.bufw.Flush(); err != nil {
		return nil, nil, err
	}

	return wire.ReadMessage(h.bufr)
}
