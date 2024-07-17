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

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// RecentEntries implements log records interception
// and stores the last 1024 entries in circular buffer in memory.
var RecentEntries = NewCircularBuffer(1024)

// circularBuffer is a storage of log records in memory.
type circularBuffer struct {
	mu      sync.RWMutex
	records []*zapcore.Entry
	index   int
}

// NewCircularBuffer creates a circular buffer for log records in memory.
func NewCircularBuffer(size int) *circularBuffer {
	if size < 1 {
		panic(fmt.Sprintf("buffer size must be at least 1, but %d provided", size))
	}

	return &circularBuffer{
		records: make([]*zapcore.Entry, size),
	}
}

// add adds an entry in circularBuffer.
func (cb *circularBuffer) add(record zapcore.Entry) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.records[cb.index] = &record
	cb.index = (cb.index + 1) % len(cb.records)
}

// get returns entries from circularBuffer.
func (cb *circularBuffer) get() []zapcore.Entry {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	l := len(cb.records)
	res := make([]zapcore.Entry, 0, l)

	for n := range l {
		i := (cb.index + n) % l

		if r := cb.records[i]; r != nil {
			res = append(res, *r)
		}
	}

	return res
}

// GetArray is a version of Get that returns an array as expected by mongosh.
func (cb *circularBuffer) GetArray() (*bson.Array, error) {
	records := cb.get()
	res := bson.MakeArray(len(records))

	for _, r := range records {
		b, err := json.Marshal(map[string]any{
			"t": map[string]time.Time{
				"$date": r.Time,
			},
			"l":   r.Level,
			"msg": r.Message,
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if err = res.Add(string(b)); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return res, nil
}
