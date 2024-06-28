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
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelsdkresource "go.opentelemetry.io/otel/sdk/resource"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)

// OtelConfig is the configuration for OpenTelemetry.
type OtelConfig struct {
	Service  string
	Version  string
	Endpoint string
}

// Otel represents the OpenTelemetry system.
type Otel struct {
	l  *zap.Logger
	tp *otelsdktrace.TracerProvider
}

// NewOtel sets up OTLP exporter and tracer provider.
func NewOtel(config *OtelConfig, l *zap.Logger) (*Otel, error) {
	if config.Endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	var exporter *otlptrace.Exporter
	var err error

	// Exporter and tracer are configured with the particular params on purpose.
	// We don't want to let them being set through OTEL_* environment variables,
	// but we set them explicitly.
	exporter, err = otlptracehttp.New(
		context.TODO(),
		otlptracehttp.WithEndpoint(config.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := otelsdktrace.NewTracerProvider(
		otelsdktrace.WithBatcher(exporter, otelsdktrace.WithBatchTimeout(time.Second)),
		otelsdktrace.WithSampler(otelsdktrace.AlwaysSample()),
		otelsdktrace.WithResource(otelsdkresource.NewSchemaless(
			otelsemconv.ServiceName(config.Service),
			otelsemconv.ServiceVersion(config.Version),
		)),
	)

	otel.SetTracerProvider(tp)

	return &Otel{
		l:  l,
		tp: tp,
	}, nil
}

// Run runs the OpenTelemetry system.
func (o *Otel) Run(ctx context.Context) {
	o.l.Info("OpenTelemetry system started successfully.")

	<-ctx.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:mnd // Simple timeout
	defer stopCancel()

	if err := o.tp.Shutdown(stopCtx); err != nil {
		o.l.Error("Error while shutdown OpenTelemetry system.", zap.Error(err))
		return
	}

	o.l.Info("OpenTelemetry system stopped successfully.")
}
