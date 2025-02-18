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
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/FerretDB/wire"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/tlsutil"
)

// Listener listens on one or multiple interfaces (TCP, Unix, TLS sockets)
// and accepts incoming client connections.
type Listener struct {
	*ListenerOpts

	ll *slog.Logger

	tcpListener  net.Listener
	unixListener net.Listener
	tlsListener  net.Listener

	listenersClosed chan struct{}
}

// ListenerOpts represents listener configuration.
type ListenerOpts struct {
	Handler *handler.Handler
	Metrics *connmetrics.ListenerMetrics
	Logger  *slog.Logger

	TCP  string // empty value disables TCP listener
	Unix string // empty value disables Unix listener

	TLS         string // empty value disables TLS listener
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string

	Mode             Mode
	ProxyAddr        string
	ProxyTLSCertFile string
	ProxyTLSKeyFile  string
	ProxyTLSCAFile   string

	TestRecordsDir string // if empty, no records are created
}

// Listen creates a new listener and starts listening on configured interfaces.
// [Listener.Run] must be collect on the returned value.
func Listen(opts *ListenerOpts) (l *Listener, err error) {
	ll := logging.WithName(opts.Logger, "listener")
	l = &Listener{
		ListenerOpts:    opts,
		ll:              ll,
		listenersClosed: make(chan struct{}),
	}

	defer func() {
		if err != nil {
			l.close()
			l = nil
		}
	}()

	ctx := context.Background()

	if l.TCP != "" {
		if l.tcpListener, err = net.Listen("tcp", l.TCP); err != nil {
			err = lazyerrors.Error(err)
			return
		}

		ll.InfoContext(ctx, fmt.Sprintf("Listening on TCP %s...", l.TCPAddr()))
	}

	if l.Unix != "" {
		if l.unixListener, err = net.Listen("unix", l.Unix); err != nil {
			err = lazyerrors.Error(err)
			return
		}

		ll.InfoContext(ctx, fmt.Sprintf("Listening on Unix %s...", l.UnixAddr()))
	}

	if l.TLS != "" {
		var config *tls.Config

		if config, err = tlsutil.Config(l.TLSCertFile, l.TLSKeyFile, l.TLSCAFile); err != nil {
			err = lazyerrors.Error(err)
			return
		}

		if l.tlsListener, err = tls.Listen("tcp", l.TLS, config); err != nil {
			err = lazyerrors.Error(err)
			return
		}

		ll.InfoContext(ctx, fmt.Sprintf("Listening on TLS %s...", l.TLSAddr()))
	}

	return
}

// close closes all listeners.
func (l *Listener) close() {
	if l.tcpListener != nil {
		_ = l.tcpListener.Close()
	}

	if l.unixListener != nil {
		_ = l.unixListener.Close()
	}

	if l.tlsListener != nil {
		_ = l.tlsListener.Close()
	}

	close(l.listenersClosed)
}

// Listening returns true if the listener is currently listening and accepting new connection.
//
// It returns false when listener is stopped
// or when it is still running with established connections.
func (l *Listener) Listening() bool {
	select {
	case <-l.listenersClosed:
		return false
	default:
		return true
	}
}

// Run runs the listener (and handler) until ctx is canceled.
//
// When this method returns, listener and all connections are closed, and handler is stopped.
func (l *Listener) Run(ctx context.Context) {
	// inherit ctx's values
	handlerCtx, handlerCancel := context.WithCancel(context.WithoutCancel(ctx))
	handlerDone := make(chan struct{})

	go func() {
		defer close(handlerDone)
		l.Handler.Run(handlerCtx)
	}()

	var wg sync.WaitGroup

	if l.tcpListener != nil {
		wg.Add(1)

		go func() {
			defer func() {
				l.ll.InfoContext(ctx, fmt.Sprintf("%s stopped", l.TCPAddr()))
				wg.Done()
			}()

			acceptLoop(ctx, l.tcpListener, &wg, l)
		}()
	}

	if l.unixListener != nil {
		wg.Add(1)

		go func() {
			defer func() {
				l.ll.InfoContext(ctx, fmt.Sprintf("%s stopped", l.UnixAddr()))
				wg.Done()
			}()

			acceptLoop(ctx, l.unixListener, &wg, l)
		}()
	}

	if l.tlsListener != nil {
		wg.Add(1)

		go func() {
			defer func() {
				l.ll.InfoContext(ctx, fmt.Sprintf("%s stopped", l.TLSAddr()))
				wg.Done()
			}()

			acceptLoop(ctx, l.tlsListener, &wg, l)
		}()
	}

	<-ctx.Done()
	l.close()
	l.ll.InfoContext(ctx, "Waiting for all connections to close")
	wg.Wait()

	// to properly handle last client commands like endSession,
	// stop handler only after the last client disconnects
	handlerCancel()
	<-handlerDone
}

