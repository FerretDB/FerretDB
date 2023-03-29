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

package types

import (
	"sync/atomic"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// arrayIterator represents an iterator for an Array.
type arrayIterator struct {
	n   atomic.Uint32
	arr *Array
}

// newArrayIterator returns a new arrayIterator.
func newArrayIterator(array *Array) iterator.Interface[int, any] {
	iter := &arrayIterator{
		arr: array,
	}

	resource.Track(iter)

	return iter
}

// Next implements iterator.Interface.
func (iter *arrayIterator) Next() (int, any, error) {
	n := int(iter.n.Add(1)) - 1

	if n >= iter.arr.Len() {
		return 0, nil, iterator.ErrIteratorDone
	}

	return n, iter.arr.s[n], nil
}

// Close implements iterator.Interface.
func (iter *arrayIterator) Close() {
	iter.n.Store(uint32(iter.arr.Len()))

	resource.Untrack(iter)
}

// check interfaces
var (
	_ iterator.Interface[int, any] = (*arrayIterator)(nil)
)
