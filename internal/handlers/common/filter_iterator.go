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
)

func FilterIterator(iter types.DocumentsIterator, filter *types.Document) types.DocumentsIterator {
	return &filterIterator{
		iter:   iter,
		filter: filter,
	}
}

type filterIterator struct {
	iter   types.DocumentsIterator
	filter *types.Document
}

// Next implements iterator.Interface.
func (iter *filterIterator) Next() (int, *types.Document, error) {
	for {
		i, doc, err := iter.iter.Next()
		if err != nil {
			return 0, nil, err
		}

		matches, err := FilterDocument(doc, iter.filter)
		if err != nil {
			return 0, nil, err
		}
		if matches {
			return i, doc, nil
		}
	}
}

// Next implements iterator.Interface.
func (iter *filterIterator) Close() {
	iter.iter.Close()
}

// check interfaces
var (
	_ types.DocumentsIterator = (*filterIterator)(nil)
)
