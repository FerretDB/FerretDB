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

// DistinctIterator returns an iterator that returns a single document containing
// the distinct values from documents on the specified key.
// It will be added to the given closer.
//
// Next method return the distinct value of all documents in iterator.
//
// Close method closes the underlying iterator.
func DistinctIterator(iter types.DocumentsIterator, closer *iterator.MultiCloser, key string) (types.DocumentsIterator, error) {
	res := &distinctIterator{
		iter: iter,
	}

	path, err := types.NewPathFromString(key)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	res.path = path

	closer.Add(res)

	return res, nil
}

// distinctIterator is returned by DistinctIterator.
type distinctIterator struct {
	iter types.DocumentsIterator
	path types.Path
	done atomic.Bool
}

// Next implements Iterator Interface.
// The first call returns the array of distinct documents from underlyin iterator,
// subsequent calls return ErrIteratorDone.
// If iterator contains no document, it returns empty array.
func (iter *distinctIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	distinct := types.MakeArray(0)

	done := iter.done.Swap(true)
	if done {
		// subsequent calls return error.
		return unused, nil, iterator.ErrIteratorDone
	}

	for {
		_, doc, err := iter.iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return unused, nil, lazyerrors.Error(err)
		}

		suffix, docsAtSuffix := getDocumentsAtSuffix(doc, iter.path)

		for _, doc := range docsAtSuffix {
			val, err := doc.Get(suffix)
			if err != nil {
				continue
			}

			switch val := val.(type) {
			case *types.Array:
				valIter := val.Iterator()
				defer valIter.Close()

				for {
					_, el, err := valIter.Next()
					if errors.Is(err, iterator.ErrIteratorDone) {
						break
					}
					if err != nil {
						return unused, nil, lazyerrors.Error(err)
					}

					if !distinct.Contains(el) {
						distinct.Append(el)
					}
				}

			default:
				if !distinct.Contains(val) {
					distinct.Append(val)
				}
			}
		}
	}

	SortArray(distinct, types.Ascending)

	return unused, must.NotFail(types.NewDocument(iter.path.Suffix(), distinct)), nil
}

func (iter *distinctIterator) Close() {
	iter.iter.Close()
}
