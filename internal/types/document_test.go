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

		err := doc.Set("Foo", Null)
		assert.NoError(t, err)
		value, err := doc.Get("Foo")
		assert.NoError(t, err)
		assert.Equal(t, Null, value)

		err = doc.Set("bar", 42)
		assert.EqualError(t, err, `types.Document.Set: types.validateValue: unsupported type: int (42)`)

		err = doc.Set("bar", nil)
		assert.EqualError(t, err, `types.Document.Set: types.validateValue: unsupported type: <nil> (<nil>)`)

		err = doc.Set("$k", int32(42))
		assert.EqualError(t, err, `types.Document.Set: types.validateDocumentKey: `+
			`short keys that start with '$' are not supported: "$k"`)

		assert.Equal(t, "foo", doc.Command())
	})

	t.Run("NewDocument", func(t *testing.T) {
		t.Parallel()

		doc, err := NewDocument(42, 42)
		assert.Nil(t, doc)
		assert.EqualError(t, err, `types.NewDocument: invalid key type: int`)

		doc, err = NewDocument("foo", 42)
		assert.Nil(t, doc)
		assert.EqualError(t, err, `types.NewDocument: types.Document.add: types.validateValue: unsupported type: int (42)`)

		doc, err = NewDocument("$k", int32(42))
		assert.Nil(t, doc)
		assert.EqualError(t, err, `types.NewDocument: types.Document.add: types.validateDocumentKey: `+
			`short keys that start with '$' are not supported: "$k"`)
	})
}
