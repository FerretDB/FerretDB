package handler

import (
	"math"
	"testing"
	"time"

	"github.com/FerretDB/wire/wirebson"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRenameTimeLimit(t *testing.T) {
	for _, value := range []any{int32(100), int64(100), float64(100)} {
		limit, err := renameTimeLimit(wirebson.MustDocument("renameCollection", "a.b", "maxTimeMS", value))
		if err != nil || limit != 100*time.Millisecond {
			t.Fatalf("invalid deadline for %T: %v %v", value, limit, err)
		}
	}
	for _, value := range []any{int32(-1), float64(1.5), math.Inf(1), math.NaN(), int64(math.MaxInt32) + 1, "100", true} {
		if _, err := renameTimeLimit(wirebson.MustDocument("renameCollection", "a.b", "maxTimeMS", value)); err == nil {
			t.Fatalf("invalid deadline accepted: %v", value)
		}
	}
	limit, err := renameTimeLimit(wirebson.MustDocument("renameCollection", "a.b"))
	if err != nil || limit != 0 {
		t.Fatal("missing deadline changed default behavior")
	}
}

func TestRenameDecimalDeadline(t *testing.T) {
	for _, tc := range []struct {
		value   string
		want    time.Duration
		invalid bool
	}{
		{"100", 100 * time.Millisecond, false}, {"100.5", 100 * time.Millisecond, false},
		{"-0.1", 0, false}, {"NaN", 0, false}, {"Infinity", 0, true},
		{"-1.1", 0, true}, {"2147483647.9", 0, true},
	} {
		decimal, err := bson.ParseDecimal128(tc.value)
		if err != nil {
			t.Fatal(err)
		}
		high, low := decimal.GetBytes()
		limit, err := renameTimeLimit(wirebson.MustDocument("renameCollection", "a.b", "maxTimeMS", wirebson.Decimal128{H: high, L: low}))
		if (err != nil) != tc.invalid || !tc.invalid && limit != tc.want {
			t.Fatalf("invalid decimal deadline %s: %v %v", tc.value, limit, err)
		}
	}
}
