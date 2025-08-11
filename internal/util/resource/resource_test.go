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

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil/teststress"
)

// Object represents a tracked object for tests.
type Object struct {
	token *Token
}

// globalObj is a global object that is never cleaned up.
var globalObj *Object

// See https://go.dev/doc/gc-guide#Testing_object_death
// and https://pkg.go.dev/cmd/compile#hdr-Line_Directives.
//
//nolint:godot // line directives are not normal comments
func TestTrackUntrack(t *testing.T) {
	name := profileName(globalObj)

	runtime.GC()

	t.Run("LocalUntrack", func(t *testing.T) {
		obj := &Object{token: NewToken()}

//line testtrack.go:100
		Track(obj, obj.token)

		assert.Equal(t, 1, pprof.Lookup(name).Count())

		Untrack(obj, obj.token)

		runtime.GC()

		assert.Equal(t, 0, pprof.Lookup(name).Count())
	})

	t.Run("LocalCleanup", func(t *testing.T) {
		token := NewToken()
		ch := make(chan string, 1)
		token.cleanup = func(t *Token) {
			ch <- t.msg
		}

		obj := &Object{token: token}

//line testtrack.go:200
		Track(obj, obj.token)

		assert.Equal(t, 1, pprof.Lookup(name).Count())

		tt := obj.token
		runtime.KeepAlive(obj)

		runtime.GC()
		msg := <-ch

		assert.Contains(t, msg, "testtrack.go:200")
		assert.Equal(t, 1, pprof.Lookup(name).Count())

		// remove profile manually to support `go test -count=X`
		pprof.Lookup(name).Remove(tt)
		assert.Equal(t, 0, pprof.Lookup(name).Count())
	})

	t.Run("Global", func(t *testing.T) {
		globalObj = &Object{token: NewToken()}

//line testtrack.go:300
		Track(globalObj, globalObj.token)

		assert.Equal(t, 1, pprof.Lookup(name).Count())

		runtime.GC()

		assert.Equal(t, 1, pprof.Lookup(name).Count())

		Untrack(globalObj, globalObj.token)

		runtime.GC()

		assert.Equal(t, 0, pprof.Lookup(name).Count())
	})

	runtime.GC()
}

func TestUntrackConcurrently(t *testing.T) {
	obj := &Object{token: NewToken()}
	Track(obj, obj.token)

	name := profileName(obj)
	assert.Equal(t, 1, pprof.Lookup(name).Count())

	teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
		ready <- struct{}{}

		<-start

		Untrack(obj, obj.token)
	})

	assert.Equal(t, 0, pprof.Lookup(name).Count())

	runtime.GC()
}
