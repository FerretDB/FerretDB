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

package logging

import (
	"fmt"
	"sync"

	"go.uber.org/zap/zapcore"
)

// RecentEntries implements zap logging entries interception
// and stores the last 1024 entries in circular buffer in memory.
var RecentEntries *circularBuffer

// circularBuffer is a storage of log records in memory.
type circularBuffer struct {
	mu    sync.RWMutex
	log   []*zapcore.Entry
	index int64
}

// NewCircularBuffer creates a circular buffer for log entries in memory.
func NewCircularBuffer(size int64) *circularBuffer {
	if size < 1 {
		panic(fmt.Sprintf("buffer size must be at least 1, but %d provided", size))
	}

	return &circularBuffer{
		log: make([]*zapcore.Entry, size),
	}
}

// append adds an entry in circularBuffer.
func (l *circularBuffer) append(entry *zapcore.Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.log[l.index] = entry
	l.index = (l.index + 1) % int64(len(l.log))
}

// Get returns entries from circularBuffer.
func (l *circularBuffer) Get() []*zapcore.Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	n := len(l.log)
	entries := make([]*zapcore.Entry, 0, n)
	for i := int64(0); i < int64(len(l.log)); i++ {
		k := (i + l.index) % int64(len(l.log))

		if l.log[k] != nil {
			entries = append(entries, l.log[k])
		}
	}

	return entries
}
