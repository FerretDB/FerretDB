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
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

var test_entrys = []zapcore.Entry{}

func init() {
	for i := 0; i < 2048; i++ {
		l := zapcore.Level(rand.Intn(6) - 1)
		en := zapcore.Entry{
			Level:      l,
			Time:       time.Now(),
			LoggerName: "logger_" + l.String(),
			Message:    "Test message in logger " + l.String(),
		}

		test_entrys = append(test_entrys, en)
	}
}

func TestLogRAM(t *testing.T) {
	t.Parallel()

	size := int64(512)
	logram := NewLogRAM(size)

	for j := 0; j < 100; j++ {
		logram.append(test_entrys[j])
	}

	ln_ram := int64(len(logram.log))
	ln_log := int64(len(logram.getLogRAM()))

	assert.Equal(t, size, ln_ram)
	assert.Equal(t, int64(100), ln_log)

	for j := 100; j < 2048; j++ {
		logram.append(test_entrys[j])
	}

	ln_ram = int64(len(logram.log))
	ln_log = int64(len(logram.getLogRAM()))

	assert.Equal(t, size, ln_ram)
	assert.Equal(t, size, ln_log)
}
