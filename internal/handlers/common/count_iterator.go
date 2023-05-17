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

package common

import (
	"errors"
	"sync/atomic"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// CountIterator returns an iterator that returns a single document containing
// the number of documents on the specified field.
// It will be added to the given closer.
//
// Next method returns the next document that matches the filter.
//
// Close method closes the underlying iterator.
func CountIterator(iter types.DocumentsIterator, closer *iterator.MultiCloser, field string) types.DocumentsIterator {
	res := &countIterator{
		iter:  iter,
		field: field,
	}

	closer.Add(res)

	return res
}

// countIterator is returned by CountIterator.
type countIterator struct {
	iter  types.DocumentsIterator
	field string
	done  atomic.Bool
}

// Next implements Iterator interface.
// The first call returns the number of documents, subsequent calls return ErrIteratorDone.
// If iterator contains no document, it returns ErrIteratorDone.
func (iter *countIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	done := iter.done.Swap(true)
	if done {
		// subsequent calls return error.
		return unused, nil, iterator.ErrIteratorDone
	}

	// only first call reaches here, safe to use local variable for count.
	var count int32

	for {
		_, _, err := iter.iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return unused, nil, lazyerrors.Error(err)
		}
		count++
	}

	if count == 0 {
		return unused, nil, iterator.ErrIteratorDone
	}

	return unused, must.NotFail(types.NewDocument(iter.field, count)), nil
}

// Close implements iterator.Interface.
func (iter *countIterator) Close() {
	iter.iter.Close()
}

// check interfaces
var (
	_ types.DocumentsIterator = (*countIterator)(nil)
)
