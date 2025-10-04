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

package documentdb

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	otelcodes "go.opentelemetry.io/otel/codes"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// contextKey is a named unexported type for the safe use of [context.WithValue].
type contextKey struct{}

// queryKey is used for setting and getting a value with [context.WithValue].
var queryKey = contextKey{}

// tracer implements various pgx interfaces to provide Prometheus metrics and
// OpenTelemetry traces.
//
// See:
//   - https://pkg.go.dev/github.com/jackc/pgx/v5/tracelog
//   - https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#AcquireTracer
//   - https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#ReleaseTracer
//   - https://pkg.go.dev/github.com/jackc/pgx/v5#hdr-Tracing_and_Logging
//   - https://pkg.go.dev/github.com/jackc/pgx/v5/multitracer
//
// TODO https://github.com/FerretDB/FerretDB/issues/3554
type tracer struct {
	tl       *tracelog.TraceLog
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// newTracer creates a new tracer.
func newTracer(l *slog.Logger) *tracer {
	return &tracer{
		// try to log everything; logger's configuration will skip extra levels if needed
		tl: &tracelog.TraceLog{
			Logger:   logging.NewPgxLogger(l),
			LogLevel: tracelog.LogLevelTrace,
			Config: &tracelog.TraceLogConfig{
				TimeKey: slog.TimeKey,
			},
		},
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "The cumulative count of the total queries to PostgreSQL.",
			},
			[]string{},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "responses_duration_seconds",
				Help:      "The duration taken for PostgreSQL query response in seconds.",
				Buckets: []float64{
					(1 * time.Millisecond).Seconds(),
					(5 * time.Millisecond).Seconds(),
					(10 * time.Millisecond).Seconds(),
					(25 * time.Millisecond).Seconds(),
					(50 * time.Millisecond).Seconds(),
					(100 * time.Millisecond).Seconds(),
					(250 * time.Millisecond).Seconds(),
					(500 * time.Millisecond).Seconds(),
					(1000 * time.Millisecond).Seconds(),
					(2500 * time.Millisecond).Seconds(),
					(5000 * time.Millisecond).Seconds(),
					(10000 * time.Millisecond).Seconds(),
				},
			},
			[]string{},
		),
	}
}

// TraceAcquireStart implements [pgxpool.AcquireTracer].
//
// It is called at the beginning of [pgxpool.Pool.Acquire].
// The returned context is used for the rest of the call and will be passed to the [tracer.TraceAcquireEnd].
func (t *tracer) TraceAcquireStart(ctx context.Context, pool *pgxpool.Pool, data pgxpool.TraceAcquireStartData) context.Context {
	return t.tl.TraceAcquireStart(ctx, pool, data)
}

// TraceAcquireEnd implements [pgxpool.AcquireTracer].
//
// It is called when a connection has been acquired.
func (t *tracer) TraceAcquireEnd(ctx context.Context, pool *pgxpool.Pool, data pgxpool.TraceAcquireEndData) {
	t.tl.TraceAcquireEnd(ctx, pool, data)
}

// TraceRelease implements [pgxpool.ReleaseTracer].
//
// It is called at the beginning of [pgxpool.Conn.Release].
func (t *tracer) TraceRelease(pool *pgxpool.Pool, data pgxpool.TraceReleaseData) {
	t.tl.TraceRelease(pool, data)
}

// TraceConnectStart implements [pgx.ConnectTracer].
//
// It is called at the beginning of [pgx.Connect] and [pgx.ConnectConfig] calls.
// The returned context is used for the rest of the call and will be passed to [tracer.TraceConnectEnd].
func (t *tracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	return t.tl.TraceConnectStart(ctx, data)
}

// TraceConnectEnd implements [pgx.ConnectTracer].
func (t *tracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	t.tl.TraceConnectEnd(ctx, data)
}

// TracePrepareStart implements [pgx.PrepareTracer].
//
// It is called at the beginning of [pgx.Conn.Prepare] calls.
// The returned context is used for the rest of the call and will be passed to [tracer.TracePrepareEnd].
func (t *tracer) TracePrepareStart(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	ctx, _ = otel.Tracer("").Start(
		ctx,
		"documentdb.Prepare",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(
			otelsemconv.DBQueryText(data.SQL),
		),
	)

	return t.tl.TracePrepareStart(ctx, conn, data)
}

// TracePrepareEnd implements [pgx.PrepareTracer].
func (t *tracer) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData) {
	t.tl.TracePrepareEnd(ctx, conn, data)

	span := oteltrace.SpanFromContext(ctx)

	if data.Err == nil {
		span.SetStatus(otelcodes.Ok, "")
	} else {
		span.SetStatus(otelcodes.Error, "")
		span.RecordError(data.Err)
	}

	span.End()
}

// TraceQueryStart implements [pgx.QueryTracer].
//
// It is called at the beginning of [pgx.Conn.Query], [pgx.Conn.QueryRow], and [pgx.Conn.Exec] calls.
// The returned context is used for the rest of the call and will be passed to [tracer.TraceQueryEnd].
func (t *tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx = context.WithValue(ctx, queryKey, time.Now())

	t.requests.With(prometheus.Labels{}).Inc()

	ctx, _ = otel.Tracer("").Start(
		ctx,
		"documentdb.Query",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(
			otelsemconv.DBQueryText(data.SQL),
		),
	)

	return t.tl.TraceQueryStart(ctx, conn, data)
}

// TraceQueryEnd implements [pgx.QueryTracer].
func (t *tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	duration := time.Since(ctx.Value(queryKey).(time.Time))

	t.duration.With(prometheus.Labels{}).Observe(duration.Seconds())

	t.tl.TraceQueryEnd(ctx, conn, data)

	span := oteltrace.SpanFromContext(ctx)

	if data.Err == nil {
		span.SetStatus(otelcodes.Ok, "")
	} else {
		span.SetStatus(otelcodes.Error, "")
		span.RecordError(data.Err)
	}

	span.End()
}

// Describe implements prometheus.Collector.
func (t *tracer) Describe(ch chan<- *prometheus.Desc) {
	t.requests.Describe(ch)
	t.duration.Describe(ch)
}

// Collect implements prometheus.Collector.
func (t *tracer) Collect(ch chan<- prometheus.Metric) {
	t.requests.Collect(ch)
	t.duration.Collect(ch)
}

// check interfaces
var (
	_ pgxpool.AcquireTracer = (*tracer)(nil)
	_ pgxpool.ReleaseTracer = (*tracer)(nil)
	_ pgx.ConnectTracer     = (*tracer)(nil)
	_ pgx.PrepareTracer     = (*tracer)(nil)
	_ pgx.QueryTracer       = (*tracer)(nil)
	_ prometheus.Collector  = (*tracer)(nil)
)
