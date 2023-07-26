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

// Package commonpath contains functions used for path.
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
	// FindArrayElements searches by iterating through the whole array to find documents that contains path key.
	// Using path `v.foo` and `v` is an array, it returns all document which has key `foo`.
	// If `v` is not an array, FindArrayElements has no impact.
	FindArrayElements bool
	// FindArrayIndex gets an element at the specified index of an array.
	// Using path `v.0` and `v` is an array, it returns 0-th index element of the array.
	// If `v` is not an array, FindArrayIndex has no impact.
	FindArrayIndex bool
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
				if opts.FindArrayIndex {
					res, err := findArrayIndex(input, key)
					if err != nil {
						return nil, lazyerrors.Error(err)
					}

					values = append(values, res...)
				}

				if opts.FindArrayElements {
					res, err := findArrayElements(input, key)
					if err != nil {
						return nil, lazyerrors.Error(err)
					}

					values = append(values, res...)
				}
			default:
				// key does not exist in scalar values
			}
		}

		inputs = values
	}

	return values, nil
}

// findArrayIndex checks if key can be used as an index of array and returns
// the element found at that index.
// Empty array is returned if the key cannot be used as an index
// or key is not an existing index of array.
func findArrayIndex(array *types.Array, key string) ([]any, error) {
	index, err := strconv.Atoi(key)
	if err != nil {
		return []any{}, nil
	}

	// key is an integer, check if that integer is an index of the array
	v, _ := array.Get(index)
	if v != nil {
		return []any{v}, nil
	}

	return []any{}, nil
}

// findArrayElements finds document elements containing the key and returns the values.
// Empty array is returned if no document elements containing the key was found.
func findArrayElements(array *types.Array, key string) ([]any, error) {
	iter := array.Iterator()
	defer iter.Close()

	res := []any{}

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

		v, _ = doc.Get(key)
		if v != nil {
			res = append(res, v)
		}
	}

	return res, nil
}
