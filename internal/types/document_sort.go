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

import "sort"

// SortByKeys sorts the document by its keys. It modifies the document in place.
func (d *Document) SortByKeys() {
	sort.Sort(d)
}

// Swap is part of sort.Interface.
func (d *Document) Swap(i, j int) {
	d.fields[i], d.fields[j] = d.fields[j], d.fields[i]
}

// Less is part of sort.Interface. It compares two fields by their keys.
func (d *Document) Less(i, j int) bool {
	return d.fields[i].key < d.fields[j].key
}
