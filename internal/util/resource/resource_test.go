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
)

type TestTrackObject struct {
	token *Token
}

// runGC forces several GC cycles to give the runtime a chance to run cleanups.
func runGC(t *testing.T) {
	t.Helper()

	for i := 0; i < 8; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}
}

// entryCount returns the number of entries for obj in its pprof profile.
func entryCount(t *testing.T, obj any) int {
	t.Helper()

	p := pprof.Lookup(profileName(obj))
	if p != nil {
		return p.Count()
	}

	return 0
}

func TestTrackProfileEntryAdded(t *testing.T) {
	obj := &TestTrackObject{token: NewToken()}
	Track(obj, obj.token)
	t.Cleanup(func() { Untrack(obj, obj.token) })

	assert.Equal(t, 1, entryCount(t, obj), "profile should have exactly one entry")
}

func TestTrackNoCleanupWhileReachable(t *testing.T) {
	obj := &TestTrackObject{token: NewToken()}
	Track(obj, obj.token)
	t.Cleanup(func() { Untrack(obj, obj.token) })

	// GC should not run cleanup because obj is still reachable.
	runGC(t)

	runtime.KeepAlive(obj)

	assert.Equal(t, 1, entryCount(t, obj), "cleanup shouldn't run while object is reachable")
}

func TestTrackCleanupRunsWhenAbandoned(t *testing.T) {
	t.Skip("cleanup panic occurs in another goroutine; needs sub‑process test")

	obj := &TestTrackObject{token: NewToken()}
	Track(obj, obj.token)

	// Drop the reference; GC will trigger cleanup.
	obj = nil

	runGC(t)
}

func TestUntrackProfileEntryRemoved(t *testing.T) {
	obj := &TestTrackObject{token: NewToken()}
	Track(obj, obj.token)

	Untrack(obj, obj.token)
	assert.Equal(t, 0, entryCount(t, obj), "profile entry should be removed after Untrack")

	token := obj.token
	runtime.KeepAlive(obj)
	obj = nil
	// Object already unassigned from memory so cleanup is not necessary
	runGC(t)

	// Fixed field name to match implementation (cleanup not cleanupHandler)
	assert.Nil(t, token.cleanup.Load(), "cleanup handle should be nil after Untrack")
}
