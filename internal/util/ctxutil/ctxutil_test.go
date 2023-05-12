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

package ctxutil

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurationWithJitter(t *testing.T) {
	t.Parallel()

	t.Run("OneRetry", func(t *testing.T) {
		sleep := DurationWithJitter(time.Second, 1)
		assert.GreaterOrEqual(t, sleep, 3*time.Millisecond)
		assert.LessOrEqual(t, sleep, 1*time.Second)
	})

	t.Run("ManyRetries", func(t *testing.T) {
		sleep := DurationWithJitter(time.Second, 100000)
		assert.GreaterOrEqual(t, sleep, 3*time.Millisecond)
		assert.LessOrEqual(t, sleep, time.Second)
	})

	t.Run("TooLowCap", func(t *testing.T) {
		assert.Panics(t, func() {
			DurationWithJitter(2*time.Millisecond, 10000)
		})
		assert.Panics(t, func() {
			DurationWithJitter(3*time.Millisecond, 10000)
		})
	})

	t.Run("RetryMultipleTimes", func(t *testing.T) {
		t.Skip("test used only to generate data")

		dir := filepath.Join("result")
		err := os.MkdirAll(dir, 0o754)
		require.NoError(t, err)

		filename := filepath.Join(dir, "multiple-retry-jitter.txt")
		f, err := os.Create(filename)
		require.NoError(t, err)

		defer f.Close()

		for i := 1; i <= 100; i++ {
			t.Logf("simulating %d clients...", i)
			total := simulateCompetingClients(i)
			fmt.Fprintf(f, "%d\t%d\n", i, total)
		}
	})
}

// simulateCompetingClients simulates amount of clients competing on a single resource.
// They use duration returned by DurationWithJitter before calling the resource again.
// It returns total amount of calls done by all clients.
func simulateCompetingClients(clients int) int64 {
	ch := make(chan struct{})

	go func() {
		for {
			ch <- struct{}{}
			time.Sleep(time.Duration(rand.Intn(18)+2) * time.Millisecond)
		}
	}()

	totalCalls := atomic.Int64{}

	call := func() bool {
		totalCalls.Add(1)
		select {
		case <-ch:
			return true
		default:
			return false
		}
	}

	wg := sync.WaitGroup{}

	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func() {
			for retry := 1; retry < 1000; retry++ {
				if call() {
					wg.Done()
					return
				}
				time.Sleep(DurationWithJitter(200*time.Millisecond, int64(retry)))
			}
		}()
	}

	wg.Wait()

	return totalCalls.Load()
}
