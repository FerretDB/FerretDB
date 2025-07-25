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

// Package middleware provides connection between listeners and handlers.
package middleware

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/prometheus/client_golang/prometheus"
)

// Handler is a common interface for handlers and middleware.
type Handler interface {
	// Handle processes a single request.
	//
	// The passed context is canceled when the client disconnects.
	//
	// Response is a normal or error response produced by the handler.
	//
	// Error is returned when the handler cannot process the request;
	// for example, when connection with PostgreSQL or proxy is lost.
	// Returning an error generally means that the listener should close the client connection.
	// Error should not be [*mongoerrors.Error] or have that type in its chain.
	Handle(ctx context.Context, req *Request) (resp *Response, err error)

	prometheus.Collector
}

// lastRequestID stores last generated request ID.
var lastRequestID atomic.Int32

type Middleware struct {
	*NewOpts
	m *metrics
}

// NewOpts represents middleware configuration.
//
//nolint:vet // for readability
type NewOpts struct {
	Mode  Mode
	DocDB Handler
	Proxy Handler
	L     *slog.Logger
}

// New returns a new middleware.
func New(opts *NewOpts) (*Middleware, error) {
	return &Middleware{
		NewOpts: opts,
		m:       newMetrics(),
	}, nil
}

func (m *Middleware) Handle(ctx context.Context, req *Request) (*Response, error) {
	docdb, proxy, docdbErr, proxyErr := m.handle(ctx, req)

	switch m.Mode {
	case NormalMode, DiffNormalMode:
		return docdb, docdbErr
	case ProxyMode, DiffProxyMode:
		return proxy, proxyErr
	default:
		panic("not reached")
	}
}

func (m *Middleware) handle(ctx context.Context, req *Request) (docdb, proxy *Response, docdbErr, proxyErr error) {
	// FIXME opcode
	opcode := req.WireHeader().OpCode.String()
	command := req.Document().Command()
	m.m.Requests.WithLabelValues(opcode, command).Inc()

	m.m.Responses.MustCurryWith(prometheus.Labels{
		"opcode":  opcode,
		"command": command,
	})

	var wg sync.WaitGroup

	if m.DocDB != nil {
		wg.Add(1)
		go func() {
			docdb, docdbErr = (&dispatcher{
				h:         m.DocDB,
				l:         logging.WithName(m.L, "documentdb"),
				responses: m.m.Responses, // FIXME
			}).Handle(ctx, req)
			wg.Done()
		}()
	}

	if m.Proxy != nil {
		wg.Add(1)
		go func() {
			proxy, proxyErr = (&dispatcher{
				h:         m.Proxy,
				l:         logging.WithName(m.L, "proxy"),
				responses: m.m.Responses, // FIXME
			}).Handle(ctx, req)

			wg.Done()
		}()
	}

	wg.Wait()

	return
}

// Describe implements [prometheus.Collector].
func (m *Middleware) Describe(ch chan<- *prometheus.Desc) {
	m.m.Describe(ch)
}

// Collect implements [prometheus.Collector].
func (m *Middleware) Collect(ch chan<- prometheus.Metric) {
	m.m.Collect(ch)
}

// check interfaces
var (
	_ Handler              = (*Middleware)(nil)
	_ prometheus.Collector = (*Middleware)(nil)
)
