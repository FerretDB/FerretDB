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
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// errDelayed is returned by [context.Cause] when [WithDelay]'s context is canceled after delay.
var errDelayed = errors.New("context canceled after delay")

// WithDelay returns a copy of the parent context (with its values), which is canceled
// when returned [context.CancelCauseFunc] is called (without any delay),
// or when the parent is canceled and 3 seconds have passed.
func WithDelay(parent context.Context) (context.Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(context.WithoutCancel(parent))

	go func() {
		select {
		case <-ctx.Done():
			cancel(nil)

		case <-parent.Done():
			t := time.NewTimer(3 * time.Second)
			defer t.Stop()

			select {
			case <-ctx.Done():
				cancel(nil)
			case <-t.C:
				cancel(errDelayed)
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
//
// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/811
func SleepWithJitter(ctx context.Context, d time.Duration, attempt int64) {
	sleepCtx, cancel := context.WithTimeout(ctx, durationWithJitter(d, attempt))
	defer cancel()
	<-sleepCtx.Done()
}

// durationWithJitter returns an exponential backoff duration based on attempt with random "full jitter".
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
//
// The maximum sleep is the cap. The minimum sleep is at least 3 milliseconds.
// Provided cap must be larger than minimum sleep, and attempt number must be a positive number.
func durationWithJitter(cap time.Duration, attempt int64) time.Duration {
	const base = 100      // ms
	const minDuration = 3 // ms

	capDuration := cap.Milliseconds()

	if attempt < 1 {
		panic("attempt must be positive number")
	}

	if capDuration <= minDuration {
		panic(fmt.Sprintf("cap must be larger than min duration (%dms)", minDuration))
	}

	// calculate base backoff based on base duration and amount of attempts
	backoff := float64(base * math.Pow(2, float64(attempt)))
	// cap is a max limit of possible durations returned
	maxDuration := int64(math.Min(float64(capDuration), backoff))

	// Math/rand is good enough because we don't need the randomness to be cryptographically secure.
	sleep := rand.Int63n(maxDuration-minDuration) + minDuration

	return time.Duration(sleep) * time.Millisecond
}
