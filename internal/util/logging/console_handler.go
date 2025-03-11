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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"runtime"
	"slices"
	"strconv"
	"sync"

	"golang.org/x/term"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// timeLayout is the format of date time used by the console handler.
const timeLayout = "2006-01-02T15:04:05.000Z0700"

var (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	gray    = "\033[37m"
	white   = "\033[97m"
)

// consoleHandler is a [slog.Handler] that writes logs to the console.
// The format is intended to be more human-readable than [slog.TextHandler]'s logfmt.
// The format is not stable.
//
// See https://golang.org/s/slog-handler-guide.
//
//nolint:vet // for readability
type consoleHandler struct {
	opts *NewHandlerOpts

	ga []groupOrAttrs

	testAttrs map[string]any

	m   *sync.Mutex
	out io.Writer
	t   *term.Terminal
}

// newConsoleHandler creates a new console handler.
func newConsoleHandler(out io.Writer, opts *NewHandlerOpts, testAttrs map[string]any) *consoleHandler {
	must.NotBeZero(opts)

	f := out.(*os.File)

	if !term.IsTerminal(int(f.Fd())) {
		panic(1)
	}

	t := term.NewTerminal(f, "")

	return &consoleHandler{
		opts:      opts,
		testAttrs: testAttrs,
		m:         new(sync.Mutex),
		t:         t,
	}
}

// Enabled implements [slog.Handler].
func (ch *consoleHandler) Enabled(_ context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if ch.opts.Level != nil {
		minLevel = ch.opts.Level.Level()
	}

	return l >= minLevel
}

func (ch *consoleHandler) coloredLevel(l slog.Level) string {
	switch l {
	case slog.LevelInfo:
		return string(ch.t.Escape.Green) + l.String() + string(ch.t.Escape.Reset)
	case slog.LevelWarn:
		return yellow + l.String() + reset
	case slog.LevelError:
		return red + l.String() + reset
	case slog.LevelDebug:
		return magenta + l.String() + reset
	}

	return l.String()
}

// Handle implements [slog.Handler].
func (ch *consoleHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf bytes.Buffer

	if !ch.opts.RemoveTime && !r.Time.IsZero() {
		t := r.Time.Format(timeLayout)
		buf.WriteString(t)
		buf.WriteRune('\t')

		if ch.testAttrs != nil {
			ch.testAttrs[slog.TimeKey] = t
		}
	}

	if !ch.opts.RemoveLevel {
		// l := r.Level.String()
		buf.WriteString(ch.coloredLevel(r.Level))
		buf.WriteRune('\t')

		if ch.testAttrs != nil {
			ch.testAttrs[slog.LevelKey] = r.Level.String()
		}
	}

	if !ch.opts.RemoveSource {
		f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		if f.File != "" {
			s := shortPath(f.File) + ":" + strconv.Itoa(f.Line)
			buf.WriteString(s)
			buf.WriteRune('\t')

			if ch.testAttrs != nil {
				ch.testAttrs[slog.SourceKey] = s
			}
		}
	}

	if r.Message != "" {
		buf.WriteString(r.Message)

		if ch.testAttrs != nil {
			ch.testAttrs[slog.MessageKey] = r.Message
		}
	}

	if m := attrs(r, ch.ga); len(m) > 0 {
		buf.WriteRune('\t')

		var b bytes.Buffer
		encoder := json.NewEncoder(&b)
		encoder.SetEscapeHTML(false)

		err := encoder.Encode(m)
		if !ch.opts.SkipChecks {
			must.NoError(err)
		}

		if err == nil {
			buf.Write(bytes.TrimSuffix(b.Bytes(), []byte{'\n'}))
		} else {
			// last resort
			buf.WriteString(fmt.Sprintf("%#v", m))
		}

		if ch.testAttrs != nil {
			maps.Copy(ch.testAttrs, m)
		}
	}

	buf.WriteRune('\n')

	ch.m.Lock()
	defer ch.m.Unlock()

	_, err := buf.WriteTo(ch.t)

	return err
}

// WithAttrs implements [slog.Handler].
func (ch *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return ch
	}

	return &consoleHandler{
		opts:      ch.opts,
		ga:        append(slices.Clone(ch.ga), groupOrAttrs{attrs: attrs}),
		m:         ch.m,
		out:       ch.out,
		t:         ch.t,
		testAttrs: ch.testAttrs,
	}
}

// WithGroup implements [slog.Handler].
func (ch *consoleHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return ch
	}

	return &consoleHandler{
		opts:      ch.opts,
		ga:        append(slices.Clone(ch.ga), groupOrAttrs{group: name}),
		m:         ch.m,
		out:       ch.out,
		t:         ch.t,
		testAttrs: ch.testAttrs,
	}
}

// check interfaces
var (
	_ slog.Handler = (*consoleHandler)(nil)
)
