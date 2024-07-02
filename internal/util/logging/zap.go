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
	"log"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
)

// logLevels maps zap log levels to slog levels.
var logLevels = map[zapcore.Level]slog.Level{
	zapcore.DebugLevel:  slog.LevelDebug,
	zapcore.InfoLevel:   slog.LevelInfo,
	zapcore.WarnLevel:   slog.LevelWarn,
	zapcore.ErrorLevel:  slog.LevelError,
	zapcore.DPanicLevel: slog.LevelError,
	zapcore.PanicLevel:  slog.LevelError,
	zapcore.FatalLevel:  slog.LevelError,
}

// Setup initializes logging with a given level.
func Setup(level zapcore.Level, encoding, uuid string) {
	setupSlog(&NewHandlerOpts{
		Base:  encoding,
		Level: logLevels[level],
	}, uuid)

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       debugbuild.Enabled,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          encoding,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:          "M",
			LevelKey:            "L",
			TimeKey:             "T",
			NameKey:             "N",
			CallerKey:           "C",
			FunctionKey:         zapcore.OmitKey,
			StacktraceKey:       "S",
			LineEnding:          zapcore.DefaultLineEnding,
			EncodeLevel:         zapcore.CapitalLevelEncoder,
			EncodeTime:          zapcore.ISO8601TimeEncoder,
			EncodeDuration:      zapcore.StringDurationEncoder,
			EncodeCaller:        zapcore.ShortCallerEncoder,
			EncodeName:          nil,
			NewReflectedEncoder: nil,
			ConsoleSeparator:    "\t",
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    nil,
	}

	if uuid != "" {
		config.InitialFields = map[string]any{"uuid": uuid}
	}

	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}

	SetupWithZapLogger(WithHooks(logger))
}

// WithHooks returns a logger with recent entries hooks.
func WithHooks(logger *zap.Logger) *zap.Logger {
	return logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		RecentEntries.append(&entry)
		return nil
	}))
}

// SetupWithZapLogger initializes zap logging with a given logger and its level.
func SetupWithZapLogger(logger *zap.Logger) {
	zap.ReplaceGlobals(logger)

	if _, err := zap.RedirectStdLogAt(logger, zap.InfoLevel); err != nil {
		log.Fatal(err)
	}
}
