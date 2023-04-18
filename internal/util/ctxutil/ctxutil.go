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

// Package ctxutil provides context helpers.
package ctxutil

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// WithDelay returns a context that is canceled after a given amount of time after done channel is closed.
func WithDelay(done <-chan struct{}, delay time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		select {
		case <-ctx.Done():
			return

		case <-done:
			t := time.NewTimer(delay)
			defer t.Stop()

			select {
			case <-ctx.Done():
				return
			case <-t.C:
				cancel()
			}
		}
	}()

	return ctx, cancel
}

// Sleep pauses the current goroutine until d has passed or ctx is canceled.
func Sleep(ctx context.Context, d time.Duration) {
	sleepCtx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	<-sleepCtx.Done()
}

// SleepWithJitter pauses the current goroutine until d + jitter has passed or ctx is canceled.
func SleepWithJitter(ctx context.Context, d time.Duration, attempts int64) {
	sleepCtx, cancel := context.WithTimeout(ctx, DurationWithJitter(d, attempts))
	defer cancel()
	<-sleepCtx.Done()
}

// DurationWithJitter returns an exponential backoff duration based on retry with random jitter.
// The maximum sleep is the cap. The minimum duration is at least 100 milliseconds.
//
// Math/rand is good enough because we don't need the randomness to be cryptographically secure.
func DurationWithJitter(cap time.Duration, retry int64) time.Duration {
	const base = time.Millisecond * 100

	if retry < 1 {
		panic("retry must be nonzero positive number")
	}

	maxMilliseconds := float64(base.Milliseconds()) * math.Pow(2, float64(retry))
	capMilliseconds := float64(cap.Milliseconds())
	lowestValue := int64(math.Min(capMilliseconds, maxMilliseconds))

	return time.Duration(rand.Int63n(lowestValue)) * time.Millisecond
}
