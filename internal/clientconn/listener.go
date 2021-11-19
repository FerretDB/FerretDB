// Copyright 2021 Baltoro OÜ.
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

	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

type Listener struct {
	opts *NewListenerOpts
}

type NewListenerOpts struct {
	ListenAddr string
	TLS        bool
	ProxyAddr  string
	Mode       Mode
	PgPool     *pg.Pool
	Logger     *zap.Logger

	TestConnTimeout time.Duration
}

func NewListener(opts *NewListenerOpts) *Listener {
	return &Listener{
		opts: opts,
	}
}

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
		l.opts.Logger.Sugar().Info("Insecure self-signed cerificate generated.")

		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{*cert},
			InsecureSkipVerify: true,
		}
		lis = tls.NewListener(lis, tlsConfig)
	}

	go func() {
		<-ctx.Done()
		lis.Close()
	}()

	var wg sync.WaitGroup
	for ctx.Err() == nil {
		netConn, err := lis.Accept()
		if err != nil {
			l.opts.Logger.Warn("Failed to accept connection", zap.Error(err))
			if !errors.Is(err, net.ErrClosed) {
				time.Sleep(time.Second)
			}
			continue
		}

		wg.Add(1)
		go func() {
			defer func() {
				netConn.Close()
				wg.Done()
			}()

			conn, e := newConn(netConn, l.opts.PgPool, l.opts.ProxyAddr, l.opts.Mode)
			if e != nil {
				l.opts.Logger.Warn("Failed to create connection", zap.Error(e))
				return
			}

			runCtx := ctx
			var runCancel context.CancelFunc
			if l.opts.TestConnTimeout != 0 {
				runCtx, runCancel = context.WithTimeout(ctx, l.opts.TestConnTimeout)
				defer runCancel()
			}

			e = conn.run(runCtx)
			if e == io.EOF {
				l.opts.Logger.Info("Connection stopped")
			} else {
				l.opts.Logger.Warn("Connection stopped", zap.Error(e))
			}
		}()
	}

	wg.Wait()

	return nil
}
