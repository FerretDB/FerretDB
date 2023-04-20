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
	"errors"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// ConsumeValues consumes all values from iterator until it is done.
// ErrIteratorDone error is returned as nil; any other error is returned as-is.
//
// Iterator is always closed at the end.
func ConsumeValues[K, V any](iter Interface[K, V]) ([]V, error) {
	defer iter.Close()

	var res []V

	for {
		_, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, ErrIteratorDone) {
				return res, nil
			}

			return nil, lazyerrors.Error(err)
		}

		res = append(res, v)
	}
}

// ConsumeValuesN consumes up to n values from iterator or until it is done.
// ErrIteratorDone error is returned as nil; any other error is returned as-is.
//
// Iterator is closed when it is done or on any error.
// Simply consuming n values does not close the iterator.
//
// Consuming already done iterator returns (nil, nil).
// The same result is returned for n = 0.
func ConsumeValuesN[K, V any](iter Interface[K, V], n int) ([]V, error) {
	var res []V

	for i := 0; i < n; i++ {
		_, v, err := iter.Next()
		if err != nil {
			iter.Close()

			if errors.Is(err, ErrIteratorDone) {
				break
			}

			return nil, lazyerrors.Error(err)
		}

		if res == nil {
			res = make([]V, 0, n)
		}

		res = append(res, v)
	}

	return res, nil
}

// Values returns an iterator over values of another iterator.
//
// Close method closes the underlying iterator.
// For that reason, there is no need to track both iterators.
func Values[K, V any](iter Interface[K, V]) Interface[struct{}, V] {
	return &valuesIterator[K, V]{
		iter: iter,
	}
}

// valuesIterator implements iterator.Interface.
type valuesIterator[K, V any] struct {
	iter Interface[K, V]
}

// Next implements iterator.Interface.
func (iter *valuesIterator[K, V]) Next() (struct{}, V, error) {
	_, v, err := iter.iter.Next()
	return struct{}{}, v, err
}

// Close implements iterator.Interface.
func (iter *valuesIterator[K, V]) Close() {
	iter.iter.Close()
}

// check interfaces
var (
	_ Interface[struct{}, any] = (*valuesIterator[any, any])(nil)
	_ Closer                   = (*valuesIterator[any, any])(nil)
)
