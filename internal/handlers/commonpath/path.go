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
	"strconv"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// FindValuesOpts sets options for FindValues.
type FindValuesOpts struct {
	// IgnoreArrayIndex ignores index dot notation for array
	IgnoreArrayIndex bool
	// IgnoreArrayElement does not iterate array elements
	IgnoreArrayElement bool
}

// FindValues go through each key of the path iteratively to
// find values that exist at suffix.
// An array may return multiple values.
// At each key of the path, it checks:
//   - if the document has the key;
//   - if the array contains an index that is equal to the key, and;
//   - if the array contains documents which has the key.
//
// It returns a slice of values at suffix. An empty array is returned
// if no value was found.
func FindValues(doc *types.Document, path types.Path, opts FindValuesOpts) []any {
	keys := path.Slice()
	vals := []any{doc}

	for _, key := range keys {
		// embeddedVals are the values found at current key.
		embeddedVals := []any{}

		for _, valAtKey := range vals {
			switch val := valAtKey.(type) {
			case *types.Document:
				embeddedVal, err := val.Get(key)
				if err != nil {
					continue
				}

				embeddedVals = append(embeddedVals, embeddedVal)
			case *types.Array:
				if !opts.IgnoreArrayIndex {
					if index, err := strconv.Atoi(key); err == nil {
						// key is an integer, check if that integer is an index of the array.
						embeddedVal, err := val.Get(index)
						if err != nil {
							// index does not exist.
							continue
						}

						// key is the index of the array, add embedded value to the next iteration.
						embeddedVals = append(embeddedVals, embeddedVal)

						continue
					}
				}

				if opts.IgnoreArrayElement {
					continue
				}

				// iterate elements to get documents that contain the key.
				for j := 0; j < val.Len(); j++ {
					elem := must.NotFail(val.Get(j))

					docElem, isDoc := elem.(*types.Document)
					if !isDoc {
						continue
					}

					embeddedVal, err := docElem.Get(key)
					if err != nil {
						continue
					}

					embeddedVals = append(embeddedVals, embeddedVal)
				}

			default:
				// not a document or array, do nothing
			}
		}

		vals = embeddedVals
	}

	return vals
}
