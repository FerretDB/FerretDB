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

// LogRAM implements zap logging entry interception
// and storage of last 1024 entry in ring buffer in memory.
package logging

import (
	"fmt"
	"sync"

	"go.uber.org/zap/zapcore"
)

var LogRAM *logRAM

// logRAM structure storage of log records in memory.
type logRAM struct {
	mu    sync.RWMutex
	log   []*zapcore.Entry
	index int64
}

// NewLogRAM is creating entries log in memory.
func NewLogRAM(size int64) *logRAM {
	if size < 1 {
		panic(fmt.Sprintf("logram size must be at least 1, but %d provided", size))
	}

	return &logRAM{
		log: make([]*zapcore.Entry, size),
	}
}

// append is adding entry in logram.
func (l *logRAM) append(entry *zapcore.Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.log[l.index] = entry
	l.index = (l.index + 1) % int64(len(l.log))
}

// Get returns entrys from logRAM.
func (l *logRAM) Get() []*zapcore.Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var entries []*zapcore.Entry
	for i := int64(0); i < int64(len(l.log)); i++ {
		k := (i + l.index) % int64(len(l.log))

		if l.log[k] != nil {
			entries = append(entries, l.log[k])
		}
	}

	return entries
}
