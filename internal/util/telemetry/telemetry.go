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

type Flag struct {
	v *bool
}

func (s *Flag) UnmarshalText(text []byte) error {
	v, err := parseValue(string(text))
	if err != nil {
		return err
	}

	*s = Flag{v: v}
	return nil
}

func State(f *Flag, dnt string, execName string, l *zap.Logger) (*bool, error) {
	var disabled bool

	// https://consoledonottrack.com is not entirely clear about accepted values.
	// Assume that "1", "t", "true", etc. mean that telemetry should be disabled,
	// and other valid values, including empty string, mean undecided.
	v, err := parseValue(dnt)
	if err != nil {
		return nil, err
	}
	if pointer.GetBool(v) {
		l.Info(fmt.Sprintf("Telemetry is disabled by DO_NOT_TRACK=%s environment variable.", dnt))
		disabled = true
	}

	if strings.Contains(strings.ToLower(execName), "donottrack") {
		l.Info(fmt.Sprintf("Telemetry is disabled by %q executable name.", execName))
		disabled = true
	}

	if disabled {
		// check for conflicts
		if f.v != nil && *f.v {
			return nil, fmt.Errorf("telemetry is disabled by DO_NOT_TRACK environment variable or executable name")
		}

		return pointer.ToBool(false), nil
	}

	return f.v, nil
}

// check interfaces
var (
	_ encoding.TextUnmarshaler = (*Flag)(nil)
)
