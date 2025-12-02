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
	"net/url"
	"sync"

	"github.com/AlekSi/lazyerrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb/cursor"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/resource"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "pool"
)

// Pool represent a pool of PostgreSQL connections.
type Pool struct {
	p     map[string]*pgxpool.Pool
	r     *cursor.Registry
	l     *slog.Logger
	token *resource.Token
	uri   *url.URL
	sp    *state.Provider
	rw    sync.RWMutex
}

// NewPool creates a new pool of PostgreSQL connections.
// No actual connections are established.
func NewPool(uri string, l *slog.Logger, sp *state.Provider) (*Pool, error) {
	must.NotBeZero(uri)
	must.NotBeZero(l)
	must.NotBeZero(sp)

	u, err := url.Parse(uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	p, err := newPgxPool(uri, logging.WithName(l, "pgx"), sp)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := &Pool{
		p:     map[string]*pgxpool.Pool{uri: p},
		r:     cursor.NewRegistry(logging.WithName(l, "cursors")),
		l:     l,
		token: resource.NewToken(),
		uri:   u,
		sp:    sp,
	}
	resource.Track(res, res.token)

	return res, nil
}

// Close closes all connections in the pool.
func (p *Pool) Close() {
	p.r.Close(todoCtx)

	p.rw.Lock()
	defer p.rw.Unlock()

	for _, pool := range p.p {
		pool.Close()
	}

	p.p = nil

	resource.Untrack(p, p.token)
}

// Acquire acquires a connection from the pool.
//
// It is caller's responsibility to call [Conn.Release].
// Most callers should use [Pool.WithConn] instead.
func (p *Pool) Acquire(ctx context.Context) (*Conn, error) {
	connInfo := conninfo.Get(ctx)

	uri := p.uri
	if connInfo.Conv() != nil && connInfo.Conv().Succeed() {
		uri.User = url.UserPassword(connInfo.Conv().Username(), "")
	}

	p.rw.Lock()
	defer p.rw.Unlock()

	pool, ok := p.p[uri.String()]
	if !ok {
		var err error
		if pool, err = newPgxPool(uri.String(), logging.WithName(p.l, "pgx"), p.sp); err != nil {
			return nil, lazyerrors.Error(err)
		}

		p.p[uri.String()] = pool
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return newConn(conn), nil
}

// WithConn acquires a connection from the pool and calls the provided function with it.
// The connection is automatically released after the function returns.
func (p *Pool) WithConn(ctx context.Context, f func(*pgx.Conn) error) error {
	conn, err := p.Acquire(ctx)
	if err != nil {
		return lazyerrors.Error(err)
	}

	defer conn.Release()

	if err = f(conn.Conn()); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// Describe implements [prometheus.Collector].
func (p *Pool) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(p, ch)
}

// Collect implements [prometheus.Collector].
func (p *Pool) Collect(ch chan<- prometheus.Metric) {
	p.r.Collect(ch)
	p.tracer.Collect(ch)

	var poolsStats []*pgxpool.Stat

	p.rw.Lock()

	for _, pool := range p.p {
		poolsStats = append(poolsStats, pool.Stat())
	}

	p.rw.Unlock()

	var acquireDuration, emptyAcquireWaitTime float64
	var acquireConns, constructingConns, idleConns, maxConns, totalConns int32
	var acquireCount, canceledAcquireCount, emptyAcquireCount, newConnsCount, maxLifetimeDestroyCount, maxIdleDestroyCount int64

	for _, stat := range poolsStats {
		acquireCount += stat.AcquireCount()
		acquireDuration += stat.AcquireDuration().Seconds()
		acquireConns += stat.AcquiredConns()
		canceledAcquireCount += stat.CanceledAcquireCount()
		constructingConns += stat.ConstructingConns()
		emptyAcquireCount += stat.EmptyAcquireCount()
		idleConns += stat.IdleConns()
		maxConns += stat.MaxConns()
		totalConns += stat.TotalConns()
		newConnsCount += stat.NewConnsCount()
		maxLifetimeDestroyCount += stat.MaxLifetimeDestroyCount()
		maxIdleDestroyCount += stat.MaxIdleDestroyCount()
		emptyAcquireWaitTime += stat.EmptyAcquireWaitTime().Seconds()
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_total"),
			"The cumulative count of successful connection acquires from the pool.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(acquireCount),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_duration_seconds_total"),
			"The total duration of all successful connection acquires from the pool.",
			nil, nil,
		),
		prometheus.CounterValue,
		acquireDuration,
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquired"),
			"The number of currently acquired connections in the pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(acquireConns),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_canceled_total"),
			"The cumulative count of connection acquires from the pool that were canceled.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(canceledAcquireCount),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "constructing"),
			"The number of connections with construction in progress in the pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(constructingConns),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_empty_total"),
			"The cumulative count of successful connection acquires from the pool "+
				"that waited for a resource to be released or constructed because the pool was empty.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(emptyAcquireCount),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "idle"),
			"The number of currently idle connections in the pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(idleConns),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "max_size"),
			"The maximum size of the connection pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(maxConns),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "size"),
			"Total number of connections currently in the pool. "+
				"Should be a sum of constructing, acquired, and idle.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(totalConns),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "opened_total"),
			"The cumulative count of new connections opened.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(newConnsCount),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "destroyed_maxlifetime_total"),
			"The cumulative count of connections destroyed because they exceeded pool_max_conn_lifetime.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(maxLifetimeDestroyCount),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "destroyed_maxidle_total"),
			"The cumulative count of connections destroyed because they exceeded pool_max_conn_idle_time.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(maxIdleDestroyCount),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_empty_duration_seconds_total"),
			"The cumulative time waited for successful acquires from the pool "+
				"for a resource to be released or constructed because the pool was empty.",
			nil, nil,
		),
		prometheus.CounterValue,
		emptyAcquireWaitTime,
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*Pool)(nil)
)
