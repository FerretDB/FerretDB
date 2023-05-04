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

package iterator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSliceValues(t *testing.T) {
	expected := []int{1, 2, 3}
	actual, err := ConsumeValues(ForSlice(expected))
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	actual, err = ConsumeValues(Values(ForSlice(expected)))
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestConsumeValuesN(t *testing.T) {
	s := []int{1, 2, 3}
	iter := ForSlice(s)

	actual, err := ConsumeValuesN(iter, 2)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, actual)

	actual, err = ConsumeValuesN(iter, 2)
	require.NoError(t, err)
	assert.Equal(t, []int{3}, actual)

	actual, err = ConsumeValuesN(iter, 2)
	require.NoError(t, err)
	assert.Nil(t, actual)

	iter.Close()

	actual, err = ConsumeValuesN(iter, 2)
	require.NoError(t, err)
	assert.Nil(t, actual)
}

func TestGetKeyValueAtN(t *testing.T) {
	values := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	iter := ForSlice(values)

	k, v, err := GetKeyValueAtN(iter, 2)
	require.NoError(t, err)
	assert.Equal(t, values[1], k)
	assert.Equal(t, values[1], v)

	k, v, err = GetKeyValueAtN(iter, 2)
	require.NoError(t, err)
	assert.Equal(t, values[3], k)
	assert.Equal(t, values[3], v)

	k, v, err = GetKeyValueAtN(iter, 2)
	require.NoError(t, err)
	assert.Equal(t, values[5], k)
	assert.Equal(t, values[5], v)

	k, v, err = GetKeyValueAtN(iter, 2)
	require.NoError(t, err)
	assert.Equal(t, values[7], k)
	assert.Equal(t, values[7], v)

	k, v, err = GetKeyValueAtN(iter, 2)
	require.NoError(t, err)
	assert.Equal(t, values[9], k)
	assert.Equal(t, values[9], v)

	k, v, err = GetKeyValueAtN(iter, 2)
	require.Error(t, err)
	assert.Equal(t, 0, k)
	assert.Equal(t, 0, v)

	iter.Close()
}

func TestForFunc(t *testing.T) {
	var i int
	f := func() (struct{}, int, error) {
		var k struct{}

		i++
		if i > 3 {
			return k, 0, ErrIteratorDone
		}

		return k, i, nil
	}

	iter1 := ForFunc(f)
	iter2 := ForFunc(f)

	defer iter1.Close()
	defer iter2.Close()

	_, v, err := iter1.Next()
	require.NoError(t, err)
	assert.Equal(t, 1, v)

	_, v, err = iter2.Next()
	require.NoError(t, err)
	assert.Equal(t, 2, v)

	_, v, err = iter1.Next()
	require.NoError(t, err)
	assert.Equal(t, 3, v)

	_, v, err = iter2.Next()
	require.ErrorIs(t, err, ErrIteratorDone)
	assert.Equal(t, 0, v)
}
