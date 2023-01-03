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
	"fmt"
	"sort"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// SortDocuments sorts given documents in place according to the given sorting conditions.
func SortDocuments(docs []*types.Document, sort *types.Document) error {
	if sort.Len() == 0 {
		return nil
	}

	if sort.Len() > 32 {
		return lazyerrors.Errorf("maximum sort keys exceeded: %v", sort.Len())
	}

	sortFuncs := make([]sortFunc, len(sort.Keys()))
	for i, sortKey := range sort.Keys() {
		sortField := must.NotFail(sort.Get(sortKey))
		sortType, err := getSortType(sortKey, sortField)
		if err != nil {
			return err
		}

		sortFuncs[i] = lessFunc(sortKey, sortType)
	}

	sorter := &docsSorter{docs: docs, sorts: sortFuncs}
	sorter.Sort(docs)

	return nil
}

// SortArray sorts the values of given array to use it in the response.
func SortArray(arr *types.Array, sortType types.SortType) error {
	sortFuncs := make([]arrSortFunc, 1)
	sortFuncs[0] = arrLessFunc("", sortType)

	sorter := &arraySorter{arr: arr, sorts: sortFuncs}
	sorter.Sort(arr)

	return nil
}

// lessFunc takes sort key and type and returns sort.Interface's Less function which
// compares selected key of 2 documents.
func lessFunc(sortKey string, sortType types.SortType) func(a, b *types.Document) bool {
	return func(a, b *types.Document) bool {
		aField, err := a.Get(sortKey)
		if err != nil {
			// sort order treats null and non-existent field equivalent,
			// hence use null for sorting.
			aField = types.Null
		}

		bField, err := b.Get(sortKey)
		if err != nil {
			return false
		}

		result := types.CompareOrderForSort(aField, bField, sortType)

		return result == types.Less
	}
}

// arrLessFunc takes sort .
func arrLessFunc(sortKey string, sortType types.SortType) func(a, b any) bool {
	return func(a, b any) bool {
		if a == nil {
			// sort order treats null and non-existent field equivalent,
			// hence use null for sorting.
			a = types.Null
		}

		result := types.CompareOrderForSort(a, b, sortType)

		return result == types.Less
	}
}

type sortFunc func(a, b *types.Document) bool

type docsSorter struct {
	docs  []*types.Document
	sorts []sortFunc
}

func (ds *docsSorter) Sort(docs []*types.Document) {
	ds.docs = docs
	sort.Sort(ds)
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

type arrSortFunc func(a, b any) bool

type arraySorter struct {
	arr   *types.Array
	sorts []arrSortFunc
}

func (as *arraySorter) Sort(arr *types.Array) {
	as.arr = arr
	sort.Sort(as)
}

func (as *arraySorter) Len() int {
	return as.arr.Len()
}

func (as *arraySorter) Swap(i, j int) {
	p, q := must.NotFail(as.arr.Get(i)), must.NotFail(as.arr.Get(j))

	must.NoError(as.arr.Set(i, q))
	must.NoError(as.arr.Set(j, p))
}

func (as *arraySorter) Less(i, j int) bool {
	p, q := must.NotFail(as.arr.Get(i)), must.NotFail(as.arr.Get(j))
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(as.sorts)-1; k++ {
		arrSortFunc := as.sorts[k]
		switch {
		case arrSortFunc(p, q):
			// p < q, so we have a decision.
			return true
		case arrSortFunc(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return as.sorts[k](p, q)
}

// getSortType determines SortType from input sort value.
func getSortType(key string, value any) (types.SortType, error) {
	sortValue, err := GetWholeNumberParam(value)
	if err != nil {
		switch err {
		case errUnexpectedType:
			if _, ok := value.(types.NullType); ok {
				value = "null"
			}

			return 0, NewCommandErrorMsgWithArgument(
				ErrSortBadValue,
				fmt.Sprintf(`Illegal key in $sort specification: %v: %v`, key, value),
				"$sort",
			)
		case errNotWholeNumber:
			return 0, NewCommandErrorMsgWithArgument(ErrBadValue, "$sort must be a whole number", "$sort")
		default:
			return 0, err
		}
	}

	switch sortValue {
	case 1:
		return types.Ascending, nil
	case -1:
		return types.Descending, nil
	default:
		return 0, NewCommandErrorMsgWithArgument(
			ErrSortBadOrder,
			"$sort key ordering must be 1 (for ascending) or -1 (for descending)",
			"$sort",
		)
	}
}
