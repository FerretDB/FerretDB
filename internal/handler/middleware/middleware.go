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
	"strconv"
	"strings"
	"sync"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/observability"
)

// Middleware connects listeners and handlers.
//
//nolint:vet // for readability
type Middleware struct {
	opts *NewOpts

	runM   sync.Mutex
	runCtx context.Context
	runWG  sync.WaitGroup
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
func New(opts *NewOpts) *Middleware {
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
		opts: opts,
	}
}

// Run implements [Handler].
func (m *Middleware) Run(ctx context.Context) {
	m.runM.Lock()
	m.runCtx = ctx
	m.runM.Unlock()

	<-ctx.Done()
	m.runWG.Wait()
}

// Handle implements [Handler],
// except that it returns nil for unrecoverable errors.
func (m *Middleware) Handle(ctx context.Context, req *Request) (resp *Response) {
	if ctx.Err() != nil {
		m.opts.L.WarnContext(ctx, "Not handling request: client already disconnected", logging.Error(ctx.Err()))
		return nil
	}

	m.runM.Lock()

	if rc := m.runCtx; rc != nil && rc.Err() != nil {
		m.runM.Unlock()
		m.opts.L.WarnContext(ctx, "Not handling request: already stopping", logging.Error(ctx.Err()))

		return nil
	}

	// we need to use Add under a lock to avoid a race with Wait in Run
	m.runWG.Add(1)
	m.runM.Unlock()

	defer m.runWG.Done()

	labels := prometheus.Labels{
		"opcode":  req.WireHeader().OpCode.String(),
		"command": req.Document().Command(),
	}
	m.opts.Metrics.requests.With(labels).Inc()

	ctx = m.startSpan(ctx, req)
	defer func() {
		m.endSpan(ctx, resp)
	}()

	if m.opts.L.Enabled(ctx, slog.LevelDebug) {
		m.opts.L.DebugContext(ctx, fmt.Sprintf("<<< %s\n%s", req.WireHeader(), req.WireBody().StringIndent()))
	}

	docdb, proxy := m.dispatch(ctx, req)

	switch m.opts.Mode {
	case NormalMode:
		resp = docdb
	case ProxyMode:
		resp = proxy
	case DiffNormalMode:
		m.logDiff(ctx, docdb, proxy)
		resp = docdb
	case DiffProxyMode:
		m.logDiff(ctx, docdb, proxy)
		resp = proxy
	default:
		panic("not reached")
	}

	return
}

// dispatch sends the request to both handlers.
// It returns nil for the given handler if it is not enabled, or if unrecoverable error occurs in it.
func (m *Middleware) dispatch(ctx context.Context, req *Request) (docdb, proxy *Response) {
	var wg sync.WaitGroup

	if m.opts.DocDB != nil {
		wg.Add(1)

		go func() {
			dCtx, _ := otel.Tracer("").Start(ctx, "middleware.Handle/documentdb")

			//exhaustruct:enforce
			d := &dispatcher{
				h:         m.opts.DocDB,
				l:         m.opts.L.With("handler", "documentdb"),
				responses: m.opts.Metrics.responses,
			}
			docdb = d.Dispatch(dCtx, req)

			wg.Done()
		}()
	}

	if m.opts.Proxy != nil {
		wg.Add(1)

		go func() {
			dCtx, _ := otel.Tracer("").Start(ctx, "middleware.Handle/proxy")

			//exhaustruct:enforce
			d := &dispatcher{
				h:         m.opts.Proxy,
				l:         m.opts.L.With("handler", "proxy"),
				responses: m.opts.Metrics.responses,
			}
			proxy = d.Dispatch(dCtx, req)

			wg.Done()
		}()
	}

	wg.Wait()

	return
}

// startSpan starts a new OpenTelemetry span for the request and returns the derived context.
func (m *Middleware) startSpan(ctx context.Context, req *Request) context.Context {
	comment, _ := req.Document().Get("comment").(string)

	spanCtx, err := observability.SpanContextFromComment(comment)
	if err != nil {
		m.opts.L.DebugContext(ctx, "Failed to extract span context from comment", logging.Error(err))
	}

	if spanCtx.IsValid() {
		ctx = oteltrace.ContextWithSpanContext(ctx, spanCtx)
	}

	command := req.Document().Command()
	database, _ := req.Document().Get("$db").(string)

	var collection string
	if command != "" {
		collection, _ = req.Document().Get(command).(string)
	}

	ctx, _ = otel.Tracer("").Start(
		ctx,
		"middleware.Handle",
		oteltrace.WithAttributes(
			otelsemconv.DBOperationName(command),
			otelsemconv.DBNamespace(database),
			otelsemconv.DBCollectionName(collection),
			otelattribute.Int("db.ferretdb.request_id", int(req.WireHeader().RequestID)),
		),
	)

	// Created span might be invalid, not sampled, and/or not recording,
	// if OpenTelemetry wasn't set up (for example, by the user of embeddable package).
	// We can't check span.SpanContext().IsValid(), span.SpanContext().IsSampled(), and span.IsRecording().

	return ctx
}

// endSpan ends the OpenTelemetry span from the context.
func (m *Middleware) endSpan(ctx context.Context, resp *Response) {
	span := oteltrace.SpanFromContext(ctx)

	if resp != nil {
		if c := resp.ErrorCode(); c != mongoerrors.ErrUnset {
			span.SetStatus(otelcodes.Error, resp.ErrorName())

			span.SetAttributes(
				otelsemconv.DBResponseStatusCode(strconv.Itoa(int(c))),
			)
		}

		span.SetAttributes(
			otelattribute.Int("db.ferretdb.response_id", int(resp.WireHeader().RequestID)),
		)
	}

	span.End()
}

// logDiff logs the diff between the DocumentDB and proxy responses.
// It does nothing if either response is nil or logging for the given level is disabled.
func (m *Middleware) logDiff(ctx context.Context, docdb, proxy *Response) {
	if docdb == nil || proxy == nil {
		return
	}

	const logLevel = slog.LevelInfo
	if !m.opts.L.Enabled(ctx, logLevel) {
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
		m.opts.L.Log(ctx, slog.LevelWarn, "Failed to get header diff", logging.Error(err))
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
		m.opts.L.Log(ctx, slog.LevelWarn, "Failed to get body diff", logging.Error(err))
		return
	}

	diffBody = strings.TrimSpace(diffBody)
	if diffBody == "" {
		diffBody = " none"
	} else {
		diffBody = "\n" + diffBody
	}

	msg := "Header diff:" + diffHeader + "\nBody diff:" + diffBody
	m.opts.L.Log(ctx, logLevel, msg)
}

// Describe implements [prometheus.Collector].
func (m *Middleware) Describe(ch chan<- *prometheus.Desc) {
	// m.opts.Metrics is not owned by the middleware; it exposes its own metrics.
}

// Collect implements [prometheus.Collector].
func (m *Middleware) Collect(ch chan<- prometheus.Metric) {
	// m.opts.Metrics is not owned by the middleware; it exposes its own metrics.
}

// check interfaces
var (
	_ prometheus.Collector = (*Middleware)(nil)
)
