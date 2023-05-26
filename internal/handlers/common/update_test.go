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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
)

func TestPathValidator(t *testing.T) {
	t.Parallel()

	for i, tc := range []struct {
		paths []string
		ok    bool // defaults to false
		errOn int  // index of path we expect the validation failure on
	}{
		{paths: []string{"v.foo.bar", "v.bar.foo", "v.foo"}, errOn: 2},
		{paths: []string{"v.foo", "v.foo.bar"}, errOn: 1},
		{paths: []string{"v.foo", "v.bar"}, ok: true},
		{paths: []string{"v.foo", "v"}, errOn: 1},
		{paths: []string{"v", "v"}, errOn: 1},
		{paths: []string{"v", "foo"}, ok: true},
	} {
		i, tc := i, tc
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tree := newPathValidator()

			for i, p := range tc.paths {
				path, err := types.NewPathFromString(p)
				require.NoError(t, err)

				ok := tree.validate(path.Slice())
				if tc.ok {
					require.True(t, ok)
					continue
				}

				if tc.errOn == i {
					require.False(t, ok, "Error expected at index %d", tc.errOn)
				} else {
					require.False(t, ok, "Error expected at index %d, but occurred at %d", tc.errOn, i)
				}
			}
		})
	}
}
