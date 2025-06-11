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
	"sync"
)

// Embeddable handler is a [slog.Handler] for custom logger.
// It is used to make sure that the config.LogLevel is effective.
//
//nolint:vet // for readability
type embeddableHandler struct {
	opts *NewHandlerOpts
	m    *sync.Mutex
}

// newEmbeddableHandler creates a new embeddable handler.
func newEmbeddableHandler(opts *NewHandlerOpts) *embeddableHandler {
	return &embeddableHandler{opts: opts, m: &sync.Mutex{}}
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
func (h *embeddableHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.opts.Handler.Handle(ctx, r)
}

// WithAttrs implements [slog.Handler].
func (h *embeddableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h.m.Lock()
	defer h.m.Unlock()

	h.opts.Handler = h.opts.Handler.WithAttrs(attrs)

	return &embeddableHandler{opts: h.opts, m: h.m}
}

// WithGroup implements [slog.Handler].
func (h *embeddableHandler) WithGroup(name string) slog.Handler {
	h.m.Lock()
	defer h.m.Unlock()

	h.opts.Handler = h.opts.Handler.WithGroup(name)

	return &embeddableHandler{opts: h.opts, m: h.m}
}
