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

// Package teststress provides a helper for stress testing.
//
// It is in a separate package to avoid import cycles.
package teststress

import (
	"runtime"
	"sync"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// NumGoroutines is the total count of goroutines created in Stress function.
var NumGoroutines = runtime.GOMAXPROCS(-1) * 10

// Stress runs function f in multiple goroutines.
//
// Function f should do a needed setup, send a message to ready channel when it is ready to start,
// wait for start channel to be closed, and then do the actual work.
func Stress(tb testutil.TB, f func(ready chan<- struct{}, start <-chan struct{})) {
	tb.Helper()

	// do a bit more work to reduce a chance that one goroutine would finish
	// before the other one is still being created
	var wg sync.WaitGroup
	readyCh := make(chan struct{}, NumGoroutines)
	startCh := make(chan struct{})

	for i := 0; i < NumGoroutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			f(readyCh, startCh)
		}()
	}

	for i := 0; i < NumGoroutines; i++ {
		<-readyCh
	}

	close(startCh)

	wg.Wait()
}
