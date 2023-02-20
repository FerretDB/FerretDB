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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestDocument(t *testing.T) {
	t.Parallel()

	t.Run("MethodsOnNil", func(t *testing.T) {
		t.Parallel()

		var doc *Document
		assert.Zero(t, doc.Len())
		assert.Nil(t, doc.Map())
		assert.Nil(t, doc.Keys())
		assert.Equal(t, "", doc.Command())
	})

	t.Run("ZeroValues", func(t *testing.T) {
		t.Parallel()

		// to avoid {} != nil in tests
		assert.Nil(t, must.NotFail(NewDocument()).fields)

		var doc Document
		assert.Equal(t, 0, doc.Len())
		assert.Nil(t, doc.fields)
		assert.Equal(t, "", doc.Command())

		doc.Set("foo", Null)
		value, err := doc.Get("foo")
		assert.NoError(t, err)
		assert.Equal(t, Null, value)
		assert.Equal(t, "foo", doc.Command())
	})

	t.Run("NewDocument", func(t *testing.T) {
		t.Parallel()

		doc, err := NewDocument(42, 42)
		assert.Nil(t, doc)
		assert.EqualError(t, err, `types.NewDocument: invalid key type: int`)
	})

	t.Run("DeepCopy", func(t *testing.T) {
		t.Parallel()

		a := must.NotFail(NewDocument("foo", int32(42)))
		b := a.DeepCopy()
		assert.Equal(t, a, b)
		assert.NotSame(t, a, b)

		a.Set("foo", "bar")
		assert.NotEqual(t, a, b)
		assert.Equal(t, int32(42), must.NotFail(b.Get("foo")))
	})

	t.Run("moveIDToTheFirstIndex", func(t *testing.T) {
		t.Parallel()

		doc := must.NotFail(NewDocument(
			"_id", int32(42),
			"foo", "bar",
			"baz", "qux",
		))
		assert.Equal(t, []string{"_id", "foo", "baz"}, doc.Keys())

		doc = must.NotFail(NewDocument(
			"foo", "bar",
			"_id", int32(42),
			"baz", "qux",
		))
		doc.moveIDToTheFirstIndex()
		assert.Equal(t, []string{"_id", "foo", "baz"}, doc.Keys())

		doc = must.NotFail(NewDocument(
			"foo", "bar",
			"baz", "qux",
			"_id", int32(42),
		))
		doc.moveIDToTheFirstIndex()
		assert.Equal(t, []string{"_id", "foo", "baz"}, doc.Keys())

		doc = must.NotFail(NewDocument(
			"foo", "bar",
			"baz", "qux",
		))
		doc.moveIDToTheFirstIndex()
		assert.Equal(t, []string{"foo", "baz"}, doc.Keys())
	})

	t.Run("SetByPath", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			document *Document
			expected *Document
			key      string
			value    any
			err      error
		}{
			{
				name:     "path exists",
				document: must.NotFail(NewDocument("foo", must.NotFail(NewDocument("bar", int32(42))))),
				key:      "foo.bar",
				value:    "baz",
				expected: must.NotFail(NewDocument("foo", must.NotFail(NewDocument("bar", "baz")))),
			},
			{
				name:     "key not exists",
				document: must.NotFail(NewDocument("foo", must.NotFail(NewDocument("bar", int32(42))))),
				key:      "foo.baz",
				value:    "bar",
				expected: must.NotFail(NewDocument("foo", must.NotFail(NewDocument("bar", int32(42), "baz", "bar")))),
			},
			{
				name:     "path not exist",
				document: must.NotFail(NewDocument("foo", must.NotFail(NewDocument("bar", int32(42))))),
				key:      "foo.baz.bar",
				value:    "bar",
				expected: must.NotFail(NewDocument(
					"foo", must.NotFail(NewDocument(
						"bar", int32(42),
						"baz", must.NotFail(NewDocument("bar", "bar")),
					)),
				)),
			},
			{
				name:     "extend empty array with document",
				document: must.NotFail(NewDocument("v", must.NotFail(NewArray()))),
				key:      "v.2.foo",
				value:    "bar",
				expected: must.NotFail(NewDocument(
					"v", must.NotFail(NewArray(Null, Null, must.NotFail(NewDocument("foo", "bar")))),
				)),
			},
			{
				name:     "extend non-empty array with document",
				document: must.NotFail(NewDocument("v", must.NotFail(NewArray("a")))),
				key:      "v.2.foo",
				value:    "bar",
				expected: must.NotFail(NewDocument(
					"v", must.NotFail(NewArray("a", Null, must.NotFail(NewDocument("foo", "bar")))),
				)),
			},
			{
				name:     "extend non-empty array with scalar",
				document: must.NotFail(NewDocument("v", must.NotFail(NewArray("a")))),
				key:      "v.2",
				value:    "bar",
				expected: must.NotFail(NewDocument(
					"v", must.NotFail(NewArray("a", Null, "bar")),
				)),
			},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				path, err := NewPathFromString(tc.key)
				require.NoError(t, err)

				err = tc.document.SetByPath(path, tc.value)

				if tc.err != nil {
					assert.Equal(t, tc.err, err)
					return
				}
				require.NoError(t, err)

				assert.Equal(t, tc.expected, tc.document)
			})
		}
	})
}
