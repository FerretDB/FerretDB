package cursor

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/FerretDB/wire/wirebson"
)

func TestCloseNamespacePreservesUnrelatedCursors(t *testing.T) {
	ctx := context.Background()
	r := NewRegistry(slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer r.Close(ctx)
	continuation, err := wirebson.MustDocument("state", int32(1)).Encode()
	if err != nil {
		t.Fatal(err)
	}
	r.NewCursor(ctx, 1, continuation, nil, "source", "collection")
	r.NewCursor(ctx, 2, continuation, nil, "source", "unrelated")
	r.NewCursor(ctx, 3, continuation, nil, "target", "collection")
	r.NewCursor(ctx, 4, continuation, nil, "source", "")
	r.CloseNamespace(ctx, "source", "collection")
	if !r.ConsumeInvalidation(1) || r.ConsumeInvalidation(1) {
		t.Fatal("rename invalidation should be reported once")
	}
	for _, id := range []int64{1, 4} {
		if value, _ := r.GetCursor(id); value != nil {
			t.Fatalf("cursor %d was not invalidated", id)
		}
	}
	for _, id := range []int64{2, 3} {
		if value, _ := r.GetCursor(id); value == nil {
			t.Fatalf("unrelated cursor %d was invalidated", id)
		}
	}
}

func TestCursorConnectionBudgetReleasedOnEveryRemovalPath(t *testing.T) {
	ctx := context.Background()
	r := NewRegistry(slog.New(slog.NewTextHandler(io.Discard, nil)))
	closed := false
	defer func() {
		if !closed {
			r.Close(ctx)
		}
	}()
	continuation, err := wirebson.MustDocument("state", int32(1)).Encode()
	if err != nil {
		t.Fatal(err)
	}
	released := 0
	release := func() { released++ }
	r.NewCursor(ctx, 1, continuation, nil, "a", "first", release)
	r.CloseNamespace(ctx, "a", "first")
	r.CloseNamespace(ctx, "a", "first")
	if released != 1 {
		t.Fatal("namespace invalidation did not release exactly once")
	}
	r.NewCursor(ctx, 2, continuation, nil, "a", "second", release)
	r.NewCursor(ctx, 2, continuation, nil, "a", "second", release)
	if released != 2 {
		t.Fatal("cursor replacement leaked its connection budget")
	}
	r.UpdateCursor(ctx, 2, nil)
	if released != 3 {
		t.Fatal("cursor exhaustion leaked its connection budget")
	}
	r.NewCursor(ctx, 3, continuation, nil, "a", "third", release)
	r.CloseCursor(ctx, 3)
	r.CloseCursor(ctx, 3)
	if released != 4 {
		t.Fatal("cursor kill did not release exactly once")
	}
	r.NewCursor(ctx, 4, continuation, nil, "a", "last", release)
	r.Close(ctx)
	closed = true
	if released != 5 {
		t.Fatal("registry shutdown leaked its connection budget")
	}
}

func TestCursorGenerationDoesNotSurviveReplacementOrRemoval(t *testing.T) {
	ctx := context.Background()
	r := NewRegistry(slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer r.Close(ctx)
	continuation, err := wirebson.MustDocument("state", int32(1)).Encode()
	if err != nil {
		t.Fatal(err)
	}
	r.NewCursor(ctx, 1, continuation, nil, "source", "collection")
	r.SetGeneration(1, "original-uuid")
	db, collection, generation, tracked := r.Generation(1)
	if db != "source" || collection != "collection" || generation != "original-uuid" || !tracked {
		t.Fatal("cursor lost its namespace identity")
	}
	r.NewCursor(ctx, 1, continuation, nil, "target", "replacement")
	if _, _, _, tracked := r.Generation(1); tracked {
		t.Fatal("replacement inherited a stale identity")
	}
	// An absent catalog entry is still a tracked identity, not an opt-out.
	r.SetGeneration(1, "")
	if _, _, _, tracked := r.Generation(1); !tracked {
		t.Fatal("empty identity disabled validation")
	}
	r.CloseCursor(ctx, 1)
	r.SetGeneration(1, "late-update")
	if _, _, _, tracked := r.Generation(1); tracked {
		t.Fatal("removed cursor was resurrected")
	}
}

func TestCursorDependencySnapshots(t *testing.T) {
	ctx := context.Background()
	r := NewRegistry(slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer r.Close(ctx)
	continuation, err := wirebson.MustDocument("state", int32(1)).Encode()
	if err != nil {
		t.Fatal(err)
	}
	r.NewCursor(ctx, 1, continuation, nil, "source", "main")
	dependencies := []NamespaceGeneration{{Database: "source", Collection: "joined", UUID: "original"}}
	r.SetDependencies(1, dependencies)
	dependencies[0].UUID = "caller-mutated"
	snapshot := r.Dependencies(1)
	if len(snapshot) != 1 || snapshot[0].UUID != "original" {
		t.Fatal("caller changed the retained namespace identity")
	}
	snapshot[0].UUID = "reader-mutated"
	if r.Dependencies(1)[0].UUID != "original" {
		t.Fatal("reader changed the retained namespace identity")
	}
	r.NewCursor(ctx, 1, continuation, nil, "source", "replacement")
	if len(r.Dependencies(1)) != 0 {
		t.Fatal("replacement inherited old dependencies")
	}
	r.CloseCursor(ctx, 1)
	r.SetDependencies(1, dependencies)
	if len(r.Dependencies(1)) != 0 {
		t.Fatal("late update resurrected removed dependencies")
	}
}
