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

// withCloseIterator wraps an iterator with a custom close function.
type withCloseIterator[K, V any] struct {
	iter  Interface[K, V]
	close func()
}

// WithClose wraps an iterator with a custom close function.
//
// That function should call Close() method of the wrapped iterator.
func WithClose[K, V any](iter Interface[K, V], close func()) Interface[K, V] {
	return &withCloseIterator[K, V]{
		iter:  iter,
		close: close,
	}
}

// Next implements iterator.Interface by calling Next() method of the wrapped iterator.
func (iter *withCloseIterator[K, V]) Next() (K, V, error) {
	return iter.iter.Next()
}

// Close implements iterator.Interface by calling the provided close function.
func (iter *withCloseIterator[K, V]) Close() {
	// we might want to wrap it with sync.Once if needed
	iter.close()
}

// check interfaces
var (
	_ Interface[any, any] = (*withCloseIterator[any, any])(nil)
	_ Closer              = (*withCloseIterator[any, any])(nil)
)
