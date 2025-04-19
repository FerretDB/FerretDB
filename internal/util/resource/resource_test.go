package resource

import (
	"runtime"
	"runtime/pprof"
	"testing"
	"time"
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
	if p := pprof.Lookup(profileName(obj)); p != nil {
		return p.Count()
	}
	return 0
}

func TestTrackProfileEntryAdded(t *testing.T) {
	obj := &TestTrackObject{token: NewToken()}
	Track(obj, obj.token)
	t.Cleanup(func() { Untrack(obj, obj.token) })

	if c := entryCount(t, obj); c != 1 {
		t.Fatalf("want profile count 1, got %d", c)
	}
}

func TestTrackNoCleanupWhileReachable(t *testing.T) {
	obj := &TestTrackObject{token: NewToken()}
	Track(obj, obj.token)
	t.Cleanup(func() { Untrack(obj, obj.token) })

	// GC should not run cleanup because obj is still reachable.
	runGC(t)
	runtime.KeepAlive(obj)

	if c := entryCount(t, obj); c != 1 {
		t.Fatalf("cleanup ran too early; profile count = %d", c)
	}
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
	if c := entryCount(t, obj); c != 0 {
		t.Fatalf("profile entry still present after Untrack; count = %d", c)
	}

	token := obj.token
	runtime.KeepAlive(obj)
	obj = nil
	// Object already unassigned from memory so cleanup is not necessary
	runGC(t)

	// cleanupByToken should no longer contain the token.
	if _, ok := cleanupByToken.LoadAndDelete(token); ok {
		t.Fatalf("cleanup handle leaked in map after Untrack")
	}
}
