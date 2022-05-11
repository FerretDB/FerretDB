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
)

type record struct {
	Level      string
	Time       time.Time
	LoggerName string
	Message    string
	Stack      string
}

// logRAM structure storage of log records in memory.
type logRAM struct {
	counter int64
	log     [1024]*record
	mu      sync.RWMutex
}

// NewLogRAM is creating entries log in memory.
func NewLogRAM() *logRAM {
	return &logRAM{
		log: [1024]*record{},
		mu:  sync.RWMutex{},
	}
}

// append adding entry in logram.
func (l *logRAM) append(rec *record) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.counter = (l.counter + 1) % 1024
	l.log[l.counter] = rec
}
