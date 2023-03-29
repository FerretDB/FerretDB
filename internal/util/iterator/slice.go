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

package iterator

import (
	"sync"

	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// ForSlice returns an iterator over a slice.
func ForSlice[V any](s []V) Interface[int, V] {
	res := &sliceIterator[V]{
		s: s,
	}

	resource.Track(res)

	return res
}

// sliceIterator implements iterator.Interface.
//
//nolint:vet // golangci-lint's govet and gopls's govet could not agree on alignment
type sliceIterator[V any] struct {
	m sync.Mutex
	n uint32
	s []V
}

// Next implements iterator.Interface.
func (iter *sliceIterator[V]) Next() (int, V, error) {
	iter.m.Lock()
	defer iter.m.Unlock()

	iter.n++
	n := int(iter.n) - 1

	var zero V
	if n >= len(iter.s) {
		return 0, zero, ErrIteratorDone
	}

	return n, iter.s[n], nil
}

// Close implements iterator.Interface.
func (iter *sliceIterator[V]) Close() {
	iter.m.Lock()
	defer iter.m.Unlock()

	iter.s = nil

	resource.Untrack(iter)
}

// check interfaces
var (
	_ Interface[int, any] = (*sliceIterator[any])(nil)
)
