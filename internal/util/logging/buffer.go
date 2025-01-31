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
	"log/slog"
	"sync"
	"time"

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// circularBuffer is a storage of log records in memory.
type circularBuffer struct {
	mu      sync.RWMutex
	records []*slog.Record
	index   int
}

// newCircularBuffer creates a circular buffer for log records in memory.
func newCircularBuffer(size int) *circularBuffer {
	if size < 1 {
		panic(fmt.Sprintf("buffer size must be at least 1, but %d provided", size))
	}

	return &circularBuffer{
		records: make([]*slog.Record, size),
	}
}

// add adds an entry in circularBuffer.
func (cb *circularBuffer) add(record *slog.Record) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.records[cb.index] = record
	cb.index = (cb.index + 1) % len(cb.records)
}

// get returns entries from circularBuffer.
func (cb *circularBuffer) get() []*slog.Record {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	l := len(cb.records)
	res := make([]*slog.Record, 0, l)

	for n := range l {
		i := (cb.index + n) % l

		if r := cb.records[i]; r != nil {
			res = append(res, r)
		}
	}

	return res
}

// getArray is a version of [circularBuffer.get] that returns an array as expected by mongosh.
func (cb *circularBuffer) getArray() (*wirebson.Array, error) {
	records := cb.get()
	res := wirebson.MakeArray(len(records))

	for _, r := range records {
		// TODO https://github.com/FerretDB/FerretDB/issues/4347
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
