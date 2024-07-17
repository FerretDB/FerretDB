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
	"log/slog"
	"testing"

	"github.com/neilotoole/slogt"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// Logger returns zap test logger with valid configuration.
func Logger(tb testtb.TB) *zap.Logger {
	return LevelLogger(tb, zap.NewAtomicLevelAt(zap.DebugLevel))
}

// LevelLogger returns zap test logger with given level and valid configuration.
func LevelLogger(tb testtb.TB, level zap.AtomicLevel) *zap.Logger {
	opts := []zaptest.LoggerOption{
		zaptest.Level(level),
		zaptest.WrapOptions(zap.AddCaller(), zap.Development()),
	}

	return zaptest.NewLogger(tb, opts...)
}

// SLogger returns slog test logger.
//
// TODO https://github.com/FerretDB/FerretDB/issues/4013
func SLogger(tb testtb.TB) *slog.Logger {
	t := tb.(testing.TB)
	return slogt.New(t)
}
