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

	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/resource"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
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

	rw    sync.RWMutex
	pools map[string]*fsql.DB

	token *resource.Token
}

// New creates a new Pool.
func New(u string, l *zap.Logger, sp *state.Provider) (*Pool, error) {
	baseURI, err := url.Parse(u)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	p := &Pool{
		baseURI: *baseURI,
		l:       l,
		sp:      sp,
		pools:   map[string]*fsql.DB{},
		token:   resource.NewToken(),
	}

	resource.Track(p, p.token)

	return p, nil
}

// Close closes all connections in the pool.
func (p *Pool) Close() {
	p.rw.Lock()
	defer p.rw.Unlock()

	for _, pool := range p.pools {
		_ = pool.Close()
	}

	p.pools = nil

	resource.Untrack(p, p.token)
}

func (p *Pool) Get(username, password string) (*fsql.DB, error) {
	uri := p.baseURI
	if username != "" {
		uri.User = url.UserPassword(username, password)
	}

	u := uri.String()

	p.rw.RLock()
	res := p.pools[u]
	p.rw.RUnlock()

	if res != nil {
		p.l.Debug("Pool: found existing pool", zap.String("username", username))
		return res, nil
	}

	p.rw.Lock()
	defer p.rw.Unlock()

	// check again to see if a concurrent connection created a pool already
	if res = p.pools[u]; res != nil {
		p.l.Debug("Pool: found existing pool (after acquiring lock)", zap.String("username", username))
		return res, nil
	}

	res, err := openDB(u, p.l, p.sp)
	if err != nil {
		p.l.Warn("Pool: connection failed", zap.String("username", username), zap.Error(err))
		return nil, lazyerrors.Error(err)
	}

	p.l.Info("Pool: connection succeeded", zap.String("username", username))
	p.pools[u] = res

	return res, nil
}

// Describe implements prometheus.Collector
func (p *Pool) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(p, ch)
}

// Collect implements prometheus.Collector
func (p *Pool) Collect(ch chan<- prometheus.Metric) {
	p.rw.RLock()
	defer p.rw.RUnlock()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "size"),
			"The current number of pools.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(len(p.pools)),
	)

	for _, pool := range p.pools {
		pool.Collect(ch)
	}
}

// check interfaces
var (
	_ prometheus.Collector = (*Pool)(nil)
)
