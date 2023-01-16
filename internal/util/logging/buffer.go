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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// RecentEntries implements zap logging entries interception
// and stores the last 1024 entries in circular buffer in memory.
var RecentEntries = NewCircularBuffer(1024)

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

// get returns entries from circularBuffer with level at minLevel or above.
func (l *circularBuffer) get(minLevel zapcore.Level) []*zapcore.Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	n := len(l.log)
	entries := make([]*zapcore.Entry, 0, n)
	for i := int64(0); i < int64(len(l.log)); i++ {
		k := (i + l.index) % int64(len(l.log))

		if l.log[k] != nil && l.log[k].Level >= minLevel {
			entries = append(entries, l.log[k])
		}
	}

	return entries
}

// GetArray is a version of Get that returns an array as expected by mongosh.
func (l *circularBuffer) GetArray(minLevel zapcore.Level) (*types.Array, error) {
	entries := l.get(minLevel)
	res := types.MakeArray(len(entries))

	for _, e := range entries {
		b, err := json.Marshal(map[string]any{
			"t": map[string]time.Time{
				"$date": e.Time,
			},
			"l":   e.Level,
			"ln":  e.LoggerName,
			"msg": e.Message,
			"c":   e.Caller,
			"s":   e.Stack,
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res.Append(string(b))
	}

	return res, nil
}
