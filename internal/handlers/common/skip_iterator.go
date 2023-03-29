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
)

// SkipIterator returns an iterator that skips a number of documents returned by the underlying iterator.
//
// Next method returns the next document after skipping a number of documents.
//
// Close method closes the underlying iterator.
// For that reason, there is no need to track both iterators.
func SkipIterator(iter types.DocumentsIterator, skip int64) types.DocumentsIterator {
	switch {
	case skip == 0:
		return iter
	case skip < 0:
		// handled by GetSkipParam
		panic(fmt.Sprintf("invalid skip value: %d", skip))
	default:
		return &skipIterator{
			iter: iter,
			skip: uint32(skip),
		}
	}
}

// skipIterator is returned by SkipIterator.
type skipIterator struct {
	iter types.DocumentsIterator
	n    atomic.Uint32
	skip uint32
}

// Next implements iterator.Interface. See SkipIterator for details.
func (iter *skipIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	for {
		if n := iter.n.Add(1) - 1; n >= iter.skip {
			break
		}

		if _, _, err := iter.iter.Next(); err != nil {
			return unused, nil, err
		}
	}

	return iter.iter.Next()
}

// Close implements iterator.Interface. See SkipIterator for details.
func (iter *skipIterator) Close() {
	iter.iter.Close()
	iter.n.Store(iter.skip)
}

// check interfaces
var (
	_ types.DocumentsIterator = (*skipIterator)(nil)
)
