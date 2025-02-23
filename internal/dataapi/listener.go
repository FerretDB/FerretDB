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
	"log/slog"
	"net"
	"net/http"

	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
	"github.com/FerretDB/FerretDB/v2/internal/dataapi/server"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// Listener represents dataapi listener.
type Listener struct {
	opts *ListenOpts
	lis  net.Listener
	srv  *server.Server
}

// ListenOpts represents [Listen] options.
type ListenOpts struct {
	L       *slog.Logger
	Handler *handler.Handler
	TCPAddr string
}

// Listen creates a new dataapi handler and starts listener on the given TCP address.
func Listen(opts *ListenOpts) (*Listener, error) {
	lis, err := net.Listen("tcp", opts.TCPAddr)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Listener{
		opts: opts,
		lis:  lis,
		srv:  server.New(opts.L, opts.Handler),
	}, nil
}

// Run runs dataapi handler until ctx is canceled.
//
// It exits when handler is stopped and listener closed.
func (lis *Listener) Run(ctx context.Context) {
	srvHandler := api.HandlerFromMux(lis.srv, http.NewServeMux())

	if lis.opts.Handler.Auth {
		srvHandler = lis.srv.AuthMiddleware(srvHandler)
	}

	srv := &http.Server{
		Handler: srvHandler,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		if err := srv.Serve(lis.lis); !errors.Is(err, http.ErrServerClosed) {
			lis.opts.L.LogAttrs(ctx, logging.LevelDPanic, "ListenAndServe exited with unexpected error", logging.Error(err))
		}
	}()

	<-ctx.Done()
}
