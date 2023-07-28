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

func TestFindValues(t *testing.T) {
	t.Parallel()

	t.Run("Array", func(t *testing.T) {
		array := must.NotFail(types.NewDocument("foo", must.NotFail(types.NewArray(
			must.NotFail(types.NewDocument("bar", 0)),
			must.NotFail(types.NewDocument("bar", 1)),
		))))

		for name, tc := range map[string]struct {
			doc  *types.Document
			path types.Path
			opts *FindValuesOpts
			res  []any
		}{
			"PathNestedFindIndexAndDocuments": {
				doc:  array,
				path: types.NewStaticPath("foo", "1", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: true,
				},
				res: []any{1},
			},
			"PathNestedFindIndex": {
				doc:  array,
				path: types.NewStaticPath("foo", "1", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: false,
				},
				res: []any{1},
			},
			"PathNestedFindDocuments": {
				doc:  array,
				path: types.NewStaticPath("foo", "1", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: true,
				},
				res: []any{},
			},
			"PathNested": {
				doc:  array,
				path: types.NewStaticPath("foo", "1", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: false,
				},
				res: []any{},
			},
			"PathIndexDotNotationFindIndexAndDocuments": {
				doc:  array,
				path: types.NewStaticPath("foo", "1"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: true,
				},
				res: []any{must.NotFail(types.NewDocument("bar", 1))},
			},
			"PathIndexDotNotationFindIndex": {
				doc:  array,
				path: types.NewStaticPath("foo", "1"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: false,
				},
				res: []any{must.NotFail(types.NewDocument("bar", 1))},
			},
			"PathIndexDotNotationFindDocuments": {
				doc:  array,
				path: types.NewStaticPath("foo", "1"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: true,
				},
				res: []any{},
			},
			"PathIndexDotNotation": {
				doc:  array,
				path: types.NewStaticPath("foo", "1"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: false,
				},
				res: []any{},
			},
			"PathDotNotationFindIndexAndDocuments": {
				doc:  array,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: true,
				},
				res: []any{0, 1},
			},
			"PathDotNotationFindIndex": {
				doc:  array,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: false,
				},
				res: []any{},
			},
			"PathDotNotationFindDocuments": {
				doc:  array,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: true,
				},
				res: []any{0, 1},
			},
			"PathDotNotation": {
				doc:  array,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: false,
				},
				res: []any{},
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
	})

	t.Run("Document", func(t *testing.T) {
		doc := must.NotFail(types.NewDocument("foo", must.NotFail(types.NewDocument("bar", 0))))

		for name, tc := range map[string]struct {
			doc  *types.Document
			path types.Path
			opts *FindValuesOpts
			res  []any
		}{
			"Empty": {
				doc:  new(types.Document),
				path: types.NewStaticPath("foo", "bar"),
				res:  []any{},
			},
			"PathDotNotationFindIndexAndDocuments": {
				doc:  doc,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: true,
				},
				res: []any{0},
			},
			"PathDotNotationFindIndex": {
				doc:  doc,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: false,
				},
				res: []any{0},
			},
			"PathDotNotationFindDocuments": {
				doc:  doc,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: true,
				},
				res: []any{0},
			},
			"PathDotNotation": {
				doc:  doc,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: false,
				},
				res: []any{0},
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
	})
}
