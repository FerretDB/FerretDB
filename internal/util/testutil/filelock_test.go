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

package testutil

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil/teststress"
)

func TestFileLockStress(t *testing.T) {
	var ids atomic.Int32
	var actual int

	expected := teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
		id := ids.Add(1)
		fl := newFileLock(t)

		ready <- struct{}{}

		<-start

		fl.Lock()
		defer fl.Unlock()

		actual++
		t.Logf("goroutine: %3d, actual: %3d", id, actual)
	})

	require.Equal(t, expected, actual)
}
