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
	"fmt"
	"sync"

	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// NextFunc is a part of Interface for the Next method.
type NextFunc[K, V any] func() (K, V, error)

// valuesIterator implements iterator.Interface.
//
//nolint:vet // for readability
type funcIterator[K, V any] struct {
	m sync.Mutex
	f NextFunc[K, V]

	token *resource.Token
}

// ForFunc returns an iterator for the given function.
func ForFunc[K, V any](f NextFunc[K, V]) Interface[K, V] {
	iter := &funcIterator[K, V]{
		f:     f,
		token: resource.NewToken(),
	}

	resource.Track(iter, iter.token)

	return iter
}

// Next implements iterator.Interface.
func (iter *funcIterator[K, V]) Next() (K, V, error) {
	iter.m.Lock()
	defer iter.m.Unlock()

	if iter.f == nil {
		var k K
		var v V

		return k, v, fmt.Errorf("%w (f is nil)", ErrIteratorDone)
	}

	return iter.f()
}

// Close implements iterator.Interface.
func (iter *funcIterator[K, V]) Close() {
	iter.m.Lock()
	defer iter.m.Unlock()

	iter.f = nil

	resource.Untrack(iter, iter.token)
}

// check interfaces
var (
	_ Interface[any, any] = (*funcIterator[any, any])(nil)
	_ NextFunc[any, any]  = (*funcIterator[any, any])(nil).Next
	_ Closer              = (*funcIterator[any, any])(nil)
)
