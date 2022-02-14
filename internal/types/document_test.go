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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// lastErr returns the last error in error chain.
func lastErr(err error) error {
	for {
		e := errors.Unwrap(err)
		if e == nil {
			return err
		}
		err = e
	}
}

func TestDocument(t *testing.T) {
	t.Parallel()

	t.Run("MethodsOnNil", func(t *testing.T) {
		t.Parallel()

		var doc *Document
		assert.Zero(t, doc.Len())
		assert.Nil(t, doc.Map())
		assert.Nil(t, doc.Keys())
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

		err := doc.Set("foo", Null)
		assert.NoError(t, err)
		value, err := doc.Get("foo")
		assert.NoError(t, err)
		assert.Equal(t, Null, value)

		err = doc.Set("bar", 42)
		assert.EqualError(t, err, `types.Document.validate: types.validateValue: unsupported type: int (42)`)

		err = doc.Set("bar", nil)
		assert.EqualError(t, err, `types.Document.validate: types.validateValue: unsupported type: <nil> (<nil>)`)
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
			err  string // unwrapped
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
			err: `types.Document.validate: key not found: "0"`,
		}, {
			name: "duplicate keys",
			doc: Document{
				keys: []string{"0", "0"},
				m:    map[string]any{"0": "foo"},
			},
			err: `types.Document.validate: keys and values count mismatch: 1 != 2`,
		}, {
			name: "duplicate and different keys",
			doc: Document{
				keys: []string{"0", "0"},
				m:    map[string]any{"0": "foo", "1": "bar"},
			},
			err: `types.Document.validate: duplicate key: "0"`,
		}, {
			name: "fjson keys",
			doc: Document{
				keys: []string{"$k"},
				m:    map[string]any{"$k": "foo"},
			},
			err: `types.validateDocumentKey: short keys that start with '$' are not supported: "$k"`,
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
				if tc.err == "" {
					assert.NoError(t, err)
				} else {
					assert.Equal(t, tc.err, lastErr(err).Error())
				}
			})
		}
	})
}
