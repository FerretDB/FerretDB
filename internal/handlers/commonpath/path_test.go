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

	empty := new(types.Document)
	doc := must.NotFail(types.NewDocument("foo", must.NotFail(types.NewDocument("bar", 1))))
	arrayDocOne := must.NotFail(types.NewDocument("foo", must.NotFail(types.NewArray(
		must.NotFail(types.NewDocument("bar", 1)),
	))))
	arrayDocTwo := must.NotFail(types.NewDocument("foo", must.NotFail(types.NewArray(
		must.NotFail(types.NewDocument("bar", 1)),
		must.NotFail(types.NewDocument("bar", 2)),
	))))
	arrayScalarThree := must.NotFail(types.NewDocument("foo", must.NotFail(types.NewArray(0, 1, 2))))

	for name, tc := range map[string]struct {
		doc  *types.Document
		path types.Path
		opts *FindValuesOpts
		res  []any
	}{
		"Empty": {
			doc:  empty,
			path: types.NewStaticPath("foo", "bar"),
			res:  []any{},
		},
		"Document": {
			doc:  doc,
			path: types.NewStaticPath("foo"),
			res:  []any{must.NotFail(types.NewDocument("bar", 1))},
		},
		"DocumentDotNotation": {
			doc:  doc,
			path: types.NewStaticPath("foo", "bar"),
			res:  []any{1},
		},
		"ArrayIndexDoc": {
			doc:  arrayDocOne,
			path: types.NewStaticPath("foo", "0", "bar"),
			opts: &FindValuesOpts{FindArrayIndex: true},
			res:  []any{1},
		},
		"ArrayIndexDocFindArrayIndexFalse": {
			doc:  arrayDocOne,
			path: types.NewStaticPath("foo", "0", "bar"),
			opts: &FindValuesOpts{FindArrayIndex: false},
			res:  []any{},
		},
		"ArrayIndexScalar": {
			doc:  arrayScalarThree,
			path: types.NewStaticPath("foo", "1"),
			opts: &FindValuesOpts{FindArrayIndex: true},
			res:  []any{1},
		},
		"ArrayDocument": {
			doc:  arrayDocOne,
			path: types.NewStaticPath("foo", "bar"),
			opts: &FindValuesOpts{SearchArray: true},
			res:  []any{1},
		},
		"ArrayDocumentSearchArrayFalse": {
			doc:  arrayDocOne,
			path: types.NewStaticPath("foo", "bar"),
			opts: &FindValuesOpts{SearchArray: false},
			res:  []any{},
		},
		"ArrayDocumentTwo": {
			doc:  arrayDocTwo,
			path: types.NewStaticPath("foo", "bar"),
			opts: &FindValuesOpts{SearchArray: true},
			res:  []any{1, 2},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res, err := FindValues(tc.doc, tc.path, tc.opts)
			require.NoError(t, err)
			require.Equal(t, tc.res, res)
		})
	}
}
