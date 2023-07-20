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

package commonpath

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestGetPathValue(t *testing.T) {
	t.Parallel()

	emptyDoc := new(types.Document)
	docDoc := must.NotFail(types.NewDocument("foo", must.NotFail(types.NewDocument("bar", 1))))
	docArrayOne := must.NotFail(types.NewDocument("foo", must.NotFail(types.NewArray(
		must.NotFail(types.NewDocument("bar", 1)),
	))))
	docArrayTwo := must.NotFail(types.NewDocument("foo", must.NotFail(types.NewArray(
		must.NotFail(types.NewDocument("bar", 1)),
		must.NotFail(types.NewDocument("bar", 2)),
	))))

	for name, tc := range map[string]struct {
		res  []any
		doc  *types.Document
		path types.Path
		opts GetPathOpts
	}{
		"EmptyDocument": {
			doc:  emptyDoc,
			path: types.NewStaticPath("foo", "bar"),
			res:  []any{},
		},
		"Document": {
			doc:  docDoc,
			path: types.NewStaticPath("foo", "bar"),
			res:  []any{1},
		},
		"ArrayIndex": {
			doc:  docArrayOne,
			path: types.NewStaticPath("foo", "0", "bar"),
			res:  []any{1},
		},
		"ArrayDocument": {
			doc:  docArrayOne,
			path: types.NewStaticPath("foo", "bar"),
			res:  []any{1},
		},
		"ArrayDocuments": {
			doc:  docArrayTwo,
			path: types.NewStaticPath("foo", "bar"),
			res:  []any{1, 2},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := GetPathValue(tc.doc, tc.path, tc.opts)
			require.Equal(t, tc.res, res)
		})
	}
}
