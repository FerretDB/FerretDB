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
	"io"
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/pg"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Listener accepts incoming client connections.
type Listener struct {
	opts            *NewListenerOpts
	metrics         *ListenerMetrics
	handlersMetrics *pg.Metrics
	startTime       time.Time
	listener        net.Listener
	listening       chan struct{}
}

// NewListenerOpts represents listener configuration.
type NewListenerOpts struct {
	ListenAddr      string
	ProxyAddr       string
	Mode            Mode
	PgPool          *pgdb.Pool
	Logger          *zap.Logger
	TestConnTimeout time.Duration
}

// NewListener returns a new listener, configured by the NewListenerOpts argument.
func NewListener(opts *NewListenerOpts) *Listener {
	return &Listener{
		opts:            opts,
		metrics:         NewListenerMetrics(),
		handlersMetrics: pg.NewMetrics(),
		startTime:       time.Now(),
		listening:       make(chan struct{}),
	}
}

// Run runs the listener until ctx is canceled or some unrecoverable error occurs.
func (l *Listener) Run(ctx context.Context) error {
	var err error
	if l.listener, err = net.Listen("tcp", l.opts.ListenAddr); err != nil {
		return lazyerrors.Error(err)
	}

	close(l.listening)
	l.opts.Logger.Sugar().Infof("Listening on %s ...", l.Addr())

	// handle ctx cancelation
	go func() {
		<-ctx.Done()
		l.listener.Close()
	}()

	const delay = 3 * time.Second

	var wg sync.WaitGroup
	for {
		netConn, err := l.listener.Accept()
		if err != nil {
			l.metrics.accepts.WithLabelValues("1").Inc()

			if ctx.Err() != nil {
				break
			}

			l.opts.Logger.Warn("Failed to accept connection", zap.Error(err))
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

			opts := &newConnOpts{
				netConn:         netConn,
				pgPool:          l.opts.PgPool,
				proxyAddr:       l.opts.ProxyAddr,
				mode:            l.opts.Mode,
				handlersMetrics: l.handlersMetrics,
				startTime:       l.startTime,
			}
			conn, e := newConn(opts)
			if e != nil {
				l.opts.Logger.Warn("Failed to create connection", zap.Error(e))
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
				l.opts.Logger.Info("Connection stopped")
			} else {
				l.opts.Logger.Warn("Connection stopped", zap.Error(e))
			}
		}()
	}

	l.opts.Logger.Info("Waiting for all connections to stop...")
	wg.Wait()

	return ctx.Err()
}

// Addr returns listener's address.
// It can be used to determine an actually used port, if it was zero.
func (l *Listener) Addr() net.Addr {
	<-l.listening
	return l.listener.Addr()
}

// Describe implements prometheus.Collector.
func (l *Listener) Describe(ch chan<- *prometheus.Desc) {
	l.metrics.Describe(ch)
	l.handlersMetrics.Describe(ch)
}

// Collect implements prometheus.Collector.
func (l *Listener) Collect(ch chan<- prometheus.Metric) {
	l.metrics.Collect(ch)
	l.handlersMetrics.Collect(ch)
}
