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
	"maps"
	"runtime"
	"slices"
	"strconv"
	"sync"
)

// Embeddable handler is a [slog.Handler] for custom logger.
// It is used to make sure that the config.LogLevel is effective.
//
//nolint:vet // for readability
type embeddableHandler struct {
	opts *NewHandlerOpts
	ga   []groupOrAttrs

	// m is used to protect opts.Handler from race conditions.
	m *sync.Mutex

	testAttrs map[string]any
}

// newEmbeddableHandler creates a new embeddable handler.
func newEmbeddableHandler(opts *NewHandlerOpts, attrs map[string]any) *embeddableHandler {
	return &embeddableHandler{opts: opts, testAttrs: attrs, m: new(sync.Mutex)}
}

// Enabled implements [slog.Handler].
func (h *embeddableHandler) Enabled(ctx context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}

	return l >= minLevel
}

// Handle implements [slog.Handler].
func (h *embeddableHandler) Handle(ctx context.Context, record slog.Record) error {
	if h.testAttrs != nil {
		if !h.opts.RemoveTime && !record.Time.IsZero() {
			t := record.Time.Format(timeLayout)
			h.testAttrs[slog.TimeKey] = t
		}

		if !h.opts.RemoveLevel {
			h.testAttrs[slog.LevelKey] = record.Level.String()
		}

		if !h.opts.RemoveSource {
			f, _ := runtime.CallersFrames([]uintptr{record.PC}).Next()
			if f.File != "" {
				s := shortPath(f.File) + ":" + strconv.Itoa(f.Line)
				h.testAttrs[slog.SourceKey] = s
			}
		}

		if record.Message != "" {
			h.testAttrs[slog.MessageKey] = record.Message
		}

		if m := attrs(record, h.ga); len(m) > 0 {
			maps.Copy(h.testAttrs, m)
		}
	}

	return h.opts.Handler.Handle(ctx, record)
}

// WithAttrs implements [slog.Handler].
func (h *embeddableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h.m.Lock()
	defer h.m.Unlock()
	h.opts.Handler = h.opts.Handler.WithAttrs(attrs)

	return &embeddableHandler{
		m:         h.m,
		opts:      h.opts,
		ga:        append(slices.Clone(h.ga), groupOrAttrs{attrs: attrs}),
		testAttrs: h.testAttrs,
	}
}

// WithGroup implements [slog.Handler].
func (h *embeddableHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	h.m.Lock()
	defer h.m.Unlock()
	h.opts.Handler = h.opts.Handler.WithGroup(name)

	return &embeddableHandler{
		m:         h.m,
		opts:      h.opts,
		ga:        append(slices.Clone(h.ga), groupOrAttrs{group: name}),
		testAttrs: h.testAttrs,
	}
}
