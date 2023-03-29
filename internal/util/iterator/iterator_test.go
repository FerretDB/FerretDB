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
