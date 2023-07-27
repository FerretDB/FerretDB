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
			"PathNested1": {
				doc:  array,
				path: types.NewStaticPath("foo", "1", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: true,
				},
				res: []any{1},
			},
			"PathNested2": {
				doc:  array,
				path: types.NewStaticPath("foo", "1", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: false,
				},
				res: []any{1},
			},
			"PathNested3": {
				doc:  array,
				path: types.NewStaticPath("foo", "1", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: true,
				},
				res: []any{},
			},
			"PathNested4": {
				doc:  array,
				path: types.NewStaticPath("foo", "1", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: false,
				},
				res: []any{},
			},
			"PathIndexDotNotation1": {
				doc:  array,
				path: types.NewStaticPath("foo", "1"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: true,
				},
				res: []any{must.NotFail(types.NewDocument("bar", 1))},
			},
			"PathIndexDotNotation2": {
				doc:  array,
				path: types.NewStaticPath("foo", "1"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: false,
				},
				res: []any{must.NotFail(types.NewDocument("bar", 1))},
			},
			"PathIndexDotNotation3": {
				doc:  array,
				path: types.NewStaticPath("foo", "1"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: true,
				},
				res: []any{},
			},
			"PathIndexDotNotation4": {
				doc:  array,
				path: types.NewStaticPath("foo", "1"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: false,
				},
				res: []any{},
			},
			"PathDotNotation1": {
				doc:  array,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: true,
				},
				res: []any{0, 1},
			},
			"PathDotNotation2": {
				doc:  array,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: false,
				},
				res: []any{},
			},
			"PathDotNotation3": {
				doc:  array,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: true,
				},
				res: []any{0, 1},
			},
			"PathDotNotation4": {
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
			"PathDotNotation1": {
				doc:  doc,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: true,
				},
				res: []any{0},
			},
			"PathDotNotation2": {
				doc:  doc,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     true,
					FindArrayDocuments: false,
				},
				res: []any{0},
			},
			"PathDotNotation3": {
				doc:  doc,
				path: types.NewStaticPath("foo", "bar"),
				opts: &FindValuesOpts{
					FindArrayIndex:     false,
					FindArrayDocuments: true,
				},
				res: []any{0},
			},
			"PathDotNotation4": {
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
