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
	"time"

	"go.uber.org/zap/zapcore"
)

type record struct {
	Level      zapcore.Level
	Time       time.Time
	LoggerName string
	Message    string
	Stack      string
}

// logRAM structure storage of log records in memory.
type logRAM struct {
	size    int64
	counter int64
	log     []*zapcore.Entry
	mu      sync.RWMutex
}

// NewLogRAM is creating entries log in memory.
func NewLogRAM(size int64) *logRAM {
	if size < 1 {
		panic(fmt.Sprintf("logram size %d", size))
	}

	return &logRAM{
		size: size,
		log:  make([]*zapcore.Entry, size),
		mu:   sync.RWMutex{},
	}
}

// append is adding entry in logram.
func (l *logRAM) append(entry zapcore.Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	rec := &zapcore.Entry{
		Level:      entry.Level,
		Time:       entry.Time,
		LoggerName: entry.LoggerName,
		Message:    entry.Message,
		Stack:      entry.Stack,
	}

	l.log[l.counter] = rec
	l.counter = (l.counter + 1) % l.size
}

// getLogRAM returns entrys from logRAM.
func (l *logRAM) getLogRAM() (entrys []*zapcore.Entry) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for i := int64(0); i < l.size; i++ {
		k := (i + l.counter) % l.size

		if l.log[k] != nil {
			e := &zapcore.Entry{
				Level:      l.log[k].Level,
				Time:       l.log[k].Time,
				LoggerName: l.log[k].LoggerName,
				Message:    l.log[k].Message,
				Stack:      l.log[k].Stack,
			}
			entrys = append(entrys, e)
		}
	}

	return
}
