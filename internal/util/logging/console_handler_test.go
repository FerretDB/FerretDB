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
	"bytes"
	"log/slog"
	"sync"
	"testing"
	"testing/slogtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/term"
)

func TestConsoleHandler(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	var testAttrs map[string]any

	newHandler := func(t *testing.T) slog.Handler {
		t.Helper()

		buf.Reset()

		testAttrs = map[string]any{}

		opts := &NewHandlerOpts{
			Level: slog.LevelDebug,
		}

		return newConsoleHandler(&buf, opts, testAttrs)
	}

	result := func(t *testing.T) map[string]any {
		t.Helper()

		return testAttrs
	}

	slogtest.Run(t, newHandler, result)
}

func TestConsoleHandlerEscapeCodes(t *testing.T) {
	t.Parallel()

	esc := term.NewTerminal(nil, "").Escape

	for name, tc := range map[string]struct {
		expected string
		level    slog.Level
	}{
		"Debug-2": {
			level:    slog.LevelDebug - 2,
			expected: "\033[34mDEBUG-2\033[0m\tfoobar\n",
		},
		"Debug-1": {
			level:    slog.LevelDebug - 1,
			expected: "\033[34mDEBUG-1\033[0m\tfoobar\n",
		},
		"Debug": {
			level:    slog.LevelDebug,
			expected: "\033[34mDEBUG\033[0m\tfoobar\n",
		},
		"Debug+1": {
			level:    slog.LevelDebug + 1,
			expected: "\033[34mDEBUG+1\033[0m\tfoobar\n",
		},
		"Debug+2": {
			level:    slog.LevelDebug + 2,
			expected: "\033[34mDEBUG+2\033[0m\tfoobar\n",
		},
		"Info": {
			level:    slog.LevelInfo,
			expected: "\033[32mINFO\033[0m\tfoobar\n",
		},
		"Info+1": {
			level:    slog.LevelInfo + 1,
			expected: "\033[32mINFO+1\033[0m\tfoobar\n",
		},
		"Info+2": {
			level:    slog.LevelInfo + 2,
			expected: "\033[32mINFO+2\033[0m\tfoobar\n",
		},
		"Info+3": {
			level:    slog.LevelInfo + 3,
			expected: "\033[32mINFO+3\033[0m\tfoobar\n",
		},
		"Warn": {
			level:    slog.LevelWarn,
			expected: "\033[33mWARN\033[0m\tfoobar\n",
		},
		"Warn+1": {
			level:    slog.LevelWarn + 1,
			expected: "\033[33mWARN+1\033[0m\tfoobar\n",
		},
		"Warn+2": {
			level:    slog.LevelWarn + 2,
			expected: "\033[33mWARN+2\033[0m\tfoobar\n",
		},
		"Warn+3": {
			level:    slog.LevelWarn + 3,
			expected: "\033[33mWARN+3\033[0m\tfoobar\n",
		},
		"Error": {
			level:    slog.LevelError,
			expected: "\033[31mERROR\033[0m\tfoobar\n",
		},
		"DPanic": {
			level:    LevelDPanic,
			expected: "\033[31mERROR+1\033[0m\tfoobar\n",
		},
		"Panic": {
			level:    LevelPanic,
			expected: "\033[31mERROR+2\033[0m\tfoobar\n",
		},
		"Fatal": {
			level:    LevelFatal,
			expected: "\033[31mERROR+3\033[0m\tfoobar\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			ch := &consoleHandler{
				opts: &NewHandlerOpts{Level: slog.LevelInfo},
				m:    new(sync.Mutex),
				out:  &buf,
				esc:  esc,
			}

			r := slog.Record{
				Level:   tc.level,
				Message: "foobar",
			}

			require.NoError(t, ch.Handle(t.Context(), r))
			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

func TestConsoleHandlerNoTTYMode(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	ch := newConsoleHandler(&buf, &NewHandlerOpts{Level: slog.LevelInfo}, nil)

	expected := "INFO\tfoobar\n"

	r := slog.Record{
		Level:   slog.LevelInfo,
		Message: "foobar",
	}

	require.NoError(t, ch.Handle(t.Context(), r))
	assert.Equal(t, expected, buf.String())
}
