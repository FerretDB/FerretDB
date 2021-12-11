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

package testutil

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
)

func SetByPath(tb testing.TB, str any, value any, path ...string) {
	tb.Helper()

	l := len(path)
	require.NotZero(tb, l, "path is empty")

	for i, p := range path {
		last := i == l-1
		switch s := str.(type) {
		case types.Array:
			index, err := strconv.Atoi(p)
			require.NoError(tb, err)

			if !last {
				str, err = s.Get(index)
			} else {
				err = s.Set(index, value)
			}
			require.NoError(tb, err)

		case types.Document:
			var err error
			if !last {
				str, err = s.Get(p)
			} else {
				err = s.Set(p, value)
			}
			require.NoError(tb, err)

		default:
			tb.Fatalf("can't access %T by path %q", str, p)
		}
	}
}
