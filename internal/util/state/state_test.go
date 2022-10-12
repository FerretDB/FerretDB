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

package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "state.json")
	p1, err := NewProvider(filename)
	require.NoError(t, err)

	s1, err := p1.Get()
	require.NoError(t, err)
	assert.NotEmpty(t, s1.UUID)

	s2, err := p1.Get()
	require.NoError(t, err)
	assert.Equal(t, s1, s2)
	assert.NotSame(t, s1, s2)

	s3, err := p1.Get()
	require.NoError(t, err)
	assert.Equal(t, s1, s3)
	assert.NotSame(t, s1, s3)

	p2, err := NewProvider(filename)
	require.NoError(t, err)

	s4, err := p2.Get()
	require.NoError(t, err)
	assert.Equal(t, s1, s4)
	assert.NotSame(t, s1, s4)

	require.NoError(t, os.Remove(filename))

	p3, err := NewProvider(filename)
	require.NoError(t, err)

	s5, err := p3.Get()
	require.NoError(t, err)
	assert.NotEqual(t, s1, s5)
}
