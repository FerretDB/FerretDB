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

// Package telemetry provides basic telemetry facilities.
package telemetry

import (
	"encoding"
	"fmt"
	"log/slog"
	"strings"

	"github.com/AlekSi/pointer"
)

// parseValue parses a string value into true, false, or nil.
func parseValue(s string) (*bool, error) {
	switch strings.ToLower(s) {
	case "1", "t", "true", "y", "yes", "on", "enable", "enabled", "optin", "opt-in", "allow":
		return pointer.ToBool(true), nil
	case "0", "f", "false", "n", "no", "off", "disable", "disabled", "optout", "opt-out", "disallow", "forbid":
		return pointer.ToBool(false), nil
	case "", "undecided":
		return nil, nil
	default:
		return nil, fmt.Errorf("failed to parse %s", s)
	}
}

// Flag represents a Kong flag with three states: true, false, and undecided (nil).
type Flag struct {
	v *bool
}

// UnmarshalText is used by Kong to parse a flag value.
func (s *Flag) UnmarshalText(text []byte) error {
	v, err := parseValue(string(text))
	if err != nil {
		return err
	}

	*s = Flag{v: v}

	return nil
}

// initialState returns initial telemetry state based on:
//   - Kong flag value (including `FERRETDB_TELEMETRY` environment variable);
//   - common DO_NOT_TRACK environment variable;
//   - executable name;
//   - and the previously saved state.
//
// The second returned value is true if the telemetry state should be locked, because of
// setting telemetry via a command-line flag, an environment variable, or a filename.
func initialState(f *Flag, dnt string, execName string, prev *bool, l *slog.Logger) (state *bool, locked bool, err error) {
	// https://consoledonottrack.com is not entirely clear about accepted values.
	// Assume that "1", "t", "true", etc. mean that telemetry should be disabled,
	// and other valid values, including "0" and empty string, mean undecided.
	dntV, err := parseValue(dnt)
	if err != nil {
		return
	}

	if pointer.GetBool(dntV) {
		l.Info(fmt.Sprintf("Telemetry is disabled by DO_NOT_TRACK=%s environment variable.", dnt))
		state = pointer.ToBool(false)
		locked = true
	}

	if strings.Contains(strings.ToLower(execName), "donottrack") {
		l.Info(fmt.Sprintf("Telemetry is disabled by %q executable name.", execName))
		state = pointer.ToBool(false)
		locked = true
	}

	// telemetry state is disabled and locked via flag, dnt env or binary name
	if state != nil {
		// check for conflicts
		if f.v != nil && *f.v {
			err = fmt.Errorf("telemetry can't be enabled")
		}

		return
	}

	// if flag is unset, use previous unlocked state
	if f.v == nil {
		state = prev

		if state == nil {
			// undecided state, reporter would log about it during run
			return
		}

		if *state {
			l.Info("Telemetry is enabled because it was enabled previously.")
		} else {
			l.Info("Telemetry is disabled because it was disabled previously.")
		}

		return
	}

	// flag is set, use it as locked state
	state = f.v
	locked = true

	if *state {
		l.Info("Telemetry enabled.")
	} else {
		l.Info("Telemetry disabled.")
	}

	return
}

// check interfaces
var (
	_ encoding.TextUnmarshaler = (*Flag)(nil)
)
