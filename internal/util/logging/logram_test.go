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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

var test_entrys = []zapcore.Entry{}

func init() {
	for i := 0; i < 20; i++ {
		l := zapcore.Level(i%7 - 1)
		en := zapcore.Entry{
			Level:      l,
			Time:       time.Now(),
			LoggerName: "logger_" + l.String(),
			Message:    "message " + strconv.Itoa(i+1),
		}

		test_entrys = append(test_entrys, en)
	}
}

func TestLogRAM(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		size        int64
		numEntries  int64
		msgPanic    string
		bufferMsg   []any
		expectedMsg []any
	}{
		"PanicNegativSize": {
			size:     -2,
			msgPanic: "logram size -2",
		},
		"PanicZeroSize": {
			size:     0,
			msgPanic: "logram size 0",
		},
		"Append3of6": {
			size:        6,
			numEntries:  3,
			bufferMsg:   []any{"message 1", "message 2", "message 3"},
			expectedMsg: []any{"message 1", "message 2", "message 3"},
		},
		"Append20of6": {
			size:        6,
			numEntries:  20,
			bufferMsg:   []any{"message 19", "message 20", "message 15", "message 16", "message 17", "message 18"},
			expectedMsg: []any{"message 15", "message 16", "message 17", "message 18", "message 19", "message 20"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.msgPanic != "" {
				assert.PanicsWithValue(t, tc.msgPanic, func() { NewLogRAM(tc.size) })
				return
			}

			logram := NewLogRAM(tc.size)

			for i := int64(0); i < tc.numEntries; i++ {
				logram.append(test_entrys[i])
			}

			assert.Len(t, logram.log, int(tc.size))

			actualLog := logram.getLogRAM()
			assert.Len(t, logram.getLogRAM(), len(tc.expectedMsg))

			actualBufferMsg := getMsg(logram.log)
			actualMsg := getMsg(actualLog)

			assert.Equal(t, tc.bufferMsg, actualBufferMsg)
			assert.Equal(t, tc.expectedMsg, actualMsg)
		})
	}
}

func getMsg(rs []*zapcore.Entry) (actual []any) {
	for _, r := range rs {
		if r != nil {
			actual = append(actual, r.Message)
		}
	}
	return
}
