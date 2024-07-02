// Package slogt implements a bridge between stdlib testing pkg and
// the slog logging library. Use slogt.New(t) to get a *slog.Logger.
package slogt

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
)

var _ slog.Handler = (*Bridge)(nil)

// Default sets the default handler. This
// can be changed by the client.
var Default = Text()

// Option is a functional option type that is used
// with New to configure the logger's underlying handler.
type Option func(b *Bridge)

// Text specifies a text handler.
//
//	log := slogt.New(t, slogt.Text())
func Text() Option {
	return func(b *Bridge) {
		hOpts := &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelDebug,
		}
		// The opts may have already set the handler.
		b.Handler = slog.NewTextHandler(b.buf, hOpts)
	}
}

// JSON specifies a JSON handler.
//
//	log := slogt.New(t, slogt.JSON())
func JSON() Option {
	return func(b *Bridge) {
		hOpts := &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelDebug,
		}
		// The opts may have already set the handler.
		b.Handler = slog.NewJSONHandler(b.buf, hOpts)
	}
}

// Factory is specifies a custom factory function for
// creating the logger's underlying handler.
func Factory(fn func(w io.Writer) slog.Handler) Option {
	return func(b *Bridge) {
		b.Handler = fn(b.buf)
	}
}

// New returns a new *slog.Logger whose logging methods
// pipe output to t.Log.
func New(t testing.TB, opts ...Option) *slog.Logger {
	h := &Bridge{
		t:   t,
		buf: &bytes.Buffer{},
		mu:  &sync.Mutex{},
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.Handler == nil {
		// No handler set yet, use the default handler.
		Default(h)
	}

	return slog.New(h)
}

// Bridge is an implementation of slog.Handler that works
// with the stdlib testing pkg.
type Bridge struct {
	slog.Handler
	t   testing.TB
	buf *bytes.Buffer
	mu  *sync.Mutex
}

// Handle implements slog.Handler.
func (b *Bridge) Handle(ctx context.Context, rec slog.Record) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	err := b.Handler.Handle(ctx, rec)
	if err != nil {
		return err
	}

	output, err := io.ReadAll(b.buf)
	if err != nil {
		return err
	}

	// The output comes back with a newline, which we need to
	// trim before feeding to t.Log.
	output = bytes.TrimSuffix(output, []byte("\n"))

	// Add calldepth. But it won't be enough, and the internal slog
	// callsite will be printed. See discussion in README.md.
	b.t.Helper()

	b.t.Log(string(output))

	return nil
}

// WithAttrs implements slog.Handler.
func (b *Bridge) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Bridge{
		t:       b.t,
		buf:     b.buf,
		mu:      b.mu,
		Handler: b.Handler.WithAttrs(attrs),
	}
}

// WithGroup implements slog.Handler.
func (b *Bridge) WithGroup(name string) slog.Handler {
	return &Bridge{
		t:       b.t,
		buf:     b.buf,
		mu:      b.mu,
		Handler: b.Handler.WithGroup(name),
	}
}
