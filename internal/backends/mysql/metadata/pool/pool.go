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

// Package pool provides access to MySQL connections.
//
// It should be used only by the metadata package.
package pool

import (
	"net/url"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/resource"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "mysql_pool"
)

// Pool provides access to MySQL connections.
//
//nolint:vet // for readability
type Pool struct {
	baseURI url.URL
	l       *zap.Logger
	sp      *state.Provider

	rw  sync.RWMutex
	dbs map[string]*fsql.DB // by full URI

	token *resource.Token
}

// New creates a new Pool.
func New(u string, l *zap.Logger, sp *state.Provider) (*Pool, error) {
	return nil, lazyerrors.New("not yet implemented")
}

// Close closes all connections in the pool.
func (p *Pool) Close() {
	p.rw.Lock()
	defer p.rw.Unlock()

	for _, db := range p.dbs {
		// we need to properly handle error on failed closure
		_ = db.Close()
	}

	p.dbs = nil

	resource.Untrack(p, p.token)
}

// Get returns a pool of connections to MySQL database for that username/password combination.
func (p *Pool) Get(username, password string) (*fsql.DB, error) {
	return nil, lazyerrors.New("not yet implemented")
}

// Describe implements Prometheus.Collector.
func (p *Pool) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(p, ch)
}

// Collect implements Prometheus.Collector.
func (p *Pool) Collect(ch chan<- prometheus.Metric) {
	p.rw.RLock()
	defer p.rw.RUnlock()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "size"),
			"The current number of pools.",
			nil, nil),
		prometheus.GaugeValue,
		float64(len(p.dbs)),
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*Pool)(nil)
)
