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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLogRAM(t *testing.T) {
	for name, tc := range map[string]struct {
		size     int64
		msgPanic string
	}{
		"PanicNegativSize": {
			size:     -2,
			msgPanic: "logram size -2",
		},
		"PanicZeroSize": {
			size:     0,
			msgPanic: "logram size 0",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			assert.PanicsWithValue(t, tc.msgPanic, func() { NewLogRAM(tc.size) })
		})
	}

	logram := NewLogRAM(3)
	for n, tc := range []struct {
		inLog    zapcore.Entry
		expected []zapcore.Entry
	}{
		{
			inLog: zapcore.Entry{
				Level:      1,
				Time:       time.Date(2022, 12, 31, 11, 59, 1, 0, time.UTC),
				LoggerName: "logger_1",
				Message:    "message 1",
			},
			expected: []zapcore.Entry{
				{
					Level:      1,
					Time:       time.Date(2022, 12, 31, 11, 59, 1, 0, time.UTC),
					LoggerName: "logger_1",
					Message:    "message 1",
				},
			},
		}, {
			inLog: zapcore.Entry{
				Level:      2,
				Time:       time.Date(2022, 12, 31, 11, 59, 2, 0, time.UTC),
				LoggerName: "logger_2",
				Message:    "message 2",
			},
			expected: []zapcore.Entry{
				{
					Level:      1,
					Time:       time.Date(2022, 12, 31, 11, 59, 1, 0, time.UTC),
					LoggerName: "logger_1",
					Message:    "message 1",
				},
				{
					Level:      2,
					Time:       time.Date(2022, 12, 31, 11, 59, 2, 0, time.UTC),
					LoggerName: "logger_2",
					Message:    "message 2",
				},
			},
		}, {
			inLog: zapcore.Entry{
				Level:      3,
				Time:       time.Date(2022, 12, 31, 11, 59, 3, 0, time.UTC),
				LoggerName: "logger_3",
				Message:    "message 3",
			},
			expected: []zapcore.Entry{
				{
					Level:      1,
					Time:       time.Date(2022, 12, 31, 11, 59, 1, 0, time.UTC),
					LoggerName: "logger_1",
					Message:    "message 1",
				},
				{
					Level:      2,
					Time:       time.Date(2022, 12, 31, 11, 59, 2, 0, time.UTC),
					LoggerName: "logger_2",
					Message:    "message 2",
				},
				{
					Level:      3,
					Time:       time.Date(2022, 12, 31, 11, 59, 3, 0, time.UTC),
					LoggerName: "logger_3",
					Message:    "message 3",
				},
			},
		}, {
			inLog: zapcore.Entry{
				Level:      4,
				Time:       time.Date(2022, 12, 31, 11, 59, 4, 0, time.UTC),
				LoggerName: "logger_4",
				Message:    "message 4",
			},
			expected: []zapcore.Entry{
				{
					Level:      2,
					Time:       time.Date(2022, 12, 31, 11, 59, 2, 0, time.UTC),
					LoggerName: "logger_2",
					Message:    "message 2",
				},
				{
					Level:      3,
					Time:       time.Date(2022, 12, 31, 11, 59, 3, 0, time.UTC),
					LoggerName: "logger_3",
					Message:    "message 3",
				},
				{
					Level:      4,
					Time:       time.Date(2022, 12, 31, 11, 59, 4, 0, time.UTC),
					LoggerName: "logger_4",
					Message:    "message 4",
				},
			},
		}, {
			inLog: zapcore.Entry{
				Level:      5,
				Time:       time.Date(2022, 12, 31, 11, 59, 5, 0, time.UTC),
				LoggerName: "logger_5",
				Message:    "message 5",
			},
			expected: []zapcore.Entry{
				{
					Level:      3,
					Time:       time.Date(2022, 12, 31, 11, 59, 3, 0, time.UTC),
					LoggerName: "logger_3",
					Message:    "message 3",
				},
				{
					Level:      4,
					Time:       time.Date(2022, 12, 31, 11, 59, 4, 0, time.UTC),
					LoggerName: "logger_4",
					Message:    "message 4",
				},
				{
					Level:      5,
					Time:       time.Date(2022, 12, 31, 11, 59, 5, 0, time.UTC),
					LoggerName: "logger_5",
					Message:    "message 5",
				},
			},
		},
	} {
		name := fmt.Sprintf("AppendGet_%d", n)
		tc := tc
		t.Run(name, func(t *testing.T) {
			logram.append(&tc.inLog)
			actual := logram.getLogRAM()
			assert.Equal(t, tc.expected, actual)
		})
	}
}
