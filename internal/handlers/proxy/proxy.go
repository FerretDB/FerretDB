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

// Parts of Prometheus metric names.
// TODO https://github.com/FerretDB/FerretDB/issues/4965
const (
	namespace = "ferretdb_unstable"
	subsystem = "proxy"
)

// Handler handles requests by sending them to another wire protocol compatible service.
//
//nolint:vet // for readability
type Handler struct {
	opts *NewOpts

	connsRW  sync.RWMutex
	connsGet map[*conninfo.ConnInfo]func() (*conn, error)

	runM   sync.Mutex
	runCtx context.Context
	runWG  sync.WaitGroup
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
		opts:     opts,
		connsGet: map[*conninfo.ConnInfo]func() (*conn, error){},
	}, nil
}

// Run implements [middleware.Handler].
func (h *Handler) Run(ctx context.Context) {
	h.runM.Lock()
	h.runCtx = ctx
	h.runM.Unlock()

	<-ctx.Done()
	h.opts.L.InfoContext(ctx, "Stopping")

	h.runWG.Wait()

	h.connsRW.Lock()

	for _, cg := range h.connsGet {
		if c, _ := cg(); c != nil {
			c.close()
		}
	}
	h.connsGet = nil

	h.connsRW.Unlock()

	h.opts.L.InfoContext(ctx, "Stopped")
}

// Handle implements [middleware.Handler] by sending it to another wire protocol compatible service.
func (h *Handler) Handle(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
	if ctx.Err() != nil {
		return nil, lazyerrors.Error(ctx.Err())
	}

	h.runM.Lock()

	if rc := h.runCtx; rc != nil && rc.Err() != nil {
		h.runM.Unlock()
		return nil, lazyerrors.Error(rc.Err())
	}

	// we need to use Add under a lock to avoid a race with Wait in Run
	h.runWG.Add(1)
	h.runM.Unlock()

	defer h.runWG.Done()

	ci := conninfo.Get(ctx)

	c, err := h.getConn(ctx, ci)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return c.handle(ctx, req)
}

// getConn returns a proxy connection for the given client connection info,
// establishing it if necessary, while preserving one-to-one mapping.
func (h *Handler) getConn(ctx context.Context, ci *conninfo.ConnInfo) (*conn, error) {
	// fast path
	h.connsRW.RLock()
	cg := h.connsGet[ci]
	h.connsRW.RUnlock()

	if cg != nil {
		return cg()
	}

	// slow path

	h.connsRW.Lock()

	// a concurrent call might have started creating connection already; check again
	if cg = h.connsGet[ci]; cg != nil {
		h.connsRW.Unlock()
		return cg()
	}

	cg = sync.OnceValues(func() (*conn, error) {
		c, err := newConn(ctx, h.opts)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		ci.OnClose(h.closeConn)

		return c, nil
	})

	h.connsGet[ci] = cg

	h.connsRW.Unlock()

	return cg()
}

// closeConn closes the proxy connection for the given client connection info.
func (h *Handler) closeConn(ci *conninfo.ConnInfo) {
	h.connsRW.Lock()
	defer h.connsRW.Unlock()

	if cg := h.connsGet[ci]; cg != nil {
		delete(h.connsGet, ci)

		if c, _ := cg(); c != nil {
			c.close()
		}
	}
}

// Describe implements [prometheus.Collector].
func (h *Handler) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(h, ch)
}

// Collect implements [prometheus.Collector].
func (h *Handler) Collect(ch chan<- prometheus.Metric) {
	h.connsRW.RLock()
	defer h.connsRW.RUnlock()

	// We should have counters for connects/disconnects, not gauge for the current number.
	// TODO https://github.com/FerretDB/FerretDB/issues/1997

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "conns"),
			"Unstable: The current number of connections.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(len(h.connsGet)),
	)
}

// check interfaces
var (
	_ middleware.Handler = (*Handler)(nil)
)
