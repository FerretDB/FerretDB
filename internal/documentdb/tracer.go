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

package documentdb

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// port tracing, tweak logging
// See:
//   - https://pkg.go.dev/github.com/jackc/pgx/v5/tracelog
//   - https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#AcquireTracer
//   - https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#ReleaseTracer
//   - https://pkg.go.dev/github.com/jackc/pgx/v5#hdr-Tracing_and_Logging
//   - https://pkg.go.dev/github.com/jackc/pgx/v5/multitracer
//
// TODO https://github.com/FerretDB/FerretDB/issues/3554
type tracer struct {
	tl *tracelog.TraceLog
}

func newTracer(l *slog.Logger) *tracer {
	return &tracer{
		// try to log everything; logger's configuration will skip extra levels if needed
		tl: &tracelog.TraceLog{
			Logger:   logging.NewPgxLogger(l),
			LogLevel: tracelog.LogLevelTrace,
			Config: &tracelog.TraceLogConfig{
				TimeKey: slog.TimeKey,
			},
		},
	}
}

// TraceAcquireStart implements [pgxpool.AcquireTracer].
//
// It is called at the beginning of [pgxpool.Pool.Acquire].
// The returned context is used for the rest of the call and will be passed to the [tracer.TraceAcquireEnd].
func (t *tracer) TraceAcquireStart(ctx context.Context, pool *pgxpool.Pool, data pgxpool.TraceAcquireStartData) context.Context {
	return t.tl.TraceAcquireStart(ctx, pool, data)
}

// TraceAcquireEnd implements [pgxpool.AcquireTracer].
//
// It is called when a connection has been acquired.
func (t *tracer) TraceAcquireEnd(ctx context.Context, pool *pgxpool.Pool, data pgxpool.TraceAcquireEndData) {
	t.tl.TraceAcquireEnd(ctx, pool, data)
}

// TraceRelease implements [pgxpool.ReleaseTracer].
//
// It is called at the beginning of [pgxpool.Conn.Release].
func (t *tracer) TraceRelease(pool *pgxpool.Pool, data pgxpool.TraceReleaseData) {
	t.tl.TraceRelease(pool, data)
}

// TraceConnectStart implements [pgx.ConnectTracer].
//
// It is called at the beginning of [pgx.Connect] and [pgx.ConnectConfig] calls.
// The returned context is used for the rest of the call and will be passed to [tracer.TraceConnectEnd].
func (t *tracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	return t.tl.TraceConnectStart(ctx, data)
}

// TraceConnectEnd implements [pgx.ConnectTracer].
func (t *tracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	t.tl.TraceConnectEnd(ctx, data)
}

// TracePrepareStart implements [pgx.PrepareTracer].
//
// It is called at the beginning of [pgx.Conn.Prepare] calls.
// The returned context is used for the rest of the call and will be passed to [tracer.TracePrepareEnd].
func (t *tracer) TracePrepareStart(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	return t.tl.TracePrepareStart(ctx, conn, data)
}

// TracePrepareEnd implements [pgx.PrepareTracer].
func (t *tracer) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData) {
	t.tl.TracePrepareEnd(ctx, conn, data)
}

// TraceQueryStart implements [pgx.QueryTracer].
//
// It is called at the beginning of [pgx.Conn.Query], [pgx.Conn.QueryRow], and [pgx.Conn.Exec] calls.
// The returned context is used for the rest of the call and will be passed to [tracer.TraceQueryEnd].
func (t *tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return t.tl.TraceQueryStart(ctx, conn, data)
}

// TraceQueryEnd implements [pgx.QueryTracer].
func (t *tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	t.tl.TraceQueryEnd(ctx, conn, data)
}

// check interfaces
var (
	_ pgxpool.AcquireTracer = (*tracer)(nil)
	_ pgxpool.ReleaseTracer = (*tracer)(nil)
	_ pgx.ConnectTracer     = (*tracer)(nil)
	_ pgx.PrepareTracer     = (*tracer)(nil)
	_ pgx.QueryTracer       = (*tracer)(nil)
)
