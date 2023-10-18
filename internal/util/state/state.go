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

// Package state stores FerretDB process state.
package state

import (
	"time"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
)

// State represents FerretDB process state.
type State struct {
	UUID      string `json:"uuid"`
	Telemetry *bool  `json:"telemetry,omitempty"` // nil for undecided

	// all following fields are never persisted

	TelemetryLocked bool      `json:"-"`
	Start           time.Time `json:"-"`

	// may be empty if FerretDB did not connect to the backend yet
	BackendName    string `json:"-"`
	BackendVersion string `json:"-"`

	// as reported by beacon, if known
	LatestVersion   string `json:"-"`
	UpdateAvailable bool   `json:"-"`
}

// TelemetryString returns "enabled", "disabled" or "undecided".
func (s *State) TelemetryString() string {
	if s.Telemetry == nil {
		return "undecided"
	}

	if *s.Telemetry {
		return "enabled"
	}

	return "disabled"
}

// DisableTelemetry disables telemetry.
//
// It also sets LatestVersion and UpdateAvailable to zero values
// to avoid stale values when telemetry is re-enabled.
func (s *State) DisableTelemetry() {
	s.Telemetry = pointer.ToBool(false)
	s.LatestVersion = ""
	s.UpdateAvailable = false
}

// EnableTelemetry enables telemetry.
func (s *State) EnableTelemetry() {
	s.Telemetry = pointer.ToBool(true)
}

// fill replaces all unset or invalid values with default.
func (s *State) fill() {
	if _, err := uuid.Parse(s.UUID); err != nil {
		s.UUID = uuid.NewString()
	}

	if s.Start.IsZero() {
		s.Start = time.Now()
	}
}

// deepCopy returns a deep copy of the state.
func (s *State) deepCopy() *State {
	var telemetry *bool
	if s.Telemetry != nil {
		telemetry = pointer.ToBool(*s.Telemetry)
	}

	return &State{
		UUID:            s.UUID,
		Telemetry:       telemetry,
		TelemetryLocked: s.TelemetryLocked,
		Start:           s.Start,
		BackendName:     s.BackendName,
		BackendVersion:  s.BackendVersion,
		LatestVersion:   s.LatestVersion,
		UpdateAvailable: s.UpdateAvailable,
	}
}
