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
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

type Middleware struct {
	*NewOpts
	wg sync.WaitGroup
}

// NewOpts represents middleware configuration.
//
//nolint:vet // for readability
type NewOpts struct {
	Mode    Mode
	DocDB   Handler
	Proxy   Handler
	Metrics *Metrics
	L       *slog.Logger
}

// New returns a new middleware.
// [Middleware.Run] must be called on the returned value.
func New(opts *NewOpts) (*Middleware, error) {
	must.NotBeZero(opts)

	switch opts.Mode {
	case NormalMode:
		must.NotBeZero(opts.DocDB)
		must.BeZero(opts.Proxy)
	case ProxyMode:
		must.BeZero(opts.DocDB)
		must.NotBeZero(opts.Proxy)
	case DiffNormalMode:
		must.NotBeZero(opts.DocDB)
		must.NotBeZero(opts.Proxy)
	case DiffProxyMode:
		must.NotBeZero(opts.DocDB)
		must.NotBeZero(opts.Proxy)
	default:
		panic("not reached")
	}

	return &Middleware{
		NewOpts: opts,
	}, nil
}

// Run implements [middleware.Handler].
func (m *Middleware) Run(ctx context.Context) {
	<-ctx.Done()
	m.wg.Wait()
}

func (m *Middleware) Handle(ctx context.Context, req *Request) (*Response, error) {
	m.wg.Add(1)
	defer m.wg.Done()

	docdb, proxy, docdbErr, proxyErr := m.handle(ctx, req)

	switch m.Mode {
	case NormalMode:
		return docdb, docdbErr
	case ProxyMode:
		return proxy, proxyErr
	case DiffNormalMode:
		m.logDiff(ctx, docdb, proxy, slog.LevelDebug)
		return docdb, docdbErr
	case DiffProxyMode:
		m.logDiff(ctx, docdb, proxy, slog.LevelDebug)
		return proxy, proxyErr
	default:
		panic("not reached")
	}
}

func (m *Middleware) handle(ctx context.Context, req *Request) (docdb, proxy *Response, docdbErr, proxyErr error) {
	// FIXME opcode
	opcode := req.WireHeader().OpCode.String()
	command := req.Document().Command()
	m.Metrics.requests.WithLabelValues(opcode, command).Inc()

	m.Metrics.responses.MustCurryWith(prometheus.Labels{
		"opcode":  opcode,
		"command": command,
	})

	if m.L.Enabled(ctx, slog.LevelDebug) {
		m.L.DebugContext(ctx, fmt.Sprintf("<<< %s\n%s", req.WireHeader(), req.WireBody().StringIndent()))
	}

	var wg sync.WaitGroup

	if m.DocDB != nil {
		wg.Add(1)
		go func() {
			docdb, docdbErr = (&dispatcher{
				h: m.DocDB,
				l: logging.WithName(m.L, "docdb"),
				m: m.Metrics,
			}).Handle(ctx, req)
			wg.Done()
		}()
	}

	if m.Proxy != nil {
		wg.Add(1)
		go func() {
			proxy, proxyErr = (&dispatcher{
				h: m.Proxy,
				l: logging.WithName(m.L, "proxy"),
				m: m.Metrics,
			}).Handle(ctx, req)

			wg.Done()
		}()
	}

	wg.Wait()

	return
}

// logDiff logs the diff between the DocumentDB and proxy responses.
// It does nothing if either response is nil or logging for the given level is disabled.
func (m *Middleware) logDiff(ctx context.Context, docdb, proxy *Response, logLevel slog.Level) {
	if docdb == nil || proxy == nil {
		return
	}

	if !m.L.Enabled(ctx, logLevel) {
		return
	}

	diffHeader, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(docdb.WireHeader().String()),
		FromFile: "docdb header",
		B:        difflib.SplitLines(proxy.WireHeader().String()),
		ToFile:   "proxy header",
		Context:  1,
	})
	if err != nil {
		m.L.Log(ctx, slog.LevelWarn, "Failed to get header diff", logging.Error(err))
		return
	}

	diffHeader = strings.TrimSpace(diffHeader)
	if diffHeader == "" {
		diffHeader = " none"
	} else {
		diffHeader = "\n" + diffHeader
	}

	diffBody, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(docdb.WireBody().StringIndent()),
		FromFile: "docdb body",
		B:        difflib.SplitLines(proxy.WireBody().StringIndent()),
		ToFile:   "proxy body",
		Context:  1,
	})
	if err != nil {
		m.L.Log(ctx, slog.LevelWarn, "Failed to get body diff", logging.Error(err))
		return
	}

	diffBody = strings.TrimSpace(diffBody)
	if diffBody == "" {
		diffBody = " none"
	} else {
		diffBody = "\n" + diffBody
	}

	msg := "Header diff:" + diffHeader + "\nBody diff:" + diffBody
	m.L.Log(ctx, logLevel, msg)
}

// Describe implements [prometheus.Collector].
func (m *Middleware) Describe(ch chan<- *prometheus.Desc) {
	// FIXME
}

// Collect implements [prometheus.Collector].
func (m *Middleware) Collect(ch chan<- prometheus.Metric) {
	// FIXME
}

// check interfaces
var (
	_ Handler              = (*Middleware)(nil)
	_ prometheus.Collector = (*Middleware)(nil)
)
