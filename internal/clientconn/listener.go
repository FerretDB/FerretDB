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
	"math/rand"
	"net"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handler"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/tlsutil"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Listener listens on one or multiple interfaces (TCP, Unix, TLS sockets)
// and accepts incoming client connections.
type Listener struct {
	*NewListenerOpts

	tcpListener  net.Listener
	unixListener net.Listener
	tlsListener  net.Listener

	tcpListenerReady  chan struct{}
	unixListenerReady chan struct{}
	tlsListenerReady  chan struct{}
}

// NewListenerOpts represents listener configuration.
type NewListenerOpts struct {
	TCP  string
	Unix string

	TLS         string
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string

	ProxyAddr        string
	ProxyTLSCertFile string
	ProxyTLSKeyFile  string
	ProxyTLSCAFile   string

	Mode           Mode
	Metrics        *connmetrics.ListenerMetrics
	Handler        *handler.Handler
	Logger         *zap.Logger
	TestRecordsDir string // if empty, no records are created
}

// NewListener returns a new listener, configured by the NewListenerOpts argument.
func NewListener(opts *NewListenerOpts) *Listener {
	return &Listener{
		NewListenerOpts:   opts,
		tcpListenerReady:  make(chan struct{}),
		unixListenerReady: make(chan struct{}),
		tlsListenerReady:  make(chan struct{}),
	}
}

// Wait blocks until all of the expected listeners have started or returned the error.
func (l *Listener) Wait() {
	var wg sync.WaitGroup

	if l.TCP != "" {
		wg.Add(1)

		go func() {
			<-l.tcpListenerReady
			wg.Done()
		}()
	}

	if l.Unix != "" {
		wg.Add(1)

		go func() {
			<-l.unixListenerReady
			wg.Done()
		}()
	}

	if l.TLS != "" {
		wg.Add(1)

		go func() {
			<-l.tlsListenerReady
			wg.Done()
		}()
	}

	wg.Wait()
}

// Run runs the listener until ctx is canceled or some unrecoverable error occurs.
//
// When this method returns, listener and all connections, as well as handler are closed.
func (l *Listener) Run(ctx context.Context) error {
	defer l.Handler.Close()

	logger := l.Logger.Named("listener")

	if l.TCP != "" {
		var err error
		l.tcpListener, err = net.Listen("tcp", l.TCP)

		close(l.tcpListenerReady)

		if err != nil {
			return err
		}

		logger.Sugar().Infof("Listening on TCP %s ...", l.TCPAddr())
	}

	if l.Unix != "" {
		var err error
		l.unixListener, err = net.Listen("unix", l.Unix)

		close(l.unixListenerReady)

		if err != nil {
			return err
		}

		logger.Sugar().Infof("Listening on Unix %s ...", l.UnixAddr())
	}

	if l.TLS != "" {
		var err error
		l.tlsListener, err = setupTLSListener(&setupTLSListenerOpts{
			addr:     l.TLS,
			certFile: l.TLSCertFile,
			keyFile:  l.TLSKeyFile,
			caFile:   l.TLSCAFile,
		})

		close(l.tlsListenerReady)

		if err != nil {
			return err
		}

		logger.Sugar().Infof("Listening on TLS %s ...", l.TLSAddr())
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()

		if l.tcpListener != nil {
			l.tcpListener.Close()
		}

		if l.unixListener != nil {
			l.unixListener.Close()
		}

		if l.tlsListener != nil {
			l.tlsListener.Close()
		}
	}()

	if l.TCP != "" {
		wg.Add(1)

		go func() {
			defer func() {
				logger.Sugar().Infof("%s stopped.", l.TCPAddr())
				wg.Done()
			}()

			acceptLoop(ctx, l.tcpListener, &wg, l, logger)
		}()
	}

	if l.Unix != "" {
		wg.Add(1)

		go func() {
			defer func() {
				logger.Sugar().Infof("%s stopped.", l.UnixAddr())
				wg.Done()
			}()

			acceptLoop(ctx, l.unixListener, &wg, l, logger)
		}()
	}

	if l.TLS != "" {
		wg.Add(1)

		go func() {
			defer func() {
				logger.Sugar().Infof("%s stopped.", l.tlsListener.Addr())
				wg.Done()
			}()

			acceptLoop(ctx, l.tlsListener, &wg, l, logger)
		}()
	}

	<-ctx.Done()
	logger.Info("Waiting for all connections to stop...")
	wg.Wait()

	return context.Cause(ctx)
}

