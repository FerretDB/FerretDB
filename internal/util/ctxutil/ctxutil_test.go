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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurationWithJitter(t *testing.T) {
	t.Parallel()

	t.Run("larger or equal then 1ms", func(t *testing.T) {
		sleep := DurationWithJitter(time.Second, 1)
		assert.GreaterOrEqual(t, sleep, time.Millisecond)
	})

	t.Run("less or equal then duration input", func(t *testing.T) {
		sleep := DurationWithJitter(time.Second, 100000)
		assert.LessOrEqual(t, sleep, time.Second)
	})

	t.Run("multiple tasks retry multiple times", func(t *testing.T) {
		// This test outputs a file for duration it took all nTasks to retry nRetries.
		// In reality not all tasks will retry, but this is good enough for visualising it.

		nTasks := 100
		nRetries := 5

		durations := make([][]time.Duration, nTasks) // task -> retry count -> duration

		for i := 0; i < nTasks; i++ {
			durations[i] = make([]time.Duration, nRetries)
			for j := 0; j < nRetries; j++ {
				durations[i][j] = DurationWithJitter(time.Second, int64(j+1))
			}
		}

		dir := filepath.Join("result")
		err := os.MkdirAll(dir, 0o777)
		require.NoError(t, err)

		filename := filepath.Join(dir, "multiple-retry-jitter.txt")
		f, err := os.Create(filename)
		require.NoError(t, err)

		defer f.Close()

		for _, task := range durations {
			for j, duration := range task {
				// each line has retry count (j+1) and duration waited in milliseconds.
				fmt.Fprintln(f, j+1, duration.Milliseconds())
			}
		}
	})
}

func TestDurationWithJitter(t *testing.T) {
	t.Parallel()

	t.Run("larger or equal then 1ms", func(t *testing.T) {
		sleep := DurationWithJitter(time.Second, 1)
		assert.GreaterOrEqual(t, sleep, time.Millisecond)
	})

	t.Run("less or equal then duration input", func(t *testing.T) {
		sleep := DurationWithJitter(time.Second, 100000)
		assert.LessOrEqual(t, sleep, time.Second)
	})

	t.Run("attempt cannot be less then 1", func(t *testing.T) {
		sleep := DurationWithJitter(time.Second, 0)
		assert.LessOrEqual(t, sleep, time.Second)
	})
}
