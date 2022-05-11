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
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

type record struct {
	Level      string
	Time       time.Time
	LoggerName string
	Message    string
	Caller     zapcore.EntryCaller
	Stack      string
}

// logRAM structure storage of log records in memory.
type logRAM struct {
	records map[int64]*record
	mu      *sync.RWMutex
}

// NewLogRAM is creating entries log in memory.
func NewLogRAM() *logRAM {
	return &logRAM{
		records: make(map[int64]*record),
		mu:      &sync.RWMutex{},
	}
}

// append adding a log entry.
func (l *logRAM) append(id int64, r *record) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.records) > 1024 {
		l.delete()
	}

	l.records[id] = r
}

// deleting log entry with minimal id.
func (l *logRAM) delete() {
	var i int64 = 0
	for k := range l.records {
		if k > i || i == 0 {
			i = k
		}
	}
	delete(l.records, i)
}
