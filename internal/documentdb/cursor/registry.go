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
	"slices"
	"sync"
	"time"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
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
	rw          sync.RWMutex
	cursors     map[int64]*cursor
	invalidated map[int64]time.Time

	l     *slog.Logger
	token *resource.Token

	created  *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// NewRegistry creates a new cursor registry.
func NewRegistry(l *slog.Logger) *Registry {
	res := &Registry{
		cursors:     map[int64]*cursor{},
		invalidated: map[int64]time.Time{},
		l:           l,
		token:       resource.NewToken(),

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

	res.created.With(prometheus.Labels{"type": "normal"})
	res.duration.With(prometheus.Labels{"type": "normal"})

	resource.Track(res, res.token)

	return res
}

// Close closes all cursors in the registry.
func (r *Registry) Close(ctx context.Context) {
	r.rw.Lock()
	defer r.rw.Unlock()

	for id := range r.cursors {
		c := r.removeCursor(ctx, id)
		must.NotBeZero(c)
		c.close(ctx)
	}

	r.cursors = nil
	r.invalidated = nil
	resource.Untrack(r, r.token)
}

// NewCursor stores a cursor with given continuation and connection (if any).
//
// Passed context is used for logging/tracing, and for closing existing cursor, if any.
// See [Registry.CloseCursor].
func (r *Registry) NewCursor(ctx context.Context, id int64, continuation wirebson.RawDocument, conn *pgx.Conn, database, collection string, release ...func()) {
	must.NotBeZero(id)
	must.BeTrue(len(continuation) > 0)

	r.rw.Lock()

	c := newCursor(continuation, conn)
	c.database, c.collection = database, collection
	if len(release) > 0 {
		c.release = release[0]
	}

	existing := r.cursors[id]
	if existing != nil {
		r.l.WarnContext(
			ctx, "Replacing existing cursor",
			slog.Int64("id", id), slog.Any("existing", existing), slog.Any("cursor", c),
		)
		r.removeCursor(ctx, id)
	}

	r.l.DebugContext(ctx, "Storing new cursor", slog.Int64("id", id), slog.Any("cursor", c))

	r.created.With(prometheus.Labels{"type": c.Type()}).Inc()

	r.cursors[id] = c
	delete(r.invalidated, id)

	r.rw.Unlock()

	if existing != nil {
		existing.close(ctx)
	}
}

// SetGeneration binds a cursor to the catalog identity observed before its query.
func (r *Registry) SetGeneration(id int64, generation string) {
	r.rw.Lock()
	defer r.rw.Unlock()
	if c := r.cursors[id]; c != nil {
		c.generation, c.tracked = generation, true
	}
}

// Generation returns the cursor namespace and its original catalog identity.
func (r *Registry) Generation(id int64) (database, collection, generation string, tracked bool) {
	r.rw.RLock()
	defer r.rw.RUnlock()
	if c := r.cursors[id]; c != nil {
		return c.database, c.collection, c.generation, c.tracked
	}
	return "", "", "", false
}

// NamespaceGeneration identifies an additional collection read by a pipeline.
type NamespaceGeneration struct {
	Database   string
	Collection string
	UUID       string
	ErrorCode  int32
}

// SetDependencies retains an independent snapshot of a cursor's read dependencies.
func (r *Registry) SetDependencies(id int64, dependencies []NamespaceGeneration) {
	r.rw.Lock()
	defer r.rw.Unlock()
	if c := r.cursors[id]; c != nil {
		c.dependencies = slices.Clone(dependencies)
	}
}

// Dependencies returns an independent snapshot, empty after replacement or removal.
func (r *Registry) Dependencies(id int64) []NamespaceGeneration {
	r.rw.RLock()
	defer r.rw.RUnlock()
	if c := r.cursors[id]; c != nil {
		return slices.Clone(c.dependencies)
	}
	return nil
}

// CloseNamespace invalidates cursors for a renamed namespace and its metadata.
func (r *Registry) CloseNamespace(ctx context.Context, database, collection string) {
	r.rw.Lock()
	for id, when := range r.invalidated {
		if time.Since(when) > 10*time.Minute {
			delete(r.invalidated, id)
		}
	}
	var closing []*cursor
	for id, c := range r.cursors {
		if c.database == database && (c.collection == collection || c.collection == "") {
			closing = append(closing, r.removeCursor(ctx, id))
			r.invalidated[id] = time.Now()
		}
	}
	r.rw.Unlock()
	for _, c := range closing {
		c.close(ctx)
	}
}

// ConsumeInvalidation preserves the reason for the first getMore after rename.
// Subsequent attempts use CursorNotFound, as on the reference server.
func (r *Registry) ConsumeInvalidation(id int64) bool {
	r.rw.Lock()
	defer r.rw.Unlock()
	when, ok := r.invalidated[id]
	delete(r.invalidated, id)
	return ok && time.Since(when) <= 10*time.Minute
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

// UpdateCursor updates existing cursor with given continuation,
// or closes it if continuation is empty.
//
// Passed context is used for logging/tracing,
// and for closing the cursor (see [Registry.CloseCursor]).
func (r *Registry) UpdateCursor(ctx context.Context, id int64, continuation wirebson.RawDocument) {
	r.rw.Lock()

	c := r.cursors[id]
	if c == nil {
		r.rw.Unlock()

		r.l.WarnContext(ctx, "Cursor not found", slog.Int64("id", id))
		return
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/5445
	if len(continuation) == 0 {
		r.l.DebugContext(
			ctx, "Closing instead of updating cursor with empty continuation",
			slog.Int64("id", id), slog.Any("cursor", c),
		)
		r.removeCursor(ctx, id)

		r.rw.Unlock()

		c.close(ctx)
		return
	}

	r.l.DebugContext(
		ctx, "Updating cursor",
		slog.Int64("id", id), slog.Any("cursor", c), slog.Any("continuation", logging.LazyDeepDecoder(continuation)),
	)
	c.continuation = continuation

	r.rw.Unlock()
}

// CloseCursor removes the cursor with the given id from the registry and closes it, if any.
// It returns true if the cursor was found and removed.
//
// It attempts a clean close by sending the exit message to PostgreSQL.
// However, this could block so ctx is available to limit the time to wait (up to 3 seconds).
// The underlying connection will always be called regardless of any other errors.
func (r *Registry) CloseCursor(ctx context.Context, id int64) bool {
	r.rw.Lock()

	c := r.removeCursor(ctx, id)

	r.rw.Unlock()

	if c == nil {
		return false
	}

	c.close(ctx)

	return true
}

// removeCursor removes the cursor with the given id from the registry and returns it, if any.
// The caller is responsible for closing it.
// Registry's rw also should be held by the caller.
func (r *Registry) removeCursor(ctx context.Context, id int64) *cursor {
	c := r.cursors[id]
	if c == nil {
		r.l.WarnContext(ctx, "Cursor not found", slog.Int64("id", id))
		return nil
	}

	dur := time.Since(c.created)

	r.l.DebugContext(
		ctx, "Removing cursor",
		slog.Int64("id", id), slog.Any("cursor", c), slog.Duration("duration", dur),
	)

	r.duration.With(prometheus.Labels{"type": c.Type()}).Observe(dur.Seconds())

	delete(r.cursors, id)

	return c
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
