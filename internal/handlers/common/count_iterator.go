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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// CountIterator returns an iterator that returns a single document containing
// the number of input documents in the specified field: {field: count}.
// It will be added to the given closer.
//
// Next method returns that document, subsequent calls return ErrIteratorDone.
// If input iterator contains no document, it returns ErrIteratorDone.
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
	done  bool
}

// Next implements iterator.Interface. See FilterIterator for details.
func (iter *countIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	if iter.done {
		return unused, nil, iterator.ErrIteratorDone
	}

	var count int32
	for {
		_, _, err := iter.iter.Next()
		if err != nil {
			iter.done = true

			if errors.Is(err, iterator.ErrIteratorDone) {
				if count == 0 {
					return unused, nil, iterator.ErrIteratorDone
				}

				return unused, must.NotFail(types.NewDocument(iter.field, count)), nil
			}

			return unused, nil, lazyerrors.Error(err)
		}

		count++
	}
}

// Close implements iterator.Interface. See CountIterator for details.
func (iter *countIterator) Close() {
	iter.iter.Close()
}

// check interfaces
var (
	_ types.DocumentsIterator = (*countIterator)(nil)
)
