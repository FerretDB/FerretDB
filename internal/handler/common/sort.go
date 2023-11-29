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
	"sort"
	"strings"

	"github.com/FerretDB/FerretDB/internal/handler/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handler/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// SortDocuments sorts given documents in place according to the given sorting conditions.
//
// If sort path is invalid, it returns a possibly wrapped types.PathError.
func SortDocuments(docs []*types.Document, sortDoc *types.Document) error {
	sortFuncs, err := unwrapSortDocument(sortDoc)
	if err != nil {
		return err
	}

	if len(sortFuncs) == 0 {
		// no keys to sort by
		return nil
	}

	sorter := &docsSorter{docs: docs, sorts: sortFuncs}
	sort.Sort(sorter)

	return nil
}

// ValidateSortDocument validates sort documents, and return
// proper error if it's invalid.
func ValidateSortDocument(sortDoc *types.Document) (*types.Document, error) {
	return convertSortDocument(sortDoc)
}

// unwrapSortDocument takes sort document, validates it, and returns
// sortFuncs if validation passed.
func convertSortDocument(sortDoc *types.Document) (*types.Document, error) {
	if sortDoc.Len() == 0 {
		return nil, nil
	}

	if sortDoc.Len() > 32 {
		return nil, lazyerrors.Errorf("maximum sort keys exceeded: %v", sortDoc.Len())
	}

	res := types.MakeDocument(sortDoc.Len())

	for _, sortKey := range sortDoc.Keys() {
		fields := strings.Split(sortKey, ".")
		for _, field := range fields {
			if strings.HasPrefix(field, "$") {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFieldPathInvalidName,
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
					"sort",
				)
			}
		}

		sortField := must.NotFail(sortDoc.Get(sortKey))

		sortValue, err := getSortValue(sortKey, sortField)
		if err != nil {
			return nil, err
		}

		sortPath, err := types.NewPathFromString(sortKey)
		if err != nil {
			return nil, err
		}

		if err = res.SetByPath(sortPath, sortValue); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return res, nil
}

// unwrapSortDocument takes sort document, validates it, and returns
// sortFuncs if validation passed.
func unwrapSortDocument(sortDoc *types.Document) ([]sortFunc, error) {
	if sortDoc.Len() == 0 {
		return nil, nil
	}

	if sortDoc.Len() > 32 {
		return nil, lazyerrors.Errorf("maximum sort keys exceeded: %v", sortDoc.Len())
	}

	sortFuncs := make([]sortFunc, sortDoc.Len())

	for i, sortKey := range sortDoc.Keys() {
		fields := strings.Split(sortKey, ".")
		for _, field := range fields {
			if strings.HasPrefix(field, "$") {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFieldPathInvalidName,
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
					"sort",
				)
			}
		}

		sortField := must.NotFail(sortDoc.Get(sortKey))

		sortType, err := GetSortType(sortKey, sortField)
		if err != nil {
			return nil, err
		}

		sortPath, err := types.NewPathFromString(sortKey)
		if err != nil {
			return nil, err
		}

		sortFuncs[i] = lessFunc(sortPath, sortType)
	}

	return sortFuncs, nil
}

// lessFunc takes sort key and type and returns sort.Interface's Less function which
// compares selected key of 2 documents.
func lessFunc(sortPath types.Path, sortType types.SortType) func(a, b *types.Document) bool {
	return func(a, b *types.Document) bool {
		aField, err := a.GetByPath(sortPath)
		if err != nil {
			// sort order treats null and non-existent field equivalent,
			// hence use null for sorting.
			aField = types.Null
		}

		bField, err := b.GetByPath(sortPath)
		if err != nil {
			// same logic as above
			bField = types.Null
		}

		result := types.CompareOrderForSort(aField, bField, sortType)

		return result == types.Less
	}
}

type sortFunc func(a, b *types.Document) bool

type docsSorter struct {
	docs  []*types.Document
	sorts []sortFunc
}

func (ds *docsSorter) Len() int {
	return len(ds.docs)
}

func (ds *docsSorter) Swap(i, j int) {
	ds.docs[i], ds.docs[j] = ds.docs[j], ds.docs[i]
}

func (ds *docsSorter) Less(i, j int) bool {
	p, q := ds.docs[i], ds.docs[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ds.sorts)-1; k++ {
		sortFunc := ds.sorts[k]

		switch {
		case sortFunc(p, q):
			// p < q, so we have a decision.
			return true
		case sortFunc(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ds.sorts[k](p, q)
}

// GetSortType determines SortType from input sort value.
func GetSortType(key string, value any) (types.SortType, error) {
	sortValue, err := getSortValue(key, value)
	if err != nil {
		return 0, err
	}

	switch sortValue {
	case 1:
		return types.Ascending, nil
	case -1:
		return types.Descending, nil
	default:
		panic("unreachable")
	}
}

// getSortValue validates if the value from sort document is the proper sort field,
// and returns it.
func getSortValue(key string, value any) (int64, error) {
	sortValue, err := commonparams.GetWholeNumberParam(value)
	if err != nil {
		switch {
		case errors.Is(err, commonparams.ErrUnexpectedType):
			if _, ok := value.(types.NullType); ok {
				value = "null"
			}

			return 0, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrSortBadValue,
				fmt.Sprintf(`Illegal key in $sort specification: %v: %v`, key, value),
				"$sort",
			)
		case errors.Is(err, commonparams.ErrNotWholeNumber):
			return 0, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrSortBadOrder,
				"$sort key ordering must be 1 (for ascending) or -1 (for descending)",
				"$sort",
			)
		default:
			return 0, err
		}
	}

	if sortValue != -1 && sortValue != 1 {
		return 0, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrSortBadOrder,
			"$sort key ordering must be 1 (for ascending) or -1 (for descending)",
			"$sort",
		)
	}

	return sortValue, nil
}
