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
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// getFirstMatchingElement returns the first element of the array that matches the filter condition.
//
// Returns command error code:
//   - ErrBadValue when multiple positional operator, positional operator is not at the end or filter is empty.
//   - ErrBadValue when array is empty.
//   - ErrBadPositionalOperator when filter does not contain filter for positional operator path.
func getFirstMatchingElement(arr *types.Array, filter *types.Document, projection string) (any, error) {
	if !strings.HasSuffix(projection, "$") ||
		strings.Count(projection, "$") > 1 ||
		filter.Len() == 0 {
		// there can only be one positional operator at the end.
		// filter must not be empty.
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			"Executor error during find command :: caused by :: positional operator '.$' element mismatch",
			"projection",
		)
	}

	if arr.Len() == 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			"Executor error during find command :: caused by :: positional operator"+
				" '.$' couldn't find a matching element in the array",
			"projection",
		)
	}

	projectionArrayPath := strings.Replace(projection, ".$", "", 1)

	iter := filter.Iterator()
	defer iter.Close()

	for {
		filterKey, filterVal, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			// filterKey did not contain projection path.
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadPositionalOperator,
				"Executor error during find command :: caused by :: positional operator"+
					" '.$' couldn't find a matching element in the array",
				"projection",
			)
		}

		if err != nil {
			return nil, err
		}

		if filterKey != projectionArrayPath {
			continue
		}

		// filter is for the projection path.
		expr, ok := filterVal.(*types.Document)

		if !ok {
			// filterVal may be different number type compared to the first element in the array
			// which matched the condition, iterate array to find the first match.
			aIter := arr.Iterator()
			defer aIter.Close()

			for {
				_, elem, err := aIter.Next()
				if errors.Is(err, iterator.ErrIteratorDone) {
					panic(fmt.Sprintf("array %v does not contain %v", arr, filterVal))
				}

				if err != nil {
					return nil, err
				}

				if types.Compare(elem, filterVal) == types.Equal {
					return elem, nil
				}
			}
		}

		iter := arr.Iterator()
		defer iter.Close()

		for {
			_, elem, err := iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				panic(fmt.Sprintf("filter %v matched array %v but no element in array matches filter", expr, arr))
			}

			if err != nil {
				return nil, err
			}

			// check array element satisfies filter condition
			matched, err := filterFieldExpr(must.NotFail(types.NewDocument("$", elem)), "$", "$", expr)
			if err != nil {
				return nil, err
			}

			if matched {
				return elem, nil
			}
		}
	}
}
