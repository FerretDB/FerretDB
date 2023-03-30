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
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// SortIterator returns an iterator of sorted documents.
//
// Since sorting iterator is impossible, this function fully consumes and closes the underlying iterator,
// sorts documents in memory and returns a new iterator over the sorted slice.
// That iterator should be closed by the caller.
func SortIterator(iter types.DocumentsIterator, sort *types.Document) (types.DocumentsIterator, error) {
	// don't consume all documents if there is no sort
	if sort.Len() == 0 {
		return iter, nil
	}

	docs, err := iterator.ConsumeValues(iterator.Interface[struct{}, *types.Document](iter))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = SortDocuments(docs, sort); err != nil {
		return nil, lazyerrors.Error(err)
	}

	sliceIter := iterator.ForSlice(docs)

	return iterator.Values(sliceIter), nil
}
