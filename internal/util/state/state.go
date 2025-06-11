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
	"strconv"
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

	// may be empty if FerretDB did not connect to PostgreSQL yet
	PostgreSQLVersion string `json:"-"`
	DocumentDBVersion string `json:"-"`

	// as reported by beacon, if known
	LatestVersion   string `json:"-"`
	UpdateInfo      string `json:"-"`
	UpdateAvailable bool   `json:"-"`
}

// asMap return state as a map, including non-persisted fields.
func (s *State) asMap() map[string]any {
	return map[string]any{
		"uuid":               s.UUID,
		"telemetry":          s.TelemetryString(),
		"telemetry_locked":   strconv.FormatBool(s.TelemetryLocked),
		"start":              s.Start.Format(time.RFC3339),
		"postgresql_version": s.PostgreSQLVersion,
		"documentdb_version": s.DocumentDBVersion,
		"latest_version":     s.LatestVersion,
		"update_info":        s.UpdateInfo,
		"update_available":   strconv.FormatBool(s.UpdateAvailable),
	}
}

// deepCopy returns a deep copy of the state.
func (s *State) deepCopy() *State {
	var telemetry *bool
	if s.Telemetry != nil {
		telemetry = pointer.ToBool(*s.Telemetry)
	}

	return &State{
		UUID:              s.UUID,
		Telemetry:         telemetry,
		TelemetryLocked:   s.TelemetryLocked,
		Start:             s.Start,
		PostgreSQLVersion: s.PostgreSQLVersion,
		DocumentDBVersion: s.DocumentDBVersion,
		LatestVersion:     s.LatestVersion,
		UpdateInfo:        s.UpdateInfo,
		UpdateAvailable:   s.UpdateAvailable,
	}
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

// fill replaces all unset or invalid values with default.
func (s *State) fill() {
	if _, err := uuid.Parse(s.UUID); err != nil {
		s.UUID = uuid.NewString()
	}

	if s.Start.IsZero() {
		s.Start = time.Now()
	}
}
