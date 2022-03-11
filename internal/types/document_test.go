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
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// TestDocument example run go test -run TestDocument/Equal github.com/FerretDB/FerretDB/internal/types
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

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name     string
			a        *Document
			b        *Document
			expected bool
		}{
			{
				name:     "normal",
				expected: true,
				a: &Document{
					keys: []string{"name", "code", "games", "joined", "lastLogin"},
					m: map[string]any{
						"name":  "John Doe",
						"code":  121081,
						"score": 42.13,
						"games": []any{
							map[string]any{"game": "abc", "score": 5},
							map[string]any{"game": "def", "score": 8},
							map[string]any{"game": "xyz", "score": 9},
						},
						"joined":    time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
						"lastLogin": time.Date(2022, 03, 16, 12, 13, 14, 0, time.UTC),
					},
				},
				b: &Document{
					keys: []string{"name", "code", "games", "joined", "lastLogin"},
					m: map[string]any{
						"name":  "John Doe",
						"code":  121081,
						"score": 42.13,
						"games": []any{
							map[string]any{"game": "abc", "score": 5},
							map[string]any{"game": "def", "score": 8},
							map[string]any{"game": "xyz", "score": 9},
						},
						"joined":    time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
						"lastLogin": time.Date(2022, 03, 16, 12, 13, 14, 0, time.UTC),
					},
				},
			},
			{
				name:     "lastLoginNsecDiffers",
				expected: false,
				a: &Document{
					keys: []string{"name", "code", "games", "joined", "lastLogin"},
					m: map[string]any{
						"name":  "John Doe",
						"code":  121081,
						"score": 42.13,
						"games": []any{
							map[string]any{"game": "abc", "score": 5},
							map[string]any{"game": "def", "score": 8},
							map[string]any{"game": "xyz", "score": 9},
						},
						"joined":    time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
						"lastLogin": time.Date(2022, 03, 16, 12, 13, 14, 0, time.UTC),
					},
				},
				b: &Document{
					keys: []string{"name", "code", "games", "joined", "lastLogin"},
					m: map[string]any{
						"name":  "John Doe",
						"code":  121081,
						"score": 42.13,
						"games": []any{
							map[string]any{"game": "abc", "score": 5},
							map[string]any{"game": "def", "score": 8},
							map[string]any{"game": "xyz", "score": 9},
						},
						"joined":    time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
						"lastLogin": time.Date(2022, 03, 16, 12, 13, 14, 1, time.UTC),
					},
				},
			},
			{
				name:     "GamesGame3dValueDiffers",
				expected: false,
				a: &Document{
					keys: []string{"name", "code", "games", "joined", "lastLogin"},
					m: map[string]any{
						"name":  "John Doe",
						"code":  121081,
						"score": 42.13,
						"games": []any{
							map[string]any{"game": "abc", "score": 5},
							map[string]any{"game": "def", "score": 8},
							map[string]any{"game": "XYZ", "score": 9},
						},
						"joined":    time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
						"lastLogin": time.Date(2022, 03, 16, 12, 13, 14, 0, time.UTC),
					},
				},
				b: &Document{
					keys: []string{"name", "code", "games", "joined", "lastLogin"},
					m: map[string]any{
						"name":  "John Doe",
						"code":  121081,
						"score": 42.13,
						"games": []any{
							map[string]any{"game": "abc", "score": 5},
							map[string]any{"game": "def", "score": 8},
							map[string]any{"game": "xyz", "score": 9},
						},
						"joined":    time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
						"lastLogin": time.Date(2022, 03, 16, 12, 13, 14, 0, time.UTC),
					},
				},
			},
			{
				name:     "CodeDiffers",
				expected: false,
				a: &Document{
					keys: []string{"name", "code", "games", "joined", "lastLogin"},
					m: map[string]any{
						"name":  "John Doe",
						"code":  121082,
						"score": 42.13,
						"games": []any{
							map[string]any{"game": "abc", "score": 5},
							map[string]any{"game": "def", "score": 8},
							map[string]any{"game": "xyz", "score": 9},
						},
						"joined":    time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
						"lastLogin": time.Date(2022, 03, 16, 12, 13, 14, 0, time.UTC),
					},
				},
				b: &Document{
					keys: []string{"name", "code", "games", "joined", "lastLogin"},
					m: map[string]any{
						"name":  "John Doe",
						"code":  121081,
						"score": 42.13,
						"games": []any{
							map[string]any{"game": "abc", "score": 5},
							map[string]any{"game": "def", "score": 8},
							map[string]any{"game": "xyz", "score": 9},
						},
						"joined":    time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
						"lastLogin": time.Date(2022, 03, 16, 12, 13, 14, 0, time.UTC),
					},
				},
			},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				actual := tc.a.Equal(tc.b)
				assert.Equal(t, tc.expected, actual)
			})
		}

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

		err := doc.Set("Foo", Null)
		assert.NoError(t, err)
		value, err := doc.Get("Foo")
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
		}} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				err := tc.doc.validate()
				assert.Equal(t, tc.err, err)
			})
		}
	})
}
