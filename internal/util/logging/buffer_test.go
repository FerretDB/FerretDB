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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCircularBufferHook(t *testing.T) {
	RecentEntries = NewCircularBuffer(2)

	Setup(zap.InfoLevel, "console", "")

	for _, tc := range []struct { //nolint:vet // for readability
		msg      string
		level    zapcore.Level
		expected []zapcore.Entry
	}{
		{
			msg:   "message 1",
			level: zap.WarnLevel,
			expected: []zapcore.Entry{{
				Level:   zap.WarnLevel,
				Message: "message 1",
			}},
		},
		{
			msg:   "message 2",
			level: zap.ErrorLevel,
			expected: []zapcore.Entry{{
				Level:   zap.WarnLevel,
				Message: "message 1",
			}, {
				Level:   zap.ErrorLevel,
				Message: "message 2",
			}},
		},
		{
			msg:   "debug not added",
			level: zap.DebugLevel,
			expected: []zapcore.Entry{{
				Level:   zap.WarnLevel,
				Message: "message 1",
			}, {
				Level:   zap.ErrorLevel,
				Message: "message 2",
			}},
		},
		{
			msg:   "message 3",
			level: zap.InfoLevel,
			expected: []zapcore.Entry{{
				Level:   zap.ErrorLevel,
				Message: "message 2",
			}, {
				Level:   zap.InfoLevel,
				Message: "message 3",
			}},
		},
	} {
		t.Run(tc.msg, func(t *testing.T) {
			zap.L().Log(tc.level, tc.msg)

			actual := RecentEntries.get()

			for i, a := range actual {
				a.Time = time.Time{}
				a.Caller = zapcore.EntryCaller{}
				a.Stack = ""
				actual[i] = a
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}