// setupTLSListenerOpts represents TLS listener setup options.
type setupTLSListenerOpts struct {
	addr     string
	certFile string
	keyFile  string
	caFile   string // may be empty to skip client's certificate validation
}

// setupTLSListener returns a new TLS listener or and error.
func setupTLSListener(opts *setupTLSListenerOpts) (net.Listener, error) {
	config, err := tlsutil.Config(opts.certFile, opts.keyFile, opts.caFile)
	if err != nil {
		return nil, err
	}

	listener, err := tls.Listen("tcp", opts.addr, config)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return listener, nil
}

// acceptLoop runs listener's connection accepting loop until context is canceled.
func acceptLoop(ctx context.Context, listener net.Listener, wg *sync.WaitGroup, l *Listener, logger *zap.Logger) {
	var retry int64
	for {
		netConn, err := listener.Accept()
		if err != nil {
			// Run closed listener on context cancellation
			if context.Cause(ctx) != nil {
				return
			}

			l.Metrics.Accepts.WithLabelValues("1").Inc()

			logger.Warn("Failed to accept connection", zap.Error(err))
			if !errors.Is(err, net.ErrClosed) {
				retry++
				ctxutil.SleepWithJitter(ctx, time.Second, retry)
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

			remoteAddr := netConn.RemoteAddr().String()
			if netConn.RemoteAddr().Network() == "unix" {
				// otherwise, all of them would be "" or "@"
				remoteAddr = fmt.Sprintf("unix:%d", rand.Int())
			}

			connID := fmt.Sprintf("%s -> %s", remoteAddr, netConn.LocalAddr())

			// give clients a few seconds to disconnect after ctx is canceled
			runCtx, runCancel := ctxutil.WithDelay(ctx.Done(), 3*time.Second)
			defer runCancel()

			defer pprof.SetGoroutineLabels(runCtx)
			runCtx = pprof.WithLabels(runCtx, pprof.Labels("conn", connID))
			pprof.SetGoroutineLabels(runCtx)

			opts := &newConnOpts{
				netConn:     netConn,
				mode:        l.Mode,
				l:           l.Logger.Named("// " + connID + " "), // derive from the original unnamed logger
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
				logger.Warn("Failed to create connection", zap.String("conn", connID), zap.Error(connErr))
				return
			}

			logger.Info("Connection started", zap.String("conn", connID))

			connErr = conn.run(runCtx)
			if errors.Is(connErr, wire.ErrZeroRead) {
				connErr = nil
				logger.Info("Connection stopped", zap.String("conn", connID))
			} else {
				logger.Warn("Connection stopped", zap.String("conn", connID), zap.Error(connErr))
			}
		}()
	}
}

// TCPAddr returns TCP listener's address.
// It can be used to determine an actually used port, if it was zero.
func (l *Listener) TCPAddr() net.Addr {
	<-l.tcpListenerReady
	return l.tcpListener.Addr()
}

// UnixAddr returns Unix domain socket listener's address.
func (l *Listener) UnixAddr() net.Addr {
	<-l.unixListenerReady
	return l.unixListener.Addr()
}

// TLSAddr returns TLS listener's address.
// It can be used to determine an actually used port, if it was zero.
func (l *Listener) TLSAddr() net.Addr {
	<-l.tlsListenerReady
	return l.tlsListener.Addr()
}

// Describe implements prometheus.Collector.
func (l *Listener) Describe(ch chan<- *prometheus.Desc) {
	l.Metrics.Describe(ch)
	l.Handler.Describe(ch)
}

// Collect implements prometheus.Collector.
func (l *Listener) Collect(ch chan<- prometheus.Metric) {
	l.Metrics.Collect(ch)
	l.Handler.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*Listener)(nil)
)
