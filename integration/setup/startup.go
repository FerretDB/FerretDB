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
	"slices"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/observability"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// listenerMetrics are shared between tests.
var listenerMetrics = connmetrics.NewListenerMetrics()

// shutdownOtel is a function that stops OpenTelemetry's tracer provider.
var shutdownOtel context.CancelFunc

// Startup initializes things that should be initialized only once.
func Startup() {
	logging.Setup(zap.DebugLevel, "console", "")

	// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
	if t, _ := strconv.ParseBool(os.Getenv("RUNNER_DEBUG")); t {
		zap.S().Info("Enabling setup debug logging on GitHub Actions.")
		*debugSetupF = true
	}

	prometheus.DefaultRegisterer.MustRegister(listenerMetrics)

	// use any available port to allow running different configurations in parallel
	go debug.RunHandler(context.Background(), "127.0.0.1:0", prometheus.DefaultRegisterer, zap.L().Named("debug"))

	otOpts := observability.OtelTracerOpts{
		Service:  "integration-tests",
		Endpoint: "127.0.0.1:4318",
		Logger:   zap.L().Named("otel"),
	}

	ot, err := observability.NewOtelTracer(&otOpts)
	if err != nil {
		zap.S().Fatal(err)
	}

	go ot.Run(context.TODO())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// do basic flags validation earlier, before all tests

	if *benchDocsF <= 0 {
		zap.S().Fatal("-bench-docs must be > 0.")
	}

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

	if *targetURLF != "" {
		var err error

		*targetURLF, err = setClientPaths(*targetURLF)
		if err != nil {
			zap.S().Fatal(err)
		}

		client, err := makeClient(ctx, *targetURLF)
		if err != nil {
			zap.S().Fatalf("Failed to connect to target system %s: %s", *targetURLF, err)
		}

		client.Disconnect(ctx)

		zap.S().Infof("Target system: %s (%s).", *targetBackendF, *targetURLF)
	} else {
		zap.S().Infof("Target system: %s (built-in).", *targetBackendF)
	}

	if *compatURLF != "" {
		var err error

		*compatURLF, err = setClientPaths(*compatURLF)
		if err != nil {
			zap.S().Fatal(err)
		}

		client, err := makeClient(ctx, *compatURLF)
		if err != nil {
			zap.S().Fatalf("Failed to connect to compat system %s: %s", *compatURLF, err)
		}

		client.Disconnect(ctx)

		zap.S().Infof("Compat system: MongoDB (%s).", *compatURLF)
	} else {
		zap.S().Infof("Compat system: none, compatibility tests will be skipped.")
	}
}

// Shutdown cleans up after all tests.
func Shutdown() {
	shutdownOtel()
	// TODO how do we check that shutdown is complete?

	// to increase a chance of resource finalizers to spot problems
	runtime.GC()
	runtime.GC()
}
