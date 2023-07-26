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
	// If FindArrayElements is true, it iterates the array to find document elements containing the key.
	// If FindArrayElements is false, it does not find any element from the array.
	// Using path `v.foo` and `v` is an array:
	// 	- with FindArrayElements true, it returns documents' values of key `foo`;
	//  - with FindArrayElements false, it returns an empty array.
	// If `v` is not an array, FindArrayElements has no impact.
	FindArrayElements bool
	// If FindArrayIndex is true, it uses indexes to find an element of the array.
	// If FindArrayIndex is false, it does use indexes on the path.
	// Using path `v.0` and `v` is an array:
	//  - with FindArrayIndex true, it returns 0-th index element of the array;
	//  - with FindArrayIndex false, it returns empty array.
	// If `v` is not an array, FindArrayIndex has no impact.
	FindArrayIndex bool
}

// FindValues iterates path elements, at each path element it checks:
//   - if it is document and has the key, add it to next values to iterate;
//   - if it is array and FindArrayIndex is true, it finds element by index and add it to next values to iterate;
//   - if it is array and FindArrayElements is true, it finds elements and add it to next values to iterate;
//
// It returns a slice of values or an empty array is returned if no value was found.
func FindValues(doc *types.Document, path types.Path, opts *FindValuesOpts) ([]any, error) {
	if opts == nil {
		opts = new(FindValuesOpts)
	}

	var nextValues = []any{doc}

	for _, p := range path.Slice() {
		values := []any{}

		for _, next := range nextValues {
			switch next := next.(type) {
			case *types.Document:
				v, err := next.Get(p)
				if err != nil {
					continue
				}

				values = append(values, v)

			case *types.Array:
				if opts.FindArrayIndex {
					res, err := findArrayIndex(next, p)
					if err != nil {
						return nil, lazyerrors.Error(err)
					}

					values = append(values, res...)
				}

				if opts.FindArrayElements {
					res, err := findArrayElements(next, p)
					if err != nil {
						return nil, lazyerrors.Error(err)
					}

					values = append(values, res...)
				}

			default:
				// path does not exist in scalar values, nothing to do
			}
		}

		nextValues = values
	}

	return nextValues, nil
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

// findArrayElements finds document elements containing the key and returns the values of the key.
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
