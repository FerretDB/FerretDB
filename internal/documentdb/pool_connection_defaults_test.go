package documentdb

import (
	"net/url"
	"testing"
)

func TestUniFiConnectionDefaults(t *testing.T) {
	for _, test := range []struct {
		name     string
		enabled  string
		values   url.Values
		expected string
	}{
		{"disabled", "", url.Values{}, ""},
		{"enabled", "true", url.Values{}, "1s"},
		{"explicit-disable", "true", url.Values{"client_connection_check_interval": {"0"}}, "0"},
		{"explicit-interval", "true", url.Values{"client_connection_check_interval": {"250ms"}}, "250ms"},
		{"explicit-startup-option", "true", url.Values{"options": {"-c client_connection_check_interval=0"}}, ""},
		{"unrelated-startup-option", "true", url.Values{"options": {"-c work_mem=4MB"}}, "1s"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("FERRETDB_UNIFI_COMPAT", test.enabled)
			newPgxPoolSetDefaults(test.values)
			if got := test.values.Get("client_connection_check_interval"); got != test.expected {
				t.Fatalf("unexpected disconnect-check interval: %q", got)
			}
			if test.values.Has("statement_timeout") || test.values.Has("lock_timeout") || test.values.Has("transaction_timeout") {
				t.Fatal("connection checks introduced an unrelated query deadline")
			}
		})
	}
}
