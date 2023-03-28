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

import "sync/atomic"

// ForSlice returns an iterator over a slice.
func ForSlice[V any](s []V) Interface[int, V] {
	return &sliceIterator[V]{
		s: s,
	}
}

// sliceIterator implements iterator.Interface.
//
//nolint:vet // golangci-lint's govet and gopls's govet could not agree on alignment
type sliceIterator[V any] struct {
	n atomic.Uint32
	s []V
}

// Next implements iterator.Interface.
func (iter *sliceIterator[V]) Next() (int, V, error) {
	n := int(iter.n.Add(1)) - 1

	var zero V
	if n >= len(iter.s) {
		return 0, zero, ErrIteratorDone
	}

	return n, iter.s[n], nil
}

// Close implements iterator.Interface.
func (iter *sliceIterator[V]) Close() {
	iter.n.Store(uint32(len(iter.s)))
}

// check interfaces
var (
	_ Interface[int, any] = (*sliceIterator[any])(nil)
)
