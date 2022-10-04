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
	"fmt"
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
		assert.Nil(t, must.NotFail(NewDocument()).m)
		assert.Nil(t, must.NotFail(NewDocument()).keys)

		var doc Document
		assert.Equal(t, 0, doc.Len())
		assert.Nil(t, doc.m)
		assert.Nil(t, doc.keys)
		assert.Equal(t, "", doc.Command())

		err := doc.Set("foo", Null)
		assert.NoError(t, err)
		value, err := doc.Get("foo")
		assert.NoError(t, err)
		assert.Equal(t, Null, value)

		err = doc.Set("bar", 42)
		assert.EqualError(t, err, `types.Document.validate: types.validateValue: unsupported type: int (42)`)

		err = doc.Set("bar", nil)
		assert.EqualError(t, err, `types.Document.validate: types.validateValue: unsupported type: <nil> (<nil>)`)

		assert.Equal(t, "foo", doc.Command())
	})

	t.Run("NewDocument", func(t *testing.T) {
		t.Parallel()

		doc, err := NewDocument(42, 42)
		assert.Nil(t, doc)
		assert.EqualError(t, err, `types.NewDocument: invalid key type: int`)

		doc, err = NewDocument("foo", 42)
		assert.Nil(t, doc)
		assert.EqualError(t, err, `types.NewDocument: types.Document.validate: types.validateValue: unsupported type: int (42)`)
	})

	t.Run("DeepCopy", func(t *testing.T) {
		t.Parallel()

		a := must.NotFail(NewDocument("foo", int32(42)))
		b := a.DeepCopy()
		assert.Equal(t, a, b)
		assert.NotSame(t, a, b)

		a.m["foo"] = "bar"
		assert.NotEqual(t, a, b)
		assert.Equal(t, int32(42), b.m["foo"])
	})

	t.Run("SetID", func(t *testing.T) {
		t.Parallel()

		doc := must.NotFail(NewDocument(
			"_id", int32(42),
			"foo", "bar",
		))
		assert.Equal(t, []string{"_id", "foo"}, doc.keys)

		doc = must.NotFail(NewDocument(
			"foo", "bar",
			"_id", int32(42),
		))
		assert.Equal(t, []string{"_id", "foo"}, doc.keys)

		doc.Set("_id", "bar")
		assert.Equal(t, []string{"_id", "foo"}, doc.keys)
	})

	t.Run("Validate", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name string
			doc  Document
			err  error
		}{{
			name: "normal",
			doc: Document{
				keys: []string{"0"},
				m:    map[string]any{"0": "foo"},
			},
		}, {
			name: "empty",
			doc:  Document{},
		}, {
			name: "different keys",
			doc: Document{
				keys: []string{"0"},
				m:    map[string]any{"1": "foo"},
			},
			err: fmt.Errorf(`types.Document.validate: key not found: "0"`),
		}, {
			name: "duplicate keys",
			doc: Document{
				keys: []string{"0", "0"},
				m:    map[string]any{"0": "foo"},
			},
			err: fmt.Errorf("types.Document.validate: keys and values count mismatch: 1 != 2"),
		}, {
			name: "duplicate and different keys",
			doc: Document{
				keys: []string{"0", "0"},
				m:    map[string]any{"0": "foo", "1": "bar"},
			},
			err: fmt.Errorf(`types.Document.validate: duplicate key: "0"`),
		}, {
			name: "fjson keys",
			doc: Document{
				keys: []string{"$k"},
				m:    map[string]any{"$k": "foo"},
			},
			err: fmt.Errorf(`types.Document.validate: invalid key: "$k"`),
		}, {
			name: "dollar keys",
			doc: Document{
				keys: []string{"$db"},
				m:    map[string]any{"$db": "foo"},
			},
		}, {
			name: "empty key",
			doc: Document{
				keys: []string{""},
				m:    map[string]any{"": ""},
			},
		}} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				err := tc.doc.validate()
				assert.Equal(t, tc.err, err)
			})
		}
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
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				err := tc.document.SetByPath(NewPathFromString(tc.key), tc.value)

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
