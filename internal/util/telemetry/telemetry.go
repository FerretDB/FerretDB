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
	"strings"

	"github.com/AlekSi/pointer"
	"go.uber.org/zap"
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
func initialState(f *Flag, dnt string, execName string, prev *bool, l *zap.Logger) (*bool, bool, error) {
	var disable, locked bool

	// https://consoledonottrack.com is not entirely clear about accepted values.
	// Assume that "1", "t", "true", etc. mean that telemetry should be disabled,
	// and other valid values, including "0" and empty string, mean undecided.
	v, err := parseValue(dnt)
	if err != nil {
		return nil, locked, err
	}

	if pointer.GetBool(v) {
		l.Sugar().Infof("Telemetry is disabled by DO_NOT_TRACK=%s environment variable.", dnt)
		disable = true
	}

	if strings.Contains(strings.ToLower(execName), "donottrack") {
		l.Sugar().Infof("Telemetry is disabled by %q executable name.", execName)
		disable = true
	}

	if disable {
		// check for conflicts
		if f.v != nil && *f.v {
			return nil, locked, fmt.Errorf("telemetry can't be enabled")
		}

		// telemetry state is disabled via flag, dnt env or binary name
		locked = true

		return pointer.ToBool(false), locked, nil
	}

	if f.v == nil {
		if prev == nil {
			// undecided state, reporter would log about it during run
			return nil, locked, nil
		}

		if *prev {
			l.Info("Telemetry is enabled because it was enabled previously.")
		} else {
			l.Info("Telemetry is disabled because it was disabled previously.")
		}

		return prev, locked, nil
	}

	// telemetry state is enabled via flag, dnt env or binary name
	locked = true

	if *f.v {
		l.Info("Telemetry enabled.")
	} else {
		l.Info("Telemetry disabled.")
	}

	return f.v, locked, nil
}

// check interfaces
var (
	_ encoding.TextUnmarshaler = (*Flag)(nil)
)
