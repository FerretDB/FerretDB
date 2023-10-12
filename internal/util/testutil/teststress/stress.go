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
	"context"
	"runtime"
	"sync"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// Stress runs function f in multiple goroutines.
// It returns the number of created goroutines.
//
// Function f should do a needed setup, send a message to ready channel when it is ready to start,
// wait for start channel to be closed, and then do the actual work.
func Stress(tb testtb.TB, f func(ready chan<- struct{}, start <-chan struct{})) int {
	tb.Helper()

	n := runtime.GOMAXPROCS(-1) * 10
	StressN(tb, n, f)

	return n
}

// StressN is a variant of Stress that allows to specify the number of goroutines.
//
// Prefer using Stress when possible.
func StressN(tb testtb.TB, n int, f func(ready chan<- struct{}, start <-chan struct{})) {
	tb.Helper()

	// do a bit more work to reduce a chance that one goroutine would finish
	// before the other one is still being created
	var wg sync.WaitGroup
	readyCh := make(chan struct{}, n)
	startCh := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	tb.Cleanup(cancel)

	for i := 0; i < n; i++ {
		wg.Add(1)

		go func() {
			var ok bool

			defer func() {
				wg.Done()

				// handles f calling testify/require.XXX or `testing.TB.FailNow()`
				if !ok {
					cancel()
				}
			}()

			f(readyCh, startCh)

			ok = true
		}()
	}

	for i := 0; i < n; i++ {
		select {
		case _, ok := <-readyCh:
			if !ok {
				tb.Error("f closed ready channel")
			}
		case <-ctx.Done():
		}
	}

	close(startCh)

	wg.Wait()
}
