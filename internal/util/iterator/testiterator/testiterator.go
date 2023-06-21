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

package testiterator

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/teststress"
)

// TestIterator checks that the iterator implementation is correct.
func TestIterator[K, V any](t *testing.T, newIter func() iterator.Interface[K, V]) {
	t.Helper()

	// TODO more tests https://github.com/FerretDB/FerretDB/issues/2867

	t.Run("Close", func(t *testing.T) {
		t.Parallel()

		iter := newIter()

		teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
			ready <- struct{}{}
			<-start

			iter.Close()
		})
	})
}
