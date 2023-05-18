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
package observability

import (
	"context"
	"runtime"
	"runtime/trace"

	"github.com/FerretDB/FerretDB/internal/util/resource"
)

type funcCall struct {
	token  *resource.Token
	region *trace.Region
}

func (fc *funcCall) leave() {
	fc.region.End()
	resource.Untrack(fc, fc.token)
}

// FuncCall TODO.
func FuncCall(ctx context.Context) func() {
	fc := &funcCall{
		token: resource.NewToken(),
	}

	resource.Track(fc, fc.token)

	pc := make([]uintptr, 1)
	runtime.Callers(1, pc)
	f, _ := runtime.CallersFrames(pc).Next()
	funcName := f.Function

	fc.region = trace.StartRegion(ctx, funcName)

	return fc.leave
}
