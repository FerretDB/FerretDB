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
)

// getFirstElement returns the first element of the array that matches the condition.
// When document does not contain an array at projection path,
// it returns the value at projection path as if positional operator is not present.
// If there is more than one array in the projection path, behaviour may be undefined.
func getFirstElement(arr *types.Array, filter *types.Document, projection string) (any, error) {
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
		filterDoc, ok := filterVal.(*types.Document)
		if !ok {
			// TODO: compare with each element of array
			break
		}

		return fetchElementFromDoc(arr, filterDoc)
	}

	return arr.Get(0)
}

// fetchElementFromDoc fetches the first element from array which matches the value of doc.
// If the doc has an operator, operator is applied to array element and the first element
// that satisfies the operator is returned.
func fetchElementFromDoc(arr *types.Array, doc *types.Document) (any, error) {
	keys := doc.Keys()

	values, err := iterator.ConsumeValues(doc.Iterator())
	if err != nil {
		return nil, err
	}

	iter := arr.Iterator()
	defer iter.Close()

	for {
		_, elem, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, err
		}

		for i := 0; i < len(values); i++ {
			// check array element satisfies condition
			k := keys[i]
			v := values[i]

			if strings.HasPrefix(k, "$") {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrNotImplemented,
					fmt.Sprintf("operator %s is not implemented for projection", k),
					"projection",
				)
			}

			// TODO: check array v. elem cannot be an array because nested array is not supported
			if types.Compare(elem, v) == types.Equal {
				return elem, nil
			}
		}
	}

	panic(fmt.Sprintf("filter %v matched array %v but no element in array matches filter", doc, arr))
}
