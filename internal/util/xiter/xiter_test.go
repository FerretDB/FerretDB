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

package xiter

import (
	"iter"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunk(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		s := []int{1, 2, 3}

		next, stop := iter.Pull(Chunk(slices.Values(s), 2))
		t.Cleanup(stop)

		v, ok := next()
		assert.Equal(t, []int{1, 2}, v)
		assert.Equal(t, 2, len(v))
		assert.Equal(t, 2, cap(v))
		assert.Equal(t, true, ok)

		v, ok = next()
		assert.Equal(t, []int{3}, v)
		assert.Equal(t, 1, len(v))
		assert.Equal(t, 1, cap(v))
		assert.Equal(t, true, ok)

		v, ok = next()
		assert.Equal(t, []int(nil), v)
		assert.Equal(t, 0, len(v))
		assert.Equal(t, 0, cap(v))
		assert.Equal(t, false, ok)
	})

	t.Run("Normal2", func(t *testing.T) {
		s := []int{1, 2, 3, 4}

		next, stop := iter.Pull(Chunk(slices.Values(s), 2))
		t.Cleanup(stop)

		v, ok := next()
		assert.Equal(t, []int{1, 2}, v)
		assert.Equal(t, 2, len(v))
		assert.Equal(t, 2, cap(v))
		assert.Equal(t, true, ok)

		v, ok = next()
		assert.Equal(t, []int{3, 4}, v)
		assert.Equal(t, 2, len(v))
		assert.Equal(t, 2, cap(v))
		assert.Equal(t, true, ok)

		v, ok = next()
		assert.Equal(t, []int(nil), v)
		assert.Equal(t, 0, len(v))
		assert.Equal(t, 0, cap(v))
		assert.Equal(t, false, ok)
	})

	t.Run("Break", func(t *testing.T) {
		s := []int{1, 2, 3}

		for c := range Chunk(slices.Values(s), 2) {
			assert.Equal(t, []int{1, 2}, c)
			assert.Equal(t, 2, len(c))
			assert.Equal(t, 2, cap(c))

			break
		}
	})

	t.Run("Empty", func(t *testing.T) {
		s := []int{}

		next, stop := iter.Pull(Chunk(slices.Values(s), 2))
		t.Cleanup(stop)

		v, ok := next()
		assert.Equal(t, []int(nil), v)
		assert.Equal(t, 0, len(v))
		assert.Equal(t, 0, cap(v))
		assert.Equal(t, false, ok)
	})

	t.Run("Nil", func(t *testing.T) {
		s := []int(nil)

		next, stop := iter.Pull(Chunk(slices.Values(s), 2))
		t.Cleanup(stop)

		v, ok := next()
		assert.Equal(t, []int(nil), v)
		assert.Equal(t, 0, len(v))
		assert.Equal(t, 0, cap(v))
		assert.Equal(t, false, ok)
	})
}
