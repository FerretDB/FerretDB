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
	"fmt"
	"sync/atomic"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// LimitIterator returns an iterator that limits a number of documents returned by the underlying iterator.
// It will be added to the given closer.
//
// Next method returns the next document until the limit is reached,
// then it returns iterator.ErrIteratorDone.
//
// Close method closes the underlying iterator.
func LimitIterator(iter types.DocumentsIterator, closer *iterator.MultiCloser, limit int64) types.DocumentsIterator {
	switch {
	case limit == 0:
		return iter
	case limit < 0:
		// limit parameter range should be handled by GetLimitParam.
		// aggregation limit stage allows limit of math.MaxInt64.
		// TODO https://github.com/FerretDB/FerretDB/issues/2255
		panic(fmt.Sprintf("invalid limit value: %d", limit))
	default:
		res := &limitIterator{
			iter:  iter,
			limit: uint32(limit),
		}
		closer.Add(res)

		return res
	}
}

// limitIterator is returned by LimitIterator.
type limitIterator struct {
	iter  types.DocumentsIterator
	n     atomic.Uint32
	limit uint32
}

// Next implements iterator.Interface. See LimitIterator for details.
func (iter *limitIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	n := iter.n.Add(1) - 1

	if n >= iter.limit {
		return unused, nil, iterator.ErrIteratorDone
	}

	return iter.iter.Next()
}

// Close implements iterator.Interface. See LimitIterator for details.
func (iter *limitIterator) Close() {
	iter.iter.Close()
	iter.n.Store(iter.limit)
}

// check interfaces
var (
	_ types.DocumentsIterator = (*limitIterator)(nil)
)
