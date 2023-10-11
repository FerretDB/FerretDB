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
	"errors"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
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
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// Those are always shared between all tests.
var (
	listenerMetrics = connmetrics.NewListenerMetrics()
)

var sharedListenerURI string

// Startup initializes things that should be initialized only once.
func Startup() func() {
	logging.Setup(zap.DebugLevel, "")

	// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
	if t, _ := strconv.ParseBool(os.Getenv("RUNNER_DEBUG")); t {
		zap.S().Info("Enabling setup debug logging on GitHub Actions.")
		flags.debugSetup = true
	}

	prometheus.DefaultRegisterer.MustRegister(listenerMetrics)

	// use any available port to allow running different configurations in parallel
	// TODO https://github.com/FerretDB/FerretDB/issues/3544
	go debug.RunHandler(context.Background(), "127.0.0.1:0", prometheus.DefaultRegisterer, zap.L().Named("debug"))

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startupCancel()

	// do basic flags validation earlier, before all tests

	if flags.benchDocs <= 0 {
		zap.S().Fatal("-bench-docs must be > 0.")
	}

	for _, p := range shareddata.AllBenchmarkProviders() {
		if g, ok := p.(shareddata.BenchmarkGenerator); ok {
			g.Init(flags.benchDocs)
		}
	}

	if flags.targetBackend == "" {
		zap.S().Fatal("-target-backend must be set.")
	}

	if !slices.Contains(allBackends, flags.targetBackend) {
		zap.S().Fatalf("Unknown target backend %q.", flags.targetBackend)
	}

	if u := flags.targetURL; u != "" {
		client, err := makeClient(startupCtx, u)
		if err != nil {
			zap.S().Fatalf("Failed to connect to target system %s: %s", u, err)
		}

		client.Disconnect(startupCtx)

		zap.S().Infof("Target system: %s (%s).", flags.targetBackend, u)
	} else {
		zap.S().Infof("Target system: %s (built-in).", flags.targetBackend)
	}

	if u := flags.compatURL; u != "" {
		client, err := makeClient(startupCtx, u)
		if err != nil {
			zap.S().Fatalf("Failed to connect to compat system %s: %s", u, err)
		}

		client.Disconnect(startupCtx)

		zap.S().Infof("Compat system: MongoDB (%s).", u)
	} else {
		zap.S().Infof("Compat system: none, compatibility tests will be skipped.")
	}

	otlpExporter := must.NotFail(otlptracehttp.New(startupCtx,
		otlptracehttp.WithEndpoint("127.0.0.1:4318"),
		otlptracehttp.WithInsecure(),
	))

	tp := oteltrace.NewTracerProvider(
		oteltrace.WithSpanProcessor(
			oteltrace.NewBatchSpanProcessor(otlpExporter),
		),
		oteltrace.WithSampler(oteltrace.AlwaysSample()),
		oteltrace.WithResource(resource.NewSchemaless(
			otelsemconv.ServiceNameKey.String("FerretDB"),
		)),
	)

	otel.SetTracerProvider(tp)

	var listenerCancel context.CancelFunc
	if flags.shareServer {
		must.BeZero(flags.targetURL)

		var listenerCtx context.Context
		listenerCtx, listenerCancel = context.WithCancel(context.Background())
		setupListener(nil, listenerCtx, logger)

		testutil.Ctx(tb)

		zap.S().Info("listener initialized start", handlerType)
		var listenerCtx context.Context
		listenerCtx, sharedListenerCancelFunc = context.WithCancel(context.Background())
		handlerType, sharedListener = makeListener(listenerCtx, zap.S().Fatalf)

		runDone := make(chan struct{})

		go func() {
			close(runDone)

			err := sharedListener.Run(listenerCtx)
			if err == nil || errors.Is(err, context.Canceled) {
				zap.S().Info("Listener stopped without error")
			} else {
				zap.S().Error("Listener stopped", zap.Error(err))
			}
		}()

		// ensure that all listener's and handler's logs are written before test ends
		defer func() {
			<-runDone
		}()
		zap.S().Info("listener initialized end", handlerType)
	}

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if flags.shareServer && flags.targetURL == "" {
			sharedListenerCancelFunc()
		}

		must.NoError(otlpExporter.Shutdown(ctx))

		// to increase a chance of resource finalizers to spot problems
		runtime.GC()
		runtime.GC()
	}
}