// acceptLoop runs listener's connection accepting loop until context is canceled.
func acceptLoop(ctx context.Context, listener net.Listener, wg *sync.WaitGroup, l *Listener) {
	var attempt int64
	for {
		netConn, err := listener.Accept()
		if err != nil {
			// Run closed listener on context cancellation
			if context.Cause(ctx) != nil {
				return
			}

			l.Metrics.Accepts.WithLabelValues("1").Inc()

			l.ll.WarnContext(ctx, "Failed to accept connection", logging.Error(err))
			if !errors.Is(err, net.ErrClosed) {
				attempt++
				ctxutil.SleepWithJitter(ctx, time.Second, attempt)
			}
			continue
		}

		wg.Add(1)
		l.Metrics.Accepts.WithLabelValues("0").Inc()

		go func() {
			var connErr error
			start := time.Now()

			defer func() {
				lv := "0"
				if connErr != nil {
					lv = "1"
				}

				l.Metrics.Durations.WithLabelValues(lv).Observe(time.Since(start).Seconds())
				netConn.Close()
				wg.Done()
			}()

			// give already connected clients a few seconds to gracefully disconnect
			connCtx, connCancel := ctxutil.WithDelay(ctx)
			defer connCancel(nil)

			remoteAddr := netConn.RemoteAddr().String()
			if netConn.RemoteAddr().Network() == "unix" {
				// otherwise, all of them would be "" or "@"
				remoteAddr = fmt.Sprintf("unix:%d", rand.Int())
			}

			connID := fmt.Sprintf("%s -> %s", remoteAddr, netConn.LocalAddr())

			opts := &newConnOpts{
				netConn:     netConn,
				mode:        l.Mode,
				l:           logging.WithName(l.ll, "// "+connID+" "), // derive from the original unnamed logger
				handler:     l.Handler,
				connMetrics: l.Metrics.ConnMetrics, // share between all conns

				proxyAddr:        l.ProxyAddr,
				proxyTLSCertFile: l.ProxyTLSCertFile,
				proxyTLSKeyFile:  l.ProxyTLSKeyFile,
				proxyTLSCAFile:   l.ProxyTLSCAFile,

				testRecordsDir: l.TestRecordsDir,
			}

			conn, connErr := newConn(opts)
			if connErr != nil {
				l.ll.WarnContext(connCtx, "Failed to create connection", slog.String("conn", connID), logging.Error(connErr))
				return
			}

			l.ll.InfoContext(ctx, "Connection started", slog.String("conn", connID))

			connErr = conn.run(connCtx)
			if errors.Is(connErr, wire.ErrZeroRead) {
				connErr = nil

				l.ll.InfoContext(ctx, "Connection stopped", slog.String("conn", connID))
			} else {
				l.ll.WarnContext(ctx, "Connection stopped", slog.String("conn", connID), logging.Error(connErr))
			}
		}()
	}
}

// TCPAddr returns TCP listener's address, or nil, if TCP listener is disabled.
// It can be used to determine an actually used port, if it was zero.
func (l *Listener) TCPAddr() net.Addr {
	if l.tcpListener == nil {
		return nil
	}

	return l.tcpListener.Addr()
}

// UnixAddr returns Unix domain socket listener's address, or nil, if Unix listener is disabled.
func (l *Listener) UnixAddr() net.Addr {
	if l.unixListener == nil {
		return nil
	}

	return l.unixListener.Addr()
}

// TLSAddr returns TLS listener's address, or nil, if TLS listener is disabled.
// It can be used to determine an actually used port, if it was zero.
func (l *Listener) TLSAddr() net.Addr {
	if l.tlsListener == nil {
		return nil
	}

	return l.tlsListener.Addr()
}

// Describe implements [prometheus.Collector].
func (l *Listener) Describe(ch chan<- *prometheus.Desc) {
	l.Metrics.Describe(ch)
	l.Handler.Describe(ch)
}

// Collect implements [prometheus.Collector].
func (l *Listener) Collect(ch chan<- prometheus.Metric) {
	l.Metrics.Collect(ch)
	l.Handler.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*Listener)(nil)
)
