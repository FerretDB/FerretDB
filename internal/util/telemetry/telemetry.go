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

func State(f *Flag, dnt string, execName string, l *zap.Logger) (state *bool, locked bool, err error) {
	return
}

// check interfaces
var (
	_ encoding.TextUnmarshaler = (*Flag)(nil)
)
