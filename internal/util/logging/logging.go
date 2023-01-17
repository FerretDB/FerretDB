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

// Package logging provides logging helpers.
package logging

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/FerretDB/FerretDB/build/version"
)

// Setup initializes logging with a given level.
func Setup(level zapcore.Level, uuid string) {
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       version.Get().DebugBuild,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
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

	logger = logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		RecentEntries.append(&entry)
		return nil
	}))

	setupWithLogger(logger)
}

// setupWithLogger initializes logging with a given logger and its level.
func setupWithLogger(logger *zap.Logger) {
	zap.ReplaceGlobals(logger)

	if _, err := zap.RedirectStdLogAt(logger, zap.InfoLevel); err != nil {
		log.Fatal(err)
	}
}
