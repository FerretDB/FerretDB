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
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircularBufferHandler(t *testing.T) {
	opts := &NewHandlerOpts{
		Base:              "console",
		Level:             slog.LevelInfo,
		CheckMessages:     true,
		recentEntriesSize: 2,
	}
	Setup(opts, "")

	for _, tc := range []struct { //nolint:vet // for readability
		msg      string
		level    slog.Level
		expected []slog.Record
	}{
		{
			msg:   "message 1",
			level: slog.LevelWarn,
			expected: []slog.Record{{
				Level:   slog.LevelWarn,
				Message: "message 1",
			}},
		},
		{
			msg:   "message 2",
			level: slog.LevelError,
			expected: []slog.Record{{
				Level:   slog.LevelWarn,
				Message: "message 1",
			}, {
				Level:   slog.LevelError,
				Message: "message 2",
			}},
		},
		{
			msg:   "debug not added",
			level: slog.LevelDebug,
			expected: []slog.Record{{
				Level:   slog.LevelWarn,
				Message: "message 1",
			}, {
				Level:   slog.LevelError,
				Message: "message 2",
			}},
		},
		{
			msg:   "message 3",
			level: slog.LevelInfo,
			expected: []slog.Record{{
				Level:   slog.LevelError,
				Message: "message 2",
			}, {
				Level:   slog.LevelInfo,
				Message: "message 3",
			}},
		},
	} {
		t.Run(tc.msg, func(t *testing.T) {
			slog.Default().Log(context.Background(), tc.level, tc.msg)

			records := slog.Default().Handler().(*Handler).recentEntries.get()
			actual := make([]slog.Record, len(records))

			for i, r := range records {
				r.Time = time.Time{}
				r.PC = 0
				actual[i] = *r
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}
