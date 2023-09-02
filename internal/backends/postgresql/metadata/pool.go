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

// Package pool provides access to PostgreSQL database and schemas.
//
// PostgreSQL schemas are mapped to FerretDB databases.
//
// It should be used only by the metadata package.
package metadata

import (
	"context"
	"net/url"
	"strings"
	"sync"

	zapadapter "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "postgresql_pool"
)

// Pool provides access to PostgreSQL database and schemas.
//
//nolint:vet // for readability
type Pool struct {
	p *pgxpool.Pool
	l *zap.Logger

	rw      sync.RWMutex
	schemas []string

	token *resource.Token
}

func New(u string, l *zap.Logger, sp *state.Provider) (*Pool, error) {
	uri, err := url.Parse(u)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	values := uri.Query()
	setDefaultValues(values)
	uri.RawQuery = values.Encode()

	config, err := pgxpool.ParseConfig(uri.String())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		var v string
		if err := conn.QueryRow(ctx, `SHOW server_version`).Scan(&v); err != nil {
			return lazyerrors.Error(err)
		}

		if sp.Get().HandlerVersion != v {
			if err := sp.Update(func(s *state.State) { s.HandlerVersion = v }); err != nil {
				l.Error("failed to update state", zap.Error(err))
			}
		}

		return nil
	}

	tracers := []pgx.QueryTracer{
		// try to log everything; logger's configuration will skip extra levels if needed
		&tracelog.TraceLog{
			Logger:   zapadapter.NewLogger(l),
			LogLevel: tracelog.LogLevelTrace,
		},
	}

	if debugbuild.Enabled {
		tracers = append(tracers, new(debugTracer))
	}

	config.ConnConfig.Tracer = &multiQueryTracer{
		Tracers: tracers,
	}

	ctx := context.TODO()

	p, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = checkSettings(ctx, p, l); err != nil {
		p.Close()
		return nil, lazyerrors.Error(err)
	}

	rows, err := p.Query(ctx, "SELECT schema_name FROM information_schema.schemata")
	if err != nil {
		p.Close()
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	schemas := make([]string, 0, 2)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			p.Close()
			return nil, lazyerrors.Error(err)
		}

		if strings.HasPrefix(name, "pg_") || name == "information_schema" {
			continue
		}

		schemas = append(schemas, name)
	}
	if err = rows.Err(); err != nil {
		p.Close()
		return nil, lazyerrors.Error(err)
	}

	slices.Sort(schemas)

	res := &Pool{
		p:       p,
		l:       l,
		schemas: schemas,
		token:   resource.NewToken(),
	}

	resource.Track(res, res.token)

	return res, nil
}

// Close frees all resources.
func (p *Pool) Close() {
	p.p.Close()
	p.p = nil
	resource.Untrack(p, p.token)
}

// List returns a sorted list of FerretDB database names.
func (p *Pool) List(ctx context.Context) []string {
	defer observability.FuncCall(ctx)()

	p.rw.RLock()
	defer p.rw.RUnlock()

	return slices.Clone(p.schemas)
}

// GetExisting returns an existing FerretDB database by valid name, or nil.
func (p *Pool) GetExisting(ctx context.Context, name string) bool {
	defer observability.FuncCall(ctx)()

	p.rw.RLock()
	defer p.rw.RUnlock()

	_, found := slices.BinarySearch(p.schemas, name)

	return found
}

// GetOrCreate returns an existing FerretDB database by valid name, or creates a new one.
//
// Returned boolean value indicates whether the FerretDB database was created.
func (p *Pool) GetOrCreate(ctx context.Context, name string) (bool, error) {
	defer observability.FuncCall(ctx)()

	if p.GetExisting(ctx, name) {
		return false, nil
	}

	p.rw.Lock()
	defer p.rw.Unlock()

	// it might have been created by a concurrent call
	i, found := slices.BinarySearch(p.schemas, name)
	if found {
		return false, nil
	}

	if _, err := p.p.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{name}.Sanitize()); err != nil {
		return false, lazyerrors.Error(err)
	}

	p.schemas = slices.Insert(p.schemas, i, name)

	return true, nil
}

// Drop removes a FerretDB database by valid name.
//
// It does nothing if the FerretDB database does not exist.
//
// Returned boolean value indicates whether the FerretDB database was removed.
func (p *Pool) Drop(ctx context.Context, name string) (bool, error) {
	defer observability.FuncCall(ctx)()

	if !p.GetExisting(ctx, name) {
		return false, nil
	}

	p.rw.Lock()
	defer p.rw.Unlock()

	// it might have been dropped by a concurrent call
	i, found := slices.BinarySearch(p.schemas, name)
	if !found {
		return false, nil
	}

	if _, err := p.p.Exec(ctx, "DROP SCHEMA "+pgx.Identifier{name}.Sanitize()); err != nil {
		return false, lazyerrors.Error(err)
	}

	p.schemas = slices.Delete(p.schemas, i, i+1)

	return true, nil
}

// Describe implements prometheus.Collector.
func (p *Pool) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(p, ch)
}

// Collect implements prometheus.Collector.
func (p *Pool) Collect(ch chan<- prometheus.Metric) {
	p.rw.RLock()
	defer p.rw.RUnlock()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "databases"),
			"The current number of FerretDB databases in the pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(len(p.schemas)),
	)
}

// check interfaces
var (
	_ prometheus.Collector = (*Pool)(nil)
)
