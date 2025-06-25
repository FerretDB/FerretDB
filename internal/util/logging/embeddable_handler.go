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

// EmbeddableHandler is a [slog.Handler] that wraps another handler.
// It provides additional functionality like level filtering, attribute manipulation,
// and thread-safe access to the embedded handler.
//
//nolint:vet // for readability
type embeddableHandler struct {
	opts *NewHandlerOpts
	ga   []groupOrAttrs
	// mu is used to ensure that the embeddable handler is not modified while it is being used.
	mu *sync.Mutex

	testAttrs map[string]any
}

// newEmbeddableHandler creates a new embeddable handler.
func newEmbeddableHandler(opts *NewHandlerOpts, attrs map[string]any) *embeddableHandler {
	return &embeddableHandler{opts: opts, testAttrs: attrs, mu: new(sync.Mutex)}
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
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.opts.RemoveTime && !record.Time.IsZero() {
		t := record.Time.Format(timeLayout)

		h.opts.EmbeddedHandler = h.opts.EmbeddedHandler.WithAttrs([]slog.Attr{
			{
				Key:   slog.TimeKey,
				Value: slog.StringValue(t),
			},
		})

		if h.testAttrs != nil {
			h.testAttrs[slog.TimeKey] = t
		}
	}

	if !h.opts.RemoveLevel {
		h.opts.EmbeddedHandler = h.opts.EmbeddedHandler.WithAttrs([]slog.Attr{
			{
				Key:   slog.LevelKey,
				Value: slog.StringValue(record.Level.String()),
			},
		})
		if h.testAttrs != nil {
			h.testAttrs[slog.LevelKey] = record.Level.String()
		}
	}

	if !h.opts.RemoveSource {
		f, _ := runtime.CallersFrames([]uintptr{record.PC}).Next()
		if f.File != "" {
			s := shortPath(f.File) + ":" + strconv.Itoa(f.Line)
			h.opts.EmbeddedHandler = h.opts.EmbeddedHandler.WithAttrs([]slog.Attr{
				{
					Key:   slog.SourceKey,
					Value: slog.StringValue(s),
				},
			})

			if h.testAttrs != nil {
				h.testAttrs[slog.SourceKey] = s
			}
		}
	}

	if record.Message != "" {
		if h.testAttrs != nil {
			h.testAttrs[slog.MessageKey] = record.Message
		}
	}

	if m := attrs(record, h.ga); len(m) > 0 {
		if h.testAttrs != nil {
			maps.Copy(h.testAttrs, m)
		}
	}

	return h.opts.EmbeddedHandler.Handle(ctx, record)
}

// WithAttrs implements [slog.Handler].
func (h *embeddableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(attrs) == 0 {
		return h
	}

	h.opts.EmbeddedHandler = h.opts.EmbeddedHandler.WithAttrs(attrs)

	return &embeddableHandler{
		opts:      h.opts,
		ga:        append(slices.Clone(h.ga), groupOrAttrs{attrs: attrs}),
		testAttrs: h.testAttrs,
	}
}

// WithGroup implements [slog.Handler].
func (h *embeddableHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	if name == "" {
		return h
	}

	h.opts.EmbeddedHandler = h.opts.EmbeddedHandler.WithGroup(name)

	return &embeddableHandler{
		opts:      h.opts,
		ga:        append(slices.Clone(h.ga), groupOrAttrs{group: name}),
		testAttrs: h.testAttrs,
	}
}
