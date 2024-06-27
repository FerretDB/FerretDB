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

// RunOtel runs the OpenTelemetry system with the given configuration.
func RunOtel(ctx context.Context, config OtelConfig, logger *zap.SugaredLogger) {
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
		logger.Errorf("Failed to create OTLP exporter: %s. OpenTelemetry won't be used.", err)
		return
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

	<-ctx.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:mnd // Simple timeout
	defer stopCancel()

	if err = tp.Shutdown(stopCtx); err != nil {
		logger.Errorf("Error while shutdown OpenTelemetry system: %v", err)
		return
	}

	logger.Info("OpenTelemetry system stopped successfully.")
}
