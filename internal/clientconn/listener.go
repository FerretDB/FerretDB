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
	"io"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Listener accepts incoming client connections.
type Listener struct {
	opts *NewListenerOpts
}

// NewListenerOpts represents listener configuration.
type NewListenerOpts struct {
	ListenAddr      string
	TLS             bool
	ProxyAddr       string
	Mode            Mode
	PgPool          *pg.Pool
	Logger          *zap.Logger
	Metrics         *ListenerMetrics
	HandlersMetrics *handlers.Metrics
	TestConnTimeout time.Duration
	StartTime       time.Time
}

// NewListener returns a new listener, configured by the NewListenerOpts argument.
func NewListener(opts *NewListenerOpts) *Listener {
	return &Listener{
		opts: opts,
	}
}

// Run runs the listener until ctx is canceled or some unrecoverable error occurs.
func (l *Listener) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", l.opts.ListenAddr)
	if err != nil {
		return lazyerrors.Error(err)
	}

	l.opts.Logger.Sugar().Infof("Listening on %s ...", l.opts.ListenAddr)

	if l.opts.TLS {
		l.opts.Logger.Sugar().Info("Using insecure TLS.")
		cert, err := generateInsecureCert()
		if err != nil {
			return err
		}
		l.opts.Logger.Sugar().Info("Insecure self-signed certificate generated.")

		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{*cert},
			InsecureSkipVerify: true,
		}
		lis = tls.NewListener(lis, tlsConfig)
	}

	// handle ctx cancelation
	go func() {
		<-ctx.Done()
		lis.Close()
	}()

	const delay = 3 * time.Second

	var wg sync.WaitGroup
	for {
		netConn, err := lis.Accept()
		if err != nil {
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
		l.opts.Metrics.ConnectedClients.Inc()

		// run connection
		go func() {
			defer func() {
				netConn.Close()
				l.opts.Metrics.ConnectedClients.Dec()
				wg.Done()
			}()

			opts := &newConnOpts{
				netConn:         netConn,
				pgPool:          l.opts.PgPool,
				proxyAddr:       l.opts.ProxyAddr,
				mode:            l.opts.Mode,
				handlersMetrics: l.opts.HandlersMetrics,
				startTime:       l.opts.StartTime,
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
