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

	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// Ctx returns test context.
// It is canceled when test is finished or interrupted.
func Ctx(tb testtb.TB) context.Context {
	tb.Helper()

	signalsCtx, signalsStop := ctxutil.SigTerm(context.Background())

	testDone := make(chan struct{})

	tb.Cleanup(func() {
		close(testDone)
	})

	go func() {
		select {
		case <-testDone:
			signalsStop()

		case <-signalsCtx.Done():
			// There is a weird interaction between terminal's process group/session signal handling,
			// Task's signal handling,
			// and this attempt to handle signals gracefully.
			// It may cause tests to continue running in the background
			// while terminal shows command-line prompt already.
			//
			// Panic to surely stop tests.
			panic("Stopping everything")
		}
	}()

	ctx, span := otel.Tracer("").Start(signalsCtx, tb.Name())
	tb.Cleanup(func() {
		span.End()
	})

	return ctx
}
