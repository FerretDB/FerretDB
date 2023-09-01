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

// Package pool provides access to SQLite databases and their connections.
//
// It should be used only by the metadata package.
package pool

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	_ "modernc.org/sqlite" // register database/sql driver

	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// filenameExtension represents SQLite database filename extension.
const filenameExtension = ".sqlite"

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "sqlite_pool"
)

// Pool provides access to SQLite databases and their connections.
//
//nolint:vet // for readability
type Pool struct {
	uri *url.URL
	l   *zap.Logger

	rw  sync.RWMutex
	dbs map[string]*fsql.DB

	token *resource.Token

	sp *state.Provider
}

// openDB opens existing database or creates a new one.
//
// All valid FerretDB database names are valid SQLite database names / file names,
// so no validation is needed.
// One exception is very long full path names for the filesystem,
// but we don't check it.
func openDB(name, uri string, singleConn bool, l *zap.Logger, sp *state.Provider) (*fsql.DB, error) {
	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)

	// TODO https://github.com/FerretDB/FerretDB/issues/2755
	if singleConn {
		db.SetMaxIdleConns(1)
		db.SetMaxOpenConns(1)
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, lazyerrors.Error(err)
	}

	if err := sp.Update(func(s *state.State) {
		if s.HandlerVersion != "" {
			return
		}

		row := db.QueryRowContext(context.Background(), "SELECT sqlite_version()")
		if err := row.Scan(&s.HandlerVersion); err != nil {
			l.Error("sqlite.metadata.pool.openDB: failed to query SQLite version", zap.Error(err))
		}
	}); err != nil {
		l.Error("sqlite.metadata.pool.openDB: failed to update state", zap.Error(err))
	}

	return fsql.WrapDB(db, name, l), nil
}

// New creates a pool for SQLite databases in the directory specified by SQLite URI.
//
// All databases are opened on creation.
//
// The returned map is the initial set of existing databases.
// It should not be modified.
func New(u string, l *zap.Logger, sp *state.Provider) (*Pool, map[string]*fsql.DB, error) {
	uri, err := parseURI(u)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse SQLite URI %q: %s", u, err)
	}

	matches, err := filepath.Glob(filepath.Join(uri.Path, "*"+filenameExtension))
	if err != nil {
		return nil, nil, lazyerrors.Error(err)
	}

	p := &Pool{
		uri:   uri,
		l:     l,
		dbs:   make(map[string]*fsql.DB, len(matches)),
		token: resource.NewToken(),
		sp:    sp,
	}

	resource.Track(p, p.token)

	for _, f := range matches {
		name := p.databaseName(f)
		uri := p.databaseURI(name)

		p.l.Debug("Opening existing database.", zap.String("name", name), zap.String("uri", uri))

		db, err := openDB(name, uri, p.singleConn(), l, p.sp)
		if err != nil {
			p.Close()
			return nil, nil, lazyerrors.Error(err)
		}

		p.dbs[name] = db
	}

	return p, p.dbs, nil
}

// memory returns true if the pool is for the in-memory database.
func (p *Pool) memory() bool {
	return p.uri.Query().Get("mode") == "memory"
}

// singleConn returns true if pool size must be limited to a single connection.
func (p *Pool) singleConn() bool {
	// https://www.sqlite.org/inmemorydb.html
	// TODO https://github.com/FerretDB/FerretDB/issues/2755
	return p.memory()
}

// databaseName returns database name for given database file path.
func (p *Pool) databaseName(databaseFile string) string {
	return strings.TrimSuffix(filepath.Base(databaseFile), filenameExtension)
}

// databaseURI returns SQLite URI for the given database name.
func (p *Pool) databaseURI(databaseName string) string {
	dbURI := *p.uri
	dbURI.Path = path.Join(dbURI.Path, databaseName+filenameExtension)
	dbURI.Opaque = dbURI.Path

	return dbURI.String()
}

// databaseFile returns database file path for the given database name,
// or empty string for in-memory database.
func (p *Pool) databaseFile(databaseName string) string {
	if p.memory() {
		return ""
	}

	return filepath.Join(p.uri.Path, databaseName+filenameExtension)
}

// Close closes all databases in the pool and frees all resources.
func (p *Pool) Close() {
	p.rw.Lock()
	defer p.rw.Unlock()

	for _, db := range p.dbs {
		_ = db.Close()
	}

	p.dbs = nil

	resource.Untrack(p, p.token)
}

// List returns a sorted list of database names in the pool.
func (p *Pool) List(ctx context.Context) []string {
	defer observability.FuncCall(ctx)()

	p.rw.RLock()
	defer p.rw.RUnlock()

	res := maps.Keys(p.dbs)
	slices.Sort(res)

	return res
}

// GetExisting returns an existing database by valid name, or nil.
func (p *Pool) GetExisting(ctx context.Context, name string) *fsql.DB {
	defer observability.FuncCall(ctx)()

	p.rw.RLock()
	defer p.rw.RUnlock()

	db := p.dbs[name]
	if db == nil {
		return nil
	}

	return db
}

// GetFirst returns a database from the pool, or nil if there are no databases in the pool.
//
// As the pool is not ordered, the returned database is not guaranteed to be the same.
// This method can be used when a DB-agnostic query is needed.
func (p *Pool) GetFirst(ctx context.Context) *fsql.DB {
	defer observability.FuncCall(ctx)()

	p.rw.RLock()
	defer p.rw.RUnlock()

	for _, db := range p.dbs {
		return db
	}

	return nil
}

// GetOrCreate returns an existing database by valid name, or creates a new one.
//
// Returned boolean value indicates whether the database was created.
func (p *Pool) GetOrCreate(ctx context.Context, name string) (*fsql.DB, bool, error) {
	defer observability.FuncCall(ctx)()

	db := p.GetExisting(ctx, name)
	if db != nil {
		return db, false, nil
	}

	p.rw.Lock()
	defer p.rw.Unlock()

	// it might have been created by a concurrent call
	if db := p.dbs[name]; db != nil {
		return db, false, nil
	}

	uri := p.databaseURI(name)
	db, err := openDB(name, uri, p.singleConn(), p.l, p.sp)
	if err != nil {
		return nil, false, lazyerrors.Errorf("%s: %w", uri, err)
	}

	p.l.Debug("Database created.", zap.String("name", name), zap.String("uri", uri))

	p.dbs[name] = db

	return db, true, nil
}

// Drop closes and removes a database by valid name.
//
// It does nothing if the database does not exist.
//
// Returned boolean value indicates whether the database was removed.
func (p *Pool) Drop(ctx context.Context, name string) bool {
	defer observability.FuncCall(ctx)()

	p.rw.Lock()
	defer p.rw.Unlock()

	db, ok := p.dbs[name]
	if !ok {
		return false
	}

	if err := db.Close(); err != nil {
		p.l.Warn("Failed to close database connection.", zap.String("name", name), zap.Error(err))
	}

	delete(p.dbs, name)

	if f := p.databaseFile(name); f != "" {
		if err := os.Remove(f); err != nil {
			p.l.Warn("Failed to remove database file.", zap.String("file", f), zap.String("name", name), zap.Error(err))
		}
	}

	p.l.Debug("Database dropped.", zap.String("name", name))

	return true
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
			"The current number of database in the pool.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(len(p.dbs)),
	)

	for _, db := range p.dbs {
		db.Collect(ch)
	}
}

// check interfaces
var (
	_ prometheus.Collector = (*Pool)(nil)
)
