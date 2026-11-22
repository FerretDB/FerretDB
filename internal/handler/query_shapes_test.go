package handler

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/wire/wirebson"
)

func TestShapeValueOmitsScalarValues(t *testing.T) {
	d := wirebson.MustDocument("find", "synthetic_collection", "filter", wirebson.MustDocument(
		"token", "synthetic-secret-not-for-output", "count", int32(42), "tags", wirebson.MustArray("synthetic-value")))
	b, err := json.Marshal(shapeValue(d, 0))
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"synthetic_collection", "synthetic-secret-not-for-output", "synthetic-value", "42"} {
		if strings.Contains(string(b), value) {
			t.Fatalf("scalar value leaked into shape")
		}
	}
	for _, key := range []string{"filter", "token", "count", "tags", "int32"} {
		if !strings.Contains(string(b), key) {
			t.Fatalf("missing shape information %s", key)
		}
	}
}

func TestShapeRecorderDisabledByDefault(t *testing.T) {
	t.Setenv("FERRETDB_QUERY_SHAPES_PATH", "")
	r, err := newShapeRecorder()
	if err != nil || r != nil {
		t.Fatal("capture should be disabled")
	}
}

func TestShapeRecorderRejectsUnacknowledgedCapture(t *testing.T) {
	t.Setenv("FERRETDB_QUERY_SHAPES_PATH", t.TempDir()+"/shapes.jsonl")
	t.Setenv("FERRETDB_QUERY_SHAPES_SYNTHETIC_ONLY", "false")
	if _, err := newShapeRecorder(); err == nil {
		t.Fatal("expected explicit synthetic-only acknowledgement")
	}
}

func TestShapeRecorderSkipsAuthenticationAndExpires(t *testing.T) {
	t.Setenv("FERRETDB_QUERY_SHAPES_PEER", "")
	t.Setenv("FERRETDB_QUERY_SHAPES_PATH", t.TempDir()+"/shapes")
	t.Setenv("FERRETDB_QUERY_SHAPES_SYNTHETIC_ONLY", "true")
	r, err := newShapeRecorder()
	if err != nil {
		t.Fatal(err)
	}
	for _, cmd := range []string{"saslStart", "saslContinue", "createUser", "updateUser"} {
		req, err := middleware.RequestDoc(wirebson.MustDocument(cmd, "synthetic-secret", "$db", "admin"))
		if err != nil {
			t.Fatal(err)
		}
		r.record(context.Background(), req, nil)
	}
	b, err := os.ReadFile(r.path)
	if err != nil || len(b) != 0 {
		t.Fatal("authentication command was captured")
	}
	req, err := middleware.RequestDoc(wirebson.MustDocument("find", "synthetic", "$db", "unifi-ferretdb-test-main"))
	if err != nil {
		t.Fatal(err)
	}
	r.until = time.Now().Add(-time.Second)
	r.record(context.Background(), req, nil)
	b, err = os.ReadFile(r.path)
	if err != nil || len(b) != 0 {
		t.Fatal("expired capture wrote data")
	}
	r.until = time.Now().Add(time.Minute)
	r.record(context.Background(), req, nil)
	b, err = os.ReadFile(r.path)
	if err != nil || len(b) == 0 {
		t.Fatal("eligible request was not captured")
	}
	if strings.Contains(string(b), "synthetic-secret") {
		t.Fatal("secret leaked")
	}
}
