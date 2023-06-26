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
// TODO better cleanup (?), more metrics https://github.com/FerretDB/FerretDB/issues/2862
//
//nolint:vet // for readability
type Registry struct {
	rw sync.RWMutex
	m  map[int64]*Cursor

	l  *zap.Logger
	wg sync.WaitGroup
}

// NewRegistry creates a new Registry.
func NewRegistry(l *zap.Logger) *Registry {
	return &Registry{
		m: map[int64]*Cursor{},
		l: l,
	}
}

// Close waits for all cursors to be closed.
func (r *Registry) Close() {
	// we mainly do that for tests; see https://github.com/uber-go/zap/issues/687

	r.wg.Wait()
}

// NewParams represent parameters for NewCursor.
type NewParams struct {
	Iter       types.DocumentsIterator
	DB         string
	Collection string
	Username   string
}

// NewCursor creates and stores a new cursor.
//
// The cursor will be closed automatically when a given context is canceled,
// even if the cursor is not being used at that time.
func (r *Registry) NewCursor(ctx context.Context, params *NewParams) *Cursor {
	r.rw.Lock()
	defer r.rw.Unlock()

	// use global, sequential, positive, short cursor IDs to make debugging easier
	var id int64
	for id == 0 || r.m[id] != nil {
		id = int64(lastCursorID.Add(1))
	}

	r.l.Debug(
		"Creating",
		zap.Int64("id", id),
		zap.String("db", params.DB),
		zap.String("collection", params.Collection),
	)

	c := newCursor(id, params.DB, params.Collection, params.Username, params.Iter, r)
	r.m[id] = c

	r.wg.Add(1)

	go func() {
		defer r.wg.Done()

		select {
		case <-ctx.Done():
			c.Close()
		case <-c.closed:
		}
	}()

	return c
}

// Get returns stored cursor by ID, or nil.
func (r *Registry) Get(id int64) *Cursor {
	r.rw.RLock()
	defer r.rw.RUnlock()

	return r.m[id]
}

func (r *Registry) All() []*Cursor {
	r.rw.RLock()
	defer r.rw.RUnlock()

	return maps.Values(r.m)
}

// This method should be called only from cursor.Close().
func (r *Registry) delete(id int64) {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.l.Debug(
		"Deleting",
		zap.Int("total", len(r.m)),
		zap.Int64("id", id),
	)

	delete(r.m, id)
}

// Describe implements prometheus.Collector.
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(r, ch)
}

// Collect implements prometheus.Collector.
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	r.rw.RLock()

	current := len(r.m)

	r.rw.RUnlock()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "current"), "The current number of cursors.", nil, nil),
		prometheus.GaugeValue,
		float64(current),
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*Registry)(nil)
)
