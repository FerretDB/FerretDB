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

package iterator_test // to avoid import cycle

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/iterator/testiterator"
)

func TestTestIteratorForFunc(t *testing.T) {
	t.Parallel()

	testiterator.TestIterator(t, func() iterator.Interface[struct{}, int] {
		var i int
		var k struct{}

		f := func() (struct{}, int, error) {
			i++
			if i > 3 {
				return k, 0, iterator.ErrIteratorDone
			}

			return k, i, nil
		}

		return iterator.ForFunc(f)
	})
}

func TestTestIteratorForSlice(t *testing.T) {
	t.Parallel()

	testiterator.TestIterator(t, func() iterator.Interface[int, int] {
		return iterator.ForSlice([]int{1, 2, 3})
	})
}
