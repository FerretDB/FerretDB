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

package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/backends"
)

func TestPrepareOrderByClause(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		sort   *backends.SortField
		capped bool

		orderBy string
	}{
		"Ascending": {
			sort:    &backends.SortField{Key: "field", Descending: false},
			orderBy: "",
		},
		"Descending": {
			sort:    &backends.SortField{Key: "field", Descending: true},
			orderBy: "",
		},
		"SortNil": {
			orderBy: "",
		},
		"Capped": {
			capped:  true,
			orderBy: ` ORDER BY _ferretdb_record_id`,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			orderBy := prepareOrderByClause(tc.sort, tc.capped)
			assert.Equal(t, tc.orderBy, orderBy)
		})
	}
}
