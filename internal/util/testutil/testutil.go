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
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Ctx returns test context.
func Ctx(tb testing.TB) context.Context {
	tb.Helper()

	ctx, stop := notifyTestsTermination(context.Background())

	go func() {
		<-ctx.Done()
		tb.Log("Stopping...")
		stop()
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
