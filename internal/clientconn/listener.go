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
	opts           *NewListenerOpts
	metrics        *ListenerMetrics
	handler        handlers.Interface
	netListener    net.Listener
	socketListener net.Listener
	listening      chan struct{}
}

// NewListenerOpts represents netListener configuration.
type NewListenerOpts struct {
	ListenAddr      string
	ListenSocket    string
	ProxyAddr       string
	Mode            Mode
	Handler         handlers.Interface
	Logger          *zap.Logger
	TestConnTimeout time.Duration
}

// NewListener returns a new netListener, configured by the NewListenerOpts argument.
func NewListener(opts *NewListenerOpts) *Listener {
	return &Listener{
		opts:      opts,
		metrics:   newListenerMetrics(),
		handler:   opts.Handler,
		listening: make(chan struct{}),
	}
}

// Run runs the listeners until ctx is done or some unrecoverable error occurs.
//
// When this method returns, listeners and all connections are closed.
func (l *Listener) Run(ctx context.Context) error {
	logger := l.opts.Logger.Named("netListener")

	var err error
	if l.netListener, err = net.Listen("tcp", l.opts.ListenAddr); err != nil {
		return lazyerrors.Error(err)
	}

	if l.opts.ListenSocket != "" {
		if l.socketListener, err = net.Listen("unix", l.opts.ListenSocket); err != nil {
			return lazyerrors.Error(err)
		}
	}

	close(l.listening)
	logger.Sugar().Infof("Listening on %s ...", l.Addr())

	// handle ctx cancelation
	go func() {
		<-ctx.Done()
		l.netListener.Close()
		if l.socketListener != nil {
			l.socketListener.Close()
		}
	}()

	const delay = 3 * time.Second

	var wg sync.WaitGroup
	go l.handleConn(ctx, l.netListener, logger, &wg, delay)
	go l.handleConn(ctx, l.socketListener, logger, &wg, delay)

	logger.Info("Waiting for all connections to stop...")
	wg.Wait()

	return ctx.Err()
}

func (l *Listener) handleConn(
	ctx context.Context, listener net.Listener, logger *zap.Logger, wg *sync.WaitGroup, delay time.Duration,
) {
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

		// run connection
		go func() {
			defer func() {
				netConn.Close()
				l.metrics.connectedClients.Dec()
				wg.Done()
			}()

			prefix := fmt.Sprintf("// %s -> %s ", netConn.RemoteAddr(), netConn.LocalAddr())
			opts := &newConnOpts{
				netConn:     netConn,
				mode:        l.opts.Mode,
				l:           l.opts.Logger.Named(prefix), // original unnamed logger
				proxyAddr:   l.opts.ProxyAddr,
				handler:     l.opts.Handler,
				connMetrics: l.metrics.connMetrics,
			}
			conn, e := newConn(opts)
			if e != nil {
				logger.Warn("Failed to create connection", zap.Error(e))
				return
			}

			runCtx, runCancel := ctxutil.WithDelay(ctx.Done(), delay)
			defer runCancel()

			if l.opts.TestConnTimeout != 0 {
				runCtx, runCancel = context.WithTimeout(runCtx, l.opts.TestConnTimeout)
				defer runCancel()
			}

			e = conn.run(runCtx) //nolint:contextcheck // false positive
			if e == io.EOF {
				logger.Info("Connection stopped")
			} else {
				logger.Warn("Connection stopped", zap.Error(e))
			}
		}()
	}
}

// Addr returns listener's address.
// It can be used to determine an actually used port, if it was zero.
func (l *Listener) Addr() net.Addr {
	<-l.listening
	return l.netListener.Addr()
}

// Describe implements prometheus.Collector.
func (l *Listener) Describe(ch chan<- *prometheus.Desc) {
	l.metrics.Describe(ch)
}

// Collect implements prometheus.Collector.
func (l *Listener) Collect(ch chan<- prometheus.Metric) {
	l.metrics.Collect(ch)
}
