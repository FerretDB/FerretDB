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
	"log/slog"
	"os"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

const (
	// LevelDPanic panics in development builds.
	LevelDPanic = slog.LevelError + 1

	// LevelPanic always panics.
	LevelPanic = slog.LevelError + 2

	// LevelFatal exits with a non-zero status.
	LevelFatal = slog.LevelError + 3
)

// nameKey is a [slog.Attr] key used by [WithName].
const nameKey = "name"

// WithName returns a logger with a given period-separated name.
//
// How this name is used depends on the handler.
func WithName(l *slog.Logger, name string) *slog.Logger {
	return l.With(slog.String(nameKey, name))
}

// Error returns [slog.Attr] for the given error (that can be nil) with error's message as a value.
func Error(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "<nil>")
	}

	return slog.String("error", err.Error())
}

// LazyString is a lazily evaluated [slog.LogValuer].
type LazyString func() string

// LogValue implements [slog.LogValuer].
func (ls LazyString) LogValue() slog.Value { return slog.StringValue(ls()) }

// Setup initializes slog logging with given options and UUID.
func Setup(opts *NewHandlerOpts, uuid string) {
	must.NotBeZero(opts)

	h := NewHandler(os.Stderr, opts)

	l := slog.New(h)
	if uuid != "" {
		l = l.With(slog.String("uuid", uuid))
	}

	slog.SetDefault(l)
	slog.SetLogLoggerLevel(slog.LevelInfo + 2)
}

// check interfaces
var (
	_ slog.LogValuer = (LazyString)(nil)
)
