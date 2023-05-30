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

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// getPositionalProjection checks validity of the positional operator
// by checking the filter contains a key for the path specified by positional operator
// and returns the first element of the array that matches the filter condition.
//
// It takes following arguments:
//   - arr is the value found at the projection operator path.
//   - filter is the original query passed to get document during filtering.
//   - path contains specification of projection operator e.g. `v.$`.
//
// Command error codes:
//   - ErrBadPositionalProjection when array or filter at positional projection path is empty;
//   - ErrBadPositionalProjection when there is no filter field key for positional projection path.
//     If positional projection is `v.$`, the filter must contain `v` in the filter key such as `{v: 42}`.
func getPositionalProjection(arr *types.Array, filter *types.Document, positionalOperatorPath string) (*types.Array, error) {
	if arr.Len() == 0 || filter.Len() == 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadPositionalProjection,
			"Executor error during find command :: caused by :: positional operator"+
				" '.$' couldn't find a matching element in the array",
			"projection",
		)
	}

	// path without `.$` suffix of the positional operator
	// is used to check that the filter contains exact key for this.
	path := must.NotFail(types.NewPathFromString(positionalOperatorPath)).TrimSuffix().String()

	iter := arr.Iterator()
	defer iter.Close()

	for {
		_, elem, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			// TODO: https://github.com/FerretDB/FerretDB/issues/2522
			// when none of element satisfies all filter condition, positional
			// operator returns an arbitrary value not empty array.
			return new(types.Array), nil
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		iter := filter.Iterator()
		defer iter.Close()

		var positionalPathFound bool

		for {
			filterKey, filterVal, err := iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				if !positionalPathFound {
					return nil, commonerrors.NewCommandErrorMsgWithArgument(
						// filterKey did not contain path for positional projection.
						// For example, if the positional operator is "v.$",
						// the filter must have key `v` such as {"v": 1}.
						// For nested dot notation such as "v.foo.$",
						// filter must have key `v.foo` such as {"v.foo": 1},
						// and just {"v": 1} is not sufficient.
						commonerrors.ErrBadPositionalProjection,
						"Executor error during find command :: caused by :: positional operator"+
							" '.$' couldn't find a matching element in the array",
						"projection",
					)
				}

				return types.NewArray(elem)
			}

			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			if filterKey != path {
				continue
			}

			positionalPathFound = true

			expr, ok := filterVal.(*types.Document)
			if !ok {
				if types.Compare(elem, filterVal) != types.Equal {
					break
				}

				// elem matched to the current filter field,
				// continue to check for the next filter field.
				continue
			}

			// array element does not have a key, use positional operator `$` as key,
			// it would be used for error message but there will not be error because
			// filterFieldExpr(...) has already been call before projection.
			key := "$"
			doc := must.NotFail(types.NewDocument(key, elem))

			// In filtering filterFieldExpr was called to check if an array
			// matched the filter.
			// In this call, we already know that the array matched the filter,
			// and we want to find out which array element matched the filter.
			matched, err := filterFieldExpr(doc, key, key, expr)
			if err != nil {
				// the array already matched the filter, so it cannot fail.
				panic(err)
			}

			if !matched {
				break
			}
		}
	}
}
