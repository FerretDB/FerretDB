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

func TestArray(t *testing.T) {
	t.Parallel()

	t.Run("MethodsOnNil", func(t *testing.T) {
		t.Parallel()

		var a *Array
		assert.Zero(t, a.Len())
	})

	t.Run("ZeroValues", func(t *testing.T) {
		t.Parallel()

		// to avoid {} != nil in tests
		assert.Nil(t, must.NotFail(NewArray()).s)
		assert.Nil(t, MakeArray(0).s)

		var a Array
		assert.Equal(t, 0, a.Len())
		assert.Nil(t, a.s)

		err := a.Append(Null)
		assert.NoError(t, err)
		value, err := a.Get(0)
		assert.NoError(t, err)
		assert.Equal(t, Null, value)

		err = a.Append(42)
		assert.EqualError(t, err, `types.Array.Append: types.validateValue: unsupported type: int (42)`)

		err = a.Append(nil)
		assert.EqualError(t, err, `types.Array.Append: types.validateValue: unsupported type: <nil> (<nil>)`)
	})

	t.Run("NewArray", func(t *testing.T) {
		t.Parallel()

		a, err := NewArray(int32(42), 42)
		assert.Nil(t, a)
		assert.EqualError(t, err, `types.NewArray: index 1: types.validateValue: unsupported type: int (42)`)
	})

	t.Run("DeepCopy", func(t *testing.T) {
		t.Parallel()

		a := must.NotFail(NewArray(int32(42)))
		b := a.DeepCopy()
		assert.Equal(t, a, b)
		assert.NotSame(t, a, b)

		a.s[0] = "foo"
		assert.NotEqual(t, a, b)
		assert.Equal(t, int32(42), b.s[0])
	})
}
