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
	"errors"
	"sync"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// newUnwindIterator returns an iterator that unwinds documents returned by the underlying iterator.
// It will be added to the given closer.
//
// Next method returns the next unwind document.
//
// Close method closes the underlying iterator.
func newUnwindIterator(iter types.DocumentsIterator, closer *iterator.MultiCloser, expression types.Expression) types.DocumentsIterator { //nolint:lll // for readability
	res := &unwindIterator{
		iter:       iter,
		expression: expression,
	}
	closer.Add(res)

	return res
}

// unwindIterator is returned by UnwindIterator.
type unwindIterator struct {
	// fields are aligned to make fieldalignment linter happy
	iter       types.DocumentsIterator
	expression types.Expression

	arrayIter iterator.Interface[int, any]
	arrayID   any

	m sync.Mutex
}

// Next implements iterator.Interface.
// Unwind returns each element of an array as a document with the same _id as the array.
func (iter *unwindIterator) Next() (unused struct{}, doc *types.Document, err error) {
	iter.m.Lock()
	defer iter.m.Unlock()

	// closes iter.arrayIter upon returning any error.
	defer func(arrayIter iterator.Interface[int, any], err error) {
		if err != nil && iter.arrayIter != nil {
			iter.arrayIter.Close()
		}
	}(iter.arrayIter, err)

	unused = struct{}{}

	if iter.expression == nil {
		err = iterator.ErrIteratorDone
		return unused, nil, err
	}

	suffixKey := iter.expression.GetExpressionSuffix()

	if iter.arrayIter != nil {
		// unwind has existing array that is unwinding.
		var v any
		_, v, err = iter.arrayIter.Next()

		switch {
		case err == nil:
			return unused, must.NotFail(types.NewDocument("_id", iter.arrayID, suffixKey, v)), nil
		case errors.Is(err, iterator.ErrIteratorDone):
			iter.arrayIter.Close()
		default:
			return unused, nil, lazyerrors.Error(err)
		}
	}

	for {
		_, doc, err = iter.iter.Next()
		if err != nil {
			return unused, nil, err
		}

		d := iter.expression.Evaluate(doc)

		switch d := d.(type) {
		case *types.Array:
			// start unwinding an array, set iter.arrayIter and iter.arrayID
			// so subsequent Next() calls use iter.arrayIter to get each element
			// until entire iter.arrayIter is consumed.
			iter.arrayIter = d.Iterator()
			iter.arrayID = must.NotFail(doc.Get("_id"))

			for {
				var v any
				_, v, err = iter.arrayIter.Next()

				if errors.Is(err, iterator.ErrIteratorDone) {
					iter.arrayIter.Close()
					break
				}

				if err != nil {
					return unused, nil, lazyerrors.Error(err)
				}

				return unused, must.NotFail(types.NewDocument("_id", iter.arrayID, suffixKey, v)), nil
			}

		case types.NullType:
			// Ignore Nulls
		default:
			return unused, doc, nil
		}
	}
}

// Close implements iterator.Interface.
func (iter *unwindIterator) Close() {
	iter.iter.Close()
}

// check interfaces
var (
	_ types.DocumentsIterator = (*unwindIterator)(nil)
)
