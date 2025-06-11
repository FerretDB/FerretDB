package logging

import (
	"context"
	"log/slog"
)

type embeddableHandler struct {
	opts *NewHandlerOpts
}

func newEmbeddableHandler(opts *NewHandlerOpts) *embeddableHandler {
	return &embeddableHandler{opts: opts}
}

func (h *embeddableHandler) Enabled(ctx context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}

	return l >= minLevel
}

func (h *embeddableHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.opts.Handler.Handle(ctx, r)
}

func (h *embeddableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h.opts.Handler = h.opts.Handler.WithAttrs(attrs)
	return h
}

func (h *embeddableHandler) WithGroup(name string) slog.Handler {
	h.opts.Handler = h.opts.Handler.WithGroup(name)
	return h
}
