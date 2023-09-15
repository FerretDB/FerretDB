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

package setup

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// listenerMetrics are shared between tests.
var listenerMetrics = connmetrics.NewListenerMetrics()

// exporter is a shared OTLP http exporter for tests.
var exporter *otlptrace.Exporter

// Startup initializes things that should be initialized only once.
func Startup() {
	logging.Setup(zap.DebugLevel, "")

	// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
	if os.Getenv("RUNNER_DEBUG") == "1" {
		zap.S().Info("Enabling setup debug logging on GitHub Actions.")
		*debugSetupF = true
	}

	prometheus.DefaultRegisterer.MustRegister(listenerMetrics)

	// use any available port to allow running different configurations in parallel
	go debug.RunHandler(context.Background(), "127.0.0.1:0", prometheus.DefaultRegisterer, zap.L().Named("debug"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// do basic flags validation earlier, before all tests

	for _, p := range shareddata.AllBenchmarkProviders() {
		if g, ok := p.(shareddata.BenchmarkGenerator); ok {
			g.Init(*benchDocsF)
		}
	}

	if *targetBackendF == "" {
		zap.S().Fatal("-target-backend must be set.")
	}

	if !slices.Contains(allBackends, *targetBackendF) {
		zap.S().Fatalf("Unknown target backend %q.", *targetBackendF)
	}

	if u := *targetURLF; u != "" {
		client, err := makeClient(ctx, u)
		if err != nil {
			zap.S().Fatalf("Failed to connect to target system %s: %s", u, err)
		}

		client.Disconnect(ctx)

		zap.S().Infof("Target system: %s (%s).", *targetBackendF, u)
	} else {
		zap.S().Infof("Target system: %s (built-in).", *targetBackendF)
	}

	if u := *compatURLF; u != "" {
		client, err := makeClient(ctx, u)
		if err != nil {
			zap.S().Fatalf("Failed to connect to compat system %s: %s", u, err)
		}

		client.Disconnect(ctx)

		zap.S().Infof("Compat system: MongoDB (%s).", u)
	} else {
		zap.S().Infof("Compat system: none, compatibility tests will be skipped.")
	}

	exporter = must.NotFail(otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("127.0.0.1:4318"),
		otlptracehttp.WithInsecure(),
	))

	tp := oteltrace.NewTracerProvider(
		oteltrace.WithSpanProcessor(
			oteltrace.NewBatchSpanProcessor(exporter),
		),
		oteltrace.WithSampler(oteltrace.AlwaysSample()),
		oteltrace.WithResource(resource.NewSchemaless(
			otelsemconv.ServiceNameKey.String("FerretDB"),
		)),
	)

	otel.SetTracerProvider(tp)
}

// Shutdown cleans up after all tests.
func Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	must.NoError(exporter.Shutdown(ctx))

	// to increase a chance of resource finalizers to spot problems
	runtime.GC()
	runtime.GC()
}
