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
	Timestamp uint64
)

// timestampOperationCounter is an ordinal number for timestamps in the system.
var timestampOperationCounter uint64

// NewTimestamp returns a timestamp from seconds and an increment.
func NewTimestamp(sec, inc uint64) Timestamp {
	sec <<= 32
	sec |= inc
	return Timestamp(sec)
}

// NextTimestamp returns a timestamp from seconds and an internal ops counter.
// TODO: get low-order 4 bytes for the inc instead of internal counter?
func NextTimestamp(sec uint64) Timestamp {
	inc := atomic.AddUint64(&timestampOperationCounter, 1)
	sec <<= 32
	sec |= inc
	return Timestamp(sec)
}

// DateTime returns time.Time ignoring increment.
func DateTime(t Timestamp) time.Time {
	t >>= 32
	return time.UnixMilli(int64(t))
}
