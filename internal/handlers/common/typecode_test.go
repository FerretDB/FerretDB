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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestHasSameTypeElements(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:paralleltest // false positive
		array *types.Array
		same  bool
	}{
		"ArrayInt": {
			array: must.NotFail(types.NewArray(must.NotFail(types.NewArray(int32(1), int32(2))), int32(3))),
			same:  false,
		},
		"IntSame": {
			array: must.NotFail(types.NewArray(int32(1), int32(2))),
			same:  true,
		},
		"IntDouble": {
			array: must.NotFail(types.NewArray(int32(1), 42.3)),
			same:  false,
		},
		"IntDoubleWhole": {
			array: must.NotFail(types.NewArray(int32(1), 42.0)),
			same:  true,
		},
		"IntLong": {
			array: must.NotFail(types.NewArray(int32(1), int64(42))),
			same:  true,
		},
	} {
		tc, name := tc, name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := hasSameTypeElements(tc.array)
			assert.Equal(t, tc.same, result)
		})
	}
}
