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

package testutil

import (
	"io"
	"log/slog"
	"strings"

	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// logWriter provides [io.Writer] for [testing.TB].
type logWriter struct {
	tb testtb.TB
}

// Write implements [io.Writer].
func (lw *logWriter) Write(p []byte) (int, error) {
	// "logging.go:xx" is added by testing.TB.Log itself; there is nothing we can do about it.
	// lw.tb.Helper() does not help. See:
	// https://github.com/golang/go/issues/59928
	// https://github.com/neilotoole/slogt/tree/v1.1.0?tab=readme-ov-file#deficiency
	lw.tb.Log(strings.TrimSuffix(string(p), "\n"))
	return len(p), nil
}

// LevelLogger returns a test logger for the given level (which might be dynamic).
func LevelLogger(tb testtb.TB, level slog.Leveler) *slog.Logger {
	h := logging.NewHandler(&logWriter{tb: tb}, &logging.NewHandlerOpts{
		Base:       "console",
		Level:      level,
		RemoveTime: true,
	})

	return slog.New(h)
}

// Logger returns slog test logger.
func Logger(tb testtb.TB) *slog.Logger {
	return LevelLogger(tb, slog.LevelDebug)
}

// check interfaces
var (
	_ io.Writer = (*logWriter)(nil)
)
