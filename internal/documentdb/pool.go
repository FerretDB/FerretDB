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
	"log/slog"

	"github.com/AlekSi/lazyerrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"

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
	p      *pgxpool.Pool
	r      *cursor.Registry
	l      *slog.Logger
	tracer *tracer
	token  *resource.Token
}

// NewPool creates a new pool of PostgreSQL connections.
// No actual connections are established.
func NewPool(uri string, l *slog.Logger, sp *state.Provider) (*Pool, error) {
	must.NotBeZero(uri)
	must.NotBeZero(l)
	must.NotBeZero(sp)

	tl := logging.WithName(l, "pgx")
	t := newTracer(tl)

	p, err := newPgxPool(uri, tl, t, sp)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := &Pool{
		p:      p,
		r:      cursor.NewRegistry(logging.WithName(l, "cursors")),
		l:      l,
		tracer: t,
		token:  resource.NewToken(),
	}
	resource.Track(res, res.token)

	return res, nil
}

// Close closes all connections in the pool.
func (p *Pool) Close() {
	p.r.Close(todoCtx)

	p.p.Close()

	resource.Untrack(p, p.token)
}

// Acquire acquires a connection from the pool.
//
// It is caller's responsibility to call [Conn.Release].
// Most callers should use [Pool.WithConn] instead.
func (p *Pool) Acquire() (*Conn, error) {
	conn, err := p.p.Acquire(todoCtx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return newConn(conn), nil
}

// WithConn acquires a connection from the pool and calls the provided function with it.
// The connection is automatically released after the function returns.
func (p *Pool) WithConn(f func(*pgx.Conn) error) error {
	conn, err := p.Acquire()
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

	stats := p.p.Stat()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_total"),
			"The cumulative count of successful connection acquires from the pool.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(stats.AcquireCount()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_duration_seconds_total"),
			"The total duration of all successful connection acquires from the pool.",
			nil, nil,
		),
		prometheus.CounterValue,
		stats.AcquireDuration().Seconds(),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquired"),
			"The number of currently acquired connections in the pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(stats.AcquiredConns()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_canceled_total"),
			"The cumulative count of connection acquires from the pool that were canceled.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(stats.CanceledAcquireCount()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "constructing"),
			"The number of connections with construction in progress in the pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(stats.ConstructingConns()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_empty_total"),
			"The cumulative count of successful connection acquires from the pool "+
				"that waited for a resource to be released or constructed because the pool was empty.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(stats.EmptyAcquireCount()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "idle"),
			"The number of currently idle connections in the pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(stats.IdleConns()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "max_size"),
			"The maximum size of the connection pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(stats.MaxConns()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "size"),
			"Total number of connections currently in the pool. "+
				"Should be a sum of constructing, acquired, and idle.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(stats.TotalConns()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "opened_total"),
			"The cumulative count of new connections opened.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(stats.NewConnsCount()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "destroyed_maxlifetime_total"),
			"The cumulative count of connections destroyed because they exceeded pool_max_conn_lifetime.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(stats.MaxLifetimeDestroyCount()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "destroyed_maxidle_total"),
			"The cumulative count of connections destroyed because they exceeded pool_max_conn_idle_time.",
			nil, nil,
		),
		prometheus.CounterValue,
		float64(stats.MaxIdleDestroyCount()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "acquires_empty_duration_seconds_total"),
			"The cumulative time waited for successful acquires from the pool "+
				"for a resource to be released or constructed because the pool was empty.",
			nil, nil,
		),
		prometheus.CounterValue,
		stats.EmptyAcquireWaitTime().Seconds(),
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*Pool)(nil)
)
