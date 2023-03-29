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

// Package resource provides utilities for tracking resource lifetimes.
package resource

import (
	"fmt"
	"runtime"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
)

// TODO https://github.com/FerretDB/FerretDB/issues/2314

// Track tracks the lifetime of an object until Untrack is called on it.
func Track(obj any) {
	stack := debugbuild.Stack()

	runtime.SetFinalizer(obj, func(obj any) {
		msg := fmt.Sprintf("%T has not been finalized", obj)
		if stack != nil {
			msg += "\nObject created by " + string(stack)
		}

		panic(msg)
	})
}

// Untrack stops tracking the lifetime of an object.
func Untrack(obj any) {
	runtime.SetFinalizer(obj, nil)
}
