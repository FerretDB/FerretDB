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
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "cursors"
)

// Global last cursor ID.
var lastCursorID atomic.Uint32

func init() {
	// to make debugging easier
	if !debugbuild.Enabled {
		lastCursorID.Store(rand.Uint32())
	}
}

// Registry stores cursors.
//
//nolint:vet // for readability
type Registry struct {
	rw sync.RWMutex
	m  map[int64]*Cursor

	l  *zap.Logger
	wg sync.WaitGroup

	created  *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// NewRegistry creates a new Registry.
func NewRegistry(l *zap.Logger) *Registry {
	return &Registry{
		m: map[int64]*Cursor{},
		l: l,
		created: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "created_total",
				Help:      "Total number of cursors created.",
			},
			[]string{"type", "db", "collection", "username"},
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
			[]string{"type", "db", "collection", "username"},
		),
	}
}

// Close waits for all cursors to be closed and removed from the registry.
func (r *Registry) Close() {
	// we mainly do that for tests; see https://github.com/uber-go/zap/issues/687

	r.wg.Wait()
}

// NewParams represent parameters for NewCursor.
type NewParams struct {
	// Data stored, but not used by this package.
	// Used to pass *handler.findCursorData between `find` and `getMore` command implementations.
	// Stored as any to avoid dependency cycle.
	Data any

	// those fields are used for limited authorization checks
	// before we implement proper authz and/or sessions
	DB         string
	Collection string
	Username   string

	Type         Type
	ShowRecordID bool

	_ struct{} // prevent unkeyed literals
}

// NewCursor creates and stores a new cursor.
//
// The cursor of any type will be closed automatically when a given context is canceled,
// even if the cursor is not being used at that time.
func (r *Registry) NewCursor(ctx context.Context, iter types.DocumentsIterator, params *NewParams) *Cursor {
	r.rw.Lock()
	defer r.rw.Unlock()

	// use global, sequential, positive, short cursor IDs to make debugging easier
	var id int64
	for id == 0 || r.m[id] != nil {
		id = int64(lastCursorID.Add(1))
	}

	r.l.Debug(
		"Creating cursor",
		zap.Int64("id", id), zap.Stringer("type", params.Type),
		zap.String("db", params.DB), zap.String("collection", params.Collection), zap.String("username", params.Username),
	)

	r.created.WithLabelValues(params.Type.String(), params.DB, params.Collection, params.Username).Inc()

	c := newCursor(id, iter, params, r)
	r.m[id] = c

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		select {
		case <-ctx.Done():
			r.CloseAndRemove(c)
		case <-c.removed: // for c.Close() and normal cursors
		}

		<-c.removed
	}()

	return c
}

// Get returns stored cursor by ID, or nil.
func (r *Registry) Get(id int64) *Cursor {
	r.rw.RLock()
	defer r.rw.RUnlock()

	return r.m[id]
}

// All returns a shallow copy of all stored cursors.
func (r *Registry) All() []*Cursor {
	r.rw.RLock()
	defer r.rw.RUnlock()

	return maps.Values(r.m)
}

// CloseAndRemove closes the given cursors, then removes it from the registry.
func (r *Registry) CloseAndRemove(c *Cursor) {
	c.Close()

	r.rw.Lock()
	defer r.rw.Unlock()

	if r.m[c.ID] == nil {
		return
	}

	d := time.Since(c.created)
	r.l.Debug(
		"Removing cursor",
		zap.Int64("id", c.ID),
		zap.Stringer("type", c.Type),
		zap.Int("total", len(r.m)),
		zap.Duration("duration", d),
	)

	r.duration.WithLabelValues(c.Type.String(), c.DB, c.Collection, c.Username).Observe(d.Seconds())

	delete(r.m, c.ID)
	close(c.removed)
}

// Describe implements prometheus.Collector.
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	r.created.Describe(ch)
	r.duration.Describe(ch)
}

// Collect implements prometheus.Collector.
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	r.created.Collect(ch)
	r.duration.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*Registry)(nil)
)
