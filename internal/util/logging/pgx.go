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

//go:build !ferretdb_no_postgresql

package logging

import (
	"context"
	"log/slog"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/tracelog"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// pgxLogLevels maps pgx log levels to slog levels.
var pgxLogLevels = map[tracelog.LogLevel]slog.Level{
	// pgx's info level is used for queries, but we want to hide them by default
	tracelog.LogLevelTrace: slog.LevelDebug - 2,
	tracelog.LogLevelDebug: slog.LevelDebug,
	tracelog.LogLevelInfo:  slog.LevelDebug + 2,

	tracelog.LogLevelWarn:  slog.LevelWarn,
	tracelog.LogLevelError: slog.LevelError,
}

// pgxLogger is a pgx's [tracelog.Logger] implementation that uses slog
// with correct source code location.
type pgxLogger struct {
	l *slog.Logger
}

// NewPgxLogger creates a new [tracelog.Logger] that uses given [slog.Logger].
func NewPgxLogger(l *slog.Logger) tracelog.Logger {
	return &pgxLogger{l: l}
}

// Log implements [tracelog.Logger].
func (sl *pgxLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	now := time.Now()

	l, ok := pgxLogLevels[level]
	if !ok {
		sl.l.LogAttrs(ctx, LevelDPanic, "Invalid pgx log level for: "+msg, slog.Int("level", int(level)))
		l = slog.LevelError
	}

	if !sl.l.Enabled(ctx, l) {
		return
	}

	// We can't call runtime.Caller(s) with some fixed skip value and then inspect a single frame
	// because call Pgx calls Log method from different stack depths.

	var pc uintptr
	callers := make([]uintptr, 10)

	if n := runtime.Callers(2, callers); n != 0 {
		frames := runtime.CallersFrames(callers[:n])

		for {
			frame, more := frames.Next()

			if !strings.Contains(frame.File, "github.com/jackc/pgx") {
				pc = frame.PC
				break
			}

			if !more {
				break
			}
		}
	}

	record := slog.NewRecord(now, l, msg, pc)

	dataKeys := maps.Keys(data)
	slices.Sort(dataKeys)

	msgAttrs := make([]slog.Attr, len(data))

	for i, k := range dataKeys {
		v := data[k]

		var attr slog.Attr

		switch v := v.(type) {
		case []any:
			attrs := make([]slog.Attr, len(v))
			for i, v := range v {
				attrs[i] = slog.Attr{
					Key:   strconv.Itoa(i),
					Value: slog.AnyValue(v),
				}
			}

			attr = slog.Attr{
				Key:   k,
				Value: slog.GroupValue(attrs...),
			}

		default:
			attr = slog.Attr{
				Key:   k,
				Value: slog.AnyValue(v),
			}
		}

		msgAttrs[i] = attr
	}

	record.AddAttrs(msgAttrs...)

	err := sl.l.Handler().Handle(ctx, record)

	if debugbuild.Enabled {
		must.NoError(err)
	}
}

// check interfaces
var (
	_ tracelog.Logger = (*pgxLogger)(nil)
)
