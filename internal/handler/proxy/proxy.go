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
	"crypto/x509"
	"net"
	"os"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
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

	if certFile != "" || keyFile != "" || caFile != "" {
		var config *tls.Config

		if config, err = tlsConfig(certFile, keyFile, caFile); err != nil {
			return nil, lazyerrors.Error(err)
		}

		// TODO https://github.com/FerretDB/FerretDB/issues/5049
		conn, err = dialTLS(context.TODO(), addr, config)
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

// tlsConfig provides TLS configuration for the given certificate and key files.
func tlsConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	var config tls.Config

	if certFile != "" || keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, lazyerrors.Errorf("Failed to load client TLS cert/key for proxy: %w", err)
		}

		config.Certificates = []tls.Certificate{cert}
	}

	if caFile != "" {
		b, err := os.ReadFile(caFile)
		if err != nil {
			return nil, lazyerrors.Errorf("Failed to read CA TLS cert for proxy: %w", err)
		}

		ca := x509.NewCertPool()
		if ok := ca.AppendCertsFromPEM(b); !ok {
			return nil, lazyerrors.New("Failed to parse CA TLS cert for proxy")
		}

		config.RootCAs = ca
	}

	return &config, nil
}

// dialTLS connects to the given address using TLS.
func dialTLS(ctx context.Context, addr string, config *tls.Config) (net.Conn, error) {
	d := &tls.Dialer{
		Config: config,
	}

	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = conn.(*tls.Conn).HandshakeContext(ctx); err != nil {
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

	if err := wire.WriteMessage(h.bufw, req.WireHeader(), req.WireBody()); err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := h.bufw.Flush(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	respHeader, respBody, err := wire.ReadMessage(h.bufr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return middleware.ResponseWire(respHeader, respBody), nil
}

// check interfaces
var (
	_ middleware.HandleFunc = (*Handler)(nil).Handle
)
