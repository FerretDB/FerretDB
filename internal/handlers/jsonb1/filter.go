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

package jsonb1

import (
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// filterDocument returns true if given document matches given filter condition.
//
// Passed arguments must not be modified.
func filterDocument(doc, filter *types.Document) (bool, error) {
	filterMap := filter.Map()
	if len(filterMap) == 0 {
		return true, nil
	}

	// top-level filters are ANDed together
	for _, filterKey := range filter.Keys() {
		filterValue := filterMap[filterKey]
		res, err := filterDocumentFoo(doc, filterKey, filterValue)
		if err != nil {
			return false, err
		}
		if !res {
			return false, nil
		}
	}

	return true, nil
}

func filterDocumentFoo(doc *types.Document, filterKey string, filterValue any) (bool, error) {
	return false, lazyerrors.Errorf("filterDocumentFoo: unhandled key %q, value %v", filterKey, filterValue)
}
