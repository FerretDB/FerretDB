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
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

var jaegerExporter *jaeger.Exporter

// Startup initializes things that should be initialized only once.
func Startup() {
	logging.Setup(zap.DebugLevel, "")

	go debug.RunHandler(context.Background(), "127.0.0.1:0", prometheus.DefaultRegisterer, zap.L().Named("debug"))

	if p := *targetPortF; p == 0 {
		zap.S().Infof("Target system: in-process FerretDB with %q handler.", *handlerF)
	} else {
		zap.S().Infof("Target system: port %d.", p)
	}

	if p := *compatPortF; p == 0 {
		zap.S().Infof("Compat system: none, compatibility tests will be skipped.")
	} else {
		zap.S().Infof("Compat system: port %d.", p)
	}

	// pass options explicitly to avoid environment variables effects
	exporter := must.NotFail(jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint("http://127.0.0.1:14268/api/traces"),
		jaeger.WithUsername(""),
		jaeger.WithPassword(""),
		jaeger.WithHTTPClient(http.DefaultClient),
	)))

	tp := oteltrace.NewTracerProvider(
		oteltrace.WithBatcher(exporter),
		oteltrace.WithSampler(oteltrace.AlwaysSample()),
		oteltrace.WithResource(resource.NewSchemaless(
			otelsemconv.ServiceNameKey.String("FerretDB"),
		)),
	)

	// Register TracerProvider globally to use it by default
	otel.SetTracerProvider(tp)
}

// Shutdown cleans up after all tests.
func Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	must.NoError(jaegerExporter.Shutdown(ctx))
}
