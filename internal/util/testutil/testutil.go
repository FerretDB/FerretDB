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

// Package testutil provides testing helpers.
package testutil

import (
	"context"
	"runtime/trace"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Ctx returns test context.
func Ctx(tb testing.TB) context.Context {
	tb.Helper()

	// root span
	ctx, span := otel.Tracer("").Start(context.Background(), tb.Name())

	// make one task per top-level test for nicer histograms
	ctx, task := trace.NewTask(ctx, strings.Split(tb.Name(), "/")[0])

	tb.Cleanup(task.End)
	tb.Cleanup(func() {
		span.End()
	})

	trace.Log(ctx, "test", tb.Name())

	ctx, stop := notifyTestsTermination(ctx)

	go func() {
		<-ctx.Done()

		tb.Log("Stopping...")
		stop()

		// There is a weird interaction between terminal's process group/session,
		// Task's signal handling, and this attempt to handle signals gracefully fails miserably.
		// It may cause tests to continue running in the background
		// while terminal shows command-line prompt already.
		//
		// Panic to surely stop tests.
		panic("Stopping everything")
	}()

	return ctx
}

// Logger returns zap test logger with valid configuration.
func Logger(tb testing.TB, level zap.AtomicLevel) *zap.Logger {
	opts := []zaptest.LoggerOption{
		zaptest.Level(level),
		zaptest.WrapOptions(zap.AddCaller(), zap.Development()),
	}

	return zaptest.NewLogger(tb, opts...)
}
