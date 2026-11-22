package handler

import (
	"github.com/FerretDB/wire/wirebson"
	"testing"
)

func TestUniFiDurabilityAlias(t *testing.T) {
	for _, tc := range []struct {
		fsync, full, commit string
		want                int32
	}{
		{"on", "on", "on", 1}, {"on", "on", "local", 1},
		{"off", "on", "on", 0}, {"on", "off", "on", 0},
		{"on", "on", "off", 0}, {"on", "on", "unknown", 0},
	} {
		d := unifiDurabilityAlias(tc.fsync, tc.full, tc.commit)
		log, err := d.Get("log").(wirebson.AnyDocument).Decode()
		if err != nil {
			t.Fatal(err)
		}
		if got := log.Get("ferretdb.postgresqlDurableWAL"); got != tc.want {
			t.Fatalf("unexpected durability alias for %v: %v", tc, got)
		}
	}
}
