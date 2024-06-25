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
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelsdkresource "go.opentelemetry.io/otel/sdk/resource"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Config is the configuration for OpenTelemetry.
type Config struct {
	Service          string
	Version          string
	Endpoint         string
	TracesSampler    string
	TracesSamplerArg string
	BSPDelay         time.Duration
}

// ShutdownFunc is a function that shuts down the OpenTelemetry observability system.
type ShutdownFunc func(context.Context) error

// SetupOtel sets up OTLP exporter and tracer provider.
//
// The function returns a shutdown function that should be called when the application is shutting down.
func SetupOtel(config Config) (ShutdownFunc, error) {
	var err error
	var exporter *otlptrace.Exporter

	if exporter, err = otlptracehttp.New(
		context.TODO(),
		otlptracehttp.WithEndpoint(config.Endpoint),
		otlptracehttp.WithInsecure(),
	); err != nil {
		return nil, err
	}

	var sampler otelsdktrace.Sampler

	switch config.TracesSampler {
	case "always_on":
		sampler = otelsdktrace.AlwaysSample()
	case "always_off":
		sampler = otelsdktrace.NeverSample()
	case "traceidratio":
		var ratio float64

		if ratio, err = strconv.ParseFloat(config.TracesSamplerArg, 64); err != nil {
			return nil, errors.New("unsupported trace ID ratio: " + config.TracesSamplerArg)
		}

		sampler = otelsdktrace.TraceIDRatioBased(ratio)

	default:
		return nil, errors.New("unsupported sampler")
	}

	tp := otelsdktrace.NewTracerProvider(
		otelsdktrace.WithBatcher(exporter, otelsdktrace.WithBatchTimeout(config.BSPDelay)),
		otelsdktrace.WithSampler(sampler),
		otelsdktrace.WithResource(otelsdkresource.NewSchemaless(
			otelsemconv.ServiceName(config.Service),
			otelsemconv.ServiceVersion(config.Version),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}
