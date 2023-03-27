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
//
// Next method returns the next document that matches the filter.
//
// Close method closes the underlying iterator.
// For that reason, there is no need to track both iterators.
func FilterIterator(iter types.DocumentsIterator, filter *types.Document) types.DocumentsIterator {
	return &filterIterator{
		iter:   iter,
		filter: filter,
	}
}

// filterIterator is returned by FilterIterator.
type filterIterator struct {
	iter   types.DocumentsIterator
	filter *types.Document
}

// Next implements iterator.Interface. See FilterIterator for details.
func (iter *filterIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	for {
		_, doc, err := iter.iter.Next()
		if err != nil {
			return unused, nil, lazyerrors.Error(err)
		}

		matches, err := FilterDocument(doc, iter.filter)
		if err != nil {
			return unused, nil, lazyerrors.Error(err)
		}

		if matches {
			return unused, doc, nil
		}
	}
}

// Close implements iterator.Interface. See FilterIterator for details.
func (iter *filterIterator) Close() {
	iter.iter.Close()
}

// check interfaces
var (
	_ types.DocumentsIterator = (*filterIterator)(nil)
)
