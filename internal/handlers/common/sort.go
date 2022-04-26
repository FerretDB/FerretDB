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

//go:generate ../../../bin/stringer -linecomment -type sortType

// sortType represents sort type for $sort aggregation.
type sortType int8

const (
	ascending  sortType = 1  // asc
	descending sortType = -1 // desc
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

		sortType, err := getSortType(sortField)
		if err != nil {
			return NewErrorMsg(ErrSortBadValue, fmt.Sprintf("%v: %v: %#v", err, sortKey, sortField))
		}

		sortFuncs[i] = lessFunc(sortKey, sortType)
	}

	sorter := &docsSorter{docs: docs, sorts: sortFuncs}
	sorter.Sort(docs)

	return nil
}

// lessFunc takes sort key and type and returns sort.Interface's Less function which
// compares selected key of 2 documents.
func lessFunc(sortKey string, sortType sortType) func(a, b *types.Document) bool {
	return func(a, b *types.Document) bool {
		aField, err := a.Get(sortKey)
		if err != nil {
			return false
		}

		bField, err := b.Get(sortKey)
		if err != nil {
			return false
		}

		result := types.Compare(aField, bField)

		switch result {
		case types.Less:
			switch sortType {
			case ascending:
				return true
			case descending:
				return false
			}
		case types.Greater:
			switch sortType {
			case ascending:
				return false
			case descending:
				return true
			}
		case types.NotEqual, types.Equal:
			return false
		}

		return false
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

// getSortType determines sortType from input sort value.
func getSortType(value any) (sortType, error) {
	sortValue, err := GetWholeNumberParam(value)
	if err != nil {
		switch err {
		case errUnexpectedType:
			return 0, NewErrorMsg(ErrBadValue, `Illegal key in $sort specification`)
		case errNotWholeNumber:
			return 0, NewErrorMsg(ErrBadValue, "$sort must be a whole number")
		default:
			return 0, err
		}
	}

	switch sortValue {
	case 1:
		return ascending, nil
	case -1:
		return descending, nil
	default:
		return 0, NewErrorMsg(ErrBadValue, `Illegal key in $sort specification`)
	}
}
