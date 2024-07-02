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
	"fmt"
	"log/slog"
	"os"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// LevelDPanic panics in debug builds.
	LevelDPanic = slog.LevelError + 1

	// LevelPanic always panics.
	LevelPanic = slog.LevelError + 2

	// LevelFatal exits with a non-zero status.
	LevelFatal = slog.LevelError + 3
)

// Error returns [slog.Attr] for the given error (that can be nil) with error's message as a value.
func Error(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "<nil>")
	}

	return slog.String("error", err.Error())
}

// GoError returns [slog.Attr] for the given error (that can be nil) with error's Go representation as a value.
func GoError(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "<nil>")
	}

	return slog.String("error", fmt.Sprintf("%#v", err))
}

// setupSlog initializes slog logging with given options and UUID.
func setupSlog(opts *NewHandlerOpts, uuid string) {
	must.NotBeZero(opts)

	h := NewHandler(os.Stderr, opts)

	l := slog.New(h)
	if uuid != "" {
		l = l.With(slog.String("uuid", uuid))
	}

	slog.SetDefault(l)
	// slog.SetLogLoggerLevel(slog.LevelInfo + 2) //nolint:mnd // "strange" level to better differentiate non-slog logs
}
