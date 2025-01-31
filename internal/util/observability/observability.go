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

// Package observability provides abstractions for tracing, metrics, etc.
//
// TODO https://github.com/FerretDB/FerretDB/issues/3244
package observability

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelsdkresource "go.opentelemetry.io/otel/sdk/resource"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// setup ensures that global tracer provider is set up only once.
var setup atomic.Bool

// OTelTraceExporter represents the OTLP trace exporter using HTTP with protobuf payloads.
type OTelTraceExporter struct {
	l  *slog.Logger
	tp *otelsdktrace.TracerProvider
}

// OTelTraceExporterOpts represents [OTelTraceExporter] options.
type OTelTraceExporterOpts struct {
	Logger *slog.Logger

	Service string
	Version string
	URL     string
}

// NewOTelTraceExporter sets up [OTelTraceExporter] and global tracer provider
// that is available via `otel.Tracer("")`.
//
// It must be called only once.
func NewOTelTraceExporter(opts *OTelTraceExporterOpts) (*OTelTraceExporter, error) {
	if opts.URL == "" {
		return nil, errors.New("URL is required")
	}

	// Exporter and tracer provider are configured explicitly to avoid environment variables fallback.
	// One current exception is TLS configuration:
	// - OTEL_EXPORTER_OTLP_CERTIFICATE
	// - OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE
	// - OTEL_EXPORTER_OTLP_CLIENT_KEY
	// - OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE
	// - OTEL_EXPORTER_OTLP_TRACES_CLIENT_CERTIFICATE
	// - OTEL_EXPORTER_OTLP_TRACES_CLIENT_KEY

	exporter := otlptracehttp.NewUnstarted(
		otlptracehttp.WithEndpointURL(opts.URL),
		otlptracehttp.WithHeaders(nil),
		otlptracehttp.WithTimeout(10*time.Second),
		otlptracehttp.WithCompression(otlptracehttp.NoCompression),
	)

	if err := exporter.Start(context.TODO()); err != nil {
		return nil, err
	}

	tp := otelsdktrace.NewTracerProvider(
		otelsdktrace.WithBatcher(exporter, otelsdktrace.WithBatchTimeout(time.Second)),
		otelsdktrace.WithResource(otelsdkresource.NewSchemaless(
			otelsemconv.ServiceName(opts.Service),
			otelsemconv.ServiceVersion(opts.Version),
		)),
		otelsdktrace.WithSampler(otelsdktrace.AlwaysSample()),
		otelsdktrace.WithRawSpanLimits(otelsdktrace.SpanLimits{
			AttributeValueLengthLimit:   otelsdktrace.DefaultAttributeValueLengthLimit,
			AttributeCountLimit:         otelsdktrace.DefaultAttributeCountLimit,
			EventCountLimit:             otelsdktrace.DefaultEventCountLimit,
			LinkCountLimit:              otelsdktrace.DefaultLinkCountLimit,
			AttributePerEventCountLimit: otelsdktrace.DefaultAttributePerEventCountLimit,
			AttributePerLinkCountLimit:  otelsdktrace.DefaultAttributePerLinkCountLimit,
		}),
	)

	if setup.Swap(true) {
		panic("global tracer provider is already set up")
	}
	otel.SetTracerProvider(tp)

	opts.Logger.Info("Starting OTel trace exporter...", slog.String("url", opts.URL))

	return &OTelTraceExporter{
		l:  opts.Logger,
		tp: tp,
	}, nil
}

// Run runs OTLP trace exporter until ctx is canceled.
func (ot *OTelTraceExporter) Run(ctx context.Context) {
	<-ctx.Done()

	// ctx is already canceled, but we want to inherit its values
	shutdownCtx, shutdownCancel := ctxutil.WithDelay(ctx)
	defer shutdownCancel(nil)

	if err := ot.tp.ForceFlush(shutdownCtx); err != nil {
		ot.l.ErrorContext(ctx, "ForceFlush exited with unexpected error", logging.Error(err))
	}

	if err := ot.tp.Shutdown(shutdownCtx); err != nil {
		ot.l.ErrorContext(ctx, "Shutdown exited with unexpected error", logging.Error(err))
	}

	ot.l.InfoContext(ctx, "OTel trace exporter stopped")
}
