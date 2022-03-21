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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// sortType represents sort type
type sortType int

const (
	ascending sortType = iota
	descending
	textScore
)

// SortDocuments sorts given documents in place according to the given sorting conditions.
func SortDocuments(docs []*types.Document, sort *types.Document) error {
	// TODO

	if sort.Len() > 32 {
		return lazyerrors.Errorf("maximum sort keys exceeded: %v", sort.Len())
	}

	for i := 0; i < len(docs)-1; i++ {
	}

	return nil
}
