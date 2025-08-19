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

package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// conn represents a single connection to a wire protocol compatible service.
// It can be used concurrently from multiple goroutines.
type conn struct {
	// the order of fields is weird to make the struct smaller due to alignment

	c    net.Conn      // protected by m
	bufr *bufio.Reader // protected by m
	bufw *bufio.Writer // protected by m
	l    *slog.Logger
	m    sync.Mutex
}

// newConn creates a new connection.
// Context cancellation stops dialing, but does not affect established connection.
func newConn(ctx context.Context, opts *NewOpts) (*conn, error) {
	var c net.Conn
	var err error

	if opts.TLSCertFile == "" && opts.TLSKeyFile == "" && opts.TLSCAFile == "" {
		c, err = dialTCP(ctx, opts.Addr)
	} else {
		c, err = dialTLS(ctx, opts.Addr, opts.TLSCertFile, opts.TLSKeyFile, opts.TLSCAFile)
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &conn{
		l: opts.L.With(
			slog.String("local", c.LocalAddr().String()),
			slog.String("remote", c.RemoteAddr().String()),
		),

		c:    c,
		bufw: bufio.NewWriter(c),
		bufr: bufio.NewReader(c),
	}, nil
}

// dialTCP connects to the given address using TCP.
func dialTCP(ctx context.Context, addr string) (net.Conn, error) {
	var d net.Dialer

	c, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return c, nil
}

// dialTLS connects to the given address using TLS.
func dialTLS(ctx context.Context, addr, certFile, keyFile, caFile string) (net.Conn, error) {
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

	d := &tls.Dialer{
		Config: &config,
	}

	c, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = c.(*tls.Conn).HandshakeContext(ctx); err != nil {
		_ = c.Close()
		return nil, lazyerrors.Error(err)
	}

	return c, nil
}

// handle sends a single request to the proxy.
// It can be called concurrently from multiple goroutines.
func (c *conn) handle(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
	// It is not clear if clients actually send multiple requests in parallel over the same connection.
	// If they do, we better support that, too.
	// TODO https://github.com/FerretDB/FerretDB/issues/5049
	if !c.m.TryLock() {
		c.l.WarnContext(ctx, "Connection is busy, waiting for lock")
		c.m.Lock()
	}
	defer c.m.Unlock()

	if ctx.Err() != nil {
		return nil, lazyerrors.Error(ctx.Err())
	}

	deadline, _ := ctx.Deadline()
	_ = c.c.SetDeadline(deadline)

	if err := wire.WriteMessage(c.bufw, req.WireHeader(), req.WireBody()); err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := c.bufw.Flush(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	header, body, err := wire.ReadMessage(c.bufr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resp, err := middleware.ResponseWire(header, body)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return resp, nil
}

// close closes the connection.
func (c *conn) close() {
	_ = c.c.Close()
}
