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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// FilterIterator returns an iterator that filters out documents that don't match the filter.
func FilterIterator(iter types.DocumentsIterator, filter *types.Document) types.DocumentsIterator {
	return &filterIterator{
		iter:   iter,
		filter: filter,
	}
}

// filterIterator implements iterator.Interface by filtering out documents that don't match the filter.
type filterIterator struct {
	iter   types.DocumentsIterator
	filter *types.Document
}

// Next implements iterator.Interface by returning the next document that matches the filter.
//
// Returned indexes are the same as in the underlying iterator.
// If the document was filtered out, that index will be skipped.
func (iter *filterIterator) Next() (int, *types.Document, error) {
	for {
		i, doc, err := iter.iter.Next()
		if err != nil {
			return 0, nil, lazyerrors.Error(err)
		}

		matches, err := FilterDocument(doc, iter.filter)
		if err != nil {
			return 0, nil, lazyerrors.Error(err)
		}

		if matches {
			return i, doc, nil
		}
	}
}

// Close implements iterator.Interface by closing the underlying iterator.
func (iter *filterIterator) Close() {
	iter.iter.Close()
}

// check interfaces
var (
	_ types.DocumentsIterator = (*filterIterator)(nil)
)
