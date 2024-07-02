package driver

import (
	"context"
	"database/sql/driver"
	"flag"
	"fmt"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	protTrace atomic.Bool
	sqlTrace  atomic.Bool
)

func setTrace(b *atomic.Bool, s string) error {
	v, err := strconv.ParseBool(s)
	if err == nil {
		b.Store(v)
	}
	return err
}

func init() {
	flag.BoolFunc("hdb.protTrace", "enabling hdb protocol trace", func(s string) error { return setTrace(&protTrace, s) })
	flag.BoolFunc("hdb.sqlTrace", "enabling hdb sql trace", func(s string) error { return setTrace(&sqlTrace, s) })
}

// SQLTrace returns true if sql tracing output is active, false otherwise.
func SQLTrace() bool { return sqlTrace.Load() }

// SetSQLTrace sets sql tracing output active or inactive.
func SetSQLTrace(on bool) { sqlTrace.Store(on) }

const (
	tracePing    = "ping"
	tracePrepare = "prepare"
	traceQuery   = "query"
	traceExec    = "exec"
)

type sqlTracer struct {
	logger    *slog.Logger
	maxArg    int
	startTime time.Time
}

const defSQLTracerMaxArg = 5 // default limit of number of arguments

func newSQLTracer(logger *slog.Logger, maxArg int) *sqlTracer {
	if maxArg <= 0 {
		maxArg = defSQLTracerMaxArg
	}
	return &sqlTracer{logger: logger, maxArg: maxArg}
}

func (t *sqlTracer) begin() { t.startTime = time.Now() }

func (t *sqlTracer) _log(ctx context.Context, logKind string, query string, err error, nvargs []driver.NamedValue) {
	duration := time.Since(t.startTime).Milliseconds()
	l := len(nvargs)

	attrs := []slog.Attr{
		slog.String(logKind, query),
		slog.Int64("ms", duration),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", (err).Error()))
	}

	if l == 0 {
		t.logger.LogAttrs(ctx, slog.LevelInfo, "SQL", attrs...)
		return
	}

	var argAttrs []slog.Attr
	for i := 0; i < min(l, t.maxArg); i++ {
		name := nvargs[i].Name
		if name == "" {
			name = strconv.Itoa(nvargs[i].Ordinal)
		}
		argAttrs = append(argAttrs, slog.String(name, fmt.Sprintf("%v", nvargs[i].Value)))
	}
	if l > t.maxArg {
		argAttrs = append(argAttrs, slog.Int("numArgSkip", l-t.maxArg))
	}
	attrs = append(attrs, slog.Any("arg", slog.GroupValue(argAttrs...)))

	t.logger.LogAttrs(ctx, slog.LevelInfo, "SQL", attrs...)
}

func (t *sqlTracer) log(ctx context.Context, logKind string, query string, err error, nvargs []driver.NamedValue) {
	// split fastpath for go to inline
	if sqlTrace.Load() {
		t._log(ctx, logKind, query, err, nvargs)
	}
}
