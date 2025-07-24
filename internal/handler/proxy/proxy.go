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
	"log/slog"
	"net"
	"os"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Handler handles requests by sending them to another wire protocol compatible service.
type Handler struct {
	conn net.Conn
	bufr *bufio.Reader
	bufw *bufio.Writer
}

type NewOpts struct {
	Addr     string
	CertFile string
	KeyFile  string
	CAFile   string

	L *slog.Logger
}

// New creates a new Handler for a service with given address.
func New(opts *NewOpts) (*Handler, error) {
	must.NotBeZero(opts)

	var conn net.Conn
	var err error

	if opts.CertFile != "" || opts.KeyFile != "" || opts.CAFile != "" {
		var config *tls.Config

		if config, err = tlsConfig(opts.CertFile, opts.KeyFile, opts.CAFile); err != nil {
			return nil, lazyerrors.Error(err)
		}

		// TODO https://github.com/FerretDB/FerretDB/issues/5049
		conn, err = dialTLS(context.TODO(), opts.Addr, config)
	} else {
		conn, err = net.Dial("tcp", opts.Addr)
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

// tlsConfig provides client TLS configuration for the given certificate and key files.
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

	err := wire.WriteMessage(h.bufw, req.WireHeader(), req.WireBody())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = h.bufw.Flush(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	header, body, err := wire.ReadMessage(h.bufr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resp, err := middleware.ResponseWire(header, body)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return resp, nil
}

// check interfaces
var (
	_ middleware.Handler = (*Handler)(nil)
)
