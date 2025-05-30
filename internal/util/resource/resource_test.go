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

package resource

import (
	"runtime"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Object struct {
	token *Token
}

var globalObj *Object

//nolint:godot // line directives are not normal comments
func TestTrackUntrack(t *testing.T) {
	ch := make(chan string, 10)
	cleanup = func(msg string) {
		ch <- msg
	}

	name := profileName(globalObj)

	t.Run("LocalUntrack", func(t *testing.T) {
		require.Nil(t, pprof.Lookup(name))

		obj := &Object{token: NewToken()}

//line testtrack.go:100
		Track(obj, obj.token)

		assert.Equal(t, 1, pprof.Lookup(name).Count())

		Untrack(obj, obj.token)

		runtime.GC()
		select {
		case <-ch:
			t.Fatal("unexpected cleanup message")
		default:
		}

		assert.Equal(t, 0, pprof.Lookup(name).Count())
	})

	t.Run("LocalCleanup", func(t *testing.T) {
		require.Equal(t, 0, pprof.Lookup(name).Count())

		obj := &Object{token: NewToken()}

//line testtrack.go:200
		Track(obj, obj.token)

		assert.Equal(t, 1, pprof.Lookup(name).Count())
		runtime.KeepAlive(obj)

		runtime.GC()
		msg := <-ch

		assert.Contains(t, msg, "testtrack.go:200")
		assert.Equal(t, 1, pprof.Lookup(name).Count())
	})

	t.Run("Global", func(t *testing.T) {
		require.Equal(t, 1, pprof.Lookup(name).Count())

		globalObj = &Object{token: NewToken()}

//line testtrack.go:300
		Track(globalObj, globalObj.token)

		assert.Equal(t, 2, pprof.Lookup(name).Count())

		runtime.GC()
		select {
		case <-ch:
			t.Fatal("unexpected cleanup message")
		default:
		}

		assert.Equal(t, 2, pprof.Lookup(name).Count())
	})

	runtime.GC()
	time.Sleep(time.Second)
	require.Empty(t, ch)
}
