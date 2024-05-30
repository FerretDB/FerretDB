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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCircularBuffer(t *testing.T) {
	for name, tc := range map[string]struct {
		size     int64
		msgPanic string
	}{
		"PanicNegativSize": {
			size:     -2,
			msgPanic: "buffer size must be at least 1, but -2 provided",
		},
		"PanicZeroSize": {
			size:     0,
			msgPanic: "buffer size must be at least 1, but 0 provided",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			assert.PanicsWithValue(t, tc.msgPanic, func() { NewCircularBuffer(tc.size) })
		})
	}

	logram := NewCircularBuffer(2)
	for n, tc := range []struct {
		inLog    zapcore.Entry
		expected []zapcore.Entry
	}{{
		inLog: zapcore.Entry{
			Level:      1,
			Time:       time.Date(2022, 12, 31, 11, 59, 1, 0, time.UTC),
			LoggerName: "logger_1",
			Message:    "message 1",
		},
		expected: []zapcore.Entry{{
			Level:      1,
			Time:       time.Date(2022, 12, 31, 11, 59, 1, 0, time.UTC),
			LoggerName: "logger_1",
			Message:    "message 1",
		}},
	}, {
		inLog: zapcore.Entry{
			Level:      2,
			Time:       time.Date(2022, 12, 31, 11, 59, 2, 0, time.UTC),
			LoggerName: "logger_2",
			Message:    "message 2",
		},
		expected: []zapcore.Entry{{
			Level:      1,
			Time:       time.Date(2022, 12, 31, 11, 59, 1, 0, time.UTC),
			LoggerName: "logger_1",
			Message:    "message 1",
		}, {
			Level:      2,
			Time:       time.Date(2022, 12, 31, 11, 59, 2, 0, time.UTC),
			LoggerName: "logger_2",
			Message:    "message 2",
		}},
	}, {
		inLog: zapcore.Entry{
			Level:      3,
			Time:       time.Date(2022, 12, 31, 11, 59, 3, 0, time.UTC),
			LoggerName: "logger_3",
			Message:    "message 3",
		},
		expected: []zapcore.Entry{{
			Level:      2,
			Time:       time.Date(2022, 12, 31, 11, 59, 2, 0, time.UTC),
			LoggerName: "logger_2",
			Message:    "message 2",
		}, {
			Level:      3,
			Time:       time.Date(2022, 12, 31, 11, 59, 3, 0, time.UTC),
			LoggerName: "logger_3",
			Message:    "message 3",
		}},
	}} {
		name := fmt.Sprintf("AppendGet_%d", n)
		tc := tc
		t.Run(name, func(t *testing.T) {
			logram.append(&tc.inLog)
			actual := logram.get(zap.DebugLevel)
			for i, exp := range tc.expected {
				assert.Equal(t, exp, *actual[i])
			}
		})
	}

	Setup(zap.DebugLevel, "console", "")
	logger := zap.L()

	for n, tc := range []struct {
		addMsg   string
		expected []string
	}{{
		addMsg:   "Test message 1",
		expected: []string{"Test message 1"},
	}, {
		addMsg:   "Test message 2",
		expected: []string{"Test message 1", "Test message 2"},
	}, {
		addMsg:   "Test message 3",
		expected: []string{"Test message 1", "Test message 2", "Test message 3"},
	}, {
		addMsg:   "Test message 4",
		expected: []string{"Test message 1", "Test message 2", "Test message 3", "Test message 4"},
	}} {
		name := fmt.Sprintf("ZapHooks_%d", n)
		tc := tc
		t.Run(name, func(t *testing.T) {
			logger.Info(tc.addMsg)
			actual := RecentEntries.get(zap.DebugLevel)
			for i, exp := range tc.expected {
				assert.Equal(t, exp, actual[i].Message)
			}
		})
	}
}
