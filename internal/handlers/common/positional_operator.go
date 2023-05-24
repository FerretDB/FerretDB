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

// getPositionalProjection checks validity of the positional operator `$` and
// returns the first element of the array that matches the filter condition.
//
// Returned command error code:
//   - ErrWrongPositionalOperatorLocation when there are multiple `$` or `$` is not at the end.
//   - ErrBadPositionalProjection when array or filter is empty.
//   - ErrBadPositionalProjection when filter does not contain positional operator filter.
func getPositionalProjection(arr *types.Array, filter *types.Document, positionalOperator string) (any, error) {
	if !strings.HasSuffix(positionalOperator, "$") ||
		strings.Count(positionalOperator, "$") > 1 {
		// there can only be one positional operator at the end.
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrWrongPositionalOperatorLocation,
			"Positional projection may only be used at the end, "+
				"for example: a.b.$. If the query previously used a form "+
				"like a.b.$.d, remove the parts following the '$' and "+
				"the results will be equivalent.",
			"projection",
		)
	}

	if arr.Len() == 0 || filter.Len() == 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadPositionalProjection,
			"Executor error during find command :: caused by :: positional operator"+
				" '.$' couldn't find a matching element in the array",
			"projection",
		)
	}

	positionalOperatorPath := strings.Replace(positionalOperator, ".$", "", 1)

	iter := filter.Iterator()
	defer iter.Close()

	for {
		filterKey, filterVal, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			// filterKey did not contain projection path.
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadPositionalProjection,
				"Executor error during find command :: caused by :: positional operator"+
					" '.$' couldn't find a matching element in the array",
				"projection",
			)
		}

		if err != nil {
			return nil, err
		}

		if filterKey != positionalOperatorPath {
			continue
		}

		expr, ok := filterVal.(*types.Document)

		if !ok {
			// filterVal may be different number type compared to the
			// first element in the array which matched the condition,
			// so iterate array to find the first match.
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

			// array element does not have a key, use positional operator `$` as key,
			// it would be used for error message but there will not be error because
			// filterFieldExpr(...) has already been call before projection.
			key := "$"
			doc := must.NotFail(types.NewDocument(key, elem))

			// filterFieldExpr handles operators present in expr.
			matched, err := filterFieldExpr(doc, key, key, expr)
			if err != nil {
				return nil, err
			}

			if matched {
				return elem, nil
			}
		}
	}
}
