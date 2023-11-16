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

package debug

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
)

// gatherer wraps another Prometheus Gatherer with a one-second cache and error handling.
type gatherer struct {
	g prometheus.Gatherer
	l *zap.Logger

	rw  sync.RWMutex
	t   time.Time
	mfs []*dto.MetricFamily
	m   map[string]*dto.MetricFamily
}

// newGatherer returns a new gatherer.
func newGatherer(g prometheus.Gatherer, l *zap.Logger) *gatherer {
	return &gatherer{
		g: g,
		l: l,
	}
}

// Gather implements prometheus.Gatherer.
//
// It never returns an error.
func (g *gatherer) Gather() ([]*dto.MetricFamily, error) {
	// fast path

	g.rw.RLock()

	if time.Since(g.t) < time.Second {
		mfs := g.mfs
		g.rw.RUnlock()
		return mfs, nil
	}

	g.rw.RUnlock()

	// slow path

	g.rw.Lock()
	defer g.rw.Unlock()

	// a concurrent call might have updated metrics already
	if time.Since(g.t) < time.Second {
		return g.mfs, nil
	}

	g.l.Debug("Gathering Prometheus metrics")
	mfs, err := g.g.Gather()
	if err != nil {
		g.l.Warn("Failed to gather Prometheus metrics", zap.Error(err), zap.Int("metrics", len(mfs)))
		g.mfs, g.m, g.t = nil, nil, time.Now()
		return nil, nil
	}

	g.l.Debug("Gathered Prometheus metrics", zap.Int("metrics", len(mfs)))
	g.mfs, g.m, g.t = mfs, nil, time.Now()
	return mfs, nil
}

// GatherMap returns Prometheus metric families as a map.
func (g *gatherer) GatherMap() map[string]*dto.MetricFamily {
	// fast path

	g.rw.RLock()

	if g.m != nil && time.Since(g.t) < time.Second {
		m := g.m
		g.rw.RUnlock()
		return m
	}

	g.rw.RUnlock()

	// slow path

	g.Gather()

	g.rw.Lock()
	defer g.rw.Unlock()

	// a concurrent call might have updated metrics already
	if g.m != nil && time.Since(g.t) < time.Second {
		return g.m
	}

	g.m = make(map[string]*dto.MetricFamily, len(g.mfs))
	for _, mf := range g.mfs {
		if g.m[*mf.Name] != nil {
			g.l.Warn("Duplicate Prometheus metric name", zap.String("name", *mf.Name))
			continue
		}

		g.m[*mf.Name] = mf
	}

	return g.m
}

// check interfaces
var (
	_ prometheus.Gatherer = (*gatherer)(nil)
)
