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
	"sync/atomic"
	"time"

	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelsdkresource "go.opentelemetry.io/otel/sdk/resource"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)

// setup ensures that OTLP tracer is set up only once.
var setup atomic.Bool

// OtelTracer represents the OTLP tracer.
type OtelTracer struct {
	l  *zap.Logger
	tp *otelsdktrace.TracerProvider
}

// OtelTracerOpts is the configuration for OtelTracer.
type OtelTracerOpts struct {
	Logger *zap.Logger

	Service  string
	Version  string
	Endpoint string
}

// NewOtelTracer sets up OTLP tracer.
func NewOtelTracer(opts *OtelTracerOpts) (*OtelTracer, error) {
	if setup.Swap(true) {
		panic("OTLP tracer is already set up")
	}

	if opts.Endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	// Exporter and tracer are configured with the particular params on purpose.
	// We don't want to let them being set through OTEL_* environment variables,
	// but we set them explicitly.

	exporter, err := otlptracehttp.New(
		context.TODO(),
		otlptracehttp.WithEndpoint(opts.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := otelsdktrace.NewTracerProvider(
		otelsdktrace.WithBatcher(exporter, otelsdktrace.WithBatchTimeout(time.Second)),
		otelsdktrace.WithSampler(otelsdktrace.AlwaysSample()),
		otelsdktrace.WithResource(otelsdkresource.NewSchemaless(
			otelsemconv.ServiceName(opts.Service),
			otelsemconv.ServiceVersion(opts.Version),
		)),
	)

	otel.SetTracerProvider(tp)

	return &OtelTracer{
		l:  opts.Logger,
		tp: tp,
	}, nil
}

// Run runs OTLP tracer until ctx is canceled.
func (ot *OtelTracer) Run(ctx context.Context) {
	ot.l.Info("OTLP tracer started successfully.")

	<-ctx.Done()

	// ctx is already canceled, but we want to inherit its values
	shutdownCtx, shutdownCancel := ctxutil.WithDelay(ctx)
	defer shutdownCancel(nil)

	if err := ot.tp.ForceFlush(shutdownCtx); err != nil {
		ot.l.DPanic("ForceFlush exited with unexpected error", zap.Error(err))
	}

	if err := ot.tp.Shutdown(shutdownCtx); err != nil {
		ot.l.DPanic("Shutdown exited with unexpected error", zap.Error(err))
	}

	ot.l.Info("OTLP tracer stopped.")
}
