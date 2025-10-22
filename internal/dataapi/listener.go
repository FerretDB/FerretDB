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

// Package dataapi provides a Data API wrapper,
// which allows FerretDB to be used over HTTP instead of MongoDB wire protocol.
package dataapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/AlekSi/lazyerrors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
	"github.com/FerretDB/FerretDB/v2/internal/dataapi/server"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// Listener represents dataapi listener.
type Listener struct {
	opts    *ListenOpts
	lis     net.Listener
	handler http.Handler
}

// ListenOpts represents [Listen] options.
type ListenOpts struct {
	L       *slog.Logger
	M       *middleware.Middleware
	TCPAddr string
	Auth    bool
}

// Listen creates a new dataapi handler and starts listener on the given TCP address.
// [Listener.Run] must be called on the returned value.
func Listen(opts *ListenOpts) (*Listener, error) {
	lis, err := net.Listen("tcp", opts.TCPAddr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	s := server.New(opts.L, opts.M)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /openapi.json", s.OpenAPISpec)

	h := api.HandlerFromMux(s, mux)
	if opts.Auth {
		h = s.AuthMiddleware(h)
	}

	h = s.ConnInfoMiddleware(h)

	return &Listener{
		opts:    opts,
		lis:     lis,
		handler: h,
	}, nil
}

// Run runs dataapi handler until ctx is canceled.
//
// It exits when handler is stopped and listener closed.
func (lis *Listener) Run(ctx context.Context) {
	srv := &http.Server{
		Handler:  lis.handler,
		ErrorLog: slog.NewLogLogger(lis.opts.L.Handler(), slog.LevelError),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	lis.opts.L.InfoContext(ctx, fmt.Sprintf("Starting DataAPI server on http://%s/", lis.lis.Addr()))

	go func() {
		if err := srv.Serve(lis.lis); !errors.Is(err, http.ErrServerClosed) {
			lis.opts.L.LogAttrs(ctx, logging.LevelDPanic, "Serve exited with unexpected error", logging.Error(err))
		}
	}()

	<-ctx.Done()

	// ctx is already canceled, but we want to inherit its values
	shutdownCtx, shutdownCancel := ctxutil.WithDelay(ctx)
	defer shutdownCancel(nil)

	if err := srv.Shutdown(shutdownCtx); err != nil {
		lis.opts.L.LogAttrs(ctx, logging.LevelDPanic, "Shutdown exited with unexpected error", logging.Error(err))
	}

	if err := srv.Close(); err != nil {
		lis.opts.L.LogAttrs(ctx, logging.LevelDPanic, "Close exited with unexpected error", logging.Error(err))
	}

	lis.opts.L.InfoContext(ctx, "DataAPI server stopped")
}

// Addr returns TCP listener's address.
// It can be used to determine an actually used port, if it was zero.
func (lis *Listener) Addr() net.Addr {
	return lis.lis.Addr()
}

// Describe implements [prometheus.Collector].
func (lis *Listener) Describe(ch chan<- *prometheus.Desc) {
}

// Collect implements [prometheus.Collector].
func (lis *Listener) Collect(ch chan<- prometheus.Metric) {
}
