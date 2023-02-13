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

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// State represents FerretDB process state.
type State struct {
	UUID      string `json:"uuid"`
	Telemetry *bool  `json:"telemetry,omitempty"` // nil for undecided

	// never persisted
	TelemetryLocked bool      `json:"-"`
	Start           time.Time `json:"-"`
	LatestVersion   string    `json:"-"` // may be empty
	HandlerVersion  string    `json:"-"` // may be empty
}

// fill replaces all unset or invalid values with default.
func (s *State) fill() {
	if _, err := uuid.Parse(s.UUID); err != nil {
		s.UUID = must.NotFail(uuid.NewRandom()).String()
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
		LatestVersion:   s.LatestVersion,
		HandlerVersion:  s.HandlerVersion,
	}
}
