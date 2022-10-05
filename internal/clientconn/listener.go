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
	"errors"
	"fmt"
	"io"
	"net"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Listener accepts incoming client connections.
type Listener struct {
	opts         *NewListenerOpts
	metrics      *ListenerMetrics
	handler      handlers.Interface
	tcpListener  net.Listener
	sockListener net.Listener
	listening    chan struct{}
}

// NewListenerOpts represents listener configuration.
type NewListenerOpts struct {
	ListenAddr         string
	ListenSock         string
	BindToSock         bool
	ProxyAddr          string
	Mode               Mode
	Handler            handlers.Interface
	Logger             *zap.Logger
	TestConnTimeout    time.Duration
	TestRunCancelDelay time.Duration
	TestRecordPath     string // if empty, no records are created
}

// Default IP we are listening on
const defaultListeningIP = "127.0.0.1:27017"

// NewListener returns a new listener, configured by the NewListenerOpts argument.
func NewListener(opts *NewListenerOpts) *Listener {
	return &Listener{
		opts:      opts,
		metrics:   newListenerMetrics(),
		handler:   opts.Handler,
		listening: make(chan struct{}),
	}
}

// Run runs the listener until ctx is done or some unrecoverable error occurs.
//
// When this method returns, listener and all connections are closed.
func (l *Listener) Run(ctx context.Context) error {
	logger := l.opts.Logger.Named("listener")

	useSock := l.opts.ListenSock != "" || l.opts.BindToSock
	useTcp := !useSock || l.opts.ListenAddr != defaultListeningIP

	if useTcp {
		var err error
		if l.tcpListener, err = net.Listen("tcp", l.opts.ListenAddr); err != nil {
			return lazyerrors.Error(err)
		}
	}

	if useSock {
		var err error
		if l.sockListener, err = net.Listen("unix", l.opts.ListenSock); err != nil {
			return lazyerrors.Error(err)
		}
	}

	close(l.listening)

	if useTcp {
		logger.Sugar().Infof("Listening on %s ...", l.Addr())
	}

	if useSock {
		logger.Sugar().Infof("Listening on %s ...", l.Sock())
	}

	// handle ctx cancellation
	go func() {
		<-ctx.Done()

		if useTcp {
			l.tcpListener.Close()
		}

		if useSock {
			l.sockListener.Close()
		}
	}()

	var wg sync.WaitGroup

	// handle TCP stream
	if useTcp {
		wg.Add(1)
		go listenLoop(ctx, &wg, l, l.tcpListener, logger)
	}

	// handle UNIX stream
	if useSock {
		wg.Add(1)
		go listenLoop(ctx, &wg, l, l.sockListener, logger)
	}

	logger.Info("Waiting for all connections to stop...")
	wg.Wait()

	return ctx.Err()
}

func listenLoop(ctx context.Context, wg *sync.WaitGroup, l *Listener, listener net.Listener, logger *zap.Logger) {
	defer wg.Done()

	for {
		netConn, err := listener.Accept()
		if err != nil {
			l.metrics.accepts.WithLabelValues("1").Inc()

			if ctx.Err() != nil {
				break
			}

			logger.Warn("Failed to accept connection", zap.Error(err))
			if !errors.Is(err, net.ErrClosed) {
				time.Sleep(time.Second)
			}
			continue
		}

		wg.Add(1)
		l.metrics.accepts.WithLabelValues("0").Inc()
		l.metrics.connectedClients.Inc()

		go runConn(ctx, netConn, l, wg, logger)
	}
}

func runConn(ctx context.Context, netConn net.Conn, l *Listener, wg *sync.WaitGroup, logger *zap.Logger) {
	connID := fmt.Sprintf("%s -> %s", netConn.RemoteAddr(), netConn.LocalAddr())

	// give clients a few seconds to disconnect after ctx is canceled
	runCancelDelay := l.opts.TestRunCancelDelay
	if runCancelDelay == 0 {
		runCancelDelay = 3 * time.Second
	}
	runCtx, runCancel := ctxutil.WithDelay(ctx.Done(), runCancelDelay)
	defer runCancel()

	if l.opts.TestConnTimeout != 0 {
		runCtx, runCancel = context.WithTimeout(runCtx, l.opts.TestConnTimeout)
		defer runCancel()
	}

	defer pprof.SetGoroutineLabels(runCtx)
	runCtx = pprof.WithLabels(runCtx, pprof.Labels("conn", connID))
	pprof.SetGoroutineLabels(runCtx)

	defer func() {
		netConn.Close()
		l.metrics.connectedClients.Dec()
		wg.Done()
	}()

	opts := &newConnOpts{
		netConn:        netConn,
		mode:           l.opts.Mode,
		l:              l.opts.Logger.Named("// " + connID + " "),
		proxyAddr:      l.opts.ProxyAddr,
		handler:        l.opts.Handler,
		connMetrics:    l.metrics.connMetrics,
		testRecordPath: l.opts.TestRecordPath,
	}
	conn, e := newConn(opts)
	if e != nil {
		logger.Warn("Failed to create connection", zap.String("conn", connID), zap.Error(e))
		return
	}

	logger.Info("Connection started", zap.String("conn", connID))

	e = conn.run(runCtx)
	if e == io.EOF {
		logger.Info("Connection stopped", zap.String("conn", connID))
	} else {
		logger.Warn("Connection stopped", zap.String("conn", connID), zap.Error(e))
	}
}

// Addr returns listener's address.
// It can be used to determine an actually used port, if it was zero.
func (l *Listener) Addr() net.Addr {
	<-l.listening
	return l.tcpListener.Addr()
}

// Sock returns listener's unix domain socket address.
//
// It is a blocking call if Run was not called.
func (l *Listener) Sock() net.Addr {
	<-l.listening
	return l.sockListener.Addr()
}

// Describe implements prometheus.Collector.
func (l *Listener) Describe(ch chan<- *prometheus.Desc) {
	l.metrics.Describe(ch)
}

// Collect implements prometheus.Collector.
func (l *Listener) Collect(ch chan<- prometheus.Metric) {
	l.metrics.Collect(ch)
}
