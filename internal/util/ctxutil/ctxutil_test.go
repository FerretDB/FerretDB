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

func TestDummy(t *testing.T) {
	// we need at least one test per package to correctly calculate coverage
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
