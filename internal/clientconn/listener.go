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
	"crypto/x509"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Listener accepts incoming client connections.
type Listener struct {
	*NewListenerOpts
	tcpListener       net.Listener
	unixListener      net.Listener
	tlsListener       net.Listener
	tcpListenerReady  chan struct{}
	unixListenerReady chan struct{}
	tlsListenerReady  chan struct{}
}

// NewListenerOpts represents listener configuration.
type NewListenerOpts struct {
	Listener       ListenerOpts
	ProxyAddr      string
	Mode           Mode
	Metrics        *connmetrics.ListenerMetrics
	Handler        handlers.Interface
	Logger         *zap.Logger
	TestRecordsDir string // if empty, no records are created
}

// ListenerOpts represents listener configuration options.
type ListenerOpts struct {
	Addr        string
	Unix        string
	TLS         string
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string
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

// Run runs the listener until ctx is done or some unrecoverable error occurs.
//
// When this method returns, listener and all connections are closed.
func (l *Listener) Run(ctx context.Context) error {
	logger := l.Logger.Named("listener")

	if l.Listener.Addr != "" {
		var err error
		if l.tcpListener, err = net.Listen("tcp", l.Listener.Addr); err != nil {
			return err
		}

		close(l.tcpListenerReady)

		logger.Sugar().Infof("Listening on TCP %s ...", l.Addr())
	}

	if l.Listener.Unix != "" {
		var err error
		if l.unixListener, err = net.Listen("unix", l.Listener.Unix); err != nil {
			return err
		}

		close(l.unixListenerReady)

		logger.Sugar().Infof("Listening on Unix %s ...", l.Unix())
	}

	if l.Listener.TLS != "" {
		var err error
		if l.tlsListener, err = setupTLSListener(&setupTLSListenerOpts{
			addr:     l.Listener.TLS,
			certFile: l.Listener.TLSCertFile,
			keyFile:  l.Listener.TLSKeyFile,
			caFile:   l.Listener.TLSCAFile,
		}); err != nil {
			return err
		}

		close(l.tlsListenerReady)

		logger.Sugar().Infof("Listening on TLS %s ...", l.TLS())
	}

	// close listeners on context cancellation to exit from listenLoop
	go func() {
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

	var wg sync.WaitGroup

	if l.Listener.Addr != "" {
		wg.Add(1)

		go func() {
			defer func() {
				logger.Sugar().Infof("%s stopped.", l.Addr())
				wg.Done()
			}()

			acceptLoop(ctx, l.tcpListener, &wg, l, logger)
		}()
	}

	if l.Listener.Unix != "" {
		wg.Add(1)

		go func() {
			defer func() {
				logger.Sugar().Infof("%s stopped.", l.Unix())
				wg.Done()
			}()

			acceptLoop(ctx, l.unixListener, &wg, l, logger)
		}()
	}

	if l.Listener.TLS != "" {
		wg.Add(1)

		go func() {
			defer func() {
				logger.Sugar().Infof("%s stopped.", l.tlsListener.Addr())
				wg.Done()
			}()

			acceptLoop(ctx, l.tlsListener, &wg, l, logger)
		}()
	}

	logger.Info("Waiting for all connections to stop...")
	wg.Wait()

	return ctx.Err()
}

// setupTLSListenerOpts represents TLS listener setup options.
type setupTLSListenerOpts struct {
	addr, certFile, keyFile, caFile string
}

// setupTLSListener returns a new TLS listener or and error.
func setupTLSListener(opts *setupTLSListenerOpts) (net.Listener, error) {
	if _, err := os.Stat(opts.certFile); err != nil {
		return nil, fmt.Errorf("TLS certificate file: %w", err)
	}

	if _, err := os.Stat(opts.keyFile); err != nil {
		return nil, fmt.Errorf("TLS key file: %w", err)
	}

	cert, err := tls.LoadX509KeyPair(opts.certFile, opts.keyFile)
	if err != nil {
		return nil, err
	}

	config := tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if opts.caFile != "" {
		if _, err = os.Stat(opts.caFile); err != nil {
			return nil, fmt.Errorf("TLS CA file: %w", err)
		}

		var rootCA []byte

		rootCA, err = os.ReadFile(opts.caFile)
		if err != nil {
			return nil, err
		}

		roots := x509.NewCertPool()

		ok := roots.AppendCertsFromPEM(rootCA)
		if !ok {
			return nil, fmt.Errorf("Failed to parse root certificate")
		}

		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = roots
	}

	listener, err := tls.Listen("tcp", opts.addr, &config)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return listener, nil
}

// acceptLoop runs listener's connection accepting loop.
func acceptLoop(ctx context.Context, listener net.Listener, wg *sync.WaitGroup, l *Listener, logger *zap.Logger) {
	for {
		netConn, err := listener.Accept()
		if err != nil {
			// Run closed listener on context cancellation
			if ctx.Err() != nil {
				break
			}

			l.Metrics.Accepts.WithLabelValues("1").Inc()

			logger.Warn("Failed to accept connection", zap.Error(err))
			if !errors.Is(err, net.ErrClosed) {
				ctxutil.Sleep(ctx, time.Second)
			}
			continue
		}

		wg.Add(1)
		l.Metrics.Accepts.WithLabelValues("0").Inc()
		l.Metrics.ConnectedClients.Inc()

		go func() {
			defer func() {
				l.Metrics.ConnectedClients.Dec()
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
				netConn:        netConn,
				mode:           l.Mode,
				l:              l.Logger.Named("// " + connID + " "), // derive from the original unnamed logger
				handler:        l.Handler,
				connMetrics:    l.Metrics.ConnMetrics,
				proxyAddr:      l.ProxyAddr,
				testRecordsDir: l.TestRecordsDir,
			}
			conn, e := newConn(opts)
			if e != nil {
				logger.Warn("Failed to create connection", zap.String("conn", connID), zap.Error(e))
				return
			}

			logger.Info("Connection started", zap.String("conn", connID))

			e = conn.run(runCtx)
			if errors.Is(e, wire.ErrZeroRead) {
				logger.Info("Connection stopped", zap.String("conn", connID))
			} else {
				logger.Warn("Connection stopped", zap.String("conn", connID), zap.Error(e))
			}
		}()
	}
}

// Addr returns TCP listener's address.
// It can be used to determine an actually used port, if it was zero.
func (l *Listener) Addr() net.Addr {
	<-l.tcpListenerReady
	return l.tcpListener.Addr()
}

// Unix returns Unix domain socket address.
func (l *Listener) Unix() net.Addr {
	<-l.unixListenerReady
	return l.unixListener.Addr()
}

// TLS returns TLS listener address.
func (l *Listener) TLS() net.Addr {
	<-l.tlsListenerReady
	return l.tlsListener.Addr()
}

// Describe implements prometheus.Collector.
func (l *Listener) Describe(ch chan<- *prometheus.Desc) {
	l.Metrics.Describe(ch)
}

// Collect implements prometheus.Collector.
func (l *Listener) Collect(ch chan<- prometheus.Metric) {
	l.Metrics.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*Listener)(nil)
)
