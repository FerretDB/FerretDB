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

package observability

import (
	"context"
	"runtime"
	"runtime/trace"
	"sync"

	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// funcCall tracks function calls.
type funcCall struct {
	cancelCtx context.CancelFunc
	token     *resource.Token
	region    *trace.Region
	span      oteltrace.Span
}

// FuncCall adds observability to a function or method call.
// It returns a new derived context and its cancel function.
//
// FuncCall should be called at the very beginning of the observable function,
// and the returned cancel function should be called at exit.
// The returned context should be passed down.
//
// The only valid way to use FuncCall is:
//
//	func foo(ctx context.Context) {
//		ctx, cancel := FuncCall(ctx)
//		defer cancel()
//
//		bar(ctx)
//		// ...
//	}
//
// For the Go execution tracer, FuncCall creates a new region for the function call
// and attaches it to the task in the context (or background task).
//
// For the OpenTelemetry tracer, FuncCall creates a new span for the function call.
func FuncCall(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancelCtx := context.WithCancel(ctx)

	fc := &funcCall{
		cancelCtx: cancelCtx,
		token:     resource.NewToken(),
	}
	resource.Track(fc, fc.token)

	pc := make([]uintptr, 1)
	runtime.Callers(1, pc)
	f, _ := runtime.CallersFrames(pc).Next()
	funcName := f.Function

	if trace.IsEnabled() {
		fc.region = trace.StartRegion(ctx, funcName)
	}

	ctx, fc.span = otel.Tracer("").Start(ctx, funcName)

	return ctx, sync.OnceFunc(fc.cancel)
}

// cancel is called on function exit.
func (fc *funcCall) cancel() {
	fc.span.End()

	if fc.region != nil {
		fc.region.End()
	}

	resource.Untrack(fc, fc.token)

	fc.cancelCtx()
}
