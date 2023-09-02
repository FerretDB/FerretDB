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

// Package pool provides access to PostgreSQL databases and their connections.
//
// It should be used only by the metadata package.
package pool

import (
	"context"
	"net/url"
	"strings"

	zapadapter "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
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

var (
	// The only supported encoding in canonical form.
	supportedEncoding = "UTF8"

	// Supported locales in canonical forms.
	supportedLocales = []string{"POSIX", "C", "C.UTF8", "en_US.UTF8"}
)

// Pool provides access to PostgreSQL database.
//
//nolint:vet // for readability
type Pool struct {
	p     *pgxpool.Pool
	l     *zap.Logger
	token *resource.Token
}

func New(u string, l *zap.Logger, p *state.Provider) (*Pool, error) {
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

		if p.Get().HandlerVersion != v {
			if err := p.Update(func(s *state.State) { s.HandlerVersion = v }); err != nil {
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

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := &Pool{
		p:     pool,
		l:     l,
		token: resource.NewToken(),
	}

	resource.Track(res, res.token)

	if err = res.checkConnection(ctx); err != nil {
		res.Close()
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// simplifySetting simplifies PostgreSQL setting value for comparison.
func simplifySetting(v string) string {
	return strings.ToLower(strings.ReplaceAll(v, "-", ""))
}

// isSupportedEncoding checks `server_encoding` and `client_encoding` values.
func isSupportedEncoding(v string) bool {
	return simplifySetting(v) == simplifySetting(supportedEncoding)
}

// isSupportedLocale checks `lc_collate` and `lc_ctype` values.
func isSupportedLocale(v string) bool {
	v = simplifySetting(v)

	for _, s := range supportedLocales {
		if v == simplifySetting(s) {
			return true
		}
	}

	return false
}

// checkConnection checks PostgreSQL settings.
func (p *Pool) checkConnection(ctx context.Context) error {
	rows, err := p.p.Query(ctx, "SHOW ALL")
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		// handle variable number of columns as a workaround for https://github.com/cockroachdb/cockroach/issues/101715
		values, err := rows.Values()
		if err != nil {
			return lazyerrors.Error(err)
		}

		if len(values) < 2 {
			return lazyerrors.Errorf("invalid row: %#v", values)
		}
		name := values[0].(string)
		value := values[1].(string)

		switch name {
		case "server_encoding", "client_encoding":
			if !isSupportedEncoding(value) {
				return lazyerrors.Errorf("%q is %q; supported value is %q", name, value, supportedEncoding)
			}
		case "lc_collate", "lc_ctype":
			if !isSupportedLocale(value) {
				return lazyerrors.Errorf("%q is %q; supported values are %v", name, value, supportedLocales)
			}
		case "standard_conforming_strings": // To sanitize safely: https://github.com/jackc/pgx/issues/868#issuecomment-725544647
			if value != "on" {
				return lazyerrors.Errorf("%q is %q, want %q", name, value, "on")
			}
		default:
			continue
		}

		p.l.Debug("PostgreSQL setting", zap.String(name, value))
	}

	if err := rows.Err(); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// Close closes all databases in the pool and frees all resources.
func (p *Pool) Close() {
	p.p.Close()
	p.p = nil
	resource.Untrack(p, p.token)
}

// List returns a sorted list of database names in the pool.
func (p *Pool) List(ctx context.Context) []string {
	defer observability.FuncCall(ctx)()

	panic("not implemented")
}

// GetExisting returns an existing database by valid name, or nil.
func (p *Pool) GetExisting(ctx context.Context, name string) *fsql.DB {
	defer observability.FuncCall(ctx)()

	panic("not implemented")
}

// GetOrCreate returns an existing database by valid name, or creates a new one.
//
// Returned boolean value indicates whether the database was created.
func (p *Pool) GetOrCreate(ctx context.Context, name string) (*fsql.DB, bool, error) {
	defer observability.FuncCall(ctx)()

	panic("not implemented")
}

// Drop closes and removes a database by valid name.
//
// It does nothing if the database does not exist.
//
// Returned boolean value indicates whether the database was removed.
func (p *Pool) Drop(ctx context.Context, name string) bool {
	defer observability.FuncCall(ctx)()

	panic("not implemented")
}

// Describe implements prometheus.Collector.
func (p *Pool) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(p, ch)
}

// Collect implements prometheus.Collector.
func (p *Pool) Collect(ch chan<- prometheus.Metric) {
}

// check interfaces
var (
	_ prometheus.Collector = (*Pool)(nil)
)
