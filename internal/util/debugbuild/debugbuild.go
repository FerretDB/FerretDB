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

// Package debugbuild provides information about whatever this is a debug build or not.
//
// It is a separate package to avoid dependency cycles.
package debugbuild

import "runtime/debug"

// Stack returns a formatted stack trace of the goroutine that calls it for debug builds.
// For non-debug builds, it returns nil.
func Stack() []byte {
	if enabled {
		return debug.Stack()
	}

	return nil
}
