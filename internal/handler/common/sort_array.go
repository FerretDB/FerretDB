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
	"sort"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// SortArray sorts the values of given array.
func SortArray(arr *types.Array, sortType types.SortType) {
	sorter := &arraySorter{arr: arr, sortType: sortType}
	sort.Sort(sorter)
}

// arraySorter implements sort.Interface to sort values of arrays.
type arraySorter struct {
	arr      *types.Array
	sortType types.SortType
}

// Len implements sort.Interface.
func (as *arraySorter) Len() int {
	return as.arr.Len()
}

// Swap implements sort.Interface.
func (as *arraySorter) Swap(i, j int) {
	p, q := must.NotFail(as.arr.Get(i)), must.NotFail(as.arr.Get(j))

	must.NoError(as.arr.Set(i, q))
	must.NoError(as.arr.Set(j, p))
}

// Less implements sort.Interface.
func (as *arraySorter) Less(i, j int) bool {
	p, q := must.NotFail(as.arr.Get(i)), must.NotFail(as.arr.Get(j))

	if p == nil {
		// sort order treats null and non-existent field equivalent,
		// hence use null for sorting.
		p = types.Null
	}

	result := types.CompareOrderForSort(p, q, as.sortType)

	return result == types.Less
}
