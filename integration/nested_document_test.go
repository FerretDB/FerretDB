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

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCreateNestedDocument(t *testing.T) {
	t.Parallel()

	t.Run("0", func(t *testing.T) {
		expected := bson.D{{"v", nil}}
		actual := CreateNestedDocument(0)
		assert.Equal(t, expected, actual)
	})

	t.Run("1", func(t *testing.T) {
		expected := bson.D{{"v", bson.A{nil}}}
		actual := CreateNestedDocument(1)
		assert.Equal(t, expected, actual)
	})

	t.Run("2", func(t *testing.T) {
		expected := bson.D{{"v", bson.A{bson.D{{"v", nil}}}}}
		actual := CreateNestedDocument(2)
		assert.Equal(t, expected, actual)
	})
}
