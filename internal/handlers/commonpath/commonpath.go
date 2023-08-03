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
	// If FindArrayDocuments is true, it iterates the array to find documents that have path.
	// If FindArrayDocuments is false, it does not find documents from the array.
	// Using path `v.foo` and `v` is an array:
	//  - with FindArrayDocuments true, it finds values of `foo` of found documents;
	//  - with FindArrayDocuments false, it returns an empty array.
	// If `v` is not an array, FindArrayDocuments has no impact.
	FindArrayDocuments bool
	// If FindArrayIndex is true, it finds value at index of an array.
	// If FindArrayIndex is false, it does not find value at index of an array.
	// Using path `v.0` and `v` is an array:
	//  - with FindArrayIndex true, it finds 0-th index value of the array;
	//  - with FindArrayIndex false, it returns empty array.
	// If `v` is not an array, FindArrayIndex has no impact.
	FindArrayIndex bool
}

// FindValues returns values by path, looking up into arrays.
//
// It iterates path elements, at each path element it adds to next values to iterate:
//   - if it is a document and has path, it adds the document field value to next values;
//   - if it is an array, FindArrayIndex is true and finds value at index, it adds value to next values;
//   - if it is an array, FindArrayDocuments is true and documents in the array have path,
//     it adds field value of all documents that have path to next values.
//
// It returns next values after iterating path elements.
func FindValues(doc *types.Document, path types.Path, opts *FindValuesOpts) ([]any, error) {
	if opts == nil {
		opts = new(FindValuesOpts)
	}

	nextValues := []any{doc}

	for _, e := range path.Slice() {
		values := []any{}

		for _, next := range nextValues {
			switch next := next.(type) {
			case *types.Document:
				v, _ := next.Get(e)
				if v == nil {
					continue
				}

				values = append(values, v)

			case *types.Array:
				if opts.FindArrayIndex {
					res, err := findArrayIndex(next, e)
					if err == nil {
						values = append(values, res)
						continue
					}
				}

				if opts.FindArrayDocuments {
					res, err := lookupArrayDocuments(next, e)
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

// findArrayIndex returns the value by valid array index.
//
// Error is returned if index is not a number or index does not exist in array.
func findArrayIndex(array *types.Array, index string) (any, error) {
	i, err := strconv.Atoi(index)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	v, err := array.Get(i)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return v, nil
}

// lookupArrayDocuments returns values for the given key in array's document.
//
// Non-document array values, documents without that key, etc. are skipped.
func lookupArrayDocuments(array *types.Array, documentKey string) ([]any, error) {
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

		if v, _ = doc.Get(documentKey); v != nil {
			res = append(res, v)
		}
	}

	return res, nil
}
