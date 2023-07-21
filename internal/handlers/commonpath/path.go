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

package commonpath

import (
	"errors"
	"strconv"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// FindValuesOpts sets options for FindValues.
type FindValuesOpts struct {
	// IgnoreArrayIndex does not check if an array contains an element at the array index
	IgnoreArrayIndex bool
	// IgnoreArrayElement does not check if an array contains documents that have the key
	IgnoreArrayElement bool
}

// FindValues goes through each key of the path iteratively on doc to find values
// at the suffix of the path. At each key of the path, it checks:
//   - if the document has the key;
//   - if the array contains an element at an index where index equals to the key, and;
//   - if the array contains one or more documents which have the key.
//
// It returns a slice of values and an empty array is returned if no value was found.
func FindValues(doc *types.Document, path types.Path, opts *FindValuesOpts) ([]any, error) {
	if opts == nil {
		opts = new(FindValuesOpts)
	}

	keys := path.Slice()
	inputs := []any{doc}
	var values []any

	for _, key := range keys {
		values = []any{}

		for _, input := range inputs {
			switch input := input.(type) {
			case *types.Document:
				v, err := input.Get(key)
				if err != nil {
					continue
				}

				values = append(values, v)
			case *types.Array:
				if !opts.IgnoreArrayIndex {
					if index, err := strconv.Atoi(key); err == nil {
						// key is an integer, check if that integer is an index of the array
						v, err := input.Get(index)
						if err != nil {
							continue
						}

						values = append(values, v)

						continue
					}
				}

				if opts.IgnoreArrayElement {
					continue
				}

				iter := input.Iterator()
				defer iter.Close()

				for {
					_, v, err := iter.Next()
					if errors.Is(err, iterator.ErrIteratorDone) {
						break
					}

					if err != nil {
						return nil, lazyerrors.Error(err)
					}

					doc, ok := v.(*types.Document)
					if !ok {
						continue
					}

					v, err = doc.Get(key)
					if err != nil {
						continue
					}

					values = append(values, v)
				}
			}
		}

		inputs = values
	}

	return values, nil
}
