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

package cursor

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/resource"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "cursors"
)

// Registry provides access to DocumentDB cursors.
//
//nolint:vet // for readability
type Registry struct {
	rw      sync.RWMutex
	cursors map[int64]*cursor

	l     *slog.Logger
	token *resource.Token

	created  *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// NewRegistry creates a new cursor registry.
func NewRegistry(l *slog.Logger) *Registry {
	res := &Registry{
		cursors: map[int64]*cursor{},
		l:       l,
		token:   resource.NewToken(),

		created: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "created_total",
				Help:      "Total number of cursors created.",
			},
			[]string{"type"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "duration_seconds",
				Help:      "Cursors lifetime in seconds.",
				Buckets: []float64{
					(1 * time.Millisecond).Seconds(),
					(5 * time.Millisecond).Seconds(),
					(10 * time.Millisecond).Seconds(),
					(25 * time.Millisecond).Seconds(),
					(50 * time.Millisecond).Seconds(),
					(100 * time.Millisecond).Seconds(),
					(250 * time.Millisecond).Seconds(),
					(500 * time.Millisecond).Seconds(),
					(1000 * time.Millisecond).Seconds(),
					(2500 * time.Millisecond).Seconds(),
					(5000 * time.Millisecond).Seconds(),
					(10000 * time.Millisecond).Seconds(),
				},
			},
			[]string{"type"},
		),
	}

	res.created.WithLabelValues("normal")
	res.duration.WithLabelValues("normal")

	resource.Track(res, res.token)

	return res
}

// Close closes all cursors in the registry.
func (r *Registry) Close(ctx context.Context) {
	r.rw.Lock()
	defer r.rw.Unlock()

	for id := range r.cursors {
		r.closeCursor(ctx, id)
	}

	r.cursors = nil

	resource.Untrack(r, r.token)
}

// NewCursor stores a cursor with given continuation and connection (if any).
//
// As a special case, if continuation is empty, this method does nothing.
// That simplifies the typical usage.
func (r *Registry) NewCursor(id int64, continuation wirebson.RawDocument, conn *pgx.Conn) {
	// to have better logging for now
	var cont *wirebson.Document
	if len(continuation) > 0 {
		cont = must.NotFail(continuation.Decode())
	}

	persist := conn != nil

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/270
	if len(continuation) == 0 {
		if persist {
			r.l.Warn(
				"Not persisting connection with empty continuation",
				slog.Int64("id", id), slog.Any("continuation", cont), slog.Bool("persist", persist),
			)

			_ = conn.Close(context.TODO())
		}

		if id != 0 {
			r.l.Warn(
				"Not storing cursor with empty continuation",
				slog.Int64("id", id), slog.Any("continuation", cont), slog.Bool("persist", persist),
			)
		}

		return
	}

	must.NotBeZero(id)

	r.rw.Lock()
	defer r.rw.Unlock()

	if _, ok := r.cursors[id]; ok {
		r.l.Error("Replacing existing cursor", slog.Int64("id", id))
		r.closeCursor(context.TODO(), id)
	}

	r.l.Debug("Creating new cursor",
		slog.Int64("id", id), slog.Any("continuation", cont), slog.Bool("persist", persist),
	)

	r.cursors[id] = newCursor(continuation, conn)

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/97
	t := "normal"
	if persist {
		t = "persist"
	}

	r.created.WithLabelValues(t).Inc()
}

// GetCursor returns the continuation and the connection for the given cursor id.
func (r *Registry) GetCursor(id int64) (wirebson.RawDocument, *pgx.Conn) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	if c := r.cursors[id]; c != nil {
		return c.continuation, c.conn
	}

	return nil, nil
}

// UpdateCursor updates existing cursor with given continuation.
func (r *Registry) UpdateCursor(id int64, continuation wirebson.RawDocument) {
	// to have better logging for now
	var cont *wirebson.Document
	if len(continuation) > 0 {
		cont = must.NotFail(continuation.Decode())
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	c := r.cursors[id]
	if c == nil {
		r.l.Warn("Cursor not found", slog.Int64("id", id))
		return
	}

	persist := c.conn != nil

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/270
	if len(continuation) == 0 {
		r.l.Warn(
			"Closing instead of updating cursor with empty continuation",
			slog.Int64("id", id), slog.Any("continuation", cont), slog.Bool("persist", persist),
		)

		r.closeCursor(context.TODO(), id)
		return
	}

	r.l.Debug(
		"Updating cursor",
		slog.Int64("id", id), slog.Any("continuation", cont), slog.Bool("persist", persist),
	)
	c.continuation = continuation
}

// CloseCursor closes the cursor with the given id and removes it from the registry.
// It returns true if the cursor was found and removed.
//
// It attempts a clean close by sending the exit message to PostgreSQL.
// However, this could block so ctx is available to limit the time to wait (up to 3 seconds).
// The underlying connection will always be called regardless of any other errors.
func (r *Registry) CloseCursor(ctx context.Context, id int64) bool {
	r.rw.Lock()
	defer r.rw.Unlock()

	return r.closeCursor(ctx, id)
}

// closeCursor is a private function that is wrapped by CloseCursor.
// It doesn't block RWMutex, hence it should be used only if necessary.
func (r *Registry) closeCursor(ctx context.Context, id int64) bool {
	c := r.cursors[id]
	if c == nil {
		r.l.WarnContext(ctx, "Cursor not found", slog.Int64("id", id))
		return false
	}

	dur := time.Since(c.created)
	persist := c.conn != nil

	r.l.DebugContext(
		ctx, "Closing and removing cursor",
		slog.Int64("id", id), slog.Bool("persist", persist), slog.Duration("duration", dur),
	)
	c.close(ctx)
	delete(r.cursors, id)

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/97
	t := "normal"
	if persist {
		t = "persist"
	}

	r.duration.WithLabelValues(t).Observe(dur.Seconds())

	return true
}

// Describe implements [prometheus.Collector].
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	r.created.Describe(ch)
	r.duration.Describe(ch)
}

// Collect implements [prometheus.Collector].
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	r.created.Collect(ch)
	r.duration.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*Registry)(nil)
)
