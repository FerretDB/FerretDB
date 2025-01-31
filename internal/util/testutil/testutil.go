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
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// contextKey is a named unexported type for the safe use of [context.WithValue].
type contextKey struct{}

// Context key for [fileLock] in context returned by [Ctx].
var fileLockKey = contextKey{}

// Ctx returns test context.
// It is canceled when test is finished or interrupted.
func Ctx(tb testing.TB) context.Context {
	tb.Helper()

	signalsCtx, signalsStop := ctxutil.SigTerm(context.Background())

	start := time.Now()

	fl := newFileLock(tb)

	if d := time.Since(start); d > 1*time.Millisecond {
		fl.tb.Logf("%s got shared flock in %s.", fl.tb.Name(), d)
	}

	signalsCtx = context.WithValue(signalsCtx, fileLockKey, fl)

	testDone := make(chan struct{})

	tb.Cleanup(func() {
		fl.Unlock()
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

// Exclusive signals that test calling that function can't be run in parallel with any other test
// that uses [Ctx] to get test context, including tests in other packages.
//
// The bar for using this helper is very high.
// Most tests can run in parallel with other tests just fine by retrying operations, filtering results,
// or using different instances of system under test (collections, databases, etc).
func Exclusive(ctx context.Context, reason string) {
	fl := ctx.Value(fileLockKey).(*fileLock)
	must.NotBeZero(fl)

	fl.tb.Helper()

	require.NotEmpty(fl.tb, reason)
	fl.tb.Logf("%s waits for exclusive flock: %s.", fl.tb.Name(), reason)

	start := time.Now()

	fl.Lock()

	fl.tb.Logf("%s got exclusive flock in %s.", fl.tb.Name(), time.Since(start))
}
