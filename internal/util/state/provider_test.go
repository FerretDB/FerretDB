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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	t.Run("Get", func(t *testing.T) {
		t.Parallel()

		filename := filepath.Join(t.TempDir(), "state.json")
		p1, err := NewProvider(filename)
		require.NoError(t, err)

		s1 := p1.Get()
		assert.NotZero(t, s1.UUID)
		assert.NotZero(t, s1.Start)

		// cached state
		s2 := p1.Get()
		assert.Equal(t, s1, s2)
		assert.NotSame(t, s1, s2)

		p2, err := NewProvider(filename)
		require.NoError(t, err)

		// reread state file with a different provider should be the same except start time
		s3 := p2.Get()
		assert.NotEqual(t, s1.Start, s3.Start)
		s3.Start = s1.Start
		assert.Equal(t, s1, s3)
		assert.NotSame(t, s1, s3)

		require.NoError(t, os.Remove(filename))

		p3, err := NewProvider(filename)
		require.NoError(t, err)

		// after removing state file UUID should be different
		s4 := p3.Get()
		assert.NotZero(t, s4.UUID)
		assert.NotEqual(t, s1.UUID, s4.UUID)
	})

	t.Run("Subscribe", func(t *testing.T) {
		t.Parallel()

		filename := filepath.Join(t.TempDir(), "state.json")
		p, err := NewProvider(filename)
		require.NoError(t, err)

		ch := p.Subscribe()

		require.Len(t, ch, cap(ch), "channel should be full")

		err = p.Update(func(s *State) { *s = State{UUID: "00000000-0000-0000-0000-000000000000"} })
		require.NoError(t, err)

		expected := &State{
			UUID:  "11111111-1111-1111-1111-111111111111",
			Start: time.Now(),
		}
		err = p.Update(func(s *State) { *s = *expected })
		require.NoError(t, err)

		assert.Equal(t, expected, p.Get())
		require.Len(t, ch, cap(ch), "channel should be full")

		<-ch
		assert.Equal(t, expected, p.Get())
		require.Empty(t, ch)

		got := make(chan struct{})
		go func() {
			<-ch

			assert.Equal(t, expected, p.Get())
			require.Empty(t, ch)
			close(got)
		}()

		err = p.Update(func(s *State) { *s = *expected })
		require.NoError(t, err)

		<-got
	})
}
