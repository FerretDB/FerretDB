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
)

func TestDeepCopy(t *testing.T) {
	t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		t.Parallel()

		b1 := Binary{
			Subtype: 0x01,
			B:       []byte{0x01, 0x02, 0x03},
		}
		b2 := deepCopy(b1)

		assert.Equal(t, b1, b2)
		assert.NotSame(t, b1, b2)

		b1.B[0] = 0
		assert.NotEqual(t, b1, b2)
	})

	t.Run("ObjectID", func(t *testing.T) {
		t.Parallel()

		o1 := NewObjectID()
		o2 := deepCopy(o1)

		assert.Equal(t, o1, o2)
		assert.NotSame(t, o1, o2)

		o1[0] = 0
		assert.NotEqual(t, o1, o2)
	})
}
