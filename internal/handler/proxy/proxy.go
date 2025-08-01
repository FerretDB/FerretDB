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

// Package proxy handles requests by sending them to another wire protocol compatible service.
package proxy

import (
	"context"
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Handler handles requests by sending them to another wire protocol compatible service.
type Handler struct {
	opts *NewOpts

	rw    sync.RWMutex
	conns map[*conninfo.ConnInfo]*conn
}

// NewOpts represents handler configuration.
//
//nolint:vet // for readability
type NewOpts struct {
	Addr        string
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string

	L *slog.Logger
}

// New creates a new Handler for a service with given address.
// [Handler.Run] must be called on the returned value.
func New(opts *NewOpts) (*Handler, error) {
	must.NotBeZero(opts)
	must.NotBeZero(opts.Addr)
	must.NotBeZero(opts.L)

	return &Handler{
		opts:  opts,
		conns: make(map[*conninfo.ConnInfo]*conn),
	}, nil
}

// Run implements [middleware.Handler].
func (h *Handler) Run(ctx context.Context) {
	<-ctx.Done()

	// FIXME do we need a lock there?
	for ci := range h.conns {
		h.closeConn(ci)
	}
}

// Handle processes a request by sending it to another wire protocol compatible service.
func (h *Handler) Handle(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
	ci := conninfo.Get(ctx)

	h.rw.RLock()
	c := h.conns[ci]
	h.rw.RUnlock()

	if c == nil {
		var err error
		if c, err = newConn(ctx, h.opts); err != nil {
			return nil, lazyerrors.Error(err)
		}

		ci.OnClose(h.closeConn)

		h.rw.Lock()
		h.conns[ci] = c
		h.rw.Unlock()
	}

	return c.handle(ctx, req)
}

func (h *Handler) closeConn(ci *conninfo.ConnInfo) {
	h.rw.Lock()
	defer h.rw.Unlock()

	c := h.conns[ci]
	if c != nil {
		c.close()
		delete(h.conns, ci)
	}
}

// Describe implements [prometheus.Collector].
func (h *Handler) Describe(ch chan<- *prometheus.Desc) {
}

// Collect implements [prometheus.Collector].
func (h *Handler) Collect(ch chan<- prometheus.Metric) {
}

// check interfaces
var (
	_ middleware.Handler = (*Handler)(nil)
)
