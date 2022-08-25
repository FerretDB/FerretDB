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

type (
	// Timestamp represents BSON type Timestamp.
	Timestamp int64
)

// timestampCounter is an ordinal number for timestamps in the system.
var timestampCounter uint32

// NewTimestamp returns a timestamp from time and an increment.
func NewTimestamp(t time.Time, c uint32) Timestamp {
	sec := t.Unix()
	sec <<= 32
	sec |= int64(c)
	return Timestamp(sec)
}

// NextTimestamp returns a timestamp from time and an internal ops counter.
func NextTimestamp(t time.Time) Timestamp {
	c := atomic.AddUint32(&timestampCounter, 1)
	return NewTimestamp(t, c)
}

// Time returns time.Time ignoring increment.
func (t Timestamp) Time() time.Time {
	t >>= 32
	return time.Unix(int64(t), 0)
}
