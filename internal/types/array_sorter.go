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

package types

import "github.com/FerretDB/FerretDB/internal/util/must"

// arraySort represents sort.Interface for an Array.
type arraySort struct {
	arr *Array
}

// newArraySort returns a new arraySort.
func newArraySort(arr *Array) *arraySort {
	return &arraySort{arr: arr}
}

// Len implements sort.Interface.
func (as *arraySort) Len() int {
	return as.arr.Len()
}

// Less implements sort.Interface.
func (as *arraySort) Less(i, j int) bool {
	return Compare(must.NotFail(as.arr.Get(i)), must.NotFail(as.arr.Get(j))) == Less
}

// Swap implements sort.Interface.
func (as *arraySort) Swap(i, j int) {
	valI := must.NotFail(as.arr.Get(i))
	valJ := must.NotFail(as.arr.Get(j))

	must.NoError(as.arr.Set(i, valJ))
	must.NoError(as.arr.Set(j, valI))
}

func (as *arraySort) Arr() *Array {
	return as.arr
}
