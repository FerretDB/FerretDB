package handler

import (
	"math"
	"testing"

	"github.com/FerretDB/wire/wirebson"
)

func TestCappedIntegerValidation(t *testing.T) {
	for _, value := range []any{int32(4096), int64(4096), float64(4096)} {
		got, err := cappedInteger(wirebson.MustDocument("create", "cap", "size", value), "size", true)
		if err != nil || got != 4096 {
			t.Fatalf("valid capped size rejected: %v %v", got, err)
		}
	}
	for _, value := range []any{int32(0), int64(-1), 1.5, math.Inf(1), math.NaN(), "4096"} {
		if _, err := cappedInteger(wirebson.MustDocument("create", "cap", "size", value), "size", true); err == nil {
			t.Fatalf("invalid capped size accepted: %v", value)
		}
	}
	got, err := cappedInteger(wirebson.MustDocument("create", "cap"), "max", false)
	if err != nil || got != 0 {
		t.Fatal("missing optional maximum changed behavior")
	}
}
