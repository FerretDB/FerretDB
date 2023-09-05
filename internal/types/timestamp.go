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

package types

import (
	"sync/atomic"
	"time"
)

// Timestamp represents BSON type Timestamp.
type Timestamp uint64

// timestampCounter is a process-wide timestamp counter.
var timestampCounter atomic.Uint32

// NewTimestamp returns the timestamp for the given time and counter values.
func NewTimestamp(t time.Time, c uint32) Timestamp {
	return Timestamp((uint64(t.Unix()) << 32) | uint64(c))
}

// NextTimestamp returns the next timestamp for the given time value.
func NextTimestamp(t time.Time) Timestamp {
	// Technically, that should be a counter within a second, not a process-wide,
	// but that's good enough for us.
	c := timestampCounter.Add(1)

	return NewTimestamp(t, c)
}

// Time returns timestamp's time component.
func (ts Timestamp) Time() time.Time {
	sec := int64(ts >> 32)
	return time.Unix(sec, 0).UTC()
}
