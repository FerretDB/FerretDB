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

package stages

import (
	"sync/atomic"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// accumulationIterator iterates on single or no value.
type accumulationIterator struct {
	value *types.Document
	n     atomic.Uint32
}

// EmptyIterator creates an iterator without any value.
func EmptyIterator() types.DocumentsIterator {
	return new(accumulationIterator)
}

// SingleValueIterator creates an iterator for single value.
func SingleValueIterator(value *types.Document) types.DocumentsIterator {
	return &accumulationIterator{value: value}
}

// Next implements Iterator interface.
// It returns the value of the iterator on the first call,
// ErrIteratorDone on subsequent calls. If value is not set, returns ErrIteratorDone.
func (iter *accumulationIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	n := iter.n.Add(1) - 1

	if n >= 1 || iter.value == nil {
		return unused, nil, iterator.ErrIteratorDone
	}

	return unused, iter.value, nil
}

// Close implements iterator.Interface.
func (iter *accumulationIterator) Close() {
}

// check interfaces
var (
	_ types.DocumentsIterator = (*accumulationIterator)(nil)
)
