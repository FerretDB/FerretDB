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

// Package proxy handles requests by sending them to another wire protocol compatible service.
package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"net"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/tlsutil"
)

// Handler handles requests by sending them to another wire protocol compatible service.
type Handler struct {
	conn net.Conn
	bufr *bufio.Reader
	bufw *bufio.Writer
}

// New creates a new Handler for a service with given address.
func New(addr, certFile, keyFile, caFile string) (*Handler, error) {
	var conn net.Conn
	var err error

	if certFile != "" {
		conn, err = dialTLS(addr, certFile, keyFile, caFile)
	} else {
		conn, err = net.Dial("tcp", addr)
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Handler{
		conn: conn,
		bufr: bufio.NewReader(conn),
		bufw: bufio.NewWriter(conn),
	}, nil
}

// dialTLS connects to the given address using TLS.
func dialTLS(addr, certFile, keyFile, caFile string) (net.Conn, error) {
	config, err := tlsutil.Config(certFile, keyFile, caFile)
	if err != nil {
		return nil, err
	}

	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = conn.Handshake(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return conn, nil
}

// Run runs the handler until ctx is canceled.
//
// When this method returns, handler is stopped.
func (h *Handler) Run(ctx context.Context) {
	<-ctx.Done()
	_ = h.conn.Close()
}

// Handle processes a request by sending it to another wire protocol compatible service.
func (h *Handler) Handle(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
	deadline, _ := ctx.Deadline()
	_ = h.conn.SetDeadline(deadline)

	var body wire.MsgBody

	switch {
	case req.OpMsg != nil:
		body = req.OpMsg
	case req.OpQuery != nil:
		body = req.OpQuery
	default:
		return nil, lazyerrors.New("request body is nil")
	}

	if err := wire.WriteMessage(h.bufw, req.WireHeader(), body); err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := h.bufw.Flush(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	respHeader, respBody, err := wire.ReadMessage(h.bufr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return middleware.ResponseWire(respHeader, respBody)
}

// check interfaces
var (
	_ middleware.HandleFunc = (*Handler)(nil).Handle
)
